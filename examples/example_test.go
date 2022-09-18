package examples

import (
	"path/filepath"
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/checker"
	"github.com/aiyengar2/hull/pkg/test"

	"github.com/stretchr/testify/assert"
)

var (
	chartPath = filepath.Join("..", "testdata", "charts", "example-chart")
)

var (
	defaultReleaseName = "example-chart"
	defaultNamespace   = "default"
)

type ExampleChart struct {
	Data map[string]interface{} `jsonschema:"description=Data to be inserted into a ConfigMap"`
}

var suite = test.Suite{
	ChartPath:    chartPath,
	ValuesStruct: &ExampleChart{},

	DefaultChecks: []checker.Check{
		{
			Name: "has exactly two configmaps",
			Func: func(t *testing.T, cms Configmaps) {
				assert.Equal(t, 2, len(cms.ConfigMaps))
			},
		},
	},

	Cases: []test.Case{
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
	suite.Run(t, test.GetRancherOptions())
}

func TestCoverage(t *testing.T) {
	templateOptions := []*chart.TemplateOptions{}
	for _, c := range suite.Cases {
		templateOptions = append(templateOptions, c.TemplateOptions)
	}
	coverage, report := test.Coverage(t, ExampleChart{}, templateOptions...)
	if t.Failed() {
		return
	}
	assert.Equal(t, 1.00, coverage, report)
}
