package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	tc := NewContext()
	t.Run("Get something from unset map", func(t *testing.T) {
		val, found := MapGet[string, string, bool](tc, "unsetMap", "unset")
		assert.False(t, found)
		assert.False(t, val)
		assert.Equal(t, map[interface{}]interface{}{}, tc.Data)
	})
	t.Run("Set a map", func(t *testing.T) {
		MapSet(tc, "setMap", "set", true)
		assert.Equal(t, map[interface{}]interface{}{
			"setMap": map[string]bool{
				"set": true,
			},
		}, tc.Data)
	})
	t.Run("Get nonexistent from set map", func(t *testing.T) {
		val, found := MapGet[string, string, bool](tc, "setMap", "unset")
		assert.False(t, found)
		assert.False(t, val)
		assert.Equal(t, map[interface{}]interface{}{
			"setMap": map[string]bool{
				"set": true,
			},
		}, tc.Data)
	})
	t.Run("Get value from set map", func(t *testing.T) {
		val, found := MapGet[string, string, bool](tc, "setMap", "set")
		assert.True(t, found)
		assert.True(t, val)
		assert.Equal(t, map[interface{}]interface{}{
			"setMap": map[string]bool{
				"set": true,
			},
		}, tc.Data)
	})
	t.Run("Set another map", func(t *testing.T) {
		MapSet(tc, "setMap1", "set", true)
		assert.Equal(t, map[interface{}]interface{}{
			"setMap": map[string]bool{
				"set": true,
			},
			"setMap1": map[string]bool{
				"set": true,
			},
		}, tc.Data)
	})
	t.Run("Set another entry in new map", func(t *testing.T) {
		MapSet(tc, "setMap1", "set2", true)
		assert.Equal(t, map[interface{}]interface{}{
			"setMap": map[string]bool{
				"set": true,
			},
			"setMap1": map[string]bool{
				"set":  true,
				"set2": true,
			},
		}, tc.Data)
	})
	t.Run("Iterate through nil map", func(t *testing.T) {
		foundSetMap := map[string]bool{}
		MapFor(tc, "unsetMap", func(k string, v bool) {
			foundSetMap[k] = v
		})
		assert.Equal(t, map[string]bool{}, foundSetMap)
	})
	t.Run("Iterate through map", func(t *testing.T) {
		foundSetMap := map[string]bool{}
		MapFor(tc, "setMap", func(k string, v bool) {
			foundSetMap[k] = v
		})
		assert.Equal(t, map[string]bool{
			"set": true,
		}, foundSetMap)

		foundSetMap = map[string]bool{}
		MapFor(tc, "setMap1", func(k string, v bool) {
			foundSetMap[k] = v
		})
		assert.Equal(t, map[string]bool{
			"set":  true,
			"set2": true,
		}, foundSetMap)
	})
}
