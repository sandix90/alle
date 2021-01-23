package kube

import (
	"context"
	"errors"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"time"
)

const (
	AlleCrdManifestName = "allemanifests.alle.org"
)

var (
	ErrAlleInitialized = errors.New("alle is already initialised")
)

type KubeInitializer interface {
	Init() error
}

type kubeInitializer struct {
	client    dynamic.Interface
	namespace string
	config    *rest.Config
}

func NewKubeInitializer(client dynamic.Interface, namespace string, config *rest.Config) KubeInitializer {
	return &kubeInitializer{client: client, namespace: namespace, config: config}
}

func (ki *kubeInitializer) Init() error {

	ctx, cancelFunc := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelFunc()

	apixClient, err := apiextv1beta1.NewForConfig(ki.config)
	if err != nil {
		return err
	}

	crdClient := apixClient.ApiextensionsV1beta1().CustomResourceDefinitions()
	_, err = crdClient.Get(ctx, AlleCrdManifestName, metav1.GetOptions{})
	if err == nil {
		return ErrAlleInitialized
	}

	crdManifest := &apiextensionv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:   AlleCrdManifestName,
			Labels: map[string]string{"name": "allemanifest"},
		},
		Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
			Group: "alle.org",
			Versions: []apiextensionv1beta1.CustomResourceDefinitionVersion{{
				Name:    "v1",
				Served:  true,
				Storage: true,
			}},
			Names: apiextensionv1beta1.CustomResourceDefinitionNames{
				Plural:     "allemanifests",
				Singular:   "allemanifest",
				Kind:       "AlleManifest",
				ShortNames: []string{"am"},
			},
			Scope: apiextensionv1beta1.NamespaceScoped,
			Validation: &apiextensionv1beta1.CustomResourceValidation{
				OpenAPIV3Schema: &apiextensionv1beta1.JSONSchemaProps{
					Properties: map[string]apiextensionv1beta1.JSONSchemaProps{
						"spec": {
							Required: []string{"kind", "apiVersion"},
							Properties: map[string]apiextensionv1beta1.JSONSchemaProps{
								"kind":       {Type: "string"},
								"apiVersion": {Type: "string"},
								"metadata": {
									Properties: map[string]apiextensionv1beta1.JSONSchemaProps{
										"alle_version":  {Type: "string"},
										"manifest_name": {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = crdClient.Create(ctx, crdManifest, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
