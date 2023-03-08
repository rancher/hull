package checker

func MapGet[M comparable, K comparable, V interface{}](tc *TestContext, mapKey M, key K) (V, bool) {
	currMap, ok := Get[M, map[K]V](tc, mapKey)
	if !ok {
		return *new(V), false
	}
	val, ok := currMap[key]
	return val, ok
}

func MapSet[M comparable, K comparable, V interface{}](tc *TestContext, mapKey M, key K, value V) {
	currMap, ok := Get[M, map[K]V](tc, mapKey)
	if !ok {
		currMap = make(map[K]V)
		Store(tc, mapKey, currMap)
	}
	currMap[key] = value
}

func MapFor[M comparable, K comparable, V interface{}](tc *TestContext, mapKey M, doFunc func(K, V)) {
	currMap, exists := Get[M, map[K]V](tc, mapKey)
	if !exists {
		return
	}
	for k, v := range currMap {
		doFunc(k, v)
	}
}
