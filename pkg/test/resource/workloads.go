package resource

import (
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"

	"github.com/aiyengar2/hull/pkg/test"
)

func init() {
	appsv1.AddToScheme(test.Scheme)
	autoscalingv1.AddToScheme(test.Scheme)
	batchv1.AddToScheme(test.Scheme)
	corev1.AddToScheme(test.Scheme)
	policyv1beta1.AddToScheme(test.Scheme)
}

type Compute struct {
	CronJobs
	DaemonSets
	Deployments
	Jobs
	StatefulSets

	Configuration

	PodSecurityPolicies
	HorizontalPodAutoscalers
}

type CronJobs []*batchv1.CronJob
type DaemonSets []*appsv1.DaemonSet
type Deployments []*appsv1.Deployment
type Jobs []*batchv1.Job
type StatefulSets []*appsv1.StatefulSet

type PodSecurityPolicies []*policyv1beta1.PodSecurityPolicy
type HorizontalPodAutoscalers []*autoscalingv1.HorizontalPodAutoscaler

type Configuration struct {
	ConfigMaps
	Secrets
}

type ConfigMaps []*corev1.ConfigMap
type Secrets []*corev1.Secret
