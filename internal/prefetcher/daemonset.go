package prefetcher

import (
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace = "default"
	name      = "kube-image-prefetch"
	image     = "averagemarcus/kube-image-prefetch:latest"
)

type ContainerPatch struct {
	Op    string             `json:"op"`
	Path  string             `json:"path"`
	Value []corev1.Container `json:"value"`
}

type PullSecretsPatch struct {
	Op    string                        `json:"op"`
	Path  string                        `json:"path"`
	Value []corev1.LocalObjectReference `json:"value"`
}

func CreateDaemonset() *appsv1.DaemonSet {
	labels := map[string]string{
		"app": name,
	}

	ds := &appsv1.DaemonSet{
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
					Containers: []corev1.Container{
						{
							Name:            "pending",
							Image:           image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"/mount/sleep"},
							Args:            []string{"-command", "sleep"},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1m"),
									corev1.ResourceMemory: resource.MustParse("10M"),
								},
							},
						},
					},
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

	return ds
}

func GeneratePatch(images []string, pullSecrets []corev1.LocalObjectReference) []byte {
	containers := []corev1.Container{}
	for i, img := range images {
		containers = append(containers, buildPrefetchContainer(img, i))
	}

	payload, _ := json.Marshal([]interface{}{
		ContainerPatch{
			Op:    "replace",
			Path:  "/spec/template/spec/containers",
			Value: containers,
		},
		PullSecretsPatch{
			Op:    "replace",
			Path:  "/spec/template/spec/imagePullSecrets",
			Value: pullSecrets,
		},
	})
	return payload
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
