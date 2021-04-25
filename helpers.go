package main

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/rest"

	"github.com/mwlng/k8s_resources_sync/pkg/k8s_resources"
)

type LBService struct {
	Name         string
	LoadBalancer string
	DnsName      string
}

func ListServices(kubeConfig *rest.Config, lbType corev1.ServiceType) ([]*LBService, error) {
	service, err := k8s_resources.NewService(kubeConfig, corev1.NamespaceDefault)
	if err != nil {
		return nil, err
	}

	lbServices := []*LBService{}
	serviceList, err := service.ListServices()
	if err != nil {
		return nil, err
	}

	for _, s := range serviceList.Items {
		if s.Spec.Type == lbType {
			lbService := &LBService{
				Name:         s.Name,
				LoadBalancer: s.Status.LoadBalancer.Ingress[0].Hostname,
				DnsName:      s.Annotations["external-dns.alpha.kubernetes.io/hostname"],
			}

			lbServices = append(lbServices, lbService)
		}
	}

	return lbServices, nil
}

func ChangeLBServiceResourceRecordSet(lbService *LBService, hostedZoneId *string) {
	dnsName := lbService.DnsName
	if !strings.HasSuffix(dnsName, ".") {
		dnsName = fmt.Sprintf("%s.", dnsName)
	}
	recordSet := r53Cli.GetResourceRecordSet(&dnsName, hostedZoneId)
	fmt.Printf("%+v\n", recordSet)
}
