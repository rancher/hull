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
				NewChainedCheckFunc(func(t *TestContext, deployments []*appsv1.Deployment) error {
					assert.NotNil(t.T, deployments)
					assert.Equal(t.T, 1, len(deployments))
					t.Store("hello", "goodbye")
					return nil
				}),
				NewChainedCheckFunc(func(t *TestContext, daemonsets []*appsv1.DaemonSet) error {
					assert.NotNil(t.T, daemonsets)
					assert.Equal(t.T, 1, len(daemonsets))
					t.Store("hello", "world")
					return nil
				}),
				NewChainedCheckFunc(func(t *TestContext, all []*unstructured.Unstructured) error {
					assert.NotNil(t.T, all)
					assert.Equal(t.T, 2, len(all))
					hello, ok := t.Get("hello")
					assert.True(t.T, ok)
					assert.Equal(t.T, "world", hello)
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
