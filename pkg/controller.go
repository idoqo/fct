package pkg

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const maxRetries = 5

type Controller struct {
	kubeclientset kubernetes.Interface
	queue workqueue.RateLimitingInterface
	informer cache.SharedIndexInformer
}

type Event struct {
	key string
	eventType string
	resourceType string
	//annotations map[string]string
}
func NewController(client kubernetes.Interface, informer cache.SharedIndexInformer) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	ctl := &Controller{
		kubeclientset: client,
		queue: queue,
		informer: informer,
	}

	ctl.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.handleObject,
	})

	return ctl
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("Starting flatcartag controller")

	go c.informer.Run(stopCh)

	klog.Info("Waiting for informer caches to sync")
	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("caches failed to sync"))
	}

	klog.Info("starting flatcartag workers")
	go wait.Until(c.runWorker, time.Second, stopCh)
	klog.Info("started flatcartag workers")
	<-stopCh
	klog.Info("shutting down flatcartag workers")
}

func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	var event Event
	if object, ok = obj.(metav1.Object); !ok {
		// probably a delete event, we should probably try recovering it from tombstone
		klog.Info("error decoding object, invalid type")
	} else {
		klog.Infof("Processing object: %s", object.GetName())
		//filter by node, and then by os (so we only enqueue)
		event.eventType = "create"
		event.resourceType = "node"
		//event.annotations = object.GetAnnotations()
		c.queue.Add(event)
	}
}

// processNextItem will read a single work off the workqueue and try to process it
func (c *Controller) processNextItem() bool {
	event, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(event)

	err := c.processItem(event.(Event))
	if err == nil {
		c.queue.Forget(event)
	} else if c.queue.NumRequeues(event) < maxRetries {
		klog.Errorf("Error processing %s (will retry): %v", event.(Event).key, err)
		c.queue.AddRateLimited(event)
	} else {
		klog.Errorf("Error processing %s (giving up): %v", event.(Event).key, err)
		c.queue.Forget(event)
		utilruntime.HandleError(err)
	}
	return true
}

func (c *Controller) processItem(event Event) error {
	_, _, err := c.informer.GetIndexer().GetByKey(event.key)
	if err != nil {
		return fmt.Errorf("failed to fetch object with key: %s from store: %s", event.key, err)
	}
	//objMeta := getObjectMetaData(obj)
	klog.Infof(fmt.Sprintf("Detected resource change for %s, type: %s", event.resourceType, event.eventType))
//	klog.Infof(fmt.Sprintf("%v", event.annotations))
	return nil
}