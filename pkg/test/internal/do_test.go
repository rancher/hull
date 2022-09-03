package internal

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type genericObject interface {
	v1.Object
	runtime.Object
}

func TestConvertObjectToObjectStruct(t *testing.T) {
	testCases := []struct {
		Name                 string
		Objects              []v1.Object
		ExpectedObjectStruct interface{}
		ShouldThrowError     bool
	}{
		{
			Name:                 "No Objects And Nil Struct",
			Objects:              nil,
			ExpectedObjectStruct: nil,
			ShouldThrowError:     false,
		},
		{
			Name:    "No Objects And Valid Struct",
			Objects: nil,
			ExpectedObjectStruct: struct {
				A corev1.ServiceAccount
				B []appsv1.Deployment
				C v1.Object
				D []genericObject
			}{},
			ShouldThrowError: false,
		},
		{
			Name: "Objects And Invalid Struct",
			Objects: []v1.Object{
				&appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "hello",
						Name:      "world",
					},
				},
			},
			ExpectedObjectStruct: nil,
			ShouldThrowError:     false,
		},
		{
			Name: "Single Object With Single Object Field",
			Objects: []v1.Object{
				&appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "hello",
						Name:      "world",
					},
				},
			},
			ExpectedObjectStruct: struct {
				Deployment appsv1.Deployment
			}{
				Deployment: appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "hello",
						Name:      "world",
					},
				},
			},
			ShouldThrowError: false,
		},
		{
			Name: "Single Object With List Object Field",
			Objects: []v1.Object{
				&appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "hello",
						Name:      "world",
					},
				},
			},
			ExpectedObjectStruct: struct {
				Deployment []appsv1.Deployment
			}{
				Deployment: []appsv1.Deployment{
					{
						ObjectMeta: v1.ObjectMeta{
							Namespace: "hello",
							Name:      "world",
						},
					},
				},
			},
			ShouldThrowError: false,
		},
		{
			Name: "Single Object With Interface Field",
			Objects: []v1.Object{
				&appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "hello",
						Name:      "world",
					},
				},
			},
			ExpectedObjectStruct: struct {
				Object genericObject
			}{
				Object: &appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "hello",
						Name:      "world",
					},
				},
			},
			ShouldThrowError: false,
		},
		{
			Name: "Single Object With List Interface Field",
			Objects: []v1.Object{
				&appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "hello",
						Name:      "world",
					},
				},
			},
			ExpectedObjectStruct: struct {
				Object []genericObject
			}{
				Object: []genericObject{
					&appsv1.Deployment{
						ObjectMeta: v1.ObjectMeta{
							Namespace: "hello",
							Name:      "world",
						},
					},
				},
			},
			ShouldThrowError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			objectStructType := reflect.TypeOf(tc.ExpectedObjectStruct)
			supportedTypes, err := parseInputStruct(objectStructType)
			if err != nil {
				t.Error(err)
				return
			}
			objectStructVal, err := convertObjectToObjectStruct(tc.Objects, objectStructType, supportedTypes)
			if err != nil && !tc.ShouldThrowError {
				t.Error(err)
				return
			}
			if tc.ShouldThrowError {
				if err == nil {
					t.Errorf("expected error to be thrown")
				}
				return
			}
			if !objectStructVal.IsValid() {
				assert.Truef(t, tc.ExpectedObjectStruct == nil, "got a nil object back when expected non-nil %s", tc.ExpectedObjectStruct)
				return
			}
			objectStruct := objectStructVal.Interface()
			t.Logf("objectStruct: %s", objectStruct)
			t.Logf("expectedObjectStruct: %s", tc.ExpectedObjectStruct)
			assert.True(t, reflect.DeepEqual(objectStruct, tc.ExpectedObjectStruct), "objects were incorrectly parsed")
		})
	}
}

