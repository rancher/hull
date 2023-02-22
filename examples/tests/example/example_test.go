package example

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/checker"
	"github.com/aiyengar2/hull/pkg/test"
)

var (
	defaultReleaseName = "with-schema"
	defaultNamespace   = "default"
)

var suite = test.Suite{
	ChartPath: ChartPath,

	DefaultChecks: []checker.Check{
		{
			Name: "Noop Default Check",
			Func: func(*testing.T, struct{}) {},
		},
	},

	Cases: []test.Case{
		{
			Name: "Using Defaults",

			TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace),

			Checks: []checker.Check{
				{
					Name: "Noop Check",
					Func: func(*testing.T, struct{}) {},
				},
			},
		},
	},
}

func TestChart(t *testing.T) {
	suite.Run(t, test.GetRancherOptions())
}
