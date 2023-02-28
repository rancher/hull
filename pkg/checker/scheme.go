package checker

import (
	"k8s.io/apimachinery/pkg/runtime"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
)

var (
	Scheme = runtime.NewScheme()
)

func init() {
	// workloads
	if err := appsv1.AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if err := autoscalingv1.AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if err := batchv1.AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if err := corev1.AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if err := policyv1beta1.AddToScheme(Scheme); err != nil {
		panic(err)
	}

	// rbac
	if err := corev1.AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if err := rbacv1.AddToScheme(Scheme); err != nil {
		panic(err)
	}

	// networking
	if err := corev1.AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if err := networkingv1.AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if err := apiregistrationv1.AddToScheme(Scheme); err != nil {
		panic(err)
	}

	// globals
	if err := admissionregistrationv1.AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if err := apiextensionsv1.AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if err := corev1.AddToScheme(Scheme); err != nil {
		panic(err)
	}
}
