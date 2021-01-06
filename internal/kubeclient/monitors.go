package kubeclient

import (
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"time"
)

func MonitorPods() error {
	clientSet, err := GetKubeClient()
	if err != nil {
		return err
	}
	watchList := cache.NewListWatchFromClient(
		clientSet.CoreV1().RESTClient(),
		"pods",
		"zombie",
		fields.Everything(),
	)
	_, controller := cache.NewInformer(
		watchList,
		&corev1.Pod{},
		time.Second*30,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod := obj.(*corev1.Pod)
				log.Debugf("deployment added: %s %s", pod.Name, pod.Status.Phase)
			},
			DeleteFunc: func(obj interface{}) {
				pod := obj.(*corev1.Pod)
				log.Debugf("deployment deleted: %s %s", pod.Name, pod.Status.Phase)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				pod := oldObj.(*corev1.Pod)
				log.Debugf("deployment updated: %s %s", pod.Name, pod.Status.Phase)
			},
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)
	return nil
}
