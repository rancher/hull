package resource

import (
	"github.com/aiyengar2/hull/pkg/checker"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
)

func init() {
	if err := corev1.AddToScheme(checker.Scheme); err != nil {
		panic(err)
	}
	if err := networkingv1.AddToScheme(checker.Scheme); err != nil {
		panic(err)
	}
	if err := apiregistrationv1.AddToScheme(checker.Scheme); err != nil {
		panic(err)
	}
}

type APIServices []*apiregistrationv1.APIService
type Services []*corev1.Service
type Ingresses []*networkingv1.Ingress
type NetworkPolicies []*networkingv1.NetworkPolicy
