package checker

import (
	"fmt"
	"testing"

	"github.com/rancher/wrangler/v3/pkg/relatedresource"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKey(t *testing.T) {
	testCases := []struct {
		Name   string
		Obj    metav1.Object
		Expect relatedresource.Key
	}{
		{
			Name: "Global Resource",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "global-resource",
				},
			},
			Expect: relatedresource.NewKey("", "global-resource"),
		},
		{
			Name: "Namespaced Resource",
			Obj: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "namespaced-resource",
					Namespace: "my-namespace",
				},
			},
			Expect: relatedresource.NewKey("my-namespace", "namespaced-resource"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Expect, Key(tc.Obj))
		})
	}
}

func TestSelect(t *testing.T) {
	testCases := []struct {
		Name         string
		Obj          metav1.Object
		ObjName      string
		ObjNamespace string
		Selected     bool
	}{
		{
			Name: "Global Resource",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "global-resource",
				},
			},
			ObjName:      "global-resource",
			ObjNamespace: "",
			Selected:     true,
		},
		{
			Name: "Select Global Resource With Namespace",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "global-resource",
				},
			},
			ObjName:      "global-resource",
			ObjNamespace: "my-namespace",
			Selected:     false,
		},
		{
			Name: "Namespaced Resource",
			Obj: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "namespaced-resource",
					Namespace: "my-namespace",
				},
			},
			ObjName:      "namespaced-resource",
			ObjNamespace: "my-namespace",
			Selected:     true,
		},
		{
			Name: "Wrong Name For Resource",
			Obj: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "namespaced-resource",
					Namespace: "my-namespace",
				},
			},
			ObjName:      "wrong-name",
			ObjNamespace: "my-namespace",
			Selected:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Selected, Select(tc.ObjName, tc.ObjNamespace, tc.Obj))
		})
	}
}

func TestHasLabels(t *testing.T) {
	testCases := []struct {
		Name   string
		Obj    metav1.Object
		Labels map[string]string

		Has   bool
		Match bool
	}{
		{
			Name: "Single Label",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"hello": "world",
					},
				},
			},
			Labels: map[string]string{
				"hello": "world",
			},
			Has:   true,
			Match: true,
		},
		{
			Name: "No Labels",
			Obj:  &rbacv1.ClusterRole{},
			Labels: map[string]string{
				"hello":  "world",
				"cattle": "rancher",
			},
			Has:   false,
			Match: false,
		},
		{
			Name: "Missing Label",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"hello": "world",
					},
				},
			},
			Labels: map[string]string{
				"hello":  "world",
				"cattle": "rancher",
			},
			Has:   false,
			Match: false,
		},
		{
			Name: "Multiple Labels",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"hello":  "world",
						"cattle": "rancher",
					},
				},
			},
			Labels: map[string]string{
				"hello":  "world",
				"cattle": "rancher",
			},
			Has:   true,
			Match: true,
		},
		{
			Name: "Wrong Labels",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"hello":  "different",
						"cattle": "rancher",
					},
				},
			},
			Labels: map[string]string{
				"hello":  "world",
				"cattle": "rancher",
			},
			Has:   false,
			Match: false,
		},
		{
			Name: "Wrong Labels 2",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"hello":  "different",
						"cattle": "another",
					},
				},
			},
			Labels: map[string]string{
				"hello":  "world",
				"cattle": "rancher",
			},
			Has:   false,
			Match: false,
		},
		{
			Name: "Extra Labels",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"hello":  "world",
						"cattle": "rancher",
					},
				},
			},
			Labels: map[string]string{
				"hello": "world",
			},
			Has: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			hasErr := HasLabels(tc.Obj, tc.Labels)
			if tc.Has {
				assert.Nil(t, hasErr, fmt.Sprintf("%s", hasErr))
			} else {
				assert.NotNil(t, hasErr, fmt.Sprintf("%s", hasErr))
			}
			matchErr := MatchLabels(tc.Obj, tc.Labels)
			if tc.Match {
				assert.Nil(t, matchErr, fmt.Sprintf("%s", matchErr))
			} else {
				assert.NotNil(t, matchErr, fmt.Sprintf("%s", matchErr))
			}
		})
	}
}

func TestHasAnnotations(t *testing.T) {
	testCases := []struct {
		Name        string
		Obj         metav1.Object
		Annotations map[string]string

		Has   bool
		Match bool
	}{
		{
			Name: "Single Annotation",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"hello": "world",
					},
				},
			},
			Annotations: map[string]string{
				"hello": "world",
			},
			Has:   true,
			Match: true,
		},
		{
			Name: "No Annotations",
			Obj:  &rbacv1.ClusterRole{},
			Annotations: map[string]string{
				"hello":  "world",
				"cattle": "rancher",
			},
			Has:   false,
			Match: false,
		},
		{
			Name: "Missing Annotation",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"hello": "world",
					},
				},
			},
			Annotations: map[string]string{
				"hello":  "world",
				"cattle": "rancher",
			},
			Has:   false,
			Match: false,
		},
		{
			Name: "Multiple Annotations",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"hello":  "world",
						"cattle": "rancher",
					},
				},
			},
			Annotations: map[string]string{
				"hello":  "world",
				"cattle": "rancher",
			},
			Has:   true,
			Match: true,
		},
		{
			Name: "Wrong Annotations",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"hello":  "different",
						"cattle": "rancher",
					},
				},
			},
			Annotations: map[string]string{
				"hello":  "world",
				"cattle": "rancher",
			},
			Has:   false,
			Match: false,
		},
		{
			Name: "Wrong Annotations 2",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"hello":  "different",
						"cattle": "another",
					},
				},
			},
			Annotations: map[string]string{
				"hello":  "world",
				"cattle": "rancher",
			},
			Has:   false,
			Match: false,
		},
		{
			Name: "Extra Annotations",
			Obj: &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"hello":  "world",
						"cattle": "rancher",
					},
				},
			},
			Annotations: map[string]string{
				"hello": "world",
			},
			Has: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			hasErr := HasAnnotations(tc.Obj, tc.Annotations)
			if tc.Has {
				assert.Nil(t, hasErr, fmt.Sprintf("%s", hasErr))
			} else {
				assert.NotNil(t, hasErr, fmt.Sprintf("%s", hasErr))
			}
			matchErr := MatchAnnotations(tc.Obj, tc.Annotations)
			if tc.Match {
				assert.Nil(t, matchErr, fmt.Sprintf("%s", matchErr))
			} else {
				assert.NotNil(t, matchErr, fmt.Sprintf("%s", matchErr))
			}
		})
	}
}
