package example

import (
	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/checker"
	"github.com/aiyengar2/hull/pkg/test"
	"github.com/aiyengar2/hull/pkg/utils"
	"github.com/rancher/wrangler/pkg/relatedresource"
	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var ChartPath = utils.MustGetPathFromModuleRoot("..", "testdata", "charts", "example-chart")

var (
	DefaultReleaseName = "example-chart"
	DefaultNamespace   = "default"
)

var suite = test.Suite{
	ChartPath: ChartPath,

	TemplateChecks: []test.TemplateCheck{
		{
			Name: "All Deployments Have ServiceAccount",
			Func: checker.NewCheckFunc(
				checker.NewChainedCheckFunc(func(tc *checker.TestContext, objs struct{ Deployments []*appsv1.Deployment }) {
					serviceAccountsToCheck := map[relatedresource.Key]bool{}
					for _, deployment := range objs.Deployments {
						key := relatedresource.NewKey(
							deployment.Namespace,
							deployment.Spec.Template.Spec.ServiceAccountName,
						)
						serviceAccountsToCheck[key] = false
					}
					checker.Store(tc, "ServiceAccountsToCheck", serviceAccountsToCheck)
				}),
				checker.NewChainedCheckFunc(func(tc *checker.TestContext, objs struct{ ServiceAccounts []*corev1.ServiceAccount }) {
					serviceAccountsToCheck, ok := checker.Get[string, map[relatedresource.Key]bool](tc, "ServiceAccountsToCheck")
					if !ok {
						return
					}
					for _, serviceAccount := range objs.ServiceAccounts {
						key := relatedresource.NewKey(serviceAccount.Namespace, serviceAccount.Name)
						_, ok := serviceAccountsToCheck[key]
						if !ok {
							continue
						}
						serviceAccountsToCheck[key] = true
					}
					for key, exists := range serviceAccountsToCheck {
						if exists {
							tc.T.Logf("serviceaccount %s exists in this Helm chart", key)
						} else {
							tc.T.Errorf("serviceaccount %s is not in this chart", key)
						}
					}
				}),
			),
		},
	},

	Cases: []test.Case{
		{
			Name: "Using Defaults",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace),
		},
		{
			Name: "Setting .Values.args[0] to --debug",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).SetValue("args[0]", "--debug"),

			ValueChecks: []test.ValueCheck{
				{
					Name: "Passes --debug Flag To Deployment",
					Covers: []string{
						"templates/deployment.yaml",
					},
					Func: checker.NewCheckFunc(
						checker.NewChainedCheckFunc(func(tc *checker.TestContext, objs struct{ Deployments []*appsv1.Deployment }) {
							for _, deployment := range objs.Deployments {
								for _, container := range deployment.Spec.Template.Spec.Containers {
									assert.Equal(tc.T, []string{"--debug"}, container.Args, "container %s in Deployment %s/%s does not have debug", container.Name, deployment.Namespace, deployment.Name)
								}
							}
						}),
					),
				},
			},
		},
	},
}
