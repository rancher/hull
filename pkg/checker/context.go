package checker

import "testing"

func NewContext() *TestContext {
	return &TestContext{
		Data: make(map[interface{}]interface{}),
	}
}

type TestContext struct {
	T *testing.T

	Data map[interface{}]interface{}
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
