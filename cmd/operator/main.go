package operator

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

		ds, _ := clientset.AppsV1().DaemonSets(namespace).Get(name, metav1.GetOptions{})
		if ds == nil || ds.ObjectMeta.Name == "" {
			ds = buildDaemonSet()
		}

		ds.Spec.Template.Spec.Containers = []corev1.Container{}
		ds.Spec.Template.Spec.ImagePullSecrets = pullSecrets

		i := 0
		for img := range images {
			ds.Spec.Template.Spec.Containers = append(
				ds.Spec.Template.Spec.Containers,
				buildPrefetchContainer(img, i),
			)
			i++
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
		return nil, err
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

func parseDeployments(deployments []appsv1.Deployment) (map[string]bool, []corev1.LocalObjectReference) {
	images := map[string]bool{}
	pullSecrets := []corev1.LocalObjectReference{}
	for _, dp := range deployments {
		for _, container := range append(dp.Spec.Template.Spec.Containers, dp.Spec.Template.Spec.InitContainers...) {
			images[container.Image] = true
		}
		pullSecrets = append(pullSecrets, dp.Spec.Template.Spec.ImagePullSecrets...)
	}

	return images, pullSecrets
}

func buildDaemonSet() *appsv1.DaemonSet {
	labels := map[string]string{
		"app": name,
	}
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{{
						Name:            "init",
						Image:           image,
						ImagePullPolicy: corev1.PullAlways,
						Args: []string{
							"-command", "copy",
							"-dest", "/mount/sleep",
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "share",
							MountPath: "/mount",
						}},
					}},
					Containers:       []corev1.Container{},
					ImagePullSecrets: []corev1.LocalObjectReference{},
					Volumes: []corev1.Volume{{
						Name: "share",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					}},
				},
			},
		},
	}
}

func buildPrefetchContainer(img string, index int) corev1.Container {
	return corev1.Container{
		Name:            fmt.Sprintf("prefetch-%d", index),
		Image:           img,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/mount/sleep"},
		Args:            []string{"-command", "sleep"},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "share",
			MountPath: "/mount",
		}},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1m"),
				corev1.ResourceMemory: resource.MustParse("10M"),
			},
		},
	}
}
