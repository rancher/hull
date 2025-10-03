package checker

import (
	"fmt"
	"sort"
	"testing"

	"github.com/rancher/hull/pkg/parser"
	"github.com/rancher/wrangler/v3/pkg/objectset"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewChecker(t *testing.T) {
	example1 := "example_1.yaml"
	exampleYamlString1 := `
apiVersion: hello.cattle.io/v1
kind: World
metadata:
	name: rancher
	namespace: hull
---
apiVersion: world.cattle.io/v1
kind: Hello
metadata:
	name: rancher
	namespace: hull
`
	exampleOs1, err := parser.Parse(exampleYamlString1)
	if err != nil {
		t.Errorf("failed to parse exampleYamlString1: %s", err)
		return
	}

	exampleOsMap1 := map[string]*objectset.ObjectSet{
		example1: exampleOs1,
	}

	expectedOsMap1 := map[string]*objectset.ObjectSet{
		"":       exampleOs1,
		example1: exampleOs1,
	}

	example2 := "example_2.yaml"
	exampleYamlString2 := `
apiVersion: rancher.cattle.io/v1
kind: Hull
metadata:
	name: hello
	namespace: world
---
apiVersion: hull.cattle.io/v1
kind: Rancher
metadata:
	name: hello
	namespace: world
`
	exampleOs2, err := parser.Parse(exampleYamlString2)
	if err != nil {
		t.Errorf("failed to parse exampleYamlString2: %s", err)
		return
	}

	exampleOsMap2 := map[string]*objectset.ObjectSet{
		example2: exampleOs2,
	}

	expectedOsMap2 := map[string]*objectset.ObjectSet{
		"":       exampleOs2,
		example2: exampleOs2,
	}

	exampleOsBoth := objectset.NewObjectSet().Add(exampleOs1.All()...).Add(exampleOs2.All()...)

	exampleOsMapBoth := map[string]*objectset.ObjectSet{
		example1: exampleOs1,
		example2: exampleOs2,
	}

	expectedOsMapBoth := map[string]*objectset.ObjectSet{
		"":       exampleOsBoth,
		example1: exampleOs1,
		example2: exampleOs2,
	}

	expectedOsMapBothWithEmptyTemplates := map[string]*objectset.ObjectSet{
		"":           exampleOsBoth,
		example1:     exampleOs1,
		example2:     exampleOs2,
		"empty.yaml": objectset.NewObjectSet(),
		"nil.yaml":   nil,
	}

	testCases := []struct {
		Name         string
		TemplateName string
		Template     interface{}
		ExpectedMap  map[string]*objectset.ObjectSet
	}{
		{
			Name:         "String 1",
			TemplateName: example1,
			Template:     exampleYamlString1,
			ExpectedMap:  expectedOsMap1,
		},
		{
			Name:         "ObjectSet 1",
			TemplateName: example1,
			Template:     exampleOs1,
			ExpectedMap:  expectedOsMap1,
		},
		{
			Name:        "ObjectSetMap 1",
			Template:    exampleOsMap1,
			ExpectedMap: expectedOsMap1,
		},
		{
			Name:         "String 2",
			TemplateName: example2,
			Template:     exampleYamlString2,
			ExpectedMap:  expectedOsMap2,
		},

		{
			Name:         "ObjectSet 2",
			TemplateName: example2,
			Template:     exampleOs2,
			ExpectedMap:  expectedOsMap2,
		},
		{
			Name:        "ObjectSetMap 2",
			Template:    exampleOsMap2,
			ExpectedMap: expectedOsMap2,
		},
		{
			Name:        "ObjectSetMap Both",
			Template:    exampleOsMapBoth,
			ExpectedMap: expectedOsMapBoth,
		},
		{
			Name:        "Expected ObjectSetMap Both",
			Template:    expectedOsMapBoth,
			ExpectedMap: expectedOsMapBoth,
		},
		{
			Name:        "Expected ObjectSetMap Both With Empty Templates",
			Template:    expectedOsMapBothWithEmptyTemplates,
			ExpectedMap: expectedOsMapBoth,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var uncastChecker Checker
			var err error
			switch template := tc.Template.(type) {
			case map[string]*objectset.ObjectSet:
				uncastChecker, err = NewChecker(template)
			case *objectset.ObjectSet:
				uncastChecker, err = NewCheckerFromObjectSet(template, tc.TemplateName)
			case string:
				uncastChecker, err = NewCheckerFromString(template, tc.TemplateName)
			default:
				t.Errorf("unknown type of template: found %T", template)
				return
			}
			if err != nil {
				t.Error(err)
				return
			}
			assert.NotNil(t, uncastChecker)
			if t.Failed() {
				return
			}

			t.Run("Passes Nil Function Check", func(t *testing.T) {
				uncastChecker.Check(t, nil)
			})

			t.Run("Passes Function Check", func(t *testing.T) {
				uncastChecker.Check(t, func(*testing.T, struct{}) {})
			})

			c, ok := uncastChecker.(*checker)
			if !ok {
				t.Errorf("could not extract internal checker from Checker for tests")
				return
			}

			expectedKeys := []string{}
			expectedObjectsPerKey := make(map[string][]unstructured.Unstructured, len(tc.ExpectedMap))
			for k, os := range tc.ExpectedMap {
				if os == nil || os.Len() == 0 {
					continue
				}
				expectedKeys = append(expectedKeys, k)
				for _, obj := range os.All() {
					expectedObjectsPerKey[k] = append(expectedObjectsPerKey[k], *(obj.(*unstructured.Unstructured)))
				}
			}
			sort.Strings(expectedKeys)
			for k := range expectedObjectsPerKey {
				sort.Slice(expectedObjectsPerKey[k], func(i, j int) bool {
					objI := expectedObjectsPerKey[k][i]
					objIKey := fmt.Sprintf("%s.%s/%s/%s", objI.GetKind(), objI.GetAPIVersion(), objI.GetNamespace(), objI.GetName())
					objJ := expectedObjectsPerKey[k][j]
					objJKey := fmt.Sprintf("%s.%s/%s/%s", objJ.GetKind(), objJ.GetAPIVersion(), objJ.GetNamespace(), objJ.GetName())
					return objIKey < objJKey
				})
			}

			keys := []string{}
			objectsPerKey := make(map[string][]unstructured.Unstructured, len(c.ObjectSets))
			for k, os := range c.ObjectSets {
				keys = append(keys, k)
				for _, obj := range os.All() {
					objectsPerKey[k] = append(objectsPerKey[k], *(obj.(*unstructured.Unstructured)))
				}
			}
			sort.Strings(keys)
			for k := range objectsPerKey {
				sort.Slice(objectsPerKey[k], func(i, j int) bool {
					objI := objectsPerKey[k][i]
					objIKey := fmt.Sprintf("%s.%s/%s/%s", objI.GetKind(), objI.GetAPIVersion(), objI.GetNamespace(), objI.GetName())
					objJ := objectsPerKey[k][j]
					objJKey := fmt.Sprintf("%s.%s/%s/%s", objJ.GetKind(), objJ.GetAPIVersion(), objJ.GetNamespace(), objJ.GetName())
					return objIKey < objJKey
				})
			}

			assert.Equal(t, expectedKeys, keys, "did not generate correct set of templateFile keys")

			for osPath := range tc.ExpectedMap {
				expectedObjects := expectedObjectsPerKey[osPath]
				objects := objectsPerKey[osPath]
				assert.Equal(t, expectedObjects, objects, "objects found at %s are not the same", osPath)
			}
		})
	}

	t.Run("Empty String", func(t *testing.T) {
		nilChecker, err := NewCheckerFromString("", "")
		assert.Nil(t, err)
		assert.Nil(t, nilChecker)
	})
	t.Run("Bad String", func(t *testing.T) {
		nilChecker, err := NewCheckerFromString("i am a bad string@@: ", "")
		assert.NotNil(t, err)
		assert.Nil(t, nilChecker)
	})
	t.Run("String With Only Document Separator", func(t *testing.T) {
		nilChecker, err := NewCheckerFromString("---\n---", "")
		assert.Nil(t, err)
		assert.Nil(t, nilChecker)
	})
	t.Run("Nil ObjectSet", func(t *testing.T) {
		nilChecker, err := NewCheckerFromObjectSet(nil, "")
		assert.Nil(t, err)
		assert.Nil(t, nilChecker)
	})
	t.Run("Empty ObjectSet", func(t *testing.T) {
		nilChecker, err := NewCheckerFromObjectSet(&objectset.ObjectSet{}, "")
		assert.Nil(t, err)
		assert.Nil(t, nilChecker)
	})
	t.Run("Nil ObjectSetMap", func(t *testing.T) {
		nilChecker, err := NewChecker(nil)
		assert.Nil(t, err)
		assert.Nil(t, nilChecker)
	})
	t.Run("Empty ObjectSetMap", func(t *testing.T) {
		nilChecker, err := NewChecker(map[string]*objectset.ObjectSet{})
		assert.Nil(t, err)
		assert.Nil(t, nilChecker)
	})
	t.Run("Empty String", func(t *testing.T) {
		nilChecker, err := NewCheckerFromString("", "")
		assert.Nil(t, err)
		assert.Nil(t, nilChecker)
	})
}
