package internal

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
)

func TestFieldTrackerAddType(t *testing.T) {
	type fieldArgs struct {
		Name string
		Type reflect.Type
	}

	testCases := []struct {
		Name              string
		Fields            []fieldArgs
		NumExpectedFields int
		ShouldThrowError  bool
	}{
		{
			Name: "Nil Field",
			Fields: []fieldArgs{
				{
					Name: "A",
					Type: nil,
				},
			},
			NumExpectedFields: 1,
			ShouldThrowError:  false,
		},
		{
			Name: "Add One Type Pointing To One Field",
			Fields: []fieldArgs{
				{
					Name: "A",
					Type: reflect.TypeOf(true),
				},
			},
			NumExpectedFields: 1,
			ShouldThrowError:  false,
		},
		{
			Name: "Add Different Types Pointing To Different Fields",
			Fields: []fieldArgs{
				{
					Name: "A",
					Type: reflect.TypeOf(true),
				},
				{
					Name: "B",
					Type: reflect.TypeOf(1),
				},
			},
			NumExpectedFields: 2,
			ShouldThrowError:  false,
		},
		{
			Name: "Add Different Types Pointing To Same Fields",
			Fields: []fieldArgs{
				{
					Name: "A",
					Type: reflect.TypeOf(true),
				},
				{
					Name: "A",
					Type: reflect.TypeOf(10),
				},
			},
			NumExpectedFields: 2,
			ShouldThrowError:  true,
		},
		{
			Name: "Add Same Types Pointing To Different Fields",
			Fields: []fieldArgs{
				{
					Name: "A",
					Type: reflect.TypeOf(true),
				},
				{
					Name: "B",
					Type: reflect.TypeOf(true),
				},
			},
			NumExpectedFields: 2,
			ShouldThrowError:  true,
		},
		{
			Name: "Add Same Types Pointing To Same Fields",
			Fields: []fieldArgs{
				{
					Name: "A",
					Type: reflect.TypeOf(true),
				},
				{
					Name: "A",
					Type: reflect.TypeOf(true),
				},
			},
			NumExpectedFields: 1,
			ShouldThrowError:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tracker := newFieldTypeTracker()
			var multiErr error
			for _, field := range tc.Fields {
				err := tracker.addType(field.Name, field.Type)
				if err != nil {
					multiErr = multierror.Append(multiErr, err)
				}
			}
			t.Logf("tracker: %s", tracker)
			if multiErr == nil {
				if tc.ShouldThrowError {
					t.Errorf("expected error to be thrown")
				}
				assert.Equal(t, len(tracker.typeToField), tc.NumExpectedFields, "fieldTypeTracker added too many fields")
			} else if !tc.ShouldThrowError {
				t.Error(multiErr)
				return
			}
		})
	}
}

func TestFieldTrackerGetFieldPath(t *testing.T) {
	type fieldArgs struct {
		Name string
		Type reflect.Type
	}

	testCases := []struct {
		Name          string
		FieldsToAdd   []fieldArgs
		FieldsToCheck []fieldArgs
	}{
		{
			Name: "Get One Type Pointing To One Field",
			FieldsToAdd: []fieldArgs{
				{
					Name: "A",
					Type: reflect.TypeOf(true),
				},
			},
			FieldsToCheck: []fieldArgs{
				{
					Name: "A",
					Type: reflect.TypeOf(true),
				},
			},
		},
		{
			Name: "Get Different Types Pointing To Different Fields",
			FieldsToAdd: []fieldArgs{
				{
					Name: "A",
					Type: reflect.TypeOf(true),
				},
				{
					Name: "B",
					Type: reflect.TypeOf(1),
				},
			},
			FieldsToCheck: []fieldArgs{
				{
					Name: "A",
					Type: reflect.TypeOf(true),
				},
				{
					Name: "B",
					Type: reflect.TypeOf(1),
				},
			},
		},
		{
			Name: "Get Nonexistent Type",
			FieldsToAdd: []fieldArgs{
				{
					Name: "A",
					Type: reflect.TypeOf(true),
				},
			},
			FieldsToCheck: []fieldArgs{
				{
					Name: "",
					Type: reflect.TypeOf(1),
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tracker := newFieldTypeTracker()
			for _, field := range tc.FieldsToAdd {
				err := tracker.addType(field.Name, field.Type)
				if err != nil {
					t.Error(err)
					return
				}
			}

			t.Logf("tracker: %s", tracker)

			for _, field := range tc.FieldsToCheck {
				fieldName, exists := tracker.getFieldPath(field.Type)
				assert.Equal(t, fieldName, field.Name)
				if !exists {
					assert.Equal(t, fieldName, "")
				} else {
					assert.NotEqual(t, fieldName, "")
				}
			}
		})
	}
}
