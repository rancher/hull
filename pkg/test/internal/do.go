package internal

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	invalidFuncSignatureWarning = "Function signature must match pattern func(t *testing.T, objStruct MyStruct)"
	invalidStructTypeWarning    = "Each field of a provided struct must correspond to an object or slice of objects that implement v1.Object. " +
		"It is also expected that no two fields the struct correspond to the same underlying type."
)

var (
	objectInterface = reflect.TypeOf(new(v1.Object)).Elem()
)

type DoFunc func(testing *testing.T, objs []v1.Object)

func WrapFunc(fromFunc interface{}) DoFunc {
	caller := reflect.ValueOf(fromFunc)
	funcName := runtime.FuncForPC(caller.Pointer()).Name()
	funcType := reflect.TypeOf(fromFunc)

	err := validateFunctionSignature(funcType)
	if err != nil {
		logrus.Errorf(invalidFuncSignatureWarning)
		logrus.WithField("doFunc", funcName).Fatalf("could not wrap doFunc: %s", err)
	}

	inputType := funcType.In(1)
	supportedTypes, err := parseInputStruct(inputType)
	if err != nil {
		logrus.Errorf(invalidStructTypeWarning)
		logrus.WithField("doFunc", funcName).Fatalf("could not wrap doFunc: %s", err)
	}

	return func(t *testing.T, objs []v1.Object) {
		in := reflect.New(inputType)

		singletonFieldToObjs := map[reflect.Value][]v1.Object{}
		for _, obj := range objs {
			fieldName, err := supportedTypes.getField(obj)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"resource":       fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()),
					"supportedTypes": supportedTypes,
				}).Fatalf("Could not unmarshall %s into %s: %s", reflect.TypeOf(obj), inputType, err)
			}
			field := in.Elem().FieldByName(fieldName)
			if field.Kind() != reflect.Slice {
				// Singleton fields will be added in a separate loop to print violaters
				singletonFieldToObjs[field] = append(singletonFieldToObjs[field], obj)
				continue
			}
			objValue := reflect.ValueOf(obj)
			if field.Type().Elem().Kind() != reflect.Interface {
				objValue = objValue.Elem()
			}
			field.Set(reflect.Append(field, objValue))
		}

		// Ensure that you only find one resource if the struct is expecting a single resource
		foundMultipleSingletons := false
		for field, objs := range singletonFieldToObjs {
			if len(objs) > 0 {
				resources := make([]string, len(objs))
				for i, obj := range objs {
					resources[i] = fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName())
				}
				logrus.
					WithField("resources", resources).
					Errorf("Expected 1 resource of type %s, found %d", field.Type(), len(objs))
				foundMultipleSingletons = true
				continue
			}
			objValue := reflect.ValueOf(objs[0])
			if field.Type().Elem().Kind() != reflect.Interface {
				objValue = objValue.Elem()
			}
			field.Set(objValue)
		}

		if foundMultipleSingletons {
			logrus.
				WithField("doFunc", funcName).
				Fatalf("Failed to unmarshall objects into %s", inputType)
		}

		// Call the function with the provided *testing.T
		args := []reflect.Value{reflect.ValueOf(t), in.Elem()}
		caller.Call(args)
	}
}

func convertObjectToObjectStruct(objs []v1.Object, objectStructType reflect.Type, supportedTypes fieldTypeTracker) (reflect.Value, error) {
	if objectStructType == nil {
		return reflect.Value{}, nil
	}
	objectStruct := reflect.New(objectStructType)
	if objs == nil {
		return objectStruct.Elem(), nil
	}
	singletonFieldToObjs := map[reflect.Value][]v1.Object{}
	for _, obj := range objs {
		fieldName, err := supportedTypes.getField(obj)
		if err != nil {
			return objectStruct.Elem(), fmt.Errorf("could not unmarshall %s (%s/%s) into %s since it was not identified as a supported type %s: %s", reflect.TypeOf(obj), obj.GetNamespace(), obj.GetName(), objectStructType, supportedTypes, err)
		}
		field := objectStruct.Elem().FieldByName(fieldName)
		if field.Kind() != reflect.Slice {
			// Singleton fields will be added in a separate loop to print violaters
			singletonFieldToObjs[field] = append(singletonFieldToObjs[field], obj)
			continue
		}
		objValue := reflect.ValueOf(obj)
		if field.Type().Elem().Kind() != reflect.Interface {
			objValue = objValue.Elem()
		}
		field.Set(reflect.Append(field, objValue))
	}

	// Ensure that you only find one resource if the struct is expecting a single resource
	for field, objs := range singletonFieldToObjs {
		if len(objs) > 1 {
			resources := make([]string, len(objs))
			for i, obj := range objs {
				resources[i] = fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName())
			}
			return objectStruct.Elem(), fmt.Errorf("failed to unmarshall objects into %s: expected 1 resource of type %s, found %d: %s", objectStructType, field.Type(), len(objs), resources)
		}
		objValue := reflect.ValueOf(objs[0])
		if field.Type().Kind() != reflect.Interface {
			objValue = objValue.Elem()
		}
		field.Set(objValue)
	}
	return objectStruct.Elem(), nil
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

func parseInputStruct(inputType reflect.Type) (fieldTypeTracker, error) {
	// Each field on the input struct should either implement v1.Object or be a slice of objects that implement v1.Object
	supportedResources := fieldTypeTracker{
		types:      map[reflect.Type]string{},
		interfaces: map[reflect.Type]string{},
	}
	if inputType == nil {
		return supportedResources, nil
	}
	for i := 0; i < inputType.NumField(); i++ {
		field := inputType.Field(i)
		fieldType := field.Type
		// Check if field is a struct or a slice
		var fieldPointerType reflect.Type
		var isInterface bool
		switch fieldType.Kind() {
		case reflect.Slice:
			fieldElemType := fieldType.Elem()
			switch fieldElemType.Kind() {
			case reflect.Struct:
				fieldPointerType = reflect.PtrTo(fieldElemType)
			case reflect.Interface:
				fieldPointerType = fieldElemType
				isInterface = true
			default:
				return supportedResources, fmt.Errorf("field %s must be a slice of structs", field.Name)
			}
		case reflect.Struct:
			fieldPointerType = reflect.PtrTo(fieldType)
		case reflect.Interface:
			fieldPointerType = fieldType
			isInterface = true
		default:
			return supportedResources, fmt.Errorf("field %s must be a struct or slice", field.Name)
		}
		if isInterface {
			err := supportedResources.addInterface(fieldPointerType, field.Name)
			if err != nil {
				return supportedResources, err
			}
		} else {
			err := supportedResources.addType(fieldPointerType, field.Name)
			if err != nil {
				return supportedResources, err
			}
		}
		// Field elem must implement v1.Object
		if !fieldPointerType.Implements(objectInterface) {
			return supportedResources, fmt.Errorf("field %s contains object(s) of type %s that do not implement %s", field.Name, fieldPointerType, objectInterface)
		}
	}
	return supportedResources, nil
}
