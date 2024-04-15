package v1alpha1

import (
	"context"
	"fmt"
	"time"

	ttlcontroller "github.com/enamespace/ttl-controller/pkg/generated/clientset/versioned"
	ttlinformer "github.com/enamespace/ttl-controller/pkg/generated/informers/externalversions/ttlcontroller/v1alpha1"
	ttllister "github.com/enamespace/ttl-controller/pkg/generated/listers/ttlcontroller/v1alpha1"
	"golang.org/x/time/rate"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	cached "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	workqueue "k8s.io/client-go/util/workqueue"
)

type Controller struct {
	ttlclientset  ttlcontroller.Interface
	kubeclient    *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
	lister        ttllister.TTLLister
	syncd         cache.InformerSynced
	queue         workqueue.RateLimitingInterface
}

func New(ctx context.Context, clientset ttlcontroller.Interface, informer ttlinformer.TTLInformer, kubeclient *kubernetes.Clientset, dynamicClient *dynamic.DynamicClient) *Controller {
	logger := klog.FromContext(ctx)
	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	controller := &Controller{
		ttlclientset:  clientset,
		kubeclient:    kubeclient,
		dynamicClient: dynamicClient,
		lister:        informer.Lister(),
		syncd:         informer.Informer().HasSynced,
		queue:         workqueue.NewRateLimitingQueue(ratelimiter),
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

// enqueueTTL takes a TTL resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than TTL.
func (c *Controller) enqueueTTL(obj interface{}) {
	var key string
	var err error

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	c.queue.Add(key)
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

func (c *Controller) syncHandler(ctx context.Context, key string) error {
	logger := klog.LoggerWithValues(klog.FromContext(ctx), "resourceName", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key %s", key))
		return nil
	}

	ttlObj, err := c.lister.TTLs(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("resource key %s in work queue no long exist", key))
			return nil
		}

		return err
	}

	u := &unstructured.Unstructured{}
	u.SetAPIVersion(ttlObj.Spec.TTLTargetRef.APIVersion)
	u.SetKind(ttlObj.Spec.TTLTargetRef.Kind)
	u.SetName(ttlObj.Spec.TTLTargetRef.Name)
	gk := u.GetObjectKind().GroupVersionKind().GroupKind()
	v := u.GetObjectKind().GroupVersionKind().Version
	tgtName := ttlObj.Spec.TTLTargetRef.Name
	resourceMapper, err := restmapper.NewDeferredDiscoveryRESTMapper(cached.NewMemCacheClient(c.kubeclient.DiscoveryClient)).RESTMapping(gk, v)

	if err != nil {
		return err
	}

	var dClient dynamic.ResourceInterface

	if resourceMapper.Scope.Name() == meta.RESTScopeNameNamespace {
		dClient = c.dynamicClient.Resource(resourceMapper.Resource).Namespace(ttlObj.GetNamespace())
	} else {
		dClient = c.dynamicClient.Resource(resourceMapper.Resource)
	}

	// Test if found
	_, err = dClient.Get(context.TODO(), tgtName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("tgt resource named %s no longer exist", tgtName))
			return nil
		}

		return err
	}

	logger.Info("Found dynamic result, Delete it")

	err = dClient.Delete(context.TODO(), tgtName, metav1.DeleteOptions{})

	return err
}

func (c *Controller) processNextItem(ctx context.Context) bool {
	obj, shutdown := c.queue.Get()
	logger := klog.FromContext(ctx)
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.queue.Done(obj)

		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.queue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(ctx, key); err != nil {
			c.queue.AddRateLimited(key)
			return fmt.Errorf("error syncing %s: %s", key, err.Error())
		}

		c.queue.Forget(obj)
		logger.Info("Successful synced", "resourceName", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}
