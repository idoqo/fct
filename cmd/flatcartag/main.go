package main

import (
	"context"
	"flag"
	"gitlab.com/idoko/flatcar-tag/pkg/controller"
	"k8s.io/client-go/rest"
	"os"
	"os/signal"
	"syscall"

	api_v1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	masterURL string
	kubeconfig string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	stopCh := setupSignalHandler()
	ctx := context.TODO()
	var kubeClient kubernetes.Interface

	if _, err := rest.InClusterConfig(); err != nil {
		kubeClient = getClientOutOfCluster()
	} else {
		kubeClient = getClient()
	}

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return kubeClient.CoreV1().Nodes().List(ctx, options)
			},

			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return kubeClient.CoreV1().Nodes().Watch(ctx, options)
			},
		},
		&api_v1.Node{},
		0,
		cache.Indexers{},
	)
	ctl := controller.NewController(kubeClient, informer)
	ctl.Run(stopCh)
}

func getClientOutOfCluster() kubernetes.Interface {
	var cfg *rest.Config
	var err error
	cfgPath := os.Getenv("KUBECONFIG")
	if cfgPath == "" {
		cfgPath = os.Getenv("HOME") + "/.kube/config"
	}
	cfg, err = clientcmd.BuildConfigFromFlags("", cfgPath)
	if err != nil {
		klog.Fatalf("Failed to get kubeconfig: %s", err.Error())
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatal("Failed to build kubernetes client: %s", err.Error())
	}
	return cs
}

func getClient() kubernetes.Interface {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Can not get kubernetes config: %s", err.Error())
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Can not create kubernetes client: %s", err.Error())
	}

	return cs
}

// setupSignalHandler listens for SIGTERM and SIGINT
func setupSignalHandler() <-chan struct{} {
	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<- c
		close(stop)
		<- c
		os.Exit(1) // close if a second signal is caught
	}()
	return stop
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Required if out-of-cluster")
	flag.StringVar(&masterURL, "master", "", "Address of the kube-api-server. Overrides any value in kubeconfig. Required if out-of-cluster")
}