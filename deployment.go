// First try to create a function to deploy kubernetes apps, it relies on some structs

package main

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// func int32Ptr(i int32) *int32 { return &i }
// func boolPtr(b bool) *bool    { return &b }

// func createContainer(name, image string) (container apiv1.Container) {
// 	container = apiv1.Container{}
// 	return container
// }

func createDeployment(namespace string, name string, labels, nodeSelector map[string]string, containers []apiv1.Container, volumes []apiv1.Volume) (deployment *appsv1.Deployment) {
	deployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Containers:   containers,
					NodeSelector: nodeSelector,
					Volumes:      volumes,
				},
			},
		},
	}
	return deployment
}
