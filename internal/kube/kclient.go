package kube

//
//import (
//	log "github.com/sirupsen/logrus"
//	corev1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/client-go/kubernetes"
//	"k8s.io/client-go/tools/clientcmd"
//	//api "k8s.io/api/core/appsv1"
//	"os"
//	"os/exec"
//)
//
//func RunKubectl() error {
//	kubectlExecPath, err := exec.LookPath("kubectl")
//
//	if err != nil {
//		log.Errorf("kubectl is not found. Make sure it is installed.")
//	}
//
//	cmdKubectl := &exec.Cmd{
//		Path:   kubectlExecPath,
//		Args:   []string{kubectlExecPath, "version"},
//		Stdout: os.Stdout,
//		Stderr: os.Stdout,
//	}
//
//	if err := cmdKubectl.Run(); err != nil {
//		log.Errorf("Kubectl error: ", err)
//	}
//	return nil
//}
//
//func GetKubeClient() (*kubernetes.Clientset, error) {
//	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
//	if err != nil {
//		return nil, err
//	}
//	clientSet, err := kubernetes.NewForConfig(config)
//	if err != nil {
//		return nil, err
//	}
//	return clientSet, nil
//}
//type podPhaseResponse struct {
//	done  bool
//	phase corev1.PodPhase
//	err   error
//}
//func newPodPhaseResponse(done bool, podPhase corev1.PodPhase, err error) *podPhaseResponse {
//	return &podPhaseResponse{
//		done:  done,
//		phase: podPhase,
//		err:   err,
//	}
//}
//
//
//func isPodRunning(c *kubernetes.Clientset, pod *corev1.Pod) *podPhaseResponse {
//	c, err := GetKubeClient()
//	if err != nil {
//		return newPodPhaseResponse(false, corev1.PodUnknown, err)
//	}
//	actualPod, err := c.CoreV1().Pods("zombie").Get(pod.Name, metav1.GetOptions{})
//	if err != nil {
//		return newPodPhaseResponse(false, corev1.PodUnknown, err)
//	}
//
//	if actualPod.Status.Phase == corev1.PodRunning{
//		return newPodPhaseResponse(true, corev1.PodRunning, nil)
//	} else if actualPod.Status.Phase == corev1.PodPending{
//		return newPodPhaseResponse(false, corev1.PodPending, nil)
//	}
//
//	return newPodPhaseResponse(false, actualPod.Status.Phase, err)
//}
//
//
//func triggerPodStatus(c *kubernetes.Clientset, pod *corev1.Pod) <-chan *podPhaseResponse {
//	podResponse := make(chan *podPhaseResponse)
//	go func() {
//		defer close(podResponse)
//		podResponse <- isPodRunning(c, pod)
//	}()
//	return podResponse
//}

//func GetDepTest(manifests []string) error {
//	clientSet, err := GetKubeClient()
//	if err != nil{
//		return err
//	}
//	deploymentsClient := clientSet.AppsV1().Deployments("zombie")
//
//	//pods, err := clientSet.CoreV1().Pods("zombie").List(metav1.ListOptions{})
//	//if err != nil {
//	//	return err
//	//}
//
//	d := clientSet.CoreV1().Namespaces()
//	dl, err := d.List(metav1.ListOptions{})
//	if err != nil {
//		panic(err.Error())
//	} else {
//		for i, n := range dl.Items {
//			log.Debugf("%d %s", i, n.Name)
//		}
//	}
//
//	//jobsClient := clientSet.BatchV1().Jobs("zombie")
//
//	//watch jobsClient.Watch()
//	//decode := scheme.Codecs.UniversalDeserializer().Decode
//	//obj, _, err := decode([]byte(manifests[0]), nil, nil)
//	//if err != nil {
//	//	log.Println(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
//	//}
//	//
//	//deldep := obj.(*appsv1.Deployment)
//	//log.Debugf("Deleting %s", deldep.Name)
//	//err = deploymentsClient.Delete(deldep.Name, &metav1.DeleteOptions{})
//	//if err != nil {
//	//
//	//}
//
//	//dep, err := deploymentsClient.Create(obj.(*appsv1.Deployment))
//	//log.Debugf("Deployment: %s is deployed", dep.Name)
//
//
//	//deps, err := deploymentsClient.List(metav1.ListOptions{})
//	//if err != nil {
//	//	log.Errorf("Error deployment list")
//	//}
//	//if deps != nil {
//	//	selector := labels.Set(deps.Items[0].Spec.Selector.MatchLabels).String()
//	//	pods, err := clientSet.CoreV1().Pods("zombie").List(metav1.ListOptions{LabelSelector: selector})
//	//
//	//	if err != nil {
//	//		return err
//	//	}
//	//
//	//	for _, pod := range pods.Items{
//	//		log.Infof("Pod: %s", pod.Name)
//	//		for i:=0; i< 3; i++{
//	//			select {
//	//			case r:= <- triggerPodStatus(clientSet, &pod):
//	//				log.Infof("Pod %s ready is %t. Status %s", pod.Name, r.done, r.phase)
//	//			}
//	//		}
//	//
//	//	}
//	//}
//
//
//	//deployments, err := deploymentsClient.List(metav1.ListOptions{})
//	//if err != nil {
//	//	log.Errorf("Deployments get error")
//	//} else {
//	//	for i, e := range deployments.Items {
//	//		log.Debugf("%d: %s", i, e.ObjectMeta.Name)
//	//	}
//	//
//	//}
//	return nil
//}
