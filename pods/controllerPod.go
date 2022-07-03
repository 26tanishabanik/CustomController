package pods

import (
	"context"
	"fmt"
	"time"
	"log"
	corev1 "k8s.io/api/core/v1"
	"github.com/26tanishabanik/customController/producer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	podinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	podlisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type controller struct {
	clientset      kubernetes.Interface
	podLister      podlisters.PodLister
	podCacheSynced cache.InformerSynced
	queue          workqueue.RateLimitingInterface
	handler   	   Handler
}

type Handler interface {
	Init() error
	ObjectCreated(obj interface{})
	ObjectDeleted(obj interface{})
	ObjectUpdated(objOld, objNew interface{})
}

type TestHandler struct{}

func (t *TestHandler) Init() error {
	log.Println("TestHandler.Init")
	return nil
}

func (t *TestHandler) ObjectCreated(obj interface{}) {
	log.Println("Pod Created")
	pod := obj.(*corev1.Pod)
	log.Printf("    ResourceVersion: %s", pod.ObjectMeta.ResourceVersion)
	log.Printf("    Name: %s", pod.Name)
	log.Printf("    NameSpace: %s", pod.ObjectMeta.Namespace)
	log.Printf("    Phase: %s", pod.Status.Phase)
	words := producer.Link{
		Link:        fmt.Sprintf("Pod Created\n    ResourceVersion: %s\n    Name: %s\n    NameSpace: %s\n    Phase: %s\n", pod.ObjectMeta.ResourceVersion, 
		pod.Name, pod.ObjectMeta.Namespace, pod.Status.Phase),
		Description: "",
		Topic:   	 "",
	}
	producer.SendLinktoproducer(words)
}

func (t *TestHandler) ObjectDeleted(obj interface{}) {
	log.Println("Pod Deleted")
	pod := obj.(*corev1.Pod)
	log.Printf("    ResourceVersion: %s", pod.ObjectMeta.ResourceVersion)
	log.Printf("    Name: %s", pod.Name)
	log.Printf("    NameSpace: %s", pod.ObjectMeta.Namespace)
	words := producer.Link{
		Link:        fmt.Sprintf("Pod Deleted\n    ResourceVersion: %s\n    Name: %s\n    NameSpace: %s\n", pod.ObjectMeta.ResourceVersion, 
		pod.Name, pod.ObjectMeta.Namespace),
		Description: "",
		Topic:   	 "",
	}
	producer.SendLinktoproducer(words)
	
}

func (t *TestHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Println("Pod Updated")
	pod := objNew.(*corev1.Pod)
	log.Printf("    ResourceVersion: %s", pod.ObjectMeta.ResourceVersion)
	log.Printf("    Name: %s", pod.Name)
	log.Printf("    NameSpace: %s", pod.ObjectMeta.Namespace)
	log.Printf("    Phase: %s", pod.Status.Phase)
	words := producer.Link{
		Link:        fmt.Sprintf("Pod Updated\n    ResourceVersion: %s\n    Name: %s\n    NameSpace: %s\n    Phase: %s\n", pod.ObjectMeta.ResourceVersion, 
		pod.Name, pod.ObjectMeta.Namespace, pod.Status.Phase),
		Description: "",
		Topic:   	 "",
	}
	producer.SendLinktoproducer(words)
	
	
}


func newController(clientset kubernetes.Interface, podInformer podinformers.PodInformer) *controller {
	c := &controller{
		clientset:      clientset,
		podLister:      podInformer.Lister(),
		podCacheSynced: podInformer.Informer().HasSynced,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "yamlGenPod"),
		handler: 		&TestHandler{},
	}
	podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDel,
		},
	)
	return c
}

func (c *controller) run(ch <-chan struct{}) {
	fmt.Println("Starting Pod Controller......")
	words := producer.Link{
		Link:        "Starting Pod Controller......",
		Description: "",
		Topic:   	 "",
	}
	producer.SendLinktoproducer(words)
	if !cache.WaitForCacheSync(ch, c.podCacheSynced) {
		fmt.Println("Error in pod local cache syncing\n")
	}
	go wait.Until(c.worker, 1*time.Second, ch)
	<-ch
}

func (c *controller) worker() {
	for c.processItem() {

	}

}

func (c *controller) processItem() bool {
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

	
	
	pod, err1 := c.clientset.CoreV1().Pods(ns).Get(context.Background(), name, metav1.GetOptions{})
	if apierrors.IsNotFound(err1){
		fmt.Printf("%s pod deleted\n", name)
		words := producer.Link{
			Link:        fmt.Sprintf("%s pod deleted\n", name),
			Description: "",
			Topic:   	 "",
		}
		producer.SendLinktoproducer(words)
		c.handler.ObjectDeleted(item)
		return true
	}else{
		c.handler.ObjectCreated(item)
	}

	err = c.syncPods(ns, name, pod)
	if err != nil {
		fmt.Printf("Error in syncing pods: %s\n", err.Error())
		return false
	}

	return true	
}



func (c *controller) syncPods(ns string, name string, pod1 *corev1.Pod) error {
	pod, err := c.podLister.Pods(ns).Get(name)

	fmt.Println("Pod Name: ", pod.Name)
	words := producer.Link{
		Link:        pod.Name,
		Description: "",
		Topic:   	 "",
	}
	producer.SendLinktoproducer(words)
	
	if err != nil {
		fmt.Printf("Error in getting pod from lister: %s\n", err.Error())
	}
	

	return nil
	
}

func returnReplica(num int32) *int32{
	return &num
}



func (c *controller) handleAdd(obj interface{}) {
	_, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		fmt.Printf("Error in getting key from cache: %s\n", err.Error())
	}
	
	fmt.Println("Add was called")
	words := producer.Link{
		Link:        "Pod Add was called",
		Description: "",
		Topic:   	 "",
	}
	producer.SendLinktoproducer(words)
	c.queue.Add(obj)
	
}

func (c *controller) handleDel(obj interface{}) {
	_, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		fmt.Printf("Error in getting key from cache: %s\n", err.Error())
	}
	words := producer.Link{
		Link:        "Pod Delete was called",
		Description: "",
		Topic:   	 "",
	}
	fmt.Println("Delete was called")
	producer.SendLinktoproducer(words)
	c.queue.Add(obj)
	
}

