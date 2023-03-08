package checker

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func Once(checkFunc func(tc *TestContext)) ChainedCheckFunc {
	return func(tc *TestContext) CheckFunc {
		checkFunc(tc)
		return nil
	}
}

func OnResources[O metav1.Object](typedListCheckFunc func(tc *TestContext, objects []O)) ChainedCheckFunc {
	return func(tc *TestContext) CheckFunc {
		return func(t *testing.T, objs struct{ Objects []O }) {
			tc.T = t
			typedListCheckFunc(tc, objs.Objects)
		}
	}
}

func PerResource[O metav1.Object](typedCheckFunc func(tc *TestContext, object O)) ChainedCheckFunc {
	return OnResources(func(tc *TestContext, objs []O) {
		for _, obj := range objs {
			typedCheckFunc(tc, obj)
		}
	})
}

func OnWorkloads(typedCheckFunc func(tc *TestContext, podTemplateSpecs map[metav1.Object]corev1.PodTemplateSpec)) ChainedCheckFunc {
	return func(tc *TestContext) CheckFunc {
		return func(t *testing.T, objs struct {
			Deployments []*appsv1.Deployment
			DaemonSet   []*appsv1.DaemonSet
			StatefulSet []*appsv1.StatefulSet
			ReplicaSet  []*appsv1.ReplicaSet
			Jobs        []*batchv1.Job
			CronJobs    []*batchv1.CronJob
		}) {
			tc.T = t
			podTemplateSpecs := make(map[metav1.Object]corev1.PodTemplateSpec)
			for _, obj := range objs.Deployments {
				podTemplateSpecs[obj] = obj.Spec.Template
			}
			for _, obj := range objs.DaemonSet {
				podTemplateSpecs[obj] = obj.Spec.Template
			}
			for _, obj := range objs.StatefulSet {
				podTemplateSpecs[obj] = obj.Spec.Template
			}
			for _, obj := range objs.ReplicaSet {
				podTemplateSpecs[obj] = obj.Spec.Template
			}
			for _, obj := range objs.Jobs {
				podTemplateSpecs[obj] = obj.Spec.Template
			}
			for _, obj := range objs.CronJobs {
				podTemplateSpecs[obj] = obj.Spec.JobTemplate.Spec.Template
			}
			typedCheckFunc(tc, podTemplateSpecs)
		}
	}
}

func PerWorkload(typedCheckFunc func(tc *TestContext, obj metav1.Object, podTemplateSpec corev1.PodTemplateSpec)) ChainedCheckFunc {
	return OnWorkloads(func(tc *TestContext, podTemplateSpecs map[metav1.Object]corev1.PodTemplateSpec) {
		for obj, podTemplateSpec := range podTemplateSpecs {
			typedCheckFunc(tc, obj, podTemplateSpec)
		}
	})
}
