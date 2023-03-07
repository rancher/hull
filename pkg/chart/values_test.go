package chart

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeValues(t *testing.T) {
	first := &Values{
		ValueFiles:   []string{"testdata/values.yaml"},
		Values:       []string{"name=prod"},
		StringValues: []string{"value=1234"},
		FileValues:   []string{"myfile=testdata/values.yaml"},
		JSONValues:   []string{"myobj={\"hello\": \"world\"}"},
	}
	second := &Values{
		ValueFiles:   []string{"testdata/values-2.yaml"},
		Values:       []string{"cluster=world"},
		StringValues: []string{"hello=4321"},
		FileValues:   []string{"myscript=testdata/values-2.yaml"},
		JSONValues:   []string{"myobj2={\"hello\": \"rancher\"}"},
	}
	t.Run("Empty Merge", func(t *testing.T) {
		assert.Equal(t, first, first.MergeValues())
	})

	t.Run("Nil Merges Into First", func(t *testing.T) {
		assert.Equal(t, first, first.MergeValues(nil))
	})

	t.Run("First Merges Into Nil", func(t *testing.T) {
		var empty Values
		assert.Equal(t, first, empty.MergeValues(first))
	})

	t.Run("Second Merges Into First", func(t *testing.T) {
		expected := &Values{
			ValueFiles:   []string{"testdata/values.yaml", "testdata/values-2.yaml"},
			Values:       []string{"name=prod", "cluster=world"},
			StringValues: []string{"value=1234", "hello=4321"},
			FileValues:   []string{"myfile=testdata/values.yaml", "myscript=testdata/values-2.yaml"},
			JSONValues:   []string{"myobj={\"hello\": \"world\"}", "myobj2={\"hello\": \"rancher\"}"},
		}
		assert.Equal(t, expected, first.MergeValues(second))
	})
	t.Run("First Merges Into Second", func(t *testing.T) {
		expected := &Values{
			ValueFiles:   []string{"testdata/values-2.yaml", "testdata/values.yaml"},
			Values:       []string{"cluster=world", "name=prod"},
			StringValues: []string{"hello=4321", "value=1234"},
			FileValues:   []string{"myscript=testdata/values-2.yaml", "myfile=testdata/values.yaml"},
			JSONValues:   []string{"myobj2={\"hello\": \"rancher\"}", "myobj={\"hello\": \"world\"}"},
		}
		assert.Equal(t, expected, second.MergeValues(first))
	})
}
