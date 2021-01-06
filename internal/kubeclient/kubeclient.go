package kubeclient

import (
	"alle/internal/models"
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

const ALLEVERSION = "0.0.1"

type IKubeClient interface {
	ApplyManifest(manifest models.IManifestor) error
	IsManifestDeployed(manifest models.IManifestor) (bool, error)
	GetManifestsList() ([]models.IManifestor, error)
	DeleteManifest(manifest models.IManifestor) error
	DeleteManifestsList(manifests []models.IManifestor) error
}

type kubernetesClientImpl struct {
	client          dynamic.Interface
	config          *rest.Config
	namespace       string
	discoveryClient *discovery.DiscoveryClient
}

func NewKubeClient(client dynamic.Interface, namespace string, config *rest.Config) (IKubeClient, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	return &kubernetesClientImpl{client: client, namespace: namespace, config: config, discoveryClient: dc}, nil
}

func (k *kubernetesClientImpl) ApplyManifest(manifest models.IManifestor) error {
	obj := &unstructured.Unstructured{}

	manifestStr, err := manifest.String()
	if err != nil {
		return fmt.Errorf("cant convert manifest to string. Manifest name: %s", manifest.GetName())
	}

	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode([]byte(manifestStr), nil, obj)

	gvr, err := k.findGVR(gvk)
	if err != nil {
		return err
	}

	dr, err := k.getDynamicResource(gvr)
	if err != nil {
		return err
	}

	obj.SetLabels(map[string]string{"alle_version": ALLEVERSION})
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	err = k.createAlleCrdManifest(manifest)
	if err != nil {
		return err
	}

	_, err = dr.Patch(obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{FieldManager: "sample-controller"})
	if err != nil {
		return err
	}
	log.Debugf(`Manifest "%s" applied`, manifest.GetFullPath())
	return nil
}

func (k *kubernetesClientImpl) GetManifestsList() ([]models.IManifestor, error) {
	panic("implement me")
}

func (k *kubernetesClientImpl) DeleteManifest(manifest models.IManifestor) error {
	obj := &unstructured.Unstructured{}

	manifestStr, err := manifest.String()
	if err != nil {
		return fmt.Errorf("cant convert manifest to string. Manifest name: %s", manifest.GetName())
	}

	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode([]byte(manifestStr), nil, obj)

	gvr, err := k.findGVR(gvk)
	if err != nil {
		return err
	}

	dr, err := k.getDynamicResource(gvr)
	if err != nil {
		return err
	}

	err = dr.Delete(obj.GetName(), &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	// Delete CustomResourceDefinition AlleManifest
	err = k.deleteAlleCrdManifest(manifest)
	if err != nil {
		return err
	}
	return nil
}

func (k *kubernetesClientImpl) DeleteManifestsList(manifests []models.IManifestor) error {
	panic("implement me")
}

func (k *kubernetesClientImpl) IsManifestDeployed(manifest models.IManifestor) (bool, error) {
	log.Printf("Manifest is %s", manifest.GetName())
	return false, nil
}

// find the corresponding GVR (available in *meta.RESTMapping) for gvk
func (k *kubernetesClientImpl) findGVR(gvk *schema.GroupVersionKind) (*meta.RESTMapping, error) {

	// DiscoveryClient queries API server about the resources
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(k.discoveryClient))

	return mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
}

func (k *kubernetesClientImpl) getDynamicResource(gvr *meta.RESTMapping) (dynamic.ResourceInterface, error) {
	var dr dynamic.ResourceInterface
	if gvr.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = k.client.Resource(gvr.Resource).Namespace(k.namespace)
	} else {
		dr = k.client.Resource(gvr.Resource)
	}
	return dr, nil
}

// Used to create meta info about deployed manifests
func (k *kubernetesClientImpl) createAlleCrdManifest(manifest models.IManifestor) error {
	amGVK := &schema.GroupVersionKind{
		Group:   "alle.org",
		Version: "v1",
		Kind:    "AlleManifest",
	}
	amGvr, err := k.findGVR(amGVK)
	if err != nil {
		return err
	}
	amDynRes, err := k.getDynamicResource(amGvr)
	if err != nil {
		return err
	}

	am := map[string]interface{}{
		"apiVersion": "alle.org/v1",
		"kind":       "AlleManifest",
		"metadata": map[string]interface{}{
			"name": manifest.GetFullPath(),
			"labels": map[string]interface{}{
				"deployed": "alle",
			},
		},
		"spec": map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"alle_version": "0.0.2",
			},
		},
	}
	amData, err := json.Marshal(am)
	if err != nil {
		return err
	}
	_, err = amDynRes.Patch(manifest.GetFullPath(), types.ApplyPatchType, amData, metav1.PatchOptions{FieldManager: "sample-controller"})
	if err != nil {
		return err
	}
	log.Debugf(`CRD Alle Manifest "%s" applied`, manifest.GetFullPath())

	return nil
}

// Used to deleted meta info about deployed manifests
func (k *kubernetesClientImpl) deleteAlleCrdManifest(manifest models.IManifestor) error {
	amGVK := &schema.GroupVersionKind{
		Group:   "alle.org",
		Version: "v1",
		Kind:    "AlleManifest",
	}

	amGvr, err := k.findGVR(amGVK)
	if err != nil {
		return err
	}
	amDynRes, err := k.getDynamicResource(amGvr)
	if err != nil {
		return err
	}

	err = amDynRes.Delete(manifest.GetFullPath(), &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
