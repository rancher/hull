package parser

import (
	"strings"
	"testing"

	"github.com/rancher/wrangler/v3/pkg/objectset"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestParse(t *testing.T) {
	resource1String := strings.Join([]string{
		"apiVersion: hello.cattle.io/v1",
		"kind: World",
		"metadata:",
		"\t" + strings.Join([]string{
			"name: rancher",
			"namespace: hull",
		}, "\n\t"),
	}, "\n")
	resource1Obj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "hello.cattle.io/v1",
			"kind":       "World",
			"metadata": map[string]interface{}{
				"name":      "rancher",
				"namespace": "hull",
			},
		},
	}

	resource2String := strings.Join([]string{
		"apiVersion: world.cattle.io/v1",
		"kind: Hello",
		"metadata:",
		"\t" + strings.Join([]string{
			"name: rancher",
			"namespace: hull",
		}, "\n\t"),
	}, "\n")
	resource2Obj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "world.cattle.io/v1",
			"kind":       "Hello",
			"metadata": map[string]interface{}{
				"name":      "rancher",
				"namespace": "hull",
			},
		},
	}

	testCases := []struct {
		Name             string
		Template         string
		ExpectedObjects  []unstructured.Unstructured
		ShouldThrowError bool
	}{
		{
			Name:            "Empty",
			Template:        "",
			ExpectedObjects: nil,
		},
		{
			Name:             "Bad Resource",
			Template:         "i am a bad string",
			ExpectedObjects:  nil,
			ShouldThrowError: true,
		},
		{
			Name:     "One Resource With Tabs",
			Template: resource1String,
			ExpectedObjects: []unstructured.Unstructured{
				resource1Obj,
			},
		},
		{
			Name:     "One Resource With 2 Spaces",
			Template: strings.ReplaceAll(resource1String, "\t", "  "),
			ExpectedObjects: []unstructured.Unstructured{
				resource1Obj,
			},
		},
		{
			Name:     "One Resource With 4 Spaces",
			Template: strings.ReplaceAll(resource1String, "\t", "    "),
			ExpectedObjects: []unstructured.Unstructured{
				resource1Obj,
			},
		},
		{
			Name:     "One Resource With Preceding Newlines",
			Template: "\n\n\n" + resource1String,
			ExpectedObjects: []unstructured.Unstructured{
				resource1Obj,
			},
		},
		{
			Name:     "One Resource Ending With Newlines",
			Template: resource1String + "\n\n\n",
			ExpectedObjects: []unstructured.Unstructured{
				resource1Obj,
			},
		},
		{
			Name:     "One Resource With Newlines On Both Ends",
			Template: "\n\n\n" + resource1String + "\n\n\n",
			ExpectedObjects: []unstructured.Unstructured{
				resource1Obj,
			},
		},
		{
			Name:     "Two Resources",
			Template: resource1String + "\n---\n" + resource2String,
			ExpectedObjects: []unstructured.Unstructured{
				resource1Obj,
				resource2Obj,
			},
		},
		{
			Name:     "Two Resources With Arbitrary Newlines",
			Template: "\n\n" + resource1String + "\n\n\n---\n" + resource2String + "\n\n\n\n\n",
			ExpectedObjects: []unstructured.Unstructured{
				resource1Obj,
				resource2Obj,
			},
		},
		{
			Name:            "No Resources But Multidocument Separators",
			Template:        "\n---\n---\n",
			ExpectedObjects: []unstructured.Unstructured{},
		},
		{
			Name:            "No Resources But Some YAML Without ApiVersion Or Kind",
			Template:        "hello: world",
			ExpectedObjects: []unstructured.Unstructured{},
		},
		{
			Name:            "No Resources But Some YAML Without ApiVersion",
			Template:        "kind: world",
			ExpectedObjects: []unstructured.Unstructured{},
		},
		{
			Name:            "No Resources But Some YAML Without Kind",
			Template:        "apiVersion: world",
			ExpectedObjects: []unstructured.Unstructured{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			os, err := Parse(tc.Template)
			if tc.ShouldThrowError {
				if err == nil {
					t.Error("expected error to be thrown")
				}
				return
			}
			if os == nil {
				os = objectset.NewObjectSet()
			}
			assert.Equal(t, len(tc.ExpectedObjects), os.Len())
			for _, obj := range tc.ExpectedObjects {
				assert.True(t, os.Contains(obj.GroupVersionKind().GroupKind(), objectset.ObjectKey{
					Name:      obj.GetName(),
					Namespace: obj.GetNamespace(),
				}))
			}
		})
	}
}
