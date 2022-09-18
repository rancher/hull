package chart

import (
	"path/filepath"
	"testing"

	"github.com/aiyengar2/hull/pkg/checker"
)

const (
	defaultReleaseName = "example-chart"
	defaultNamespace   = "default"
)

var (
	chartPath = filepath.Join("..", "..", "testdata", "charts", "example-chart")
)

type ExampleChart struct {
	Data map[string]interface{} `jsonschema:"description=Data to be inserted into a ConfigMap"`
}

// convert into jsonschema to validate values.schema.json contents
// verify that template values can be marshalled into a struct of this type
// define coverage based on the number of fields that are touched in the struct

func TestTest(t *testing.T) {
	testCases := []struct {
		Name  string
		Suite *TestSuite
	}{
		{
			Name: "Default",
			Suite: &TestSuite{
				ChartPath:    chartPath,
				ValuesStruct: &ExampleChart{},
				Cases: []TestCase{
					{
						Name:            "Using Defaults",
						TemplateOptions: NewTemplateOptions(defaultReleaseName, defaultNamespace),
						Checks: []checker.Check{
							{
								Name: "Noop",
								Func: func(*testing.T, struct{}) {},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Suite.Run(t, nil)
		})
	}
}
