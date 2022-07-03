package pods


import (
	"flag"
	"log"
	"time"
	
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	//meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/client-go/tools/cache"
	//"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/watch"
	"github.com/26tanishabanik/customController/deployments"
)

// func GetkubeConfig() *string{
// 	kubeconfig := flag.String("kubeconfig", "/home/tanisha/.kube/config", "location of the kubeconfig file")
// 	return kubeconfig
// }
// var kubeconfig *string = flag.String("kubeconfig", "/home/tanisha/.kube/config", "location of the kubeconfig file")
func Both(){
	kubeconfig := flag.String("kubeconfig", "/home/tanisha/.kube/config", "location of the kubeconfig file")
	go PodsMain(kubeconfig)
	go deployments.DeploymentsMain(kubeconfig)
}


func PodsMain(kubeconfig *string) {
	//kubeconfig := flag.String("kubeconfig", "/home/tanisha/.kube/config", "location of the kubeconfig file")
	// kubeconfig := GetkubeConfig()
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
	// ch2 := make(chan struct{})
	informers1 := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	// informers2 := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	
	// c := newController(clientset, informers.Apps().V1().Deployments())

	c1 := newController(clientset,informers1.Core().V1().Pods())
	informers1.Start(ch1)
	defer close(ch1)
	c1.run(ch1)

	// c2 := newDeploymentController(clientset, informers2.Apps().V1().Deployments())
	// informers2.Start(ch2)
	// defer close(ch2)
	// c2.runDeployment(ch2)
}

