package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const ContainerNum int = 100

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Create 100 client
	for i := 0; i < ContainerNum; i++ {
		createClientPod(clientset, fmt.Sprintf("counter-%d", i), i)
	}

	fmt.Println("Wait for the all clients are runnning.")
	fmt.Println("Press Any Key")
	prompt()

	// Create 100 Spamer
	for i := 0; i < ContainerNum; i++ {
		createSpamerPod(clientset, fmt.Sprintf("counter-%d", i), i)
	}

	fmt.Println("Press Any Key")
	prompt()

	// Delete Pod
	fmt.Println("Deleting pod...")

	// Delete Clients
	deletePods(clientset, "spamer")
	deletePods(clientset, "client")

}

func deletePods(clientset *kubernetes.Clientset, header string) {
	client := clientset.CoreV1().Pods(apiv1.NamespaceDefault)
	deletePolicyPod := metav1.DeletePropagationForeground
	for i := 0; i < ContainerNum; i++ {
		if err := client.Delete(fmt.Sprintf("%s-%d", header, i), &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicyPod,
		}); err != nil {
			panic(err)
		}
		fmt.Printf("Deleted pod %s-%d\n", header, i)
	}
}

func createSpamerPod(clientset *kubernetes.Clientset, guid string, count int) {
	client := clientset.CoreV1().Pods(apiv1.NamespaceDefault)

	pdeployment := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("spamer-%d", count),
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:  "web",
					Image: "nginx:1.12",
					Env: []apiv1.EnvVar{
						{
							Name:  "guid",
							Value: guid,
						},
						{
							Name:  "queName",
							Value: "que1",
						},
						{
							Name:  "messagesCount",
							Value: "10",
						},
						{
							Name:  "StorageConnectionString",
							Value: "",
						},
					},
				},
			},
			RestartPolicy: "Never",
		},
	}

	fmt.Printf("Name: %s Guid:%s Que:%s MessageCount:%s: ConnectionString:%s¥n", fmt.Sprintf("spamer-%d", count), guid, "que1", "10", "")

	// Create Deployment
	fmt.Println("Creating a pod...")
	result, err := client.Create(pdeployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created a pod. %q.\n", result.GetObjectMeta().GetName())
}

func createClientPod(clientset *kubernetes.Clientset, guid string, count int) {
	client := clientset.CoreV1().Pods(apiv1.NamespaceDefault)

	pdeployment := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("client-%d", count),
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:  "web",
					Image: "nginx:1.12",
					Env: []apiv1.EnvVar{
						{
							Name:  "DeviceID",
							Value: guid,
						},
						{
							Name:  "StorageConnectionString",
							Value: "",
						},
					},
				},
			},
			RestartPolicy: "Never",
		},
	}

	fmt.Printf("Name: %s DeviceID:%s ConnectionString:%s\n", fmt.Sprintf("client-%d", count), guid, "")

	// Create Deployment
	fmt.Println("Creating a pod...")
	result, err := client.Create(pdeployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created a pod. %q.\n", result.GetObjectMeta().GetName())
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

func int32Ptr(i int32) *int32 { return &i }
