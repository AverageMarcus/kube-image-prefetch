package controller

import (
	"testing"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetImages_SingleImage(t *testing.T) {
	dp := appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name: "container-1",
							Image: "image:1",
						},
					},
				},
			},
		},
	}

	expected := []string{"image:1"}
	actual := getImages(dp)

	if ! reflect.DeepEqual(expected, actual) {
		t.Errorf("Unexpected images returned - %v", actual)
	}
}

func TestGetImages_MultipleImage(t *testing.T) {
	dp := appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name: "container-1",
							Image: "image:1",
						},
						corev1.Container{
							Name: "container-2",
							Image: "image:2",
						},
					},
				},
			},
		},
	}

	expected := []string{"image:1","image:2"}
	actual := getImages(dp)

	if ! reflect.DeepEqual(expected, actual) {
		t.Errorf("Unexpected images returned - %v", actual)
	}
}

func TestGetImages_InitContainers(t *testing.T) {
	dp := appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						corev1.Container{
							Name: "initcontainer-1",
							Image: "initimage:1",
						},
					},
					Containers: []corev1.Container{
						corev1.Container{
							Name: "container-1",
							Image: "image:1",
						},
					},
				},
			},
		},
	}

	expected := []string{"initimage:1","image:1"}
	actual := getImages(dp)

	if ! reflect.DeepEqual(expected, actual) {
		t.Errorf("Unexpected images returned - %v", actual)
	}
}

func TestGetImages_IgnoreDeployment(t *testing.T) {
	dp := appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-deployment",
			Annotations: map[string]string{
				"kube-image-prefetch/ignore": "true",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						corev1.Container{
							Name: "initcontainer-1",
							Image: "initimage:1",
						},
					},
					Containers: []corev1.Container{
						corev1.Container{
							Name: "container-1",
							Image: "image:1",
						},
					},
				},
			},
		},
	}

	expected := []string{}
	actual := getImages(dp)

	if ! reflect.DeepEqual(expected, actual) {
		t.Errorf("Unexpected images returned - %v", actual)
	}
}

func TestGetImages_IgnoreContainer(t *testing.T) {
	dp := appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-deployment",
			Annotations: map[string]string{
				"kube-image-prefetch/ignore-containers": "initcontainer-1",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						corev1.Container{
							Name: "initcontainer-1",
							Image: "initimage:1",
						},
					},
					Containers: []corev1.Container{
						corev1.Container{
							Name: "container-1",
							Image: "image:1",
						},
					},
				},
			},
		},
	}

	expected := []string{"image:1"}
	actual := getImages(dp)

	if ! reflect.DeepEqual(expected, actual) {
		t.Errorf("Unexpected images returned - %v", actual)
	}
}

func TestGetImages_IgnoreMultipleContainer(t *testing.T) {
	dp := appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-deployment",
			Annotations: map[string]string{
				"kube-image-prefetch/ignore-containers": "initcontainer-1,container-1",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						corev1.Container{
							Name: "initcontainer-1",
							Image: "initimage:1",
						},
					},
					Containers: []corev1.Container{
						corev1.Container{
							Name: "container-1",
							Image: "image:1",
						},
					},
				},
			},
		},
	}

	expected := []string{}
	actual := getImages(dp)

	if ! reflect.DeepEqual(expected, actual) {
		t.Errorf("Unexpected images returned - %v", actual)
	}
}

func TestGetImages_IgnoreContainerSpaces(t *testing.T) {
	dp := appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-deployment",
			Annotations: map[string]string{
				"kube-image-prefetch/ignore-containers": "initcontainer-1, container-1",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						corev1.Container{
							Name: "initcontainer-1",
							Image: "initimage:1",
						},
					},
					Containers: []corev1.Container{
						corev1.Container{
							Name: "container-1",
							Image: "image:1",
						},
					},
				},
			},
		},
	}

	expected := []string{}
	actual := getImages(dp)

	if ! reflect.DeepEqual(expected, actual) {
		t.Errorf("Unexpected images returned - %v", actual)
	}
}
