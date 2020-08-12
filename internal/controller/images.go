package controller

import (
	"strings"

	appsv1 "k8s.io/api/apps/v1"
)

func getImages(dp appsv1.Deployment) []string {
	images := []string{}

	ignoreDP := dp.ObjectMeta.Annotations["kube-image-prefetch/ignore"]
	if ignoreDP == "true" {
		return images
	}

	ignoreContainersStr := dp.ObjectMeta.Annotations["kube-image-prefetch/ignore-containers"]
	ignoreContainers := strings.Split(ignoreContainersStr, ",")

	for _, container := range append(dp.Spec.Template.Spec.InitContainers, dp.Spec.Template.Spec.Containers...) {
		if !contains(ignoreContainers, container.Name) {
			images = append(images, container.Image)
		}
	}

	return images
}

func contains(arr []string, str string) bool {
	for _, c := range arr {
		if strings.TrimSpace(strings.ToLower(str)) == strings.TrimSpace(strings.ToLower(c)) {
			return true
		}
	}
	return false
}
