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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	defaultScheme = runtime.NewScheme()
)

func init() {
	err := appsv1.AddToScheme(defaultScheme)
	if err != nil {
		panic(err)
	}
	err = batchv1.AddToScheme(defaultScheme)
	if err != nil {
		panic(err)
	}
	err = corev1.AddToScheme(defaultScheme)
	if err != nil {
		panic(err)
	}
	err = rbacv1.AddToScheme(defaultScheme)
	if err != nil {
		panic(err)
	}
	err = rbacv1beta1.AddToScheme(defaultScheme)
	if err != nil {
		panic(err)
	}
}

func TestWrapFunc(t *testing.T) {
	testCases := []struct {
		Name       string
		Objects    []runtime.Object
		Func       interface{}
		ShouldFail bool
	}{
		{
			Name:       "Bad Function Signature",
			Objects:    nil,
			Func:       func() {},
			ShouldFail: true,
		},
		{
			Name:       "Good Signature No Objects Empty Struct",
			Objects:    nil,
			Func:       func(*testing.T, struct{}) {},
			ShouldFail: false,
		},
		{
			Name: "Good Signature Some Objects Empty Struct",
			Objects: []runtime.Object{
				&appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "hello",
						Name:      "world",
					},
				},
			},
			Func:       func(*testing.T, struct{}) {},
			ShouldFail: false,
		},
		{
			Name: "Good Signature Some Objects Has Struct",
			Objects: []runtime.Object{
				&appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "hello",
						Name:      "world",
					},
				},
			},
			Func: func(*testing.T, struct {
				Deployments []*appsv1.Deployment
			}) {
			},
			ShouldFail: false,
		},
		{
			Name: "Good Signature Some Objects Invalid Struct Internals",
			Objects: []runtime.Object{
				&appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "hello",
						Name:      "world",
					},
				},
			},
			Func: func(*testing.T, struct {
				Deployments []*appsv1.Deployment
				Objects     []runtime.Object
			}) {
			},
			ShouldFail: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			doFunc := WrapFunc(tc.Func, nil)
			if tc.ShouldFail {
				mockT := &testing.T{}
				doFunc(mockT, tc.Objects)
				assert.True(t, mockT.Failed())
			} else {
				doFunc(t, tc.Objects)
			}
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
			Name:             "Not Func",
			Func:             5,
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
			Name:             "Has Non Struct Second Arg",
			Func:             func(*testing.T, int) {},
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

