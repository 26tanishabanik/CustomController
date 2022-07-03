package deployments

import (
	"context"
	"fmt"
	"time"
	"log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"github.com/26tanishabanik/customController/producer"	
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type controllerDeployment struct {
	clientset      kubernetes.Interface
	depLister      appslisters.DeploymentLister
	depCacheSynced cache.InformerSynced
	queue          workqueue.RateLimitingInterface
	handler   	   DeploymentHandler
}


type DeploymentHandler interface {
	Init() error
	DeploymentObjectCreated(obj interface{})
	DeploymentObjectDeleted(obj interface{})
	DeploymentObjectUpdated(objOld, objNew interface{})
}

type TestHandlerDeployment struct{}

func (t *TestHandlerDeployment) Init() error {
	log.Println("TestHandlerDeployment.Init")
	return nil
}

func (t *TestHandlerDeployment) DeploymentObjectCreated(obj interface{}) {
	log.Println("Deployment Created")
	deploy := obj.(*appsv1.Deployment)
	log.Printf("    ResourceVersion: %s", deploy.ObjectMeta.ResourceVersion)
	log.Printf("    Name: %s", deploy.Name)
	log.Printf("    NameSpace: %s", deploy.ObjectMeta.Namespace)
	words := producer.Link{
		Link:        fmt.Sprintf("Deployment Created\n    ResourceVersion: %s\n    Name: %s\n    NameSpace: %s\n", deploy.ObjectMeta.ResourceVersion, 
		deploy.Name, deploy.ObjectMeta.Namespace),
		Description: "",
		Topic:   	 "",
	}
	producer.SendLinktoproducer(words)

}

func (t *TestHandlerDeployment) DeploymentObjectDeleted(obj interface{}) {
	log.Println("Deployment Deleted")
	deploy := obj.(*appsv1.Deployment)
	log.Printf("    ResourceVersion: %s", deploy.ObjectMeta.ResourceVersion)
	log.Printf("    Name: %s", deploy.Name)
	log.Printf("    NameSpace: %s", deploy.ObjectMeta.Namespace)
	words := producer.Link{
		Link:        fmt.Sprintf("Deployment Deleted\n    ResourceVersion: %s\n    Name: %s\n    NameSpace: %s\n", deploy.ObjectMeta.ResourceVersion, 
		deploy.Name, deploy.ObjectMeta.Namespace),
		Description: "",
		Topic:   	 "",
	}
	producer.SendLinktoproducer(words)
	
}

func (t *TestHandlerDeployment) DeploymentObjectUpdated(objOld, objNew interface{}) {
	log.Println("Deployment Updated")
	deploy := objNew.(*appsv1.Deployment)
	log.Printf("    ResourceVersion: %s", deploy.ObjectMeta.ResourceVersion)
	log.Printf("    Name: %s", deploy.Name)
	log.Printf("    NameSpace: %s", deploy.ObjectMeta.Namespace)
	words := producer.Link{
		Link:        fmt.Sprintf("Deployment Updated\n    ResourceVersion: %s\n    Name: %s\n    NameSpace: %s\n", deploy.ObjectMeta.ResourceVersion, 
		deploy.Name, deploy.ObjectMeta.Namespace),
		Description: "",
		Topic:   	 "",
	}
	producer.SendLinktoproducer(words)
}



func newDeploymentController(clientset kubernetes.Interface, depInformer appsinformers.DeploymentInformer) *controllerDeployment {
	c := &controllerDeployment{
		clientset:      clientset,
		depLister:      depInformer.Lister(),
		depCacheSynced: depInformer.Informer().HasSynced,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "yamlGenDeployment"),
		handler: 		&TestHandlerDeployment{},
	}
	depInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAddDeployment,
			DeleteFunc: c.handleDelDeployment,
		},
	)
	return c
}

