package unstructured

import "testing"

func (u Unstructured) GetUnstructured(t *testing.T, path string) Unstructured {
	var val Unstructured
	u.On(t, path, func(t testing.T, extractedVal Unstructured) {
		val = extractedVal
	})
	return val
}

func (u Unstructured) GetString(t *testing.T, path string) string {
	var val string
	u.On(t, path, func(t testing.T, extractedVal string) {
		val = extractedVal
	})
	return val
}

func (u Unstructured) GetStringMap(t *testing.T, path string) map[string]string {
	var val map[string]string
	u.On(t, path, func(t testing.T, extractedVal map[string]string) {
		val = extractedVal
	})
	return val
}

func (u Unstructured) GetStringSlice(t *testing.T, path string) []string {
	var val []string
	u.On(t, path, func(t testing.T, extractedVal []string) {
		val = extractedVal
	})
	return val
}

func (u Unstructured) GetBool(t *testing.T, path string) bool {
	var val bool
	u.On(t, path, func(t testing.T, extractedVal bool) {
		val = extractedVal
	})
	return val
}

func (u Unstructured) GetBoolMap(t *testing.T, path string) map[string]bool {
	var val map[string]bool
	u.On(t, path, func(t testing.T, extractedVal map[string]bool) {
		val = extractedVal
	})
	return val
}

func (u Unstructured) GetBoolSlice(t *testing.T, path string) []bool {
	var val []bool
	u.On(t, path, func(t testing.T, extractedVal []bool) {
		val = extractedVal
	})
	return val
}

func (u Unstructured) GetInt64(t *testing.T, path string) int64 {
	var val int64
	u.On(t, path, func(t testing.T, extractedVal int64) {
		val = extractedVal
	})
	return val
}

func (u Unstructured) GetInt64Map(t *testing.T, path string) map[string]int64 {
	var val map[string]int64
	u.On(t, path, func(t testing.T, extractedVal map[string]int64) {
		val = extractedVal
	})
	return val
}

func (u Unstructured) GetInt64Slice(t *testing.T, path string) []int64 {
	var val []int64
	u.On(t, path, func(t testing.T, extractedVal []int64) {
		val = extractedVal
	})
	return val
}

func (u Unstructured) GetFloat64(t *testing.T, path string) float64 {
	var val float64
	u.On(t, path, func(t testing.T, extractedVal float64) {
		val = extractedVal
	})
	return val
}

func (u Unstructured) GetFloat64Map(t *testing.T, path string) map[string]float64 {
	var val map[string]float64
	u.On(t, path, func(t testing.T, extractedVal map[string]float64) {
		val = extractedVal
	})
	return val
}

func (u Unstructured) GetFloat64Slice(t *testing.T, path string) []float64 {
	var val []float64
	u.On(t, path, func(t testing.T, extractedVal []float64) {
		val = extractedVal
	})
	return val
}
