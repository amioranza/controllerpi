package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// This program lists the pods in a cluster equivalent to
//
// kubectl get pods
//
func main() {
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
		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			log.Fatalln("failed to get pods:", err)
		}

		nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			log.Fatalln("failed to get nodes:", err)
		}

		// print pods
		fmt.Println("+#+#+#+#   PODS    +#+#+#+#")
		for i, pod := range pods.Items {
			fmt.Printf("[%d] %s\n", i, pod.GetName())
		}
		fmt.Println("+#+#+#+#   NODES   +#+#+#+#")
		for j, node := range nodes.Items {
			fmt.Printf("[%d] %s - %s\n", j, node.GetName(), node.GetLabels())
		}
		log.Println("Sleeping 5s")
		time.Sleep(5000 * time.Millisecond)
	}
}
