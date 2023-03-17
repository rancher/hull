package simple

import (
	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/checker"
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
			Name: "ConfigMaps have expected data",

			Covers: []string{
				".Values.data",
			},

			Checks: test.Checks{
				checker.PerResource(func(tc *checker.TestContext, configmap *corev1.ConfigMap) {
					assert.Contains(tc.T,
						configmap.Data, "config",
						"%T %s does not have 'config' key", configmap, checker.Key(configmap),
					)
					if tc.T.Failed() {
						return
					}
					assert.Equal(tc.T,
						checker.ToYAML(checker.MustRenderValue[map[string]string](tc, ".Values.data")), configmap.Data["config"],
						"%T %s does not have correct data in 'config' key", configmap, checker.Key(configmap),
					)
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
				".Values.shouldFail",
			},

			FailureMessage: ".Values.shouldFail is set to true",
		},
		{
			Name: "Set .Values.shouldFailRequired",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				SetValue("shouldFailRequired", "true"),

			Covers: []string{
				".Values.shouldFailRequired",
			},

			FailureMessage: ".Values.shouldFailRequired is set to true",
		},
	},
}
