package unstructured

import (
	unstructured "github.com/rancher/wrangler/pkg/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func Convert(obj runtime.Object) (*Unstructured, error) {
	uObj, err := unstructured.ToUnstructured(obj)
	return (*Unstructured)(uObj), err
}
