// First try to create a function to deploy kubernetes apps, it relies on some structs

package main

import (
	"log"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// func int32Ptr(i int32) *int32 { return &i }
// func boolPtr(b bool) *bool    { return &b }

type label struct {
	name  string
	value string
}

type port struct {
	name     string
	port     int32
	protocol string
}

type container struct {
	name  string
	image string
	ports []port
}

type volumeMount struct {
	name      string
	mountPath string
}

type hostPath struct {
	name string
	path string
}

type volumeSource struct {
	name      string
	hostPaths []hostPath
}

type volume struct {
	name          string
	volumeSources []volumeSource
}

type deployment struct {
	name         string
	namespace    string
	labels       []label
	containers   []container
	volumeMounts []volumeMount
	nodeSelector map[string]string
	volumes      []volume
}

func deployApplicationK8S(deploy deployment) (err error) {

	clientset, err := getConfig()
	if err != nil {
		log.Fatalln("Failed to get the config", err)
	}
	deploymentsClient := clientset.AppsV1().Deployments(deploy.namespace)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploy.name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					deploy.labels[0].name: deploy.labels[0].value,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						deploy.labels[0].name: deploy.labels[0].value,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  deploy.containers[0].name,
							Image: deploy.containers[0].image,
							Ports: []apiv1.ContainerPort{
								{
									Name:          deploy.containers[0].ports[0].name,
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: deploy.containers[0].ports[0].port,
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Privileged: boolPtr(true),
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      deploy.volumeMounts[0].name,
									MountPath: deploy.volumeMounts[0].mountPath,
								},
								{
									Name:      deploy.volumeMounts[1].name,
									MountPath: deploy.volumeMounts[1].mountPath,
								},
							},
						},
					},
					NodeSelector: map[string]string{
						deploy.labels[0].name: deploy.labels[0].value,
					},
					Volumes: []apiv1.Volume{
						{
							Name: deploy.volumes[0].name,
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: deploy.volumes[0].volumeSources[0].hostPaths[0].path,
								},
							},
						},
						{
							Name: deploy.volumes[1].name,
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: deploy.volumes[1].volumeSources[0].hostPaths[0].path,
								},
							},
						},
					},
				},
			},
		},
	}
	log.Println("Creating deployment...")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	log.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	return
}
