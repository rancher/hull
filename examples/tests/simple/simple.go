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

	TemplateChecks: []test.TemplateCheck{},

	Cases: []test.Case{
		{
			Name: "Using Defaults",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace),
		},
		{
			Name: "Override .Values.data",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				Set("data", map[string]string{"hello": "cattle"}),

			ValueChecks: []test.ValueCheck{
				{
					Name: "Has hello: cattle in ConfigMap",
					Covers: []string{
						"templates/configmap.yaml",
					},
					Func: checker.NewCheckFunc(
						checker.NewChainedCheckFunc(func(tc *checker.TestContext, objs struct{ Configmaps []*corev1.ConfigMap }) {
							for _, configmap := range objs.Configmaps {
								assert.Equal(tc.T, map[string]string{"config": "hello: cattle"}, configmap.Data)
							}
						}),
					),
				},
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
