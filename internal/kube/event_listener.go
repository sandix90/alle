package kube

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type KubeEventListenerHandler func(obj *unstructured.Unstructured) error

func StartListening(client dynamic.Interface, stopCh <-chan struct{}, waitChannel chan *unstructured.Unstructured,
	gvr *schema.GroupVersionResource, namespace string) {

	f := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, namespace, nil)

	informer := f.ForResource(*gvr)
	s := informer.Informer()

	funcs := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			unStr, ok := obj.(*unstructured.Unstructured)
			if ok {
				waitChannel <- unStr
				//rep, err := dynRes.Get(unStr.GetName(), metav1.GetOptions{})
				//log.Debugln(rep)
				//log.Debugln(err)
				//log.Debugf("created: %s %s",unStr.GetName(), unStr.GetUID())
				//log.Debugln(handler(unStr))
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			unStr, ok := newObj.(*unstructured.Unstructured)
			if ok {
				log.Debugf("updated: %s %s", unStr.GetName(), unStr.GetUID())
				//log.Debugln(handler(unStr))
			}
		},
		DeleteFunc: func(obj interface{}) {
			unStr, ok := obj.(*unstructured.Unstructured)
			if ok {
				log.Debugf("deleted: %s %s", unStr.GetName(), unStr.GetUID())
				//log.Debugln(handler(unStr))
			}
		},
	}

	s.AddEventHandler(funcs)
	s.Run(stopCh)
}

func EventListenerHandler(obj *unstructured.Unstructured) error {
	log.Debugln(obj.GetName(), obj.GetUID())
	return nil
}

func WaitFor(client dynamic.Interface, stopCh <-chan struct{}, waitChannel chan *unstructured.Unstructured,
	gvr *schema.GroupVersionResource, namespace string) error {

	go StartListening(client, stopCh, waitChannel, gvr, namespace)

	select {
	case unStr := <-waitChannel:
		log.Debugln(unStr)
		dynRes := client.Resource(*gvr).Namespace(namespace)
		log.Debugln(dynRes)
	case <-stopCh:
		return fmt.Errorf("listening stopped")
	}

	return nil
}

//func MonitorPods() error {
//	clientSet, err := GetKubeClient()
//	if err != nil {
//		return err
//	}
//
//	watchList := cache.NewListWatchFromClient(
//		clientSet.CoreV1().RESTClient(),
//		"pods",
//		"zombie",
//		fields.Everything(),
//	)
//	_, controller := cache.NewInformer(
//		watchList,
//		&corev1.Pod{},
//		time.Second*30,
//		cache.ResourceEventHandlerFuncs{
//			AddFunc: func(obj interface{}) {
//				pod := obj.(*corev1.Pod)
//				log.Debugf("deployment added: %s %s", pod.Name, pod.Status.Phase)
//			},
//			DeleteFunc: func(obj interface{}) {
//				pod := obj.(*corev1.Pod)
//				log.Debugf("deployment deleted: %s %s", pod.Name, pod.Status.Phase)
//			},
//			UpdateFunc: func(oldObj, newObj interface{}) {
//				pod := oldObj.(*corev1.Pod)
//				log.Debugf("deployment updated: %s %s", pod.Name, pod.Status.Phase)
//			},
//		},
//	)
//	stop := make(chan struct{})
//	go controller.Run(stop)
//	return nil
//}
