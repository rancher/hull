package checker

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var exampleTestObjects = []metav1.Object{
	&rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: fmt.Sprintf("%s/v1", rbacv1.GroupName),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "global-role",
		},
	},
	&rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: fmt.Sprintf("%s/v1", rbacv1.GroupName),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "global-binding",
		},
	},
	&corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-configmap",
			Namespace: "default",
		},
	},
	&appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: fmt.Sprintf("%s/v1", appsv1.GroupName),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: "default",
		},
	},
	&appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: fmt.Sprintf("%s/v1", appsv1.GroupName),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-daemonset",
			Namespace: "default",
		},
	},
	&appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: fmt.Sprintf("%s/v1", appsv1.GroupName),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-statefulset",
			Namespace: "default",
		},
	},
	&appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicaSet",
			APIVersion: fmt.Sprintf("%s/v1", appsv1.GroupName),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-replicaset",
			Namespace: "default",
		},
	},
	&batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: fmt.Sprintf("%s/v1", batchv1.GroupName),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-job",
			Namespace: "default",
		},
	},
	&batchv1.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: fmt.Sprintf("%s/v1", batchv1.GroupName),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cronjob",
			Namespace: "default",
		},
	},
}

func TestNewCheckFunc(t *testing.T) {
	var unstructuredObjects []*unstructured.Unstructured
	for _, obj := range exampleTestObjects {
		uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			t.Fatal(err)
		}
		unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{
			Object: uObj,
		})
	}
	objs := struct {
		Unstructured []*unstructured.Unstructured
	}{
		Unstructured: unstructuredObjects,
	}

	testCases := []struct {
		Name       string
		Checks     []ChainedCheckFunc
		FinalCheck func(tc *TestContext)
	}{
		{
			Name: "Nil Funcs",
		},
		{
			Name: "Empty Func",
			Checks: []ChainedCheckFunc{
				NewChainedCheckFunc[metav1.Object](nil),
			},
		},
		{
			Name: "Basic Example",
			Checks: []ChainedCheckFunc{
				NewChainedCheckFunc(func(tc *TestContext, objs struct{ Deployments []*appsv1.Deployment }) {
					assert.NotNil(tc.T, objs.Deployments)
					assert.Equal(tc.T, 1, len(objs.Deployments))
					Store(tc, "hello", "goodbye")
				}),
				NewChainedCheckFunc(func(tc *TestContext, objs struct{ DaemonSets []*appsv1.DaemonSet }) {
					assert.NotNil(tc.T, objs.DaemonSets)
					assert.Equal(tc.T, 1, len(objs.DaemonSets))
					Store(tc, "hello", "world")
				}),
				NewChainedCheckFunc(func(tc *TestContext, objs struct{ All []*unstructured.Unstructured }) {
					assert.NotNil(tc.T, objs.All)
					assert.Equal(tc.T, len(exampleTestObjects), len(objs.All))
					hello, ok := Get[string, string](tc, "hello")
					assert.True(tc.T, ok)
					assert.Equal(tc.T, "world", hello)
				}),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			args := []reflect.Value{
				reflect.ValueOf(t),
				reflect.ValueOf(objs),
			}
			checks := tc.Checks
			if tc.FinalCheck != nil {
				checks = append(checks, Once(tc.FinalCheck))
			}
			checkFunc := NewCheckFunc(checks...)
			checkFuncVal := reflect.ValueOf(checkFunc)
			checkFuncVal.Call(args)
		})
	}

	t.Run("Do Not Continue After Failure", func(t *testing.T) {
		fakeT := &testing.T{}
		args := []reflect.Value{
			reflect.ValueOf(fakeT),
			reflect.ValueOf(objs),
		}
		var reached bool
		checks := []ChainedCheckFunc{
			Once(func(tc *TestContext) {
				assert.Fail(tc.T, "initiate failure")
			}),
			Once(func(_ *TestContext) {
				reached = true
			}),
		}
		checkFunc := NewCheckFunc(checks...)
		checkFuncVal := reflect.ValueOf(checkFunc)
		checkFuncVal.Call(args)

		assert.False(t, reached, "should not have continued")
		assert.True(t, fakeT.Failed(), "should have failed")
	})

	t.Run("Do Not Continue After Failure Again", func(t *testing.T) {
		fakeT := &testing.T{}
		args := []reflect.Value{
			reflect.ValueOf(fakeT),
			reflect.ValueOf(objs),
		}
		var reached bool
		checks := []ChainedCheckFunc{
			NewChainedCheckFunc(func(tc *TestContext, _ struct{}) {
				assert.Fail(tc.T, "initiate failure")
			}),
			Once(func(_ *TestContext) {
				reached = true
			}),
		}
		checkFunc := NewCheckFunc(checks...)
		checkFuncVal := reflect.ValueOf(checkFunc)
		checkFuncVal.Call(args)

		assert.False(t, reached, "should not have continued")
		assert.True(t, fakeT.Failed(), "should have failed")
	})

	t.Run("Continue On Failure", func(t *testing.T) {
		fakeT := &testing.T{}
		args := []reflect.Value{
			reflect.ValueOf(fakeT),
			reflect.ValueOf(objs),
		}
		var reached bool
		checks := []ChainedCheckFunc{
			Once(func(tc *TestContext) {
				assert.Fail(tc.T, "initiate failure")
				tc.Continue()
			}),
			Once(func(_ *TestContext) {
				reached = true
			}),
		}
		checkFunc := NewCheckFunc(checks...)
		checkFuncVal := reflect.ValueOf(checkFunc)
		checkFuncVal.Call(args)

		assert.True(t, reached, "did not continue")
		assert.True(t, fakeT.Failed(), "should have failed")
	})
}