func TestParseObjectsIntoStruct(t *testing.T) {
	type validMultiResourceStruct struct {
		ConfigMaps  []*corev1.ConfigMap
		Deployments []*appsv1.Deployment
	}

	type validMultiResourceStructWithUnstructured struct {
		ConfigMaps  []*corev1.ConfigMap
		Deployments []*appsv1.Deployment
		Others      []*unstructured.Unstructured
	}

	type configMapStruct struct {
		ConfigMaps []*corev1.ConfigMap
	}
	type deploymentStruct struct {
		Deployments []*appsv1.Deployment
	}
	type secretStruct struct {
		Secrets []*corev1.Secret
	}
	type validStructWithEmbedded struct {
		deploymentStruct
		ConfigMapInfo configMapStruct
	}

	type aConfigMapStruct struct {
		A []*corev1.ConfigMap
	}
	type aDeploymentStruct struct {
		A []*appsv1.Deployment
	}
	type validStructWithIdenticalEmbeddedFields struct {
		aConfigMapStruct
		aDeploymentStruct
	}

	type ambiguousStruct struct {
		Deployments1 []*appsv1.Deployment
		Deployments2 []*appsv1.Deployment
	}

	helloWorldDeployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "hello",
			Name:      "world",
		},
	}

	helloWorldConfigMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "hello",
			Name:      "world",
		},
	}

	helloWorldSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "hello",
			Name:      "world",
		},
	}

	uObj, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(helloWorldSecret)
	unstructuredHelloWorldSecret := &unstructured.Unstructured{
		Object: uObj,
	}

	testCases := []struct {
		Name                 string
		Objects              []runtime.Object
		InputObjectStruct    interface{}
		ExpectedObjectStruct interface{}
		Strict               bool
		Scheme               *runtime.Scheme
		ShouldThrowError     bool
	}{
		{
			Name:                 "Nil Struct",
			Objects:              nil,
			InputObjectStruct:    nil,
			ExpectedObjectStruct: nil,
			Strict:               false,
			Scheme:               defaultScheme,
			ShouldThrowError:     true,
		},
		{
			Name:                 "Valid Objects And Nil Struct",
			Objects:              []runtime.Object{helloWorldDeployment},
			InputObjectStruct:    nil,
			ExpectedObjectStruct: nil,
			Strict:               false,
			Scheme:               defaultScheme,
			ShouldThrowError:     true,
		},
		{
			Name:                 "Valid Objects And Empty Struct",
			Objects:              []runtime.Object{helloWorldDeployment},
			InputObjectStruct:    &struct{}{},
			ExpectedObjectStruct: &struct{}{},
			Strict:               false,
			Scheme:               defaultScheme,
			ShouldThrowError:     false,
		},
		{
			Name:                 "Valid Objects And Empty Struct In Strict Mode",
			Objects:              []runtime.Object{helloWorldDeployment},
			InputObjectStruct:    &struct{}{},
			ExpectedObjectStruct: &struct{}{},
			Strict:               true,
			Scheme:               defaultScheme,
			ShouldThrowError:     true,
		},
		{
			Name:                 "Valid Objects And Valid Non Ptr Struct",
			Objects:              []runtime.Object{helloWorldDeployment},
			InputObjectStruct:    deploymentStruct{}, // note: should be &deploymentStruct{}
			ExpectedObjectStruct: nil,
			Strict:               false,
			Scheme:               defaultScheme,
			ShouldThrowError:     true,
		},
		{
			Name:                 "Nil Objects And Valid Ptr Struct",
			Objects:              nil,
			InputObjectStruct:    &deploymentStruct{},
			ExpectedObjectStruct: &deploymentStruct{},
			Strict:               false,
			Scheme:               defaultScheme,
			ShouldThrowError:     false,
		},
		{
			Name:                 "Nil Objects And Valid Ptr Struct In Strict Mode",
			Objects:              nil,
			InputObjectStruct:    &deploymentStruct{},
			ExpectedObjectStruct: &deploymentStruct{},
			Strict:               true,
			Scheme:               defaultScheme,
			ShouldThrowError:     false,
		},
		{
			Name:              "Valid Objects And Valid Ptr Struct",
			Objects:           []runtime.Object{helloWorldDeployment},
			InputObjectStruct: &deploymentStruct{},
			ExpectedObjectStruct: &deploymentStruct{
				Deployments: []*appsv1.Deployment{helloWorldDeployment},
			},
			Strict:           false,
			Scheme:           defaultScheme,
			ShouldThrowError: false,
		},
		{
			Name:              "Unstructured Objects And Valid Ptr Struct",
			Objects:           []runtime.Object{unstructuredHelloWorldSecret},
			InputObjectStruct: &secretStruct{},
			ExpectedObjectStruct: &secretStruct{
				Secrets: []*corev1.Secret{helloWorldSecret},
			},
			Strict:           false,
			Scheme:           defaultScheme,
			ShouldThrowError: false,
		},
		{
			Name:                 "Nil Objects And Ambiguous Ptr Struct",
			Objects:              nil,
			InputObjectStruct:    &ambiguousStruct{},
			ExpectedObjectStruct: nil,
			Strict:               false,
			Scheme:               defaultScheme,
			ShouldThrowError:     true,
		},
		{
			Name:                 "Valid Objects And Ambiguous Ptr Struct",
			Objects:              []runtime.Object{helloWorldDeployment},
			InputObjectStruct:    &ambiguousStruct{},
			ExpectedObjectStruct: nil,
			Strict:               false,
			Scheme:               defaultScheme,
			ShouldThrowError:     true,
		},
		{
			Name:                 "Extra Objects And Valid Ptr Struct",
			Objects:              []runtime.Object{helloWorldDeployment},
			InputObjectStruct:    &configMapStruct{},
			ExpectedObjectStruct: &configMapStruct{},
			Strict:               false,
			Scheme:               defaultScheme,
			ShouldThrowError:     false,
		},
		{
			Name:                 "Extra Objects And Valid Ptr Struct In Strict Mode",
			Objects:              []runtime.Object{helloWorldDeployment},
			InputObjectStruct:    &configMapStruct{},
			ExpectedObjectStruct: &configMapStruct{},
			Strict:               true,
			Scheme:               defaultScheme,
			ShouldThrowError:     true,
		},
		{
			Name: "Valid Objects And Multiple Valid Ptr Struct",
			Objects: []runtime.Object{
				helloWorldDeployment,
				helloWorldConfigMap,
			},
			InputObjectStruct: &validMultiResourceStruct{},
			ExpectedObjectStruct: &validMultiResourceStruct{
				Deployments: []*appsv1.Deployment{helloWorldDeployment},
				ConfigMaps:  []*corev1.ConfigMap{helloWorldConfigMap},
			},
			Strict:           false,
			Scheme:           defaultScheme,
			ShouldThrowError: false,
		},
		{
			Name: "Valid Objects And Multiple Valid Ptr Struct In Strict Mode",
			Objects: []runtime.Object{
				helloWorldDeployment,
				helloWorldConfigMap,
			},
			InputObjectStruct: &validMultiResourceStruct{},
			ExpectedObjectStruct: &validMultiResourceStruct{
				Deployments: []*appsv1.Deployment{helloWorldDeployment},
				ConfigMaps:  []*corev1.ConfigMap{helloWorldConfigMap},
			},
			Strict:           true,
			Scheme:           defaultScheme,
			ShouldThrowError: false,
		},
		{
			Name: "Valid Objects And Multiple Valid Embedded Ptr Struct",
			Objects: []runtime.Object{
				helloWorldDeployment,
				helloWorldConfigMap,
			},
			InputObjectStruct: &validStructWithEmbedded{},
			ExpectedObjectStruct: &validStructWithEmbedded{
				ConfigMapInfo: configMapStruct{
					ConfigMaps: []*corev1.ConfigMap{helloWorldConfigMap},
				},
				deploymentStruct: deploymentStruct{
					Deployments: []*appsv1.Deployment{helloWorldDeployment},
				},
			},
			Strict:           false,
			Scheme:           defaultScheme,
			ShouldThrowError: false,
		},
		{
			Name: "Valid Objects And Multiple Valid Embedded Ptr Struct In Strict Mode",
			Objects: []runtime.Object{
				helloWorldDeployment,
				helloWorldConfigMap,
			},
			InputObjectStruct: &validStructWithEmbedded{},
			ExpectedObjectStruct: &validStructWithEmbedded{
				ConfigMapInfo: configMapStruct{
					ConfigMaps: []*corev1.ConfigMap{helloWorldConfigMap},
				},
				deploymentStruct: deploymentStruct{
					Deployments: []*appsv1.Deployment{helloWorldDeployment},
				},
			},
			Strict:           true,
			Scheme:           defaultScheme,
			ShouldThrowError: false,
		},
		{
			Name: "Valid Objects And Multiple Valid Embedded Ptr Struct With Identical Fields",
			Objects: []runtime.Object{
				helloWorldDeployment,
				helloWorldConfigMap,
			},
			InputObjectStruct: &validStructWithIdenticalEmbeddedFields{},
			ExpectedObjectStruct: &validStructWithIdenticalEmbeddedFields{
				aConfigMapStruct: aConfigMapStruct{
					A: []*corev1.ConfigMap{helloWorldConfigMap},
				},
				aDeploymentStruct: aDeploymentStruct{
					A: []*appsv1.Deployment{helloWorldDeployment},
				},
			},
			Strict:           false,
			Scheme:           defaultScheme,
			ShouldThrowError: false,
		},
		{
			Name: "Valid Objects And Multiple Valid Embedded Ptr Struct With Identical Fields In Strict Mode",
			Objects: []runtime.Object{
				helloWorldDeployment,
				helloWorldConfigMap,
			},
			InputObjectStruct: &validStructWithIdenticalEmbeddedFields{},
			ExpectedObjectStruct: &validStructWithIdenticalEmbeddedFields{
				aConfigMapStruct: aConfigMapStruct{
					A: []*corev1.ConfigMap{helloWorldConfigMap},
				},
				aDeploymentStruct: aDeploymentStruct{
					A: []*appsv1.Deployment{helloWorldDeployment},
				},
			},
			Strict:           true,
			Scheme:           defaultScheme,
			ShouldThrowError: false,
		},
		{
			Name: "Valid and Extra Objects And Multiple Valid Ptr Struct With Unstructured",
			Objects: []runtime.Object{
				helloWorldConfigMap,
				helloWorldSecret,
				helloWorldDeployment,
			},
			InputObjectStruct: &validMultiResourceStructWithUnstructured{},
			ExpectedObjectStruct: &validMultiResourceStructWithUnstructured{
				Deployments: []*appsv1.Deployment{helloWorldDeployment},
				ConfigMaps:  []*corev1.ConfigMap{helloWorldConfigMap},
				Others: []*unstructured.Unstructured{
					unstructuredHelloWorldSecret,
				},
			},
			Strict:           false,
			Scheme:           defaultScheme,
			ShouldThrowError: false,
		},
		{
			Name: "Valid and Extra Objects And Multiple Valid Ptr Struct With Unstructured In Strict Mode",
			Objects: []runtime.Object{
				helloWorldConfigMap,
				helloWorldSecret,
				helloWorldDeployment,
			},
			InputObjectStruct: &validMultiResourceStructWithUnstructured{},
			ExpectedObjectStruct: &validMultiResourceStructWithUnstructured{
				Deployments: []*appsv1.Deployment{helloWorldDeployment},
				ConfigMaps:  []*corev1.ConfigMap{helloWorldConfigMap},
				Others: []*unstructured.Unstructured{
					unstructuredHelloWorldSecret,
				},
			},
			Strict:           true,
			Scheme:           defaultScheme,
			ShouldThrowError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := parseObjectsIntoStruct(tc.Objects, tc.InputObjectStruct, &ParseOptions{
				Strict: tc.Strict,
				Scheme: tc.Scheme,
			})
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
			t.Logf("objectStruct: %s", tc.InputObjectStruct)
			t.Logf("expectedObjectStruct: %s", tc.ExpectedObjectStruct)
			assert.Equal(t, tc.ExpectedObjectStruct, tc.InputObjectStruct, "objects were incorrectly parsed")
		})
	}
}

