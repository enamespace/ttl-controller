package main

import (
	"flag"
	"time"

	ttlcontroller "github.com/enamespace/ttl-controller/internal/ttlcontroller/v1alpha1"
	generatedclientset "github.com/enamespace/ttl-controller/pkg/generated/clientset/versioned"
	generatedinformer "github.com/enamespace/ttl-controller/pkg/generated/informers/externalversions"
	"github.com/enamespace/ttl-controller/pkg/signals"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	ctx := signals.SetupSignalHandler()
	logger := klog.FromContext(ctx)

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		logger.Error(err, "Error building kubeconfig")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	ttlclientset, err := generatedclientset.NewForConfig(cfg)
	if err != nil {
		logger.Error(err, "Error building kubernetes clientset")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	ttlInformerFactory := generatedinformer.NewSharedInformerFactory(ttlclientset, time.Second*30)

	ttlinformer := ttlInformerFactory.Ttlcontroller().V1alpha1()
	controller := ttlcontroller.New(ctx, ttlclientset, ttlinformer.TTLs())

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(ctx.done())
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	ttlInformerFactory.Start(ctx.Done())
	if err = controller.Run(ctx, 2); err != nil {
		logger.Error(err, "Error running controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
