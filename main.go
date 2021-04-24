package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"

	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"

	corev1 "k8s.io/api/core/v1"

	"github.com/mwlng/k8s_resources_sync/pkg/helpers"
	"github.com/mwlng/k8s_resources_sync/pkg/utils"
)

var (
	homeDir string
)

func init() {
	klog.InitFlags(nil)
	homeDir = utils.GetHomeDir()
}

func main() {
	defer func() {
		klog.Flush()
	}()

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	environ := flag.String("e", "alpha", "Target environment")
	srcEksClusterName := flag.String("source_cluster_name", "", "Source k8s cluster name")

	flag.Set("v", "2")
	flag.Parse()

	if len(*srcEksClusterName) == 0 {
		klog.Infoln("No specified source k8s cluster name, nothing to restore, exit !")
		Usage()
		os.Exit(0)
	}

	klog.Infoln("Loading client kubeconfig ...")

	sourceKubeConfig, err := helpers.GetKubeConfig(*srcEksClusterName, *kubeconfig)
	if err != nil {
		panic(err)
	}

	klog.Infof("Starting to restore k8s externl dns for ELB from %s in %s ...", sourceKubeConfig.Host, *environ)
	lbServices, err := ListServices(sourceKubeConfig, corev1.ServiceTypeLoadBalancer)
	if err != nil {
		klog.Errorf("Failed to list services. Err was %s", err)
	}

	for _, lbService := range lbServices {
		fmt.Printf("%+v\n", lbService)
	}
}

func Usage() {
	fmt.Println()
	fmt.Printf("Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}
