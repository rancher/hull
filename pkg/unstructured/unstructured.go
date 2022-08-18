package unstructured

import (
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Unstructured unstructured.Unstructured

// type onFunc func(t *testing.T, someDataType interface{})

var (
	tPtrStruct         = reflect.TypeOf(new(*testing.T)).Elem()
	unstructuredStruct = reflect.TypeOf(new(Unstructured)).Elem()
	errInterface       = reflect.TypeOf(new(error)).Elem()

	strType     = reflect.TypeOf(new(string)).Elem()
	boolType    = reflect.TypeOf(new(bool)).Elem()
	int64Type   = reflect.TypeOf(new(int64)).Elem()
	float64Type = reflect.TypeOf(new(float64)).Elem()

	validTypes = []reflect.Type{
		unstructuredStruct,
		strType,
		reflect.SliceOf(strType),
		reflect.MapOf(strType, strType),
		boolType,
		reflect.SliceOf(boolType),
		reflect.MapOf(strType, boolType),
		int64Type,
		reflect.SliceOf(int64Type),
		reflect.MapOf(strType, int64Type),
		float64Type,
		reflect.SliceOf(float64Type),
		reflect.MapOf(strType, float64Type),
	}
)

func (u Unstructured) On(t *testing.T, path string, onFunc interface{}) {
	funcType := reflect.TypeOf(onFunc)
	err := validateFunctionSignature(funcType)
	if err != nil {
		t.Fatal(err)
	}

}

func validateFunctionSignature(funcType reflect.Type) error {
	if funcType.Kind() != reflect.Func {
		return fmt.Errorf("expected function, received %v", funcType.Kind())
	}
	// Function should take in two arguments
	if funcType.NumIn() != 2 {
		return fmt.Errorf("expected function that takes in exactly 2 arguments, found %d", funcType.NumIn())
	}
	// First function argument should be a testing.T
	tType := funcType.In(0)
	if tType.Kind() != tPtrStruct.Kind() {
		return fmt.Errorf("expected first argument to be %s, found %s", tPtrStruct, tType.Kind())
	}
	// Second function argument should be a parseable value
	parsedType := funcType.In(1)
	validType := false
	for _, t := range validTypes {
		if parsedType == t {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("expected second argument to be a valid kind (%s), found %s", validTypes, parsedType.Kind())
	}
	// Function should only output 1 argument
	if funcType.NumOut() != 1 {
		return fmt.Errorf("expected function that outputs exactly 1 return value, found %d", funcType.NumOut())
	}
	// Function should output an error
	if !funcType.Out(0).Implements(errInterface) {
		return fmt.Errorf("expected function that outputs an error, found %s", funcType.Out(0).Kind())
	}
	return nil
}