func TestValidateFuncSignature(t *testing.T) {
	testCases := []struct {
		Name             string
		Func             interface{}
		ShouldThrowError bool
	}{
		{
			Name:             "Nil Func",
			Func:             func() {},
			ShouldThrowError: true,
		},
		{
			Name:             "One Arg",
			Func:             func(struct{}) {},
			ShouldThrowError: true,
		},
		{
			Name:             "Three Args",
			Func:             func(struct{}, struct{}, struct{}) {},
			ShouldThrowError: true,
		},
		{
			Name:             "Two Invalid Args",
			Func:             func(struct{}, struct{}) {},
			ShouldThrowError: true,
		},
		{
			Name:             "Swapped Order",
			Func:             func(struct{}, *testing.T) {},
			ShouldThrowError: true,
		},
		{
			Name:             "Has Return Value",
			Func:             func(*testing.T, struct{}) bool { return true },
			ShouldThrowError: true,
		},
		{
			Name:             "Valid",
			Func:             func(*testing.T, struct{}) {},
			ShouldThrowError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			funcType := reflect.TypeOf(tc.Func)
			err := validateFunctionSignature(funcType)
			if err != nil && !tc.ShouldThrowError {
				t.Error(err)
				return
			}
			if tc.ShouldThrowError {
				if err == nil {
					t.Errorf("expected error to be thrown")
				}
				return
			}
		})
	}
}

