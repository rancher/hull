package checker

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/rancher/hull/pkg/extract"
	helmChartUtil "helm.sh/helm/v3/pkg/chartutil"
)

func NewContext() *TestContext {
	return &TestContext{
		Data: make(map[interface{}]interface{}),
	}
}

type TestContext struct {
	T *testing.T

	Data map[interface{}]interface{}

	RenderValues helmChartUtil.Values

	continueExecution bool
}

func (tc *TestContext) Continue() {
	tc.continueExecution = true
}

func Store[K comparable, V interface{}](tc *TestContext, key K, value V) {
	tc.Data[key] = value
}

func Get[K comparable, V interface{}](tc *TestContext, key K) (V, bool) {
	value, ok := tc.Data[key]
	if value == nil {
		return *new(V), ok
	}
	return value.(V), ok
}

func RenderValue[O interface{}](tc *TestContext, path string) (O, bool) {
	val, ok := extract.Field[O](tc.RenderValues, path)
	if !ok {
		return val, ok
	}
	rVal := reflect.ValueOf(val)
	if rVal.IsZero() {
		return val, false
	}
	switch rKind := rVal.Kind(); rKind {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Interface:
		if reflect.Indirect(rVal).IsZero() {
			return val, false
		}
	}
	return val, true
}

func MustRenderValue[O interface{}](tc *TestContext, path string) O {
	val, ok := extract.Field[O](tc.RenderValues, path)
	if !ok {
		panic(fmt.Sprintf("cannot extract value at path %s with type %T in values passed into TestContext for %s", path, *new(O), tc.T.Name()))
	}
	return val
}
