package example

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/checker"
	"github.com/aiyengar2/hull/pkg/checker/resource/workloads"
	"github.com/aiyengar2/hull/pkg/test"
	"github.com/aiyengar2/hull/pkg/test/coverage/deprecated"

	"github.com/stretchr/testify/assert"
)

var (
	defaultReleaseName = "with-schema"
	defaultNamespace   = "default"
)

var suite = test.Suite{
	ChartPath:    ChartPath,
	ValuesStruct: &ValuesYaml{},

	DefaultChecks: []checker.Check{
		{
			Name: "has exactly two configmaps",
			Func: workloads.EnsureNumConfigMaps(2),
		},
	},

	Cases: []test.Case{
		{
			Name: "Using Defaults",

			TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace),

			Checks: []checker.Check{
				{
					Name: "has correct default data in ConfigMaps",
					Func: workloads.EnsureConfigMapsHaveData(
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
					Func: workloads.EnsureConfigMapsHaveData(
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
	coverage, report := deprecated.CoverageOnType(t, ValuesYaml{}, templateOptions...)
	if t.Failed() {
		return
	}
	assert.Equal(t, 1.00, coverage, report)
}
