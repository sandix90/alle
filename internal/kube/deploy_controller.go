package kube

import (
	"alle/internal/models"
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"sync"
	"time"
)

var ErrPackageNotFound = errors.New("package not found")

type DeployController interface {
	ApplyPackages(packs []*models.Package) error
}

type DeployNode struct {
	name     string
	deployed bool
	parents  []*DeployNode
	pack     *models.Package
}

type deployControllerImpl struct {
	kubeClient    IKubeClient
	eventListener EventListener
	mut           sync.RWMutex
}

func NewDeployControllerFromEnv(environment string) (DeployController, error) {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return nil, err
	}

	config.QPS = 100
	dynclient, err := dynamic.NewForConfig(config)
	kubeClient, err := NewKubeClient(dynclient, environment, config)
	if err != nil {
		return nil, err
	}

	eventListener := NewEventListener(dynclient)
	return NewDeployController(kubeClient, eventListener), nil
}

func NewDeployController(kubeClient IKubeClient, eventListener EventListener) DeployController {
	return &deployControllerImpl{
		kubeClient:    kubeClient,
		eventListener: eventListener,
	}
}

func (d *deployControllerImpl) ApplyPackages(packs []*models.Package) error {

	q, err := d.getNodesQueue(packs)
	if err != nil {
		return err
	}

	errCh := make(chan error)
	wgDone := make(chan struct{})
	wg := new(sync.WaitGroup)
	wg.Add(len(q))

	ctx := context.Background()
	go d.applyDeployNodes(ctx, nil, wg, q, errCh)

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	select {
	case err := <-errCh:
		return err
	case <-wgDone:
		break
	}
	close(errCh)

	return nil
	//go StartListening(dynclient, stopCh, kube.EventListenerHandler)
}

func (d *deployControllerImpl) applyDeployNodes(ctx context.Context, caller *DeployNode, wg *sync.WaitGroup,
	nodes []*DeployNode, errCh chan error) {

	for _, n := range nodes {

		if n.deployed {
			wg.Done()
			return
		}

		if len(n.parents) > 0 {
			nestedWg := new(sync.WaitGroup)
			nestedWg.Add(len(n.parents))
			nestedErrCh := make(chan error)

			go d.applyDeployNodes(ctx, n, nestedWg, n.parents, nestedErrCh)

			c := make(chan struct{})
			go func() {
				defer close(c)
				nestedWg.Wait()
			}()

			select {
			case <-c:
				log.Debugf(`all parent packs for "%s" package has been deployed`, n.name)
			case <-ctx.Done():
				return
			case err := <-nestedErrCh:
				errCh <- err
				return
			}
		}

		var timeout time.Duration
		if caller != nil {
			timeout = time.Duration(caller.pack.Wait.Timeout) * time.Second
		} else {
			timeout = 0
		}
		err := d.deployNode(ctx, n, timeout)
		if err != nil {
			errCh <- err
			return
		}

		//if n.pack.Wait != nil {
		//	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second * time.Duration(n.pack.Wait.Timeout))
		//	d.eventListener.WaitForEventType(ctx, n.pack.Wait.For, AddEventType)
		//
		//	n.deployed = true
		//	cancelFn()
		//}
		//for _, m := range n.pack.Manifests {
		//	err := d.kubeClient.ApplyManifest(ctx, m)
		//	if err != nil {
		//		errCh <- err
		//	}
		//}
		//time.Sleep(10 * time.Second)

		wg.Done()
	}
}

func (d *deployControllerImpl) deployNode(ctx context.Context, n *DeployNode, timeout time.Duration) error {
	wg := new(sync.WaitGroup)
	wg.Add(len(n.pack.Manifests))

	errCh := make(chan error)

	// if timeout is not set, lets set it to 9 hours. Guess should be enough :-)
	if timeout == 0 {
		timeout = time.Hour * 9
	}

	go func() {
		for _, m := range n.pack.Manifests {
			err := d.kubeClient.ApplyManifest(ctx, m)
			if err != nil {
				errCh <- err
			}
			n.deployed = true
			wg.Done()
		}
	}()

	c := make(chan struct{})

	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		log.Debugf("package %s deployed", n.name)
	case err := <-errCh:
		return fmt.Errorf("error deploy node. OError: %v", err)
	case <-time.After(timeout):
		return fmt.Errorf("deploy node timeout reached")
	case <-ctx.Done():
		return fmt.Errorf("context close reached")
	}
	return nil
}

