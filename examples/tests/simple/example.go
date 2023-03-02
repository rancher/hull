package example

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
						checker.NewChainedCheckFunc(func(tc *checker.TestContext, configmaps []*corev1.ConfigMap) error {
							for _, configmap := range configmaps {
								assert.Equal(tc.T, map[string]string{"config": "hello: cattle"}, configmap.Data)
							}
							return nil
						}),
					),
				},
			},
		},
	},
}