func TestGetSupportedTypes(t *testing.T) {
	type serviceAccountStruct struct {
		ServiceAccounts []*corev1.ServiceAccount
	}

	type serviceAccountStruct2 struct {
		ServiceAccounts []*corev1.ServiceAccount
	}

	type deploymentStruct struct {
		Deployments []*appsv1.Deployment
	}

	type embeddedValidStruct struct {
		serviceAccountStruct
		deploymentStruct
	}

	type embeddedValidStructWithConflicts struct {
		serviceAccountStruct
		serviceAccountStruct2
	}

	type embeddedInception struct {
		embeddedValidStruct
	}

	type embeddedInceptionDeployment struct {
		deploymentStruct
	}

	type embeddedInceptionMax struct {
		embeddedInceptionDeployment
		serviceAccountStruct
	}

	type embeddedInceptionMaxWithConflicts struct {
		embeddedInceptionDeployment
		deploymentStruct
	}

	testCases := []struct {
		Name             string
		Object           interface{}
		ExpectedTypes    map[reflect.Type]string
		ShouldThrowError bool
	}{
		{
			Name:             "Nil Object",
			Object:           nil,
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
		{
			Name:             "Not A Struct",
			Object:           5,
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
		{
			Name: "Empty Struct",
			Object: struct {
			}{},
			ExpectedTypes:    nil,
			ShouldThrowError: false,
		},
		{
			Name: "Non-List Struct That Implements runtime.Object",
			Object: struct {
				A corev1.ServiceAccount
			}{},
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
		{
			Name: "Non-List Ptrs To Struct That Implements runtime.Object",
			Object: struct {
				A *corev1.ServiceAccount
			}{},
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
		{
			Name: "Lists of Structs that Implement runtime.Object",
			Object: struct {
				A []corev1.ServiceAccount
				B []appsv1.Deployment
				C []batchv1.Job
				D []rbacv1beta1.ClusterRoleBinding
				E []rbacv1.ClusterRoleBinding
				F []rbacv1.RoleBinding
			}{},
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
		{
			Name: "Lists of Ptr To Structs That Don't Implement runtime.Object",
			Object: struct {
				A []*struct{}
			}{},
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
		{
			Name: "Interfaces That Implement runtime.Object",
			Object: struct {
				B runtime.Object
			}{},
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
		{
			Name: "Lists of Interfaces That Implement runtime.Object",
			Object: struct {
				B []runtime.Object
			}{},
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
		{
			Name: "Mix and Match Lists of Pointers To Structs That Implement runtime.Object and Interfaces That Implement runtime.Object",
			Object: struct {
				A []*corev1.ServiceAccount
				B []*appsv1.Deployment
				D runtime.Object
			}{},
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
		{
			Name: "Mix and Match Lists of Pointers To Structs That Implement runtime.Object and Lists of Interfaces That Implement runtime.Object",
			Object: struct {
				A []*corev1.ServiceAccount
				B []*appsv1.Deployment
				D []runtime.Object
			}{},
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
		{
			Name: "Mix And Match",
			Object: struct {
				A []*corev1.ServiceAccount
				B appsv1.Deployment
				C runtime.Object
				D []runtime.Object
			}{},
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
		{
			Name: "Ambiguous Fields",
			Object: struct {
				A []*corev1.ServiceAccount
				B []*corev1.ServiceAccount
			}{},
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
		{
			Name: "Lists of Ptrs To Structs That Implement runtime.Object",
			Object: struct {
				A []*corev1.ServiceAccount
				B []*appsv1.Deployment
				C []*batchv1.Job
				D []*rbacv1beta1.ClusterRoleBinding
				E []*rbacv1.ClusterRoleBinding
				F []*rbacv1.RoleBinding
				G []*unstructured.Unstructured
			}{},
			ExpectedTypes: map[reflect.Type]string{
				reflect.TypeOf(&corev1.ServiceAccount{}):          "A",
				reflect.TypeOf(&appsv1.Deployment{}):              "B",
				reflect.TypeOf(&batchv1.Job{}):                    "C",
				reflect.TypeOf(&rbacv1beta1.ClusterRoleBinding{}): "D",
				reflect.TypeOf(&rbacv1.ClusterRoleBinding{}):      "E",
				reflect.TypeOf(&rbacv1.RoleBinding{}):             "F",
				reflect.TypeOf(&unstructured.Unstructured{}):      "G",
			},
			ShouldThrowError: false,
		},
		{
			Name:   "Embedded Struct",
			Object: embeddedValidStruct{},
			ExpectedTypes: map[reflect.Type]string{
				reflect.TypeOf(&corev1.ServiceAccount{}): "serviceAccountStruct.ServiceAccounts",
				reflect.TypeOf(&appsv1.Deployment{}):     "deploymentStruct.Deployments",
			},
			ShouldThrowError: false,
		},
		{
			Name:             "Embedded Struct With Conflicts",
			Object:           embeddedValidStructWithConflicts{},
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
		{
			Name:   "Embedded Inception",
			Object: embeddedInception{},
			ExpectedTypes: map[reflect.Type]string{
				reflect.TypeOf(&corev1.ServiceAccount{}): "embeddedValidStruct.serviceAccountStruct.ServiceAccounts",
				reflect.TypeOf(&appsv1.Deployment{}):     "embeddedValidStruct.deploymentStruct.Deployments",
			},
			ShouldThrowError: false,
		},
		{
			Name:   "Embedded Inception Max",
			Object: embeddedInceptionMax{},
			ExpectedTypes: map[reflect.Type]string{
				reflect.TypeOf(&corev1.ServiceAccount{}): "serviceAccountStruct.ServiceAccounts",
				reflect.TypeOf(&appsv1.Deployment{}):     "embeddedInceptionDeployment.deploymentStruct.Deployments",
			},
			ShouldThrowError: false,
		},
		{
			Name:             "Embedded Inception Max With Conflicts",
			Object:           embeddedInceptionMaxWithConflicts{},
			ExpectedTypes:    nil,
			ShouldThrowError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			objType := reflect.TypeOf(tc.Object)
			tracker, err := getSupportedTypes(objType)
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

			t.Logf("types: %s", tracker.typeToField)
			t.Logf("expectedTypes: %s", tc.ExpectedTypes)
			// Ensure types are the same
			for fieldType, field := range tracker.typeToField {
				assert.Equal(t, tc.ExpectedTypes[fieldType], field)
			}
			assert.Equal(t, len(tracker.typeToField), len(tc.ExpectedTypes))
		})
	}
}
