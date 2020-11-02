package kubeclient

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

func getDeploymentClient(kubeclient *kubernetes.Clientset, namespace string) v1.DeploymentInterface {
	deploymentsClient := kubeclient.AppsV1().Deployments(namespace)
	return deploymentsClient
}

func DeleteDeployment(kubeclient *kubernetes.Clientset, namespace string, manifests []string) error {
	deploymentsClient := getDeploymentClient(kubeclient, namespace)

	for _, manifest := range manifests {
		obj, err := DeserializeManifest(manifest)
		//log.Debugf("Kind is %s", obj.GetObjectKind().GroupVersionKind())
		if err != nil {
			return err
		}

		delDeployment, ok := obj.(*appsv1.Deployment)
		if !ok {
			return errors.New("Error casting manifest to Deployment")
		}

		log.Infof("Deleting %s", delDeployment.Name)
		err = deploymentsClient.Delete(delDeployment.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateDeployment(kubeclient *kubernetes.Clientset, namespace string, manifests []string) error {
	//config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	//dynclient, err := dynamic.NewForConfig(config)
	//dynclient.Resource().
	deploymentClient := getDeploymentClient(kubeclient, namespace)

	for _, manifest := range manifests {

		obj, err := DeserializeManifest(manifest)
		if err != nil {
			return err
		}

		deployment, ok := obj.(*appsv1.Deployment)
		deployment.SetAnnotations(map[string]string{"alleVersion": "0.0.1"})
		if !ok {
			return errors.New("Error casting manifest to Deployment")
		}

		log.Infof("Creating %s", deployment.Name)
		_, err = deploymentClient.Create(deployment)
		if err != nil {
			return err
		}
	}
	//err := wait.WaitFor(waitFunc, conditionFunc, make(chan struct{}))
	//if err != nil {
	//
	//}
	return nil
}

func ListDeployments(kubeclient *kubernetes.Clientset, namespace string, wr io.Writer) ([]appsv1.Deployment, error) {
	deploymentClient := getDeploymentClient(kubeclient, namespace)
	watchObj, err := deploymentClient.Watch(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for event := range watchObj.ResultChan() {
		fmt.Printf("Type: %v\n", event.Type)
		p, ok := event.Object.(*appsv1.Deployment)
		if !ok {
			log.Fatal("unexpected type")
		}
		fmt.Println(p.Status)
		//fmt.Println(p.Status.ContainerStatuses)
		//fmt.Println(p.Status.Phase)
	}

	deployments, err := deploymentClient.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for i, dep := range deployments.Items {
		out := fmt.Sprintf("%d: Deployment: %s, Image: %s, ReadyReplicas: %d\n", i+1, dep.Name, dep.Spec.Template.Spec.Containers[0].Image, dep.Status.ReadyReplicas)

		_, err = wr.Write([]byte(out))
		if err != nil {
			return nil, err
		}
	}

	return deployments.Items, err

}

//func conditionFunc() (done bool, err error) {
//
//}
//
//func waitFunc(done <-chan struct{}) <-chan struct{} {
//
//}
