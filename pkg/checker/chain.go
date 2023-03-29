package checker

import (
	"testing"

	"github.com/rancher/hull/pkg/checker/internal"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewCheckFunc(funcs ...ChainedCheckFunc) CheckFunc {
	return func(t *testing.T, u struct{ Unstructured []*unstructured.Unstructured }) {
		tc := NewContext()
		tc.T = t
		for _, f := range funcs {
			checkFunc := f(tc)
			if checkFunc == nil {
				if !tc.continueExecution && tc.T.Failed() {
					break
				}
				tc.continueExecution = false
				continue
			}
			doFunc := internal.WrapFunc(checkFunc, &internal.ParseOptions{
				Scheme: Scheme,
			})
			objs := make([]runtime.Object, len(u.Unstructured))
			for i, unstructured := range u.Unstructured {
				objs[i] = unstructured
			}
			doFunc(tc.T, objs)
			if !tc.continueExecution && tc.T.Failed() {
				break
			}
			tc.continueExecution = false
		}
	}
}

type ChainedCheckFunc func(t *TestContext) CheckFunc

func NewChainedCheckFunc[S interface{}](checkFuncWithContext func(tc *TestContext, objStruct S)) ChainedCheckFunc {
	return func(tc *TestContext) CheckFunc {
		if checkFuncWithContext == nil {
			return nil
		}
		return func(t *testing.T, objStruct S) {
			tc.T = t
			checkFuncWithContext(tc, objStruct)
		}
	}
}
