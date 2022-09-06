package resource

import (
	"github.com/aiyengar2/hull/pkg/test"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
)

func init() {
	corev1.AddToScheme(test.Scheme)
	networkingv1.AddToScheme(test.Scheme)
	apiregistrationv1.AddToScheme(test.Scheme)
}

type Networking struct {
	APIServices
	Services
	Ingresses

	NetworkPolicies
}

type APIServices []*apiregistrationv1.APIService
type Services []*corev1.Service
type Ingresses []*networkingv1.Ingress
type NetworkPolicies []*networkingv1.NetworkPolicy
