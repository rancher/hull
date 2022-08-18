package unstructured

import (
	"reflect"
	"testing"
)

func TestFunctionSignature(t *testing.T) {
	testCases := []struct {
		function    interface{}
		pass        bool
		description string
	}{
		{
			function: func(*testing.T, string) error { return nil },
			pass:     true,
		},
		{
			function: func(*testing.T, map[string]string) error { return nil },
			pass:     true,
		},
		{
			function: func(*testing.T, []string) error { return nil },
			pass:     true,
		},
		{
			function: func(*testing.T, bool) error { return nil },
			pass:     true,
		},
		{
			function: func(*testing.T, map[string]bool) error { return nil },
			pass:     true,
		},
		{
			function: func(*testing.T, []bool) error { return nil },
			pass:     true,
		},
		{
			function: func(*testing.T, int64) error { return nil },
			pass:     true,
		},
		{
			function: func(*testing.T, map[string]int64) error { return nil },
			pass:     true,
		},
		{
			function: func(*testing.T, []int64) error { return nil },
			pass:     true,
		},
		{
			function: func(*testing.T, float64) error { return nil },
			pass:     true,
		},
		{
			function: func(*testing.T, map[string]float64) error { return nil },
			pass:     true,
		},
		{
			function: func(*testing.T, []float64) error { return nil },
			pass:     true,
		},
		{
			function: func(*testing.T, Unstructured) error { return nil },
			pass:     true,
		},
		{
			function: func(*testing.T, int) error { return nil },
			pass:     false,
		},
		{
			function: func(*testing.T, int16) error { return nil },
			pass:     false,
		},
		{
			function: func(*testing.T, int32) error { return nil },
			pass:     false,
		},
		{
			function: func(*testing.T, float32) error { return nil },
			pass:     false,
		},
		{
			function: func(testing.T, string) error { return nil },
			pass:     false,
		},
		{
			function: func(*testing.T, string) bool { return false },
			pass:     false,
		},
	}

	for _, tc := range testCases {
		funcType := reflect.TypeOf(tc.function)
		err := validateFunctionSignature(funcType)
		if tc.pass && err != nil {
			t.Errorf("failed validation on function with signature %s: %s", funcType, err)
		}
		if !tc.pass && err == nil {
			t.Errorf("accidentally passed on function with signature %s", funcType)
		}
	}
}
