package main

import (
	"context"
	"fmt"
	"time"
	"log"
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	//netv1  "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	//appsinformers "k8s.io/client-go/informers/apps/v1"
	podinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	// appslisters "k8s.io/client-go/listers/apps/v1"
	podlisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// type controller struct {
// 	clientset      kubernetes.Interface
// 	depLister      appslisters.DeploymentLister
// 	depCacheSynced cache.InformerSynced
// 	queue          workqueue.RateLimitingInterface
// }
type controller struct {
	clientset      kubernetes.Interface
	podLister      podlisters.PodLister
	podCacheSynced cache.InformerSynced
	queue          workqueue.RateLimitingInterface
}

// func newController(clientset kubernetes.Interface, depInformer appsinformers.DeploymentInformer) *controller {
// 	c := &controller{
// 		clientset:      clientset,
// 		depLister:      depInformer.Lister(),
// 		depCacheSynced: depInformer.Informer().HasSynced,
// 		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "yamlGen"),
// 	}
// 	depInformer.Informer().AddEventHandler(
// 		cache.ResourceEventHandlerFuncs{
// 			AddFunc:    c.handleAdd,
// 			DeleteFunc: c.handleDel,
// 		},
// 	)
// 	return c
// }

func newController(clientset kubernetes.Interface, podInformer podinformers.PodInformer) *controller {
	c := &controller{
		clientset:      clientset,
		podLister:      podInformer.Lister(),
		podCacheSynced: podInformer.Informer().HasSynced,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "yamlGen"),
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
	fmt.Println("Starting Controller......")
	if !cache.WaitForCacheSync(ch, c.podCacheSynced) {
		fmt.Println("Error in local cache syncing\n")
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

	// _, err = c.clientset.AppsV1().Deployments(ns).Get(context.Background(), name, metav1.GetOptions{})
	if ns == "test1"{
		pod, err := c.clientset.CoreV1().Pods(ns).Get(context.Background(), name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			fmt.Printf("%s pod deleted\n", name)
			err := c.clientset.CoreV1().Services(ns).Delete(context.Background(), name, metav1.DeleteOptions{})
			if err != nil {
				fmt.Printf("deleting service %s, error %s\n", name, err.Error())
				return false
			}
			return true
		}

		err = c.syncDeployment(ns, name, pod)
		if err != nil {
			fmt.Printf("Error in syncing pods: %s\n", err.Error())
			return false
		}
		return true
	}else{
		return false
	}
	
}

func (c *controller) syncDeployment(ns string, name string, pod1 *corev1.Pod) error {
	pod, err := c.podLister.Pods(ns).Get(name)
	
	if err != nil {
		fmt.Printf("Error in getting pod from lister: %s\n", err.Error())
	}
	fmt.Println("pod namespace", ns)
	
	dep := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: pod.Name + "-api",
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": pod.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": pod.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  pod.Name,
							Image: pod1.Spec.Containers[0].Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	if ns == "test1"{
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
	return nil

	//return createIngress(context.Background(), c.clientset, s)
}

func returnReplica(num int32) *int32{
	return &num
}


func depLabels(dep appsv1.Deployment) map[string]string{
	return dep.Spec.Template.Labels
}

func (c *controller) handleAdd(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		fmt.Printf("Error in getting key from cache: %s\n", err.Error())
	}
	ns, _, _ := cache.SplitMetaNamespaceKey(key)
	
	if ns == "test1"{
		pod := obj.(*corev1.Pod)
		log.Printf("    ResourceVersion: %s", pod.ObjectMeta.ResourceVersion)
		log.Printf("    NodeName: %s", pod.Spec.NodeName)
		log.Printf("    Phase: %s", pod.Status.Phase)
			fmt.Println("Add was called")
		c.queue.Add(obj)
	}
}

func (c *controller) handleDel(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		fmt.Printf("Error in getting key from cache: %s\n", err.Error())
	}
	ns, _, _ := cache.SplitMetaNamespaceKey(key)
	if ns == "test1"{
		fmt.Println("Delete was called")
		c.queue.Add(obj)
	}
}
/*
import (
	"context"
	"fmt"
	"log"
	"time"

	core_v1 "k8s.io/api/core/v1"
	podLister "k8s.io/client-go/listers/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Controller struct defines how a controller should encapsulate
// logging, client connectivity, informing (list and watching)
// queueing, and handling of resource changes

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

	pod := obj.(*core_v1.Pod)
	log.Printf("    ResourceVersion: %s", pod.ObjectMeta.ResourceVersion)
	log.Printf("    NodeName: %s", pod.Spec.NodeName)
	log.Printf("    Phase: %s", pod.Status.Phase)
}

func (t *TestHandler) ObjectDeleted(obj interface{}) {
	log.Println("Pod Deleted")
}

func (t *TestHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Println("Pod Updated")
}

type Controller struct {
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
	podLister podLister.PodLister
	handler   Handler
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	// handle a panic with logging and exiting
	defer utilruntime.HandleCrash()
	// ignore new items in the queue but when all goroutines
	// have completed existing items then shutdown
	defer c.queue.ShutDown()

	log.Println("Starting Controller.............")

	// run the informer to start listing and watching resources
	go c.informer.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("error syncing cache"))
		return
	}

	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

func (c *Controller) runWorker() {
	//log.Println("Controller.runWorker: starting")

	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {

	key, quit := c.queue.Get()

	if quit {
		return false
	}
	//defer c.queue.Forget(key)

	defer c.queue.Done(key)

	keyRaw := key.(string)

	item, exists, err := c.informer.GetIndexer().GetByKey(keyRaw)
	if err != nil {
		if c.queue.NumRequeues(key) < 5 {
			log.Println("Error processing item with key %s with error %v, retrying", key, err)
			c.queue.AddRateLimited(key)
		} else {
			log.Println("Error processing item with key %s with error %v, no more retries", key, err)
			c.queue.Forget(key)
			utilruntime.HandleError(err)
		}
	}
	// item1, err := cache.MetaNamespaceKeyFunc(keyRaw)
	// if err != nil {
	// 	fmt.Printf("getting key from cache %s\n", err.Error())
	// }
	item.
	ns, name, err := cache.SplitMetaNamespaceKey(item.(string))
	if err != nil {
		fmt.Printf("splitting key into namespace and name %s\n", err.Error())
		return false
	}

	
	

	if !exists {
		log.Println(" object detected: %s", keyRaw)
		c.handler.ObjectDeleted(item)
		c.queue.Forget(key)
	} else {
		log.Println("object detected: %s", keyRaw)
		// dep, err := c.podLister.Pods(ns).Get(name)
		// if err != nil {
		// 	fmt.Printf("getting deployment from lister %s\n", err.Error())
		// }
		svc := core_v1.Service{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      name,
				Namespace: ns,
			},
			Spec: core_v1.ServiceSpec{
				//Selector: depLabels(*dep),
				Ports: []core_v1.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
				},
			},
		}
		_, err = c.clientset.CoreV1().Services(ns).Create(context.Background(), &svc, meta_v1.CreateOptions{})
		if err != nil {
			fmt.Printf("creating service %s\n", err.Error())
		}
		c.handler.ObjectCreated(item)
		c.queue.Forget(key)
	}

	return true
}

// func depLabels(core core_v1) map[string]string {
// 	return dep.Spec.Template.Labels
// }

*/
