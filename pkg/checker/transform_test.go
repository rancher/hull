package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToYAML(t *testing.T) {
	testCases := []struct {
		Name   string
		Obj    interface{}
		Expect string
	}{
		{
			Name:   "Empty",
			Obj:    nil,
			Expect: "null",
		},
		{
			Name: "Some Object",
			Obj: map[string]interface{}{
				"hello": []int{1, 2, 3, 4},
			},
			Expect: "hello:" + "\n" +
				"- 1" + "\n" +
				"- 2" + "\n" +
				"- 3" + "\n" +
				"- 4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Expect, ToYAML(tc.Obj))
		})
	}
}

func TestToJSON(t *testing.T) {
	testCases := []struct {
		Name   string
		Obj    interface{}
		Expect string
	}{
		{
			Name:   "Empty",
			Obj:    nil,
			Expect: "null",
		},
		{
			Name: "Some Object",
			Obj: map[string]interface{}{
				"hello": []int{1, 2, 3, 4},
			},
			Expect: `{"hello":[1,2,3,4]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Expect, ToJSON(tc.Obj))
		})
	}
}

func TestFromYAML(t *testing.T) {
	testCases := []struct {
		Name   string
		YAML   string
		Expect interface{}
	}{
		{
			Name:   "Empty",
			YAML:   "",
			Expect: nil,
		},
		{
			Name:   "Null",
			YAML:   "null",
			Expect: nil,
		},
		{
			Name:   "Invalid",
			YAML:   "*",
			Expect: nil,
		},
		{
			Name: "Some Object",
			YAML: "hello:" + "\n" +
				"- 1" + "\n" +
				"- 2" + "\n" +
				"- 3" + "\n" +
				"- 4",
			Expect: map[interface{}]interface{}{
				"hello": []interface{}{1, 2, 3, 4},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Expect == nil {
				assert.Zero(t, FromYAML[map[interface{}]interface{}](tc.YAML))
			} else {
				assert.Equal(t, tc.Expect, FromYAML[map[interface{}]interface{}](tc.YAML))
			}
		})
	}
}

func TestFromJSON(t *testing.T) {
	testCases := []struct {
		Name             string
		JSON             string
		Expect           interface{}
		ShouldThrowError bool
	}{
		{
			Name:   "Empty",
			JSON:   "",
			Expect: nil,
		},
		{
			Name:   "Null",
			JSON:   "null",
			Expect: nil,
		},
		{
			Name: "Some Object",
			JSON: `{"hello": [1,2,3,4]}`,
			Expect: map[string]interface{}{
				"hello": []interface{}{1.0, 2.0, 3.0, 4.0},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Expect == nil {
				assert.Zero(t, FromJSON[map[string]interface{}](tc.JSON))
			} else {
				assert.Equal(t, tc.Expect, FromJSON[map[string]interface{}](tc.JSON))
			}
		})
	}
}
