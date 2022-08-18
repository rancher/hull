package objectset

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/unstructured"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (os ObjectSet) GetAll(t *testing.T, gvk schema.GroupVersionKind) map[objectset.ObjectKey]*unstructured.Unstructured {
	objsByGVK := os.ObjectsByGVK()
	gvkObjs, ok := objsByGVK[gvk]
	if !ok || gvkObjs == nil {
		return nil
	}
	uMap := make(map[objectset.ObjectKey]*unstructured.Unstructured)
	for key, obj := range gvkObjs {
		uObj, err := unstructured.Convert(obj)
		if err != nil {
			t.Error(err)
			continue
		}
		uMap[key] = uObj
	}
	return uMap
}

func (os ObjectSet) Get(t *testing.T, r ObjectSetResource) (unstructured.Unstructured, bool) {
	return os.GetUnstructured(t, r, "")
}

func (os ObjectSet) GetUnstructured(t *testing.T, r ObjectSetResource, path string) (unstructured.Unstructured, bool) {
	var val unstructured.Unstructured
	exists := os.On(t, r, path, func(t testing.T, extractedVal unstructured.Unstructured) {
		val = extractedVal
	})
	return val, exists
}

func (os ObjectSet) GetString(t *testing.T, r ObjectSetResource, path string) (string, bool) {
	var val string
	exists := os.On(t, r, path, func(t testing.T, extractedVal string) {
		val = extractedVal
	})
	return val, exists
}

func (os ObjectSet) GetStringMap(t *testing.T, r ObjectSetResource, path string) (map[string]string, bool) {
	var val map[string]string
	exists := os.On(t, r, path, func(t testing.T, extractedVal map[string]string) {
		val = extractedVal
	})
	return val, exists
}

func (os ObjectSet) GetStringSlice(t *testing.T, r ObjectSetResource, path string) ([]string, bool) {
	var val []string
	exists := os.On(t, r, path, func(t testing.T, extractedVal []string) {
		val = extractedVal
	})
	return val, exists
}

func (os ObjectSet) GetBool(t *testing.T, r ObjectSetResource, path string) (bool, bool) {
	var val bool
	exists := os.On(t, r, path, func(t testing.T, extractedVal bool) {
		val = extractedVal
	})
	return val, exists
}

func (os ObjectSet) GetBoolMap(t *testing.T, r ObjectSetResource, path string) (map[string]bool, bool) {
	var val map[string]bool
	exists := os.On(t, r, path, func(t testing.T, extractedVal map[string]bool) {
		val = extractedVal
	})
	return val, exists
}

func (os ObjectSet) GetBoolSlice(t *testing.T, r ObjectSetResource, path string) ([]bool, bool) {
	var val []bool
	exists := os.On(t, r, path, func(t testing.T, extractedVal []bool) {
		val = extractedVal
	})
	return val, exists
}

func (os ObjectSet) GetInt64(t *testing.T, r ObjectSetResource, path string) (int64, bool) {
	var val int64
	exists := os.On(t, r, path, func(t testing.T, extractedVal int64) {
		val = extractedVal
	})
	return val, exists
}

func (os ObjectSet) GetInt64Map(t *testing.T, r ObjectSetResource, path string) (map[string]int64, bool) {
	var val map[string]int64
	exists := os.On(t, r, path, func(t testing.T, extractedVal map[string]int64) {
		val = extractedVal
	})
	return val, exists
}

func (os ObjectSet) GetInt64Slice(t *testing.T, r ObjectSetResource, path string) ([]int64, bool) {
	var val []int64
	exists := os.On(t, r, path, func(t testing.T, extractedVal []int64) {
		val = extractedVal
	})
	return val, exists
}

func (os ObjectSet) GetFloat64(t *testing.T, r ObjectSetResource, path string) (float64, bool) {
	var val float64
	exists := os.On(t, r, path, func(t testing.T, extractedVal float64) {
		val = extractedVal
	})
	return val, exists
}

func (os ObjectSet) GetFloat64Map(t *testing.T, r ObjectSetResource, path string) (map[string]float64, bool) {
	var val map[string]float64
	exists := os.On(t, r, path, func(t testing.T, extractedVal map[string]float64) {
		val = extractedVal
	})
	return val, exists
}

func (os ObjectSet) GetFloat64Slice(t *testing.T, r ObjectSetResource, path string) ([]float64, bool) {
	var val []float64
	exists := os.On(t, r, path, func(t testing.T, extractedVal []float64) {
		val = extractedVal
	})
	return val, exists
}
