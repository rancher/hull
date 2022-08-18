package unstructured

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		path        string
		obj         map[string]interface{}
		expected    interface{}
		description string
	}{
		{
			path: "hello",
			obj: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": "hello",
				},
			},
			expected: map[string]interface{}{
				"world": "hello",
			},
			description: "parsing map[string]interface{}",
		},
		{
			path: "hello",
			obj: map[string]interface{}{
				"hello": "world",
			},
			expected:    "world",
			description: "parsing string",
		},
		{
			path: "hello",
			obj: map[string]interface{}{
				"hello": []string{"world"},
			},
			expected:    []string{"world"},
			description: "parsing []string",
		},
		{
			path: "hello",
			obj: map[string]interface{}{
				"hello": map[string]string{
					"world": "hello",
				},
			},
			expected: map[string]string{
				"world": "hello",
			},
			description: "parsing map[string]string",
		},
		{
			path: "hello",
			obj: map[string]interface{}{
				"hello": false,
			},
			expected:    false,
			description: "parsing bool",
		},
		{
			path: "hello",
			obj: map[string]interface{}{
				"hello": []bool{true, false, true},
			},
			expected:    []bool{true, false, true},
			description: "parsing []bool",
		},
		{
			path: "hello",
			obj: map[string]interface{}{
				"hello": map[string]bool{
					"world": false,
				},
			},
			expected: map[string]bool{
				"world": false,
			},
			description: "parsing map[string]bool",
		},
		{
			path: "hello",
			obj: map[string]interface{}{
				"hello": int64(5),
			},
			expected:    int64(5),
			description: "parsing int64",
		},
		{
			path: "hello",
			obj: map[string]interface{}{
				"hello": []int64{5, 6, 7},
			},
			expected:    []int64{5, 6, 7},
			description: "parsing []int64",
		},
		{
			path: "hello",
			obj: map[string]interface{}{
				"hello": map[string]int64{
					"world": 10,
				},
			},
			expected: map[string]int64{
				"world": 10,
			},
			description: "parsing map[string]int64",
		},
		{
			path: "hello",
			obj: map[string]interface{}{
				"hello": 5.1,
			},
			expected:    float64(5.1),
			description: "parsing float64",
		},
		{
			path: "hello",
			obj: map[string]interface{}{
				"hello": []float64{5.1, 6.1, 7.1},
			},
			expected:    []float64{5.1, 6.1, 7.1},
			description: "parsing []float64",
		},
		{
			path: "hello",
			obj: map[string]interface{}{
				"hello": map[string]float64{
					"world": 10.1,
				},
			},
			expected: map[string]float64{
				"world": 10.1,
			},
			description: "parsing map[string]float64",
		},
	}
	for _, tc := range testCases {
		parsedObj, err := Parse(tc.path, Unstructured{Object: tc.obj})
		if err != nil {
			t.Errorf("[path='%s', obj=%s] %s should pass: could not parse path: %s", tc.path, tc.obj, tc.description, err)
			continue
		}
		if !reflect.DeepEqual(parsedObj, tc.expected) {
			t.Errorf("[path='%s', obj=%s] %s should pass: expected '%v', found '%v'", tc.path, tc.obj, tc.description, tc.expected, parsedObj)
			continue
		}
	}
}

func TestIsValidPath(t *testing.T) {
	testCases := []struct {
		path        string
		pass        bool
		description string
	}{
		// Valid Test Cases
		{
			path:        "",
			pass:        true,
			description: "empty string",
		},
		{
			path:        "hello",
			pass:        true,
			description: "accessing key",
		},
		{
			path:        "hello.world",
			pass:        true,
			description: "accessing nested key",
		},
		{
			path:        "hello.world[5]",
			pass:        true,
			description: "accessing list item within nested key",
		},
		{
			path:        "hello.world[_]",
			pass:        true,
			description: "accessing all list items within nested key",
		},
		{
			path:        "hello.world[55]",
			pass:        true,
			description: "accessing list item within nested key with >1 digit",
		},
		{
			path:        "hello.world[1234567890]",
			pass:        true,
			description: "accessing list item within nested key with multiple digits",
		},
		{
			path:        "hello.world[123].one.two.three",
			pass:        true,
			description: "accessing nested keys within list item with multiple digits",
		},
		{
			path:        "hello.world[_].one.two.three",
			pass:        true,
			description: "accessing nested keys within all list items",
		},
		{
			path:        "hello.world[123].one[4].two[5].three",
			pass:        true,
			description: "accessing nested list items within nested keys ending with nested key",
		},
		{
			path:        "hello.world[123].one[4].two[5].three[6]",
			pass:        true,
			description: "accessing nested list items within nested keys ending with list item",
		},
		{
			path:        "hello.world[123].one[4].two[5].three[6][12][18]",
			pass:        true,
			description: "accessing nested list items within list items ending with list item",
		},
		{
			path:        "hello.world[123].one[4].two[5].three[6][12][18].four",
			pass:        true,
			description: "accessing nested list items within list items ending with nested key",
		},
		// Invalid Test Cases
		{
			path:        "hello.",
			pass:        false,
			description: "ends with period",
		},
		{
			path:        "hello.world[",
			pass:        false,
			description: "ends with open bracket",
		},
		{
			path:        "hello.world[5",
			pass:        false,
			description: "ends with character but does not close bracket",
		},
		{
			path:        "hello.world[__]",
			pass:        false,
			description: "invalid list index using multiple underscores",
		},
		{
			path:        "hello.world[abc]",
			pass:        false,
			description: "invalid list index using alphabet",
		},
		{
			path:        "hello.world[023].one.two.three",
			pass:        false,
			description: "invalid list index using leading zero",
		},
		{
			path:        "hello.world[123].one[4].two[5].three[6",
			pass:        false,
			description: "invalid list index using multiple underscores in nested key",
		},
	}
	for _, tc := range testCases {
		if isValidPath(tc.path) != tc.pass {
			if tc.pass {
				t.Errorf("[path='%s'] %s should pass", tc.path, tc.description)
			} else {
				t.Errorf("[path='%s'] %s should fail", tc.path, tc.description)
			}
		}
	}
}
