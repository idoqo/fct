package controller

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"strings"
	"time"

	api_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const (

	FlatcarOSName         = "flatcar" //todo: figure out actual os name for flatcar container linux
	usesFlatcarLabelKey   = "k8c.io~1uses-container-linux"
	usesFlatcarLabelValue = "true"
	maxRetries            = 5
)

type Controller struct {
	kubeclientset kubernetes.Interface
	queue         workqueue.RateLimitingInterface
	informer      cache.SharedIndexInformer
}

type Event struct {
	key          string
	eventType    string
}

func NewController(client kubernetes.Interface, informer cache.SharedIndexInformer) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	ctl := &Controller{
		kubeclientset: client,
		queue:         queue,
		informer:      informer,
	}

	var object metav1.Object
	var ok bool
	var event Event
	var err error
	ctl.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if object, ok = obj.(metav1.Object); !ok {
				klog.Info("error decoding object, invalid type")
			} else {
				klog.Infof("Processing object: %s", object.GetName())
				event.key, err = cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					utilruntime.HandleError(err)
				}
				event.eventType = "create"
				queue.Add(event)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if object, ok = newObj.(metav1.Object); !ok {
				klog.Info("error decoding new object, invalid type")
			} else {
				klog.Infof("Processing object: %s", object.GetName())
				event.key, err = cache.MetaNamespaceKeyFunc(newObj)
				if err != nil {
					utilruntime.HandleError(err)
				}
				switch objType := newObj.(type) {
				case *api_v1.Node:
					if _, exists := objType.Labels["k8c.io/uses-container-linux"]; !exists {
						// only requeue if the node isn't already labelled.
						event.eventType = "update"
						queue.Add(event)
					}
					break
				default:
				}
			}
		},
	})

	return ctl
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("Starting flatcartag controller")

	go c.informer.Run(stopCh)

	klog.Info("Waiting for informer caches to sync")
	if !cache.WaitForNamedCacheSync("fct-controller", stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("caches failed to sync"))
	}

	klog.Info("starting flatcartag workers")
	go wait.Until(c.runWorker, time.Second, stopCh)
	klog.Info("started flatcartag workers")
	<-stopCh
	klog.Info("shutting down flatcartag workers")
}

func CreateNodeInformer(client kubernetes.Interface) cache.SharedIndexInformer {
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Nodes().List(context.TODO(), options)
			},

			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Nodes().Watch(context.TODO(), options)
			},
		},
		&api_v1.Node{},
		0,
		cache.Indexers{},
	)
	return informer
}

func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
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
	obj, _, err := c.informer.GetIndexer().GetByKey(event.key)
	if err != nil {
		return fmt.Errorf("failed to fetch object with key: %s from store: %s", event.key, err)
	}
	switch objType := obj.(type) {
	case *api_v1.Node:
		node := obj.(*api_v1.Node)
		nodeOS := obj.(*api_v1.Node).Status.NodeInfo.OperatingSystem
		if strings.Contains(nodeOS, FlatcarOSName) {
			klog.Infof(fmt.Sprintf("Node %s running flatcar container linux, applying label {%s: %s}",
				node.Name, usesFlatcarLabelKey, usesFlatcarLabelValue))

			labelPatch := fmt.Sprintf(
				`[{"op": "add", "path":"/metadata/labels/%s", "value":"%s" }]`,
				usesFlatcarLabelKey,
				usesFlatcarLabelValue,
				)
			_, err = c.kubeclientset.CoreV1().Nodes().Patch(context.TODO(), node.Name, types.JSONPatchType, []byte(labelPatch), metav1.PatchOptions{})
			if err != nil {
				return fmt.Errorf("failed to label node: %v", err)
			}
		}
		break
	default:
		klog.Infof(fmt.Sprintf("ignoring detected resource change for %s, type: %s", objType, event.eventType))
	}
	return nil
}