func (d *deployControllerImpl) getNodesQueue(packs []*models.Package) ([]*DeployNode, error) {
	// TODO: Check unique name of manifests
	var knownNodes []*DeployNode
	for _, pkg := range packs {
		k := pkg.Name
		var v string
		if pkg.Wait != nil {
			v = pkg.Wait.For
		}

		if k == v {
			return nil, fmt.Errorf(`"wait for" cant depend on itself. Package: "%v"`, k)
		}

		if k == "" {
			return nil, fmt.Errorf("deployNode name cant be empty")
		}
		kNode, err := d.findNode(knownNodes, k)
		if err != nil {
			kNode = &DeployNode{name: k, pack: pkg}
			knownNodes = append(knownNodes, kNode)
		}

		if v == "" {
			log.Debugf("deployNode \"%s\" has zero parent detected. Ignoring...\n", k)
			continue
		}

		vNode, err := d.findNode(knownNodes, v)
		if err != nil {
			p, err := findPackage(packs, v)
			if err != nil {
				return nil, fmt.Errorf("package \"%s\" is not found. Cant add dependency wait package \"%s\" to \"%s\"."+
					"\nPlease check the right labels are selected", v, v, k)
			}

			vNode = &DeployNode{name: v, pack: p}
			knownNodes = append(knownNodes, vNode)
		}

		// Check if parent node is already exist, so we dont need to add it into the knowsNodes list
		if existNode, _ := d.findNode(kNode.parents, v); existNode != nil {
			fmt.Println("deployNode is already exist. Ignoring add to knownNodes list")
			continue
		}

		if err := d.checkCircleDependencies(vNode, kNode.name); err != nil {
			return nil, fmt.Errorf("circle parentness detected when deployNode \"%v\" wants to add parent node with value "+
				"\"%v\".\nParent deps: %v -> %v -> X %v", kNode.name, vNode.name, vNode.name, err, vNode.name)
		}
		kNode.parents = append(kNode.parents, vNode)

	}

	var q []*DeployNode
	for _, kn := range knownNodes {
		if !d.isSeenInParentsWithExcludeSelf(knownNodes, kn) {
			q = append(q, kn)
		}
	}
	return q, nil
}

func (d *deployControllerImpl) isSeenInParentsWithExcludeSelf(nodes []*DeployNode, seekForNode *DeployNode) bool {
	var cleanedNodes []*DeployNode
	for _, n := range nodes {
		if n == seekForNode {
			continue
		}
		cleanedNodes = append(cleanedNodes, n)
	}

	return d.isSeenInParents(cleanedNodes, seekForNode)
}

func (d *deployControllerImpl) isSeenInParents(nodes []*DeployNode, n *DeployNode) bool {
	for _, kn := range nodes {

		if kn.name == n.name {
			return true
		}

		if kn.parents != nil {
			if d.isSeenInParents(kn.parents, n) {
				return true
			}
		}

	}
	return false
}

func (d *deployControllerImpl) findNode(nodes []*DeployNode, v string) (*DeployNode, error) {
	for _, n := range nodes {
		if v == n.name {
			return n, nil
		}
	}
	return nil, ErrPackageNotFound
}

func (d *deployControllerImpl) checkCircleDependencies(n *DeployNode, v string) error {
	for _, pn := range n.parents {

		if pn.parents != nil {
			err := d.checkCircleDependencies(pn, v)
			if err != nil {
				return fmt.Errorf("%v -> %v", pn.name, err)
			}
			return nil
		}
		if pn.name == v {
			return fmt.Errorf("%v", v)
		}
	}
	return nil
}

func findPackage(packs []*models.Package, v string) (*models.Package, error) {
	for _, p := range packs {
		if v == p.Name {
			return p, nil
		}
	}
	return nil, ErrPackageNotFound
}
