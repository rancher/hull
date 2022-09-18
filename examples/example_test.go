package examples

import (
	"path/filepath"
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/checker"

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

var suite = chart.TestSuite{
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

func TestCoverage(t *testing.T) {
	templateOptions := []*chart.TemplateOptions{}
	for _, c := range suite.Cases {
		templateOptions = append(templateOptions, c.TemplateOptions)
	}
	coverage := chart.Coverage(t, ExampleChart{}, templateOptions...)
	if t.Failed() {
		return
	}
	assert.Equal(t, 1.00, coverage)
}
