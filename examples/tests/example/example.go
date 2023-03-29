package example

import (
	"github.com/rancher/hull/pkg/chart"
	"github.com/rancher/hull/pkg/checker"
	"github.com/rancher/hull/pkg/test"
	"github.com/rancher/hull/pkg/utils"
	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ChartPath = utils.MustGetPathFromModuleRoot("..", "testdata", "charts", "example-chart")

var (
	DefaultReleaseName = "example-chart"
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
			Name: "Set .Values.args[0] to --debug",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).SetValue("args[0]", "--debug"),
		},
		{
			Name: "Set .Values.args[0] to --trace",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).SetValue("args[0]", "--trace"),
		},
	},

	NamedChecks: []test.NamedCheck{
		{
			Name: "Check Container Args",
			Covers: []string{
				".Values.args",
			},

			Checks: test.Checks{
				checker.PerWorkload(func(tc *checker.TestContext, obj metav1.Object, podTemplateSpec corev1.PodTemplateSpec) {
					if obj.GetNamespace() != checker.MustRenderValue[string](tc, ".Release.Namespace") {
						return
					}
					if obj.GetName() != checker.MustRenderValue[string](tc, ".Release.Name") {
						return
					}
					expectedArgs := checker.MustRenderValue[[]string](tc, ".Values.args")
					for _, container := range podTemplateSpec.Spec.Containers {
						if len(expectedArgs) == 0 {
							assert.Nil(tc.T, container.Args,
								"expected container %s in %T %s to have no args",
								container.Name, obj, checker.Key(obj),
							)
						} else {
							assert.Equal(tc.T,
								expectedArgs, container.Args,
								"container %s in %T %s does not have correct args",
								container.Name, obj, checker.Key(obj),
							)
						}
					}
				}),
			},
		},
	},
}
