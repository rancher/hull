package internal

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	unstructuredType = reflect.TypeOf(&unstructured.Unstructured{})
	objectInterface  = reflect.TypeOf(new(runtime.Object)).Elem()
)

type DoFunc func(t *testing.T, objs []runtime.Object)

type ParseOptions struct {
	Scheme *runtime.Scheme
	Strict bool
}

func (o *ParseOptions) setDefaults() *ParseOptions {
	if o == nil {
		o = &ParseOptions{}
	}
	if o.Scheme == nil {
		o.Scheme = runtime.NewScheme()
	}
	return o
}

func WrapFunc(fromFunc interface{}, opts *ParseOptions) DoFunc {
	funcType := reflect.TypeOf(fromFunc)
	err := validateFunctionSignature(funcType)
	if err != nil {
		return getBadSignatureFunc(fromFunc, err)
	}
	return func(t *testing.T, objs []runtime.Object) {
		inputType := funcType.In(1)
		objStruct := reflect.New(inputType).Interface()
		if err := parseObjectsIntoStruct(objs, objStruct, opts); err != nil {
			t.Error(err)
			return
		}
		args := []reflect.Value{
			reflect.ValueOf(t),
			reflect.ValueOf(objStruct).Elem(),
		}
		reflect.ValueOf(fromFunc).Call(args)
	}
}

func getBadSignatureFunc(fromFunc interface{}, err error) DoFunc {
	return func(t *testing.T, objs []runtime.Object) {
		t.Errorf("invalid function signature for %s: function signature must match pattern "+
			"func(t *testing.T, objStruct MyStruct), where each field of a provided "+
			"struct must correspond to a slice of pointers to structs that implement "+
			"v1.Object and runtime.Object: %s",
			reflect.TypeOf(fromFunc),
			err,
		)
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
	// First function argument should be a *testing.T
	testingType := funcType.In(0)
	expectedTestingType := reflect.TypeOf((*testing.T)(nil))
	if testingType != expectedTestingType {
		return fmt.Errorf("expected first argument of the function to be of type %s, found %s", expectedTestingType, testingType.Kind())
	}
	// Function argument should be a struct
	inputType := funcType.In(1)
	if inputType.Kind() != reflect.Struct {
		return fmt.Errorf("expected second argument of the function to be a struct, found %s", inputType.Kind())
	}
	// Function should output 0 return values
	if funcType.NumOut() != 0 {
		return fmt.Errorf("expected function that outputs no return values, found %d", funcType.NumOut())
	}
	return nil
}

func parseObjectsIntoStruct(objs []runtime.Object, objectStruct interface{}, opts *ParseOptions) error {
	opts = opts.setDefaults()
	// validate object
	if objectStruct == nil {
		return errors.New("cannot parse objects into nil object")
	}
	objectStructTypePtr := reflect.TypeOf(objectStruct)
	if objectStructTypePtr.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot parse objects into non-pointer type of object, found object of type %s", objectStructTypePtr)
	}
	objectStructType := objectStructTypePtr.Elem()
	supportedTypes, err := getSupportedTypes(objectStructType)
	if err != nil {
		return fmt.Errorf("could not get supported types from object of type %s: %s", objectStructTypePtr, err)
	}

	if objs == nil {
		// nothing to parse
		return nil
	}

	for _, obj := range objs {
		// Identify object type from scheme
		var objType reflect.Type
		gvk := obj.GetObjectKind().GroupVersionKind()
		for kind, reflectType := range opts.Scheme.KnownTypes(gvk.GroupVersion()) {
			if kind == gvk.Kind {
				objType = reflect.PtrTo(reflectType)
				newObj := reflect.New(objType.Elem()).Interface()
				opts.Scheme.Convert(obj, newObj, nil)
				obj = newObj.(runtime.Object)
				obj.GetObjectKind().SetGroupVersionKind(gvk)
				break
			}
		}
		if objType == nil {
			// fall back to provided type
			objType = reflect.TypeOf(obj)
		}
		fieldPath, exists := supportedTypes.getFieldPath(objType)
		if !exists {
			unstructuredFieldPath, exists := supportedTypes.getFieldPath(unstructuredType)
			if !exists {
				if !opts.Strict {
					continue
				}
				return fmt.Errorf("could not unmarshall object of type %s into %s since it was not identified as a supported type %s: %s", reflect.TypeOf(obj), objectStructType, supportedTypes, err)
			}
			fieldPath = unstructuredFieldPath
			uObj, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
			obj = &unstructured.Unstructured{
				Object: uObj,
			}
		}
		fieldNames := strings.Split(fieldPath, ".")
		fieldVal := reflect.ValueOf(objectStruct).Elem()
		for _, fieldName := range fieldNames {
			fieldVal = fieldVal.FieldByName(fieldName)
		}
		fieldVal.Set(reflect.Append(fieldVal, reflect.ValueOf(obj)))
	}
	return nil
}

func getSupportedTypes(objStructType reflect.Type) (*fieldTypeTracker, error) {
	supportedTypes := newFieldTypeTracker()
	return supportedTypes, addSupportedTypes(supportedTypes, objStructType, "")
}

func addSupportedTypes(supportedTypes *fieldTypeTracker, objStructType reflect.Type, path string) error {
	// Each field on the input struct should be a slice of objects that implements runtime.Object
	if objStructType == nil {
		return errors.New("input type cannot be nil")
	}
	if objStructType.Kind() != reflect.Struct {
		return fmt.Errorf("input type must be a struct, found %s", objStructType)
	}

	for i := 0; i < objStructType.NumField(); i++ {
		field := objStructType.Field(i)
		fieldType := field.Type
		fieldPath := field.Name
		if len(path) > 0 {
			fieldPath = path + "." + fieldPath
		}
		// Check if field is an embedded struct that is a slice of objects that implements runtime.Object
		if fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Struct {
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			err := addSupportedTypes(supportedTypes, fieldType, fieldPath)
			if err != nil {
				return fmt.Errorf("cannot parse supported types from pointer type %s at path %s: %s", fieldType, fieldPath, err)
			}
			continue
		}
		// Check if field is a slice
		if fieldType.Kind() != reflect.Slice {
			return fmt.Errorf("field %s must be a slice of structs that implement v1.Object and runtime.Object", field.Name)
		}
		fieldElemType := fieldType.Elem()
		if fieldElemType.Kind() != reflect.Ptr {
			return fmt.Errorf("field %s must be a slice of pointers to structs", field.Name)
		}
		// Each element in the field's type must implement v1.Object and runtime.Object
		if !fieldElemType.Implements(objectInterface) {
			return fmt.Errorf("field %s contains object(s) of type %s that do not implement %s", field.Name, fieldElemType, objectInterface)
		}
		err := supportedTypes.addType(fieldPath, fieldElemType)
		if err != nil {
			return err
		}
	}
	return nil
}
