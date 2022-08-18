package objectset

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/unstructured"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ObjectSet struct {
	*objectset.ObjectSet

	Source string
}

func (os ObjectSet) GetSource() string {
	return os.Source
}

func (os ObjectSet) OnAll(t *testing.T, path string, onFunc interface{}) {
	for _, obj := range os.All() {
		os.on(t, obj, path, onFunc)
	}
}

func (os ObjectSet) OnGVK(t *testing.T, gvk schema.GroupVersionKind, path string, onFunc interface{}) bool {
	objsByGVK := os.ObjectsByGVK()
	gvkObjs, ok := objsByGVK[gvk]
	if !ok || gvkObjs == nil {
		return false
	}
	for _, obj := range gvkObjs {
		os.on(t, obj, path, onFunc)
	}
	return true
}

func (os ObjectSet) On(t *testing.T, r ObjectSetResource, path string, onFunc interface{}) bool {
	objsByGVK := os.ObjectsByGVK()
	gvkObjs, ok := objsByGVK[r.GroupVersionKind]
	if !ok || gvkObjs == nil {
		return false
	}
	obj, ok := gvkObjs[r.ObjectKey]
	if !ok || obj == nil {
		return false
	}
	os.on(t, obj, path, onFunc)
	return true
}

func (os ObjectSet) on(t *testing.T, obj runtime.Object, path string, onFunc interface{}) {
	uObj, err := unstructured.Convert(obj)
	if err != nil {
		t.Error(err)
		return
	}
	uObj.On(t, path, onFunc)
}

type ObjectSetResource struct {
	schema.GroupVersionKind
	objectset.ObjectKey
}

func (r ObjectSetResource) Empty() bool {
	return len(r.ObjectKey.String()) == 0 || r.GroupVersionKind.Empty()
}
