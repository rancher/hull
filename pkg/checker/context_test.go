package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestContext(t *testing.T) {
	tc := NewContext()
	t.Run("Get Something Unset", func(t *testing.T) {
		_, found := Get[string, bool](tc, "unset")
		assert.False(t, found)
	})
	Store(tc, "set", true)
	t.Run("Get Something Set", func(t *testing.T) {
		set, found := Get[string, bool](tc, "set")
		assert.True(t, found)
		assert.True(t, set)
	})
	Store(tc, "set", false)
	t.Run("Get Something Else Set", func(t *testing.T) {
		set, found := Get[string, bool](tc, "set")
		assert.True(t, found)
		assert.False(t, set)
	})
	t.Run("Get Unset Map", func(t *testing.T) {
		_, found := Get[string, map[string]interface{}](tc, "nilMap")
		assert.False(t, found)
	})
	Store[string, map[string]interface{}](tc, "nilMap", nil)
	t.Run("Get Map Set Nil", func(t *testing.T) {
		set, found := Get[string, map[string]interface{}](tc, "nilMap")
		assert.True(t, found)
		assert.Nil(t, set)
	})
}
