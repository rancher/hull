package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	helmChart "helm.sh/helm/v3/pkg/chart"
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
	t.Run("RenderValue When Unset", func(t *testing.T) {
		value, found := RenderValue[string](tc, ".Chart.Name")
		assert.False(t, found)
		assert.Equal(t, "", value)
	})
	t.Run("MustRenderValue When Unset", func(t *testing.T) {
		defer func() {
			err := recover()
			if err == nil {
				assert.Fail(t, "should not have passed MustRenderValue")
			}
		}()
		MustRenderValue[string](tc, ".Chart.Name")
	})
	tc.RenderValues = map[string]interface{}{
		"Chart": helmChart.Metadata{
			Name: "my-chart",
		},
	}
	tc.RenderValues = map[string]interface{}{
		"Chart": helmChart.Metadata{
			Name: "my-chart",
		},
	}
	t.Run("RenderValue When Set Struct", func(t *testing.T) {
		value, found := RenderValue[string](tc, ".Chart.Name")
		assert.True(t, found)
		assert.Equal(t, "my-chart", value)
	})
	tc.RenderValues = map[string]interface{}{
		"Chart": helmChart.Metadata{
			Name: "my-chart-2",
		},
		"Values": map[string]interface{}{
			"data": map[string]interface{}{
				"hello": "world",
			},
		},
	}
	t.Run("MustRenderValue When Set Struct", func(t *testing.T) {
		defer func() {
			err := recover()
			if err != nil {
				assert.Fail(t, "should have passed MustRenderValue", err)
			}
		}()
		MustRenderValue[string](tc, ".Chart.Name")
	})
	t.Run("RenderValue When Updated", func(t *testing.T) {
		value, found := RenderValue[string](tc, ".Chart.Name")
		assert.True(t, found)
		assert.Equal(t, "my-chart-2", value)
	})
	t.Run("RenderValue With Set Map", func(t *testing.T) {
		value, found := RenderValue[map[string]interface{}](tc, ".Values.data")
		assert.True(t, found)
		assert.Equal(t, map[string]interface{}{"hello": "world"}, value)
	})
	t.Run("Convert RenderValue Type To String Map", func(t *testing.T) {
		value, found := RenderValue[map[string]string](tc, ".Values.data")
		assert.True(t, found)
		assert.Equal(t, map[string]string{"hello": "world"}, value)
	})
	t.Run("Convert RenderValue Type To Struct", func(t *testing.T) {
		type HelloWorld struct {
			Hello string `json:"hello"`
		}
		value, found := RenderValue[HelloWorld](tc, ".Values.data")
		assert.True(t, found)
		assert.Equal(t, HelloWorld{Hello: "world"}, value)
	})
}
