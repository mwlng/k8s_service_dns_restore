package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/mwlng/aws-go-clients/clients"
	"github.com/mwlng/aws-go-clients/service"
	"github.com/mwlng/k8s_resources_sync/pkg/helpers"
	"github.com/mwlng/k8s_resources_sync/pkg/utils"
)

const (
	defaultRegion  = "us-east-1"
	defaultEnviron = "alpha"
)

var (
	homeDir       string
	r53Cli        *clients.R53Client
	hostedZoneIds = map[string]string{
		"alpha": "Z0300709FFG0S6WMO3S",
		"qa":    "Z0176301YOHIVCTCW9TG",
		"prod":  "Z08708993EO08ZJY0OUSO",
	}
)

func init() {
	klog.InitFlags(nil)
	homeDir = utils.GetHomeDir()

	svc := service.Service{
		Region: defaultRegion,
	}
	sess := svc.NewSession()

	r53Cli = clients.NewClient("route53", sess).(*clients.R53Client)

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

	environ := flag.String("e", defaultEnviron, "Target environment")
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

	zoneId := hostedZoneIds[*environ]
	recordSets := []*route53.ResourceRecordSet{}
	for _, lbService := range lbServices {
		dnsName := lbService.DnsName
		if !strings.HasSuffix(dnsName, ".") {
			dnsName = fmt.Sprintf("%s.", dnsName)
		}
		recordSet := r53Cli.GetResourceRecordSet(&dnsName, &zoneId)

		if recordSet != nil {
			recordSets = append(recordSets, recordSet)
		} else {
			klog.Infof("Can't find route53 resource record set with name: %s, skipped\n", dnsName)
		}
	}

	action := "UPSERT"
	comment := "Restore dns records related to k8s loadbalacer service"
	result := r53Cli.ChangeResourceRecordSets(recordSets, &action, &zoneId, &comment)

	klog.Infof("%s\n", result)

}

func Usage() {
	fmt.Println()
	fmt.Printf("Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}
