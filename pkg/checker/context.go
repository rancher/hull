package checker

import (
	"fmt"
	"testing"

	"github.com/aiyengar2/hull/pkg/extract"
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
	return extract.Field[O](tc.RenderValues, path)
}

func MustRenderValue[O interface{}](tc *TestContext, path string) O {
	val, ok := extract.Field[O](tc.RenderValues, path)
	if !ok {
		panic(fmt.Sprintf("cannot extract value at path %s with type %T in values passed into TestContext for %s", path, *new(O), tc.T.Name()))
	}
	return val
}
