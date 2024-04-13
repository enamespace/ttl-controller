package v1alpha1

import (
	"context"
	"fmt"
	"time"

	ttlcontroller "github.com/enamespace/ttl-controller/pkg/generated/clientset/versioned"
	ttlinformer "github.com/enamespace/ttl-controller/pkg/generated/informers/externalversions/ttlcontroller/v1alpha1"
	ttllister "github.com/enamespace/ttl-controller/pkg/generated/listers/ttlcontroller/v1alpha1"
	"golang.org/x/time/rate"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	workqueue "k8s.io/client-go/util/workqueue"
)

type Controller struct {
	ttlclientset ttlcontroller.Interface

	lister ttllister.TTLLister
	syncd  cache.InformerSynced
	queue  workqueue.RateLimitingInterface
}

func New(ctx context.Context, clientset ttlcontroller.Interface, informer ttlinformer.TTLInformer) *Controller {
	logger := klog.FromContext(ctx)
	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	controller := &Controller{
		ttlclientset: clientset,

		lister: informer.Lister(),
		syncd:  informer.Informer().HasSynced,
		queue:  workqueue.NewRateLimitingQueue(ratelimiter),
	}
	logger.Info("Setting up event handler")
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueTTL,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueTTL(new)
		},
	})

	return controller
}

func (c *Controller) enqueueTTL(obj interface{}) {
	klog.Info("enqueueTTL: ", obj)
}

func (c *Controller) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	logger := klog.FromContext(ctx)
	logger.Info("Starting TTL Controller")
	logger.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), c.syncd); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logger.Info("Starting workers", "count", workers)

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	logger.Info("Started workers")
	<-ctx.Done()
	logger.Info("Shutting down workers")
	return nil
}

func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextItem(ctx) {

	}
}

func (c *Controller) processNextItem(ctx context.Context) bool {
	obj, shutdown := c.queue.Get()
	logger := klog.FromContext(ctx)
	if shutdown {
		return false
	}

	logger.Info("Processing", obj)
	return true
}
