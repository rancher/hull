package unit

import (
	"path/filepath"
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/checker"
	"github.com/aiyengar2/hull/pkg/checker/resource"
	"github.com/stretchr/testify/assert"
)

const (
	defaultReleaseName = "example-chart"
	defaultNamespace   = "default"
)

var (
	chartPath = filepath.Join("..", "testdata", "charts", "example-chart")
)

type ExampleChart struct {
	Data map[string]interface{} `jsonschema:"description=Data to be inserted into a ConfigMap"`
}

// convert into jsonschema to validate values.schema.json contents
// verify that template values can be marshalled into a struct of this type
// define coverage based on the number of fields that are touched in the struct

var suite = chart.TestSuite{
	ChartPath:    chartPath,
	ValuesStruct: &ExampleChart{},

	Cases: []chart.TestCase{
		{
			Name: "Using Defaults",

			TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace),

			Checks: []checker.Check{
				{
					Name: "has correct default data in ConfigMaps",
					Func: checkIfConfigMapsHaveData(
						map[string]string{"config": "hello: rancher"},
					),
				},
			},
		},
		{
			Name: "Setting .Values.Data",

			TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace).
				SetValue("data.hello", "world"),

			Checks: []checker.Check{
				{
					Name: "sets .data.config on ConfigMaps",
					Func: checkIfConfigMapsHaveData(
						map[string]string{"config": "hello: world"},
					),
				},
			},
		},
	},
}

func TestChart(t *testing.T) {
	suite.Run(t, &chart.HelmLintOptions{
		Rancher: chart.RancherHelmLintOptions{
			Enabled: true,
		},
	})
}

type Configmaps struct {
	resource.ConfigMaps
}

func checkIfConfigMapsHaveData(data map[string]string) checker.CheckFunc {
	return func(t *testing.T, cms Configmaps) {
		for _, cm := range cms.ConfigMaps {
			assert.Equal(t, cm.Data, data)
		}
	}
}

// Classes of tests that need to be tested
// Those that apply on every manifest: windows nodeSelectors and tolerations
// Those that apply

// convert go struct to openapi schema and assert equivalence to values.schema.json
// define "coverage" on an arbitrary values.yaml struct to be a test that needs to be 100
// define scaffold for "default" test cases
