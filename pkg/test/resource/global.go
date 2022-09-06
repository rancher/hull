package resource

import (
	"github.com/aiyengar2/hull/pkg/test"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func init() {
	admissionregistrationv1.AddToScheme(test.Scheme)
	apiextensionsv1.AddToScheme(test.Scheme)
	corev1.AddToScheme(test.Scheme)
}

type Global struct {
	CRDs
	Namespaces

	MutatingWebhookConfigurations
	ValidatingWebhookConfigurations
}

type CRDs []*apiextensionsv1.CustomResourceDefinition
type Namespaces []*corev1.Namespace
type MutatingWebhookConfigurations []*admissionregistrationv1.MutatingWebhookConfiguration
type ValidatingWebhookConfigurations []*admissionregistrationv1.ValidatingWebhookConfiguration
