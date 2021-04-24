package main

import (
	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/rest"

	"github.com/mwlng/k8s_resources_sync/pkg/k8s_resources"
)

type LBService struct {
	Name         string
	LoadBalancer string
	HostDnsName  string
}

func ListServices(kubeConfig *rest.Config, lbType string) ([]*LBService, error) {
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
		lbService := &LBService{
			Name:         s.Name,
			LoadBalancer: s.Status.LoadBalancer.String(),
			HostDnsName:  s.Annotations[""],
		}

		lbServices = append(lbServices, lbService)
	}

	return lbServices, nil
}
