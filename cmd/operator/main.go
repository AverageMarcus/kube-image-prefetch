package operator

import (
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"

	"kube-image-prefetch/internal/controller"
	"kube-image-prefetch/internal/prefetcher"
)

const (
	namespace = "default"
	name      = "kube-image-prefetch"
	image     = "averagemarcus/kube-image-prefetch:latest"
)

// Run triggers the operator to watch deployments and update the prefetcher daemonset
func Run() error {
	clientset, err := getClient()
	if err != nil {
		return err
	}

	ds, err := clientset.AppsV1().DaemonSets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		ds = prefetcher.CreateDaemonSet(getSelfOwnerReference(clientset)...)
		ds, err = clientset.AppsV1().DaemonSets(namespace).Create(ds)
		if err != nil {
			return err
		}
	}

	imageChan := make(chan controller.Images, 1)
	controller.Start(clientset, imageChan)

	toPrefetch := map[string]controller.Images{}
	for {
		img := <-imageChan

		if img.Images == nil {
			delete(toPrefetch, img.ID)
		} else {
			toPrefetch[img.ID] = img
		}

		images := []string{}
		pullSecrets := []corev1.LocalObjectReference{}

		for _, v := range toPrefetch {
			images = append(images, v.Images...)
			pullSecrets = append(pullSecrets, v.PullSecrets...)
		}

		ds, _ = clientset.AppsV1().DaemonSets(namespace).Get(name, metav1.GetOptions{})
		ds, err = clientset.AppsV1().DaemonSets(namespace).Patch(name, types.JSONPatchType, prefetcher.GeneratePatch(dedupe(images), pullSecrets))
		if err != nil {
			return err
		}
	}
}

func getClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfigPath := os.Getenv("KUBECONFIG")
		if kubeconfigPath == "" {
			kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	return kubernetes.NewForConfig(config)
}

func dedupe(a []string) []string {
	tempMap := map[string]bool{}
	dest := []string{}

	for _, obj := range a {
		if !tempMap[obj] {
			tempMap[obj] = true
			dest = append(dest, obj)
		}
	}

	return dest
}

func getSelfOwnerReference(clientset *kubernetes.Clientset) []metav1.OwnerReference {
	ownerReference := []metav1.OwnerReference{}

	self, err := clientset.AppsV1().Deployments("kube-system").Get("kube-image-prefetch-manager", metav1.GetOptions{})
	if err != nil {
		return ownerReference
	}

	ownerReference = append(ownerReference, metav1.OwnerReference{
		APIVersion: appsv1.SchemeGroupVersion.Identifier(),
		Kind:       "Deployment",
		Name:       self.ObjectMeta.Name,
		UID:        self.ObjectMeta.UID,
		Controller: pointer.BoolPtr(true),
	})

	return ownerReference
}
