package checker

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/checker/internal"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewCheckFunc(funcs ...ChainedCheckFunc) CheckFunc {
	return func(t *testing.T, u struct{ Unstructured []*unstructured.Unstructured }) {
		tc := NewContext()
		for _, f := range funcs {
			checkFunc := f(tc)
			doFunc := internal.WrapFunc(checkFunc, &internal.ParseOptions{
				Scheme: Scheme,
			})
			objs := make([]runtime.Object, len(u.Unstructured))
			for i, unstructured := range u.Unstructured {
				objs[i] = unstructured
			}
			doFunc(t, objs)
		}
	}
}

type ChainedCheckFunc func(t *TestContext) CheckFunc

func NewChainedCheckFunc[O runtime.Object](typedCheckFunc func(t *TestContext, objects []O) error) ChainedCheckFunc {
	return func(tc *TestContext) CheckFunc {
		return func(t *testing.T, objs struct{ Objects []O }) {
			tc.T = t
			if err := typedCheckFunc(tc, objs.Objects); err != nil {
				t.Error(err)
			}
		}
	}
}
