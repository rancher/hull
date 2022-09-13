package internal

import (
	"fmt"
	"reflect"
)

type fieldTypeTracker struct {
	typeToField map[reflect.Type]string
	fieldToType map[string]reflect.Type
}

func newFieldTypeTracker() *fieldTypeTracker {
	return &fieldTypeTracker{
		typeToField: make(map[reflect.Type]string),
		fieldToType: make(map[string]reflect.Type),
	}
}

func (r *fieldTypeTracker) addType(fieldName string, fieldType reflect.Type) error {
	if currFieldName, exists := r.typeToField[fieldType]; exists && currFieldName != fieldName {
		return fmt.Errorf("field %s and %s track the same object type %s", currFieldName, fieldName, fieldType)
	}
	if currType, exists := r.fieldToType[fieldName]; exists && currType != fieldType {
		return fmt.Errorf("field %s is already tracking %s, cannot also track %s", fieldName, currType, fieldType)
	}
	r.typeToField[fieldType] = fieldName
	r.fieldToType[fieldName] = fieldType
	return nil
}

func (r *fieldTypeTracker) getFieldPath(fieldType reflect.Type) (string, bool) {
	for supportedType, field := range r.typeToField {
		if fieldType == supportedType {
			return field, true
		}
	}
	return "", false
}

func (r fieldTypeTracker) String() string {
	return fmt.Sprintf("{typesToField: %s, fieldToType: %s}", r.typeToField, r.fieldToType)
}
