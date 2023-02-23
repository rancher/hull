package example

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/checker"
	"github.com/aiyengar2/hull/pkg/test"
	"github.com/aiyengar2/hull/pkg/utils"
)

var (
	ChartPath = utils.MustGetPathFromModuleRoot("..", "testdata", "charts", "with-schema")

	DefaultReleaseName = "with-schema"
	DefaultNamespace   = "default"
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

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace),

			Checks: []checker.Check{
				{
					Name: "Noop Check",
					Func: func(*testing.T, struct{}) {},
				},
			},
		},
	},
}
