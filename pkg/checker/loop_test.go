package checker

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestLoop(t *testing.T) {
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
			Name:       "No Checks",
			Checks:     nil,
			FinalCheck: func(tc *TestContext) {},
		},
		{
			Name: "Ensure Once only runs once",
			Checks: []ChainedCheckFunc{
				Once(func(tc *TestContext) {
					Store(tc, "i", 0)
				}),
				Once(func(tc *TestContext) {
					i, ok := Get[string, int](tc, "i")
					if !ok {
						assert.Fail(tc.T, "could not retrieve stored value for i")
					}
					Store(tc, "i", i+1)
				}),
			},
			FinalCheck: func(tc *TestContext) {
				i, ok := Get[string, int](tc, "i")
				if !ok {
					assert.Fail(tc.T, "could not retrieve stored value for i")
				}
				assert.Equal(tc.T, 1, i)
			},
		},
		{
			Name: "Ensure PerResource runs on all resources",
			Checks: []ChainedCheckFunc{
				Once(func(tc *TestContext) {
					Store(tc, "i", 0)
				}),
				PerResource(func(tc *TestContext, _ *unstructured.Unstructured) {
					i, ok := Get[string, int](tc, "i")
					if !ok {
						assert.Fail(tc.T, "could not retrieve stored value for i")
					}
					Store(tc, "i", i+1)
				}),
			},
			FinalCheck: func(tc *TestContext) {
				i, ok := Get[string, int](tc, "i")
				if !ok {
					assert.Fail(tc.T, "could not retrieve stored value for i")
				}
				assert.Equal(tc.T, len(exampleTestObjects), i)
			},
		},
		{
			Name: "Ensure PerWorkload runs on all workloads",
			Checks: []ChainedCheckFunc{
				Once(func(tc *TestContext) {
					Store(tc, "i", 0)
				}),
				PerWorkload(func(tc *TestContext, _ metav1.Object, _ corev1.PodTemplateSpec) {
					i, ok := Get[string, int](tc, "i")
					if !ok {
						assert.Fail(tc.T, "could not retrieve stored value for i")
					}
					Store(tc, "i", i+1)
				}),
			},
			FinalCheck: func(tc *TestContext) {
				i, ok := Get[string, int](tc, "i")
				if !ok {
					assert.Fail(tc.T, "could not retrieve stored value for i")
				}
				assert.Equal(tc.T, 6, i)
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
}
