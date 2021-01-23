package kube

import (
	"alle/internal/models"
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
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
	kubeClient IKubeClient
	mut        sync.RWMutex
}

func NewDeployController(kubeClient IKubeClient) DeployController {
	return &deployControllerImpl{
		kubeClient: kubeClient,
	}
}

func (d *deployControllerImpl) ApplyPackages(packs []*models.Package) error {

	q, err := d.getNodesQueue(packs)
	if err != nil {
		return err
	}
	//log.Debugln(q)
	//stopCh := make(chan struct{})

	errCh := make(chan error)
	wgDone := make(chan struct{})
	wg := new(sync.WaitGroup)
	wg.Add(len(q))
	ctx := context.Background()
	go d.applyDeployNodes(ctx, wg, q, errCh)

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
func (d *deployControllerImpl) applyDeployNodes(ctx context.Context, wg *sync.WaitGroup, nodes []*DeployNode, errCh chan error) {
	for _, n := range nodes {

		if n.deployed {
			wg.Done()
			return
		}

		if len(n.parents) > 0 {
			nestedWg := new(sync.WaitGroup)
			nestedWg.Add(len(n.parents))

			go d.applyDeployNodes(ctx, nestedWg, n.parents, errCh)
			nestedWg.Wait()

			for _, pn := range n.parents {
				if !pn.deployed {
					errCh <- fmt.Errorf("node %s is not deployed, but wait group is done", pn.name)
				}
			}
		}

		for _, m := range n.pack.Manifests {
			err := d.kubeClient.ApplyManifest(ctx, m)
			if err != nil {
				errCh <- err
			}
		}

		time.Sleep(10 * time.Second)
		n.deployed = true
		wg.Done()
	}
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
