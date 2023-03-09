package simple

import (
	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/checker"
	"github.com/aiyengar2/hull/pkg/extract"
	"github.com/aiyengar2/hull/pkg/test"
	"github.com/aiyengar2/hull/pkg/utils"
	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
)

var ChartPath = utils.MustGetPathFromModuleRoot("..", "testdata", "charts", "simple-chart")

var (
	DefaultReleaseName = "simple-chart"
	DefaultNamespace   = "default"
)

var suite = test.Suite{
	ChartPath: ChartPath,

	Cases: []test.Case{
		{
			Name: "Using Defaults",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace),
		},
		{
			Name: "Override .Values.data",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				Set("data", map[string]string{"hello": "cattle"}),
		},
	},

	NamedChecks: []test.NamedCheck{
		{
			Name:   "ConfigMaps have expected data",
			Covers: []string{"templates/configmap.yaml"},
			Checks: test.Checks{
				checker.PerResource(func(tc *checker.TestContext, configmap *corev1.ConfigMap) {
					// ensure config key always exists
					configData, configDataExists := extract.Field[string](configmap.Data, "config")
					assert.True(tc.T, configDataExists, "missing key 'config' in %T %s", configmap, checker.Key(configmap))

					// ensure .Values.data is always in ConfigMap
					valuesData := checker.ToYAML(
						checker.MustRenderValue[map[string]string](tc, ".Values.data"),
					)
					assert.Equal(tc.T, valuesData, configData, "key 'config' in %T %s has unexpected data", configmap, checker.Key(configmap))
				}),
			},
		},
	},

	FailureCases: []test.FailureCase{
		{
			Name: "Set .Values.shouldFail",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				SetValue("shouldFail", "true"),

			Covers: []string{
				"templates/configmap.yaml",
			},

			FailureMessage: ".Values.shouldFail is set to true",
		},
		{
			Name: "Set .Values.shouldFailRequired",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				SetValue("shouldFailRequired", "true"),

			Covers: []string{
				"templates/configmap.yaml",
			},

			FailureMessage: ".Values.shouldFailRequired is set to true",
		},
	},
}
