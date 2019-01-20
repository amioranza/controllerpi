package main

import (
	"flag"
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
	"k8s.io/client-go/tools/clientcmd"
)

type App struct {
	Node   string `json:"node"`
	Pin    string `json:"pin"`
	Status string `json:"status"`
}

func int32Ptr(i int32) *int32 { return &i }

func DeployApp(w http.ResponseWriter, r *http.Request) {
	log.Printf("Raw params: %v", mux.Vars(r))
	params := mux.Vars(r)
	// node := params["node"]
	// pin := params["pin"]
	status := params["status"]
	log.Printf("Params: %s\n", params)

	// Bootstrap k8s configuration from local 	Kubernetes config file
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	log.Println("Using kubeconfig file: ", kubeconfig)

	// config kubernetes access, ext cluster, uncomment below
	// config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	// config kubernetes access, intra cluster, uncomment below
	config, err := clientcmd.BuildConfigFromFlags("", "")

	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	deploymentsClient := clientset.AppsV1().Deployments("pi-system")

	if status == "true" {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "led-deployment",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "led",
					},
				},
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "led",
						},
					},
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
							{
								Name:  "led",
								Image: "nginx:1.12",
								Ports: []apiv1.ContainerPort{
									{
										Name:          "http",
										Protocol:      apiv1.ProtocolTCP,
										ContainerPort: 80,
									},
								},
							},
						},
						NodeSelector: map[string]string{
							"app": "led",
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
		fmt.Println("Deleting deployment...")
		deletePolicy := metav1.DeletePropagationForeground
		if err := deploymentsClient.Delete("led-deployment", &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}); err != nil {
			// log.Println("ERROU!!!")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "<html><h1>ERROU!!!</h1></html>")
			panic(err)
		}
		fmt.Println("Deleted deployment.")
	}

}

func SayHello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<html><h1>HELLO from MimiServer</h1></html>")
}

// This program lists the pods in a cluster equivalent to
//
// kubectl get pods
//
func main() {

	go func() {
		var ns string
		flag.StringVar(&ns, "namespace", "", "namespace")

		// Bootstrap k8s configuration from local 	Kubernetes config file
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		log.Println("Using kubeconfig file: ", kubeconfig)
		//config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		config, err := clientcmd.BuildConfigFromFlags("", "")
		if err != nil {
			log.Fatal(err)
		}

		// Create an rest client not targeting specific API version
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Fatal(err)
		}

		for {
			pods, err := clientset.CoreV1().Pods("pi-system").List(metav1.ListOptions{})
			if err != nil {
				log.Fatalln("failed to get pods:", err)
			}

			nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
			if err != nil {
				log.Fatalln("failed to get nodes:", err)
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
	// new restfull api server
	router := mux.NewRouter()
	router.HandleFunc("/{pin}/{status}", DeployApp).Methods("POST")
	router.HandleFunc("/", SayHello).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))

}