func (c *controllerDeployment) runDeployment(ch <-chan struct{}) {
	fmt.Println("Starting Deployment Controller......")
	words := producer.Link{
		Link:        "Starting Deployment Controller......",
		Description: "",
		Topic:   	 "",
	}
	producer.SendLinktoproducer(words)
	if !cache.WaitForCacheSync(ch, c.depCacheSynced) {
		fmt.Println("Error in deployment local cache syncing\n")
	}
	go wait.Until(c.workerDeployment, 1*time.Second, ch)
	<-ch
}

func (c *controllerDeployment) workerDeployment() {
	for c.processDeploymentItem() {

	}

}

func (c *controllerDeployment) processDeploymentItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		fmt.Printf("Error in getting key from cache: %s\n", err.Error())
	}
	ns, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		fmt.Printf("Error in splitting key into namespace and name: %s\n", err.Error())
	}

	
	
	
	deploy, err2 := c.clientset.AppsV1().Deployments(ns).Get(context.Background(), name, metav1.GetOptions{})
	if apierrors.IsNotFound(err2){
		fmt.Printf("%s deployment deleted\n", name)
		words := producer.Link{
			Link:        fmt.Sprintf("%s deployment deleted\n", name),
			Description: "",
			Topic:   	 "",
		}
		producer.SendLinktoproducer(words)
		err := c.clientset.CoreV1().Services(ns).Delete(context.Background(), name, metav1.DeleteOptions{})
		if err != nil {
			fmt.Printf("deleting service %s, error %s\n", name, err.Error())
			return false
		}
		c.handler.DeploymentObjectDeleted(item)
		return true
	}else{
		c.handler.DeploymentObjectCreated(item)
	}
	

	
	err = c.syncDeployment(ns, name, deploy)
	if err != nil {
		fmt.Printf("Error in syncing deployments: %s\n", err.Error())
		return false
	}

	return true	
}

func (c *controllerDeployment) syncDeployment(ns string, name string, deploy1 *appsv1.Deployment) error {
	deploy, err := c.depLister.Deployments(ns).Get(name)

	fmt.Println("Deployment Name: ", deploy.Name)
	words := producer.Link{
		Link:        deploy.Name,
		Description: "",
		Topic:   	 "",
	}
	producer.SendLinktoproducer(words)
	
	if err != nil {
		fmt.Printf("Error in getting deployment from lister: %s\n", err.Error())
	}
	_, err = c.clientset.AppsV1().Deployments(ns).Create(context.Background(), &dep, metav1.CreateOptions{})
	if err != nil{
		fmt.Printf("Error in creating deployment: %s\n", err.Error())
	}
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:		pod.Name,
			Namespace:	ns,
		},
		Spec: corev1.ServiceSpec{
			Selector: depLabels(dep),
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
			},
		},
	}
	_, err = c.clientset.CoreV1().Services(ns).Create(context.Background(), &svc, metav1.CreateOptions{})
	if err != nil{
		fmt.Printf("Error in creating service: %s\n", err.Error())
	}

	return nil
}


func depLabels(dep appsv1.Deployment) map[string]string{
	return dep.Spec.Template.Labels
}

func (c *controllerDeployment) handleAddDeployment(obj interface{}) {
	_, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		fmt.Printf("Error in getting key from cache: %s\n", err.Error())
	}
	
	fmt.Println("Deployment Add was called")
	words := producer.Link{
		Link:        "Deployment Add was called",
		Description: "",
		Topic:   	 "",
	}
	producer.SendLinktoproducer(words)
	c.queue.Add(obj)
	
}

func (c *controllerDeployment) handleDelDeployment(obj interface{}) {
	_, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		fmt.Printf("Error in getting key from cache: %s\n", err.Error())
	}
	
	fmt.Println("Deployment Delete was called")
	words := producer.Link{
		Link:        "Deployment Delete was called",
		Description: "",
		Topic:   	 "",
	}
	producer.SendLinktoproducer(words)
	c.queue.Add(obj)
	
}
