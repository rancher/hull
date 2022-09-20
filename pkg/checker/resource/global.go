package resource

import (
	"github.com/aiyengar2/hull/pkg/checker"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func init() {
	if err := admissionregistrationv1.AddToScheme(checker.Scheme); err != nil {
		panic(err)
	}
	if err := apiextensionsv1.AddToScheme(checker.Scheme); err != nil {
		panic(err)
	}
	if err := corev1.AddToScheme(checker.Scheme); err != nil {
		panic(err)
	}
}

type CRDs []*apiextensionsv1.CustomResourceDefinition
type Namespaces []*corev1.Namespace
type MutatingWebhookConfigurations []*admissionregistrationv1.MutatingWebhookConfiguration
type ValidatingWebhookConfigurations []*admissionregistrationv1.ValidatingWebhookConfiguration
