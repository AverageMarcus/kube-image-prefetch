package operator

import (
	"kube-image-prefetch/internal/prefetcher"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	namespace = "default"
	name      = "kube-image-prefetch"
	image     = "averagemarcus/kube-image-prefetch:latest"
	timeout   = 15 * time.Minute
)

func Run() error {
	clientset, err := getClient()
	if err != nil {
		return err
	}

	for {
		deployments, err := getDeployments(clientset)
		if err != nil {
			return err
		}

		images, pullSecrets := parseDeployments(deployments)

		ds := prefetcher.BuildDaemonset(images, pullSecrets)

		existingDs, err := clientset.AppsV1().DaemonSets(namespace).Get(name, metav1.GetOptions{})
		if err == nil {
			ds.ResourceVersion = existingDs.ResourceVersion
		}

		if ds.ResourceVersion == "" {
			ds, err = clientset.AppsV1().DaemonSets("default").Create(ds)
		} else {
			ds, err = clientset.AppsV1().DaemonSets("default").Update(ds)
		}
		if err != nil {
			return err
		}

		time.Sleep(timeout)
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

func getDeployments(clientset *kubernetes.Clientset) ([]appsv1.Deployment, error) {
	dps, err := clientset.AppsV1().Deployments(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return dps.Items, nil
}

func parseDeployments(deployments []appsv1.Deployment) ([]string, []corev1.LocalObjectReference) {
	imagesMap := map[string]bool{}
	pullSecrets := []corev1.LocalObjectReference{}
	for _, dp := range deployments {
		for _, container := range append(dp.Spec.Template.Spec.Containers, dp.Spec.Template.Spec.InitContainers...) {
			imagesMap[container.Image] = true
		}
		pullSecrets = append(pullSecrets, dp.Spec.Template.Spec.ImagePullSecrets...)
	}

	images := []string{}
	for img := range imagesMap {
		images = append(images, img)
	}

	return images, pullSecrets
}
