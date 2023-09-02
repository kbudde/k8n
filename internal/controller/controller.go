package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kbudde/k8n/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// Prometheus metrics.
//
//nolint:gochecknoglobals
var (
	detectedChanges = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "k8n_detected_changes_total",
		Help: "Number of detected changes",
	}, []string{"name", "operation"})
)

//nolint:ireturn
func NewIndexerInformer(prefix, kind, apiVersion, labelSelector string,
	queue workqueue.RateLimitingInterface, dynamicClient dynamic.Interface,
) (cache.Indexer, cache.Controller) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		panic(err)
	}

	gvr := gv.WithResource(kind)

	//nolint:exhaustruct
	listWatch := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.LabelSelector = labelSelector
			options.APIVersion = apiVersion
			options.Kind = kind

			return dynamicClient.Resource(gvr).Namespace(v1.NamespaceAll).List(context.Background(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.LabelSelector = labelSelector
			options.APIVersion = apiVersion
			options.Kind = kind

			return dynamicClient.Resource(gvr).Namespace(v1.NamespaceAll).Watch(context.Background(), options)
		},
	}
	resEvHndler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(fmt.Sprintf("Added %s/%s", prefix, key))
				detectedChanges.WithLabelValues(prefix, "add").Inc()
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(fmt.Sprintf("Update %s/%s", prefix, key))
				detectedChanges.WithLabelValues(prefix, "update").Inc()
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(fmt.Sprintf("Delete %s/%s", prefix, key))
				detectedChanges.WithLabelValues(prefix, "delete").Inc()
			}
		},
	}

	indexer, informer := cache.NewIndexerInformer(listWatch, nil, 0, resEvHndler, cache.Indexers{})

	return indexer, informer
}

func NewForConfig(config config.Config, restConfig *rest.Config) (*Controller, error) {
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	indexerM := map[string]cache.Indexer{}
	informerM := map[string]cache.Controller{}

	for _, w := range config.Watcher {
		indexer, informer := NewIndexerInformer(w.Name, w.Kind, w.APIVersion, w.Selector, queue, dynamicClient)
		indexerM[w.Name] = indexer
		informerM[w.Name] = informer
	}

	return NewController(queue, indexerM, informerM), nil
}

// Controller demonstrates how to implement a controller with client-go.
type Controller struct {
	indexer  map[string]cache.Indexer
	informer map[string]cache.Controller
	queue    workqueue.RateLimitingInterface
	process  chan []byte
	delay    time.Duration
	retries  int
}

// NewController creates a new Controller.
func NewController(queue workqueue.RateLimitingInterface,
	indexer map[string]cache.Indexer, informer map[string]cache.Controller,
) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
		process:  make(chan []byte),
		// might be a good idea to make this configurable
		delay:   time.Second,
		retries: 2, //nolint:gomnd
	}
}

// SetDelay between consecutive syncs. Defaults to 1 second.
func (c *Controller) SetDelay(d time.Duration) {
	c.delay = d
}

// SetRetries sets the number of retries before dropping an item out of the queue.
func (c *Controller) SetRetries(r int) {
	c.retries = r
}

// GetProcessChan returns the channel to which the processed data is sent.
func (c *Controller) GetProcessChan() <-chan []byte {
	return c.process
}

func (c *Controller) processNextItem() {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return
	}

	defer c.queue.Done(key)

	sKey, ok := key.(string)
	if !ok {
		c.handleErr(fmt.Errorf("key is not a string: %v", key), key)
	}

	err := c.doTheStuff(sKey)
	c.handleErr(err, key)
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)

		return
	}

	if c.queue.NumRequeues(key) < c.retries {
		klog.Infof("Error syncing: %v", err)
		c.queue.AddRateLimited(key)

		return
	}

	c.queue.Forget(key)
	klog.Infof("Dropping item out of the queue: %v", err)
}

// Run begins watching and syncing.
func (c *Controller) Run(stopCh chan struct{}) {
	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	klog.Info("Starting controller")

	for key, informer := range c.informer {
		klog.Infof("Starting informer %s", key)

		go informer.Run(stopCh)
	}

	klog.Info("Wait for caches to sync")
	{
		var wg sync.WaitGroup
		wg.Add(len(c.informer))
		for name, controller := range c.informer {
			go func(name string, controller cache.Controller) {
				defer wg.Done()
				if !cache.WaitForCacheSync(stopCh, controller.HasSynced) {
					panic(fmt.Sprintf("Cache sync failed for %s", name))
				}
				klog.Infof("Cache %s synced", name)
			}(name, controller)
		}

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			klog.Info("All caches synced successfully")
		case <-stopCh:
			panic("Timeout waiting for cache sync")
		}
	}

	// clear the queue and add one initial item
	for c.queue.Len() > 0 {
		key, quit := c.queue.Get()
		if quit {
			return
		}

		c.queue.Done(key)
	}
	c.queue.Add("initial sync")

	go wait.Until(c.processNextItem, c.delay, stopCh)

	<-stopCh
	klog.Info("Stopping controller")
}
