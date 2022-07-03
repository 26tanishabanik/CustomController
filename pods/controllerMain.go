package pods


import (
	"flag"
	"log"
	"time"
	
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/26tanishabanik/customController/deployments"
)


func Both(){
	kubeconfig := flag.String("kubeconfig", "/home/tanisha/.kube/config", "location of the kubeconfig file")
	go PodsMain(kubeconfig)
	go deployments.DeploymentsMain(kubeconfig)
}


func PodsMain(kubeconfig *string) {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Error in building config: %s\n", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("Error in getting cluster config: %s\n", err.Error())
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Error in creating clientset: %s\n", err.Error())
	}
	ch1 := make(chan struct{})
	informers1 := informers.NewSharedInformerFactory(clientset, 10*time.Minute)

	c1 := newController(clientset,informers1.Core().V1().Pods())
	informers1.Start(ch1)
	defer close(ch1)
	c1.run(ch1)

}

