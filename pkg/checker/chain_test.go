package checker

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewCheckFunc(t *testing.T) {
	objects := struct{ Unstructured []*unstructured.Unstructured }{
		Unstructured: []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
				},
			},
			{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "DaemonSet",
				},
			},
		},
	}
	testCases := []struct {
		Name             string
		Funcs            []ChainedCheckFunc
		ShouldThrowError bool
	}{
		{
			Name: "Basic Example",
			Funcs: []ChainedCheckFunc{
				NewChainedCheckFunc(func(tc *TestContext, deployments []*appsv1.Deployment) error {
					assert.NotNil(tc.T, deployments)
					assert.Equal(tc.T, 1, len(deployments))
					Store(tc, "hello", "goodbye")
					return nil
				}),
				NewChainedCheckFunc(func(tc *TestContext, daemonsets []*appsv1.DaemonSet) error {
					assert.NotNil(tc.T, daemonsets)
					assert.Equal(tc.T, 1, len(daemonsets))
					Store(tc, "hello", "world")
					return nil
				}),
				NewChainedCheckFunc(func(tc *TestContext, all []*unstructured.Unstructured) error {
					assert.NotNil(tc.T, all)
					assert.Equal(tc.T, 2, len(all))
					hello, ok := Get[string, string](tc, "hello")
					assert.True(tc.T, ok)
					assert.Equal(tc.T, "world", hello)
					return nil
				}),
			},
		},
		{
			Name: "Throw Error",
			Funcs: []ChainedCheckFunc{
				NewChainedCheckFunc(func(t *TestContext, all []*unstructured.Unstructured) error {
					return fmt.Errorf("should throw error")
				}),
			},
			ShouldThrowError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var args []reflect.Value
			fakeT := &testing.T{}
			if tc.ShouldThrowError {
				args = []reflect.Value{
					reflect.ValueOf(fakeT),
					reflect.ValueOf(objects),
				}
			} else {
				args = []reflect.Value{
					reflect.ValueOf(t),
					reflect.ValueOf(objects),
				}
			}
			checkFunc := NewCheckFunc(tc.Funcs...)
			checkFuncVal := reflect.ValueOf(checkFunc)
			checkFuncVal.Call(args)
			assert.Equal(t, tc.ShouldThrowError, fakeT.Failed())
		})
	}
}