func TestParseInputStruct(t *testing.T) {
	testCases := []struct {
		Name               string
		Object             interface{}
		ExpectedTypes      map[reflect.Type]string
		ExpectedInterfaces map[reflect.Type]string
		ShouldThrowError   bool
	}{
		{
			Name:               "Nil Struct",
			Object:             nil,
			ExpectedTypes:      nil,
			ExpectedInterfaces: nil,
			ShouldThrowError:   false,
		},
		{
			Name: "Empty Struct",
			Object: struct {
			}{},
			ExpectedTypes:      nil,
			ExpectedInterfaces: nil,
			ShouldThrowError:   false,
		},
		{
			Name: "Structs",
			Object: struct {
				A corev1.ServiceAccount
				B appsv1.Deployment
				C batchv1.Job
				D rbacv1beta1.ClusterRoleBinding
				E rbacv1.ClusterRoleBinding
				F rbacv1.RoleBinding
			}{},
			ExpectedTypes: map[reflect.Type]string{
				reflect.TypeOf(&corev1.ServiceAccount{}):          "A",
				reflect.TypeOf(&appsv1.Deployment{}):              "B",
				reflect.TypeOf(&batchv1.Job{}):                    "C",
				reflect.TypeOf(&rbacv1beta1.ClusterRoleBinding{}): "D",
				reflect.TypeOf(&rbacv1.ClusterRoleBinding{}):      "E",
				reflect.TypeOf(&rbacv1.RoleBinding{}):             "F",
			},
			ExpectedInterfaces: nil,
			ShouldThrowError:   false,
		},
		{
			Name: "Lists of Structs",
			Object: struct {
				A []corev1.ServiceAccount
				B []appsv1.Deployment
				C []batchv1.Job
				D []rbacv1beta1.ClusterRoleBinding
				E []rbacv1.ClusterRoleBinding
				F []rbacv1.RoleBinding
			}{},
			ExpectedTypes: map[reflect.Type]string{
				reflect.TypeOf(&corev1.ServiceAccount{}):          "A",
				reflect.TypeOf(&appsv1.Deployment{}):              "B",
				reflect.TypeOf(&batchv1.Job{}):                    "C",
				reflect.TypeOf(&rbacv1beta1.ClusterRoleBinding{}): "D",
				reflect.TypeOf(&rbacv1.ClusterRoleBinding{}):      "E",
				reflect.TypeOf(&rbacv1.RoleBinding{}):             "F",
			},
			ExpectedInterfaces: nil,
			ShouldThrowError:   false,
		},
		{
			Name: "Interfaces",
			Object: struct {
				A v1.Object
				B genericObject
			}{},
			ExpectedTypes: nil,
			ExpectedInterfaces: map[reflect.Type]string{
				reflect.TypeOf((*v1.Object)(nil)).Elem():     "A",
				reflect.TypeOf((*genericObject)(nil)).Elem(): "B",
			},
			ShouldThrowError: false,
		},
		{
			Name: "Lists of Interfaces",
			Object: struct {
				A []v1.Object
				B []genericObject
			}{},
			ExpectedTypes: nil,
			ExpectedInterfaces: map[reflect.Type]string{
				reflect.TypeOf((*v1.Object)(nil)).Elem():     "A",
				reflect.TypeOf((*genericObject)(nil)).Elem(): "B",
			},
			ShouldThrowError: false,
		},
		{
			Name: "Resources And Structs",
			Object: struct {
				A corev1.ServiceAccount
				B appsv1.Deployment
				C v1.Object
				D genericObject
			}{},
			ExpectedTypes: map[reflect.Type]string{
				reflect.TypeOf(&corev1.ServiceAccount{}): "A",
				reflect.TypeOf(&appsv1.Deployment{}):     "B",
			},
			ExpectedInterfaces: map[reflect.Type]string{
				reflect.TypeOf((*v1.Object)(nil)).Elem():     "C",
				reflect.TypeOf((*genericObject)(nil)).Elem(): "D",
			},
			ShouldThrowError: false,
		},
		{
			Name: "Lists of Resources And Lists of Structs",
			Object: struct {
				A []corev1.ServiceAccount
				B []appsv1.Deployment
				C []v1.Object
				D []genericObject
			}{},
			ExpectedTypes: map[reflect.Type]string{
				reflect.TypeOf(&corev1.ServiceAccount{}): "A",
				reflect.TypeOf(&appsv1.Deployment{}):     "B",
			},
			ExpectedInterfaces: map[reflect.Type]string{
				reflect.TypeOf((*v1.Object)(nil)).Elem():     "C",
				reflect.TypeOf((*genericObject)(nil)).Elem(): "D",
			},
			ShouldThrowError: false,
		},
		{
			Name: "Mix And Match",
			Object: struct {
				A corev1.ServiceAccount
				B []appsv1.Deployment
				C v1.Object
				D []genericObject
			}{},
			ExpectedTypes: map[reflect.Type]string{
				reflect.TypeOf(&corev1.ServiceAccount{}): "A",
				reflect.TypeOf(&appsv1.Deployment{}):     "B",
			},
			ExpectedInterfaces: map[reflect.Type]string{
				reflect.TypeOf((*v1.Object)(nil)).Elem():     "C",
				reflect.TypeOf((*genericObject)(nil)).Elem(): "D",
			},
			ShouldThrowError: false,
		},
		{
			Name: "Not A V1 Object",
			Object: struct {
				A runtime.Object
			}{},
			ExpectedTypes:      nil,
			ExpectedInterfaces: nil,
			ShouldThrowError:   true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			objType := reflect.TypeOf(tc.Object)
			tracker, err := parseInputStruct(objType)
			if err != nil && !tc.ShouldThrowError {
				t.Error(err)
				return
			}
			if tc.ShouldThrowError {
				if err == nil {
					t.Errorf("expected error to be thrown")
				}
				return
			}

			t.Logf("types: %s", tracker.types)
			t.Logf("expectedTypes: %s", tc.ExpectedTypes)
			t.Logf("interfaces: %s", tracker.interfaces)
			t.Logf("expectedInterfaces: %s", tc.ExpectedInterfaces)
			// Ensure types are the same
			for fieldType, field := range tracker.types {
				assert.Equal(t, tc.ExpectedTypes[fieldType], field)
			}
			assert.Equal(t, len(tracker.types), len(tc.ExpectedTypes))
			// Ensure interfaces are the same
			for fieldType, field := range tracker.interfaces {
				assert.Equal(t, tc.ExpectedInterfaces[fieldType], field)
			}
			assert.Equal(t, len(tracker.interfaces), len(tc.ExpectedInterfaces))
		})
	}
}
