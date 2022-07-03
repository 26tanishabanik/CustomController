package deployments


import (
	//"flag"
	"log"
	"time"
	
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	//meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	// "github.com/26tanishabanik/customController/pods"
	//"k8s.io/client-go/tools/cache"
	//"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/watch"
)

func DeploymentsMain(kubeconfig *string) {
	// kubeconfig := flag.String("kubeconfig", "/home/tanisha/.kube/config", "location of the kubeconfig file")
	// kubeconfig := pods.GetkubeConfig()
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

	ch2 := make(chan struct{})
	
	informers2 := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	
	

	c2 := newDeploymentController(clientset, informers2.Apps().V1().Deployments())
	informers2.Start(ch2)
	defer close(ch2)
	c2.runDeployment(ch2)

	

}

