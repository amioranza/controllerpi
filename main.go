package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const repository = "amioranza"
const namespace = "pi-system"
const appversion = ":v0"

func int32Ptr(i int32) *int32 { return &i }
func boolPtr(b bool) *bool    { return &b }

func nodeLabel(nodeName string, labelName string, labelValue string, op string) {
	clientset, err := getConfig()
	if err != nil {
		log.Fatalln("failed to get the config:", err)
	}

	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get nodes:", err)
	}

	if op == "add" {
		for _, node := range nodes.Items {
			if node.GetName() == nodeName {
				node.Labels[labelName] = labelValue
				fmt.Printf("%s - %s\n", node.GetName(), node.GetLabels())
				_, err := clientset.Core().Nodes().Update(&node)
				if err != nil {
					log.Fatalln("failed to get nodes:", err)
				}
			}
		}
	} else if op == "del" {
		for _, node := range nodes.Items {
			if node.GetName() == nodeName {
				delete(node.Labels, labelName)
				fmt.Printf("%s - %s\n", node.GetName(), node.GetLabels())
				_, err := clientset.Core().Nodes().Update(&node)
				if err != nil {
					log.Fatalln("failed to get nodes:", err)
				}
			}
		}
	}
}

func createContainer(name, image string) (container apiv1.Container) {
	container = apiv1.Container{
		Name:  name,
		Image: image,
		SecurityContext: &apiv1.SecurityContext{
			Privileged: boolPtr(true),
		},
		VolumeMounts: []apiv1.VolumeMount{
			{
				Name:      "mem",
				MountPath: "/dev/mem",
			},
			{
				Name:      "gpiomem",
				MountPath: "/dev/gpiomem",
			},
		},
	}
	return container
}

// DeployApp handles application deployment on k8s
func DeployApp(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	status := params["status"]
	log.Printf("Params: %s\n", params)

	clientset, err := getConfig()
	if err != nil {
		log.Fatalln("failed to get the config:", err)
	}

	deploymentsClient := clientset.AppsV1().Deployments(namespace)

	deploymentName := params["app"] + "-deployment"

	list, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get deployments:", err)
	}

	containers := []apiv1.Container{createContainer(params["app"], repository+"/"+params["app"]+appversion)}

	if status == "true" {
		for _, d := range list.Items {
			if d.Name == deploymentName && *d.Spec.Replicas > 0 {
				log.Printf("Deployment already running: %s\n", deploymentName)
				return
			}
		}

		nodeLabel(params["node"], "app", params["app"], "add")

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: deploymentName,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": params["app"],
					},
				},
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": params["app"],
						},
					},
					Spec: apiv1.PodSpec{
						Containers: containers,
						NodeSelector: map[string]string{
							"app": params["app"],
						},
						Volumes: []apiv1.Volume{
							{
								Name: "mem",
								VolumeSource: apiv1.VolumeSource{
									HostPath: &apiv1.HostPathVolumeSource{
										Path: "/dev/mem",
									},
								},
							},
							{
								Name: "gpiomem",
								VolumeSource: apiv1.VolumeSource{
									HostPath: &apiv1.HostPathVolumeSource{
										Path: "/dev/gpiomem",
									},
								},
							},
						},
					},
				},
			},
		}
		// Create Deployment
		fmt.Println("Creating deployment...")
		result, err := deploymentsClient.Create(deployment)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	} else {

		nodeLabel(params["node"], "app", params["app"], "del")

		fmt.Println("Deleting deployment...")
		deletePolicy := metav1.DeletePropagationForeground
		if err := deploymentsClient.Delete(deploymentName, &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}); err != nil {
			panic(err)
		}
		fmt.Println("Deleted deployment.")
	}

}

// SayHello says Hello
func SayHello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<html><h1>HELLO FROM PI-SYSTEM</h1></html>")
}

func getConfig() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if kubeconfig == "/root/.kube/config" {
		log.Println("Using in cluster config")
		config, err = clientcmd.BuildConfigFromFlags("", "")
		// in cluster access
	} else {
		log.Println("Using out of cluster config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func main() {

	go func() {
		// Create an rest client not targeting specific API version
		clientset, err := getConfig()
		if err != nil {
			log.Fatalln("failed to get the config:", err)
		}

		deploymentsClient := clientset.AppsV1().Deployments(namespace)

		for {
			pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
			if err != nil {
				log.Fatalln("failed to get pods:", err)
			}

			nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
			if err != nil {
				log.Fatalln("failed to get nodes:", err)
			}

			list, err := deploymentsClient.List(metav1.ListOptions{})
			if err != nil {
				log.Fatalln("failed to get deployments:", err)
			}
			fmt.Println("\n____________________  D E P L O Y S    ______________________")
			for _, d := range list.Items {
				fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
			}

			// print pods
			fmt.Println("\n______________________   P O D S    _________________________")
			for i, pod := range pods.Items {
				fmt.Printf("[%d] %s\n", i, pod.GetName())
			}
			fmt.Println("\n______________________   N O D E S    ________________________")
			for j, node := range nodes.Items {
				fmt.Printf("[%d] %s - %s\n", j, node.GetName(), node.GetLabels())
			}
			log.Println("Sleeping 5s")
			time.Sleep(5000 * time.Millisecond)
		}
	}()

	// r := gin.Default()
	// r.POST("/:app/:status/:node", deployApp1)
	// r.GET("/", SayHello1)
	// r.Run(":8000")

	// new restfull api server
	router := mux.NewRouter()
	router.HandleFunc("/{app}/{status}/{node}", DeployApp).Methods("POST")
	router.HandleFunc("/", SayHello).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))

}
