package internal

import (
	"fmt"
	"reflect"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fieldTypeTracker struct {
	types      map[reflect.Type]string
	interfaces map[reflect.Type]string
}

func (r *fieldTypeTracker) addType(typeToAdd reflect.Type, fieldName string) error {
	if currFieldName, exists := r.types[typeToAdd]; exists {
		return fmt.Errorf("field %s and %s track the same object type %s", currFieldName, fieldName, typeToAdd)
	}
	r.types[typeToAdd] = fieldName
	return nil
}

func (r *fieldTypeTracker) addInterface(interfaceToAdd reflect.Type, fieldName string) error {
	if currFieldName, exists := r.types[interfaceToAdd]; exists {
		return fmt.Errorf("field %s and %s track the same object interface %s", currFieldName, fieldName, interfaceToAdd)
	}
	r.interfaces[interfaceToAdd] = fieldName
	return nil
}

func (r *fieldTypeTracker) getField(obj v1.Object) (fieldName string, err error) {
	objType := reflect.TypeOf(obj)
	for supportedType, field := range r.types {
		if objType == supportedType {
			return field, nil
		}
	}
	implementsFields := []string{}
	for supportedInterface, field := range r.interfaces {
		if objType.Implements(supportedInterface) {
			implementsFields = append(implementsFields, field)
		}
	}
	if len(implementsFields) == 0 {
		return "", fmt.Errorf("no existing fields support %s", objType)
	}
	if len(implementsFields) > 1 {
		return "", fmt.Errorf("placement of %s is ambiguous, can be marshalled into multiple interface fields: %s", objType, implementsFields)
	}
	return implementsFields[0], nil
}

func (r fieldTypeTracker) String() string {
	supportedTypes := make([]reflect.Type, len(r.types))
	i := 0
	for supportedType := range r.types {
		supportedTypes[i] = supportedType
		i++
	}
	supportedInterfaces := make([]reflect.Type, len(r.interfaces))
	i = 0
	for supportedInterface := range r.interfaces {
		supportedInterfaces[i] = supportedInterface
		i++
	}
	return fmt.Sprintf("{types: %s, interfaces: %s}", supportedTypes, supportedInterfaces)
}
