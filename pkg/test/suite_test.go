package test

import (
	"sort"
	"strings"
	"testing"

	"github.com/rancher/hull/pkg/chart"
	"github.com/rancher/hull/pkg/checker"
	"github.com/rancher/hull/pkg/utils"
	"github.com/stretchr/testify/assert"
	helmChartUtil "helm.sh/helm/v3/pkg/chartutil"
)

const (
	defaultReleaseName = "example-chart"
	defaultNamespace   = "default"
)

var (
	chartPath        = utils.MustGetPathFromModuleRoot("testdata", "charts", "example-chart")
	simpleChartPath  = utils.MustGetPathFromModuleRoot("testdata", "charts", "simple-chart")
	badTemplatesPath = utils.MustGetPathFromModuleRoot("testdata", "charts", "bad-templates")
)

// convert into jsonschema to validate values.schema.json contents
// verify that template values can be marshalled into a struct of this type
// define coverage based on the number of fields that are touched in the struct

func TestRun(t *testing.T) {
	testCases := []struct {
		Name             string
		Suite            *Suite
		ShouldThrowError bool
	}{
		{
			Name: "No Chart",
			Suite: &Suite{
				ChartPath: "",
			},
			ShouldThrowError: true,
		},
		{
			Name: "Bad Templates",
			Suite: &Suite{
				ChartPath: badTemplatesPath,
			},
			ShouldThrowError: true,
		},
		{
			Name: "Example Chart",
			Suite: &Suite{
				ChartPath: chartPath,
			},
		},
		{
			Name: "Example Chart With Nil TemplateOptions",
			Suite: &Suite{
				ChartPath: chartPath,
				Cases: []Case{
					{
						Name:            "No Options",
						TemplateOptions: nil,
					},
				},
			},
		},
		{
			Name: "Example Chart With DefaultValues",
			Suite: &Suite{
				ChartPath:     chartPath,
				DefaultValues: chart.NewValues(),
				Cases: []Case{
					{
						Name:            "No Options",
						TemplateOptions: nil,
					},
				},
			},
		},
		{
			Name: "Example Chart With Cases",
			Suite: &Suite{
				ChartPath: chartPath,
				Cases: []Case{
					{
						Name:            "Using Defaults",
						TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace),
					},
				},
			},
		},
		{
			Name: "Example Chart With Cases And Nil Checks",
			Suite: &Suite{
				ChartPath: chartPath,
				NamedChecks: []NamedCheck{
					{
						Name: "Noop",
					},
				},
				Cases: []Case{
					{
						Name:            "Using Defaults",
						TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace),
					},
				},
			},
		},
		{
			Name: "Simple Chart With Nil TemplateOptions In FailureCase But Failing DefaultValues",
			Suite: &Suite{
				ChartPath:     simpleChartPath,
				DefaultValues: chart.NewValues().Set("shouldFail", "true"),
				FailureCases: []FailureCase{
					{
						Name:            "No Options",
						TemplateOptions: nil,
						FailureMessage:  ".Values.shouldFail is set to true",
					},
				},
			},
		},
		{
			Name: "Example Chart Override Value With PreCheck",
			Suite: &Suite{
				ChartPath: chartPath,
				PreCheck: func(tc *checker.TestContext) {
					renderValuesMap := tc.RenderValues.AsMap()
					renderValuesMap["Values"].(helmChartUtil.Values)["hello"] = "rancher"
				},
				NamedChecks: []NamedCheck{
					{
						Name: "Check Value Of .Values.hello",
						Checks: Checks{
							checker.Once(func(tc *checker.TestContext) {
								assert.Equal(t, "rancher", checker.MustRenderValue[string](tc, ".Values.hello"))
							}),
						},
					},
				},
				Cases: []Case{
					{
						Name:            "Using Defaults",
						TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace),
					},
					{
						Name:            "Using Override",
						TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace).SetValue("hello", "world"),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			opts := &SuiteOptions{
				Coverage: CoverageOptions{
					Disabled: true,
				},
			}
			if !tc.ShouldThrowError {
				tc.Suite.Run(t, opts)
				return
			}
			fakeT := &testing.T{}
			tc.Suite.Run(fakeT, opts)
			assert.True(t, fakeT.Failed(), "expected error to be thrown")
		})
	}

	t.Run("Full Coverage", func(t *testing.T) {
		suite := &Suite{
			ChartPath: simpleChartPath,

			NamedChecks: []NamedCheck{
				{
					Name: "Noop",

					Covers: []string{".Values.data"},

					Checks: Checks{
						checker.NewChainedCheckFunc[struct{}](nil),
					},
				},
			},

			Cases: []Case{
				{
					Name: "Using Defaults",

					TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace),
				},
				{
					Name: "Set Data",

					TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace).
						Set("data", map[string]string{"hello": "world"}),
				},
			},

			FailureCases: []FailureCase{
				{
					Name: "Set .Values.shouldFail",

					TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace).
						SetValue("shouldFail", "true"),

					Covers: []string{
						".Values.shouldFail",
					},

					FailureMessage: ".Values.shouldFail is set to true",
				},
				{
					Name: "Set .Values.shouldFailRequired",

					TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace).
						SetValue("shouldFailRequired", "true"),

					Covers: []string{
						".Values.shouldFailRequired",
					},

					FailureMessage: ".Values.shouldFailRequired is set to true",
				},
			},
		}
		suite.Run(t, nil)
	})

	t.Run("OmitCases", func(t *testing.T) {
		visitedTests := map[string]bool{}
		collectTest := func(tc *checker.TestContext) {
			name := strings.TrimPrefix(tc.T.Name(), "TestRun/OmitCases/")
			visitedTests[name] = true
		}
		suite := &Suite{
			ChartPath: chartPath,
			NamedChecks: []NamedCheck{
				{
					Name: "Default",
					Checks: Checks{
						checker.Once(collectTest),
					},
				},
				{
					Name: "Collector",
					Checks: Checks{
						checker.Once(collectTest),
					},
				},
			},
			Cases: []Case{
				{
					Name:            "Using Defaults",
					TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace),
				},
				{
					Name:            "Using Debug",
					TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace).SetValue("args[0]", "--debug"),
				},
				{
					Name:            "Another Case",
					TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace).SetValue("args[0]", "--debug"),
				},
				{
					Name:            "Hello World",
					TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace).SetValue("args[0]", "--debug"),
				},
				{
					Name:            "Omit Collector",
					TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace).SetValue("args[0]", "--debug"),
					OmitNamedChecks: []string{"Collector"},
				},
			},
		}
		opts := &SuiteOptions{
			Coverage: CoverageOptions{
				Disabled: true,
			},
			YAMLLint: YamlLintOptions{
				Enabled: true,
			},
		}
		suite.Run(t, opts)
		var visitedTestsSlice []string
		for visitedTest := range visitedTests {
			visitedTestsSlice = append(visitedTestsSlice, visitedTest)
		}
		sort.Strings(visitedTestsSlice)
		assert.Equal(t, []string{
			"Another_Case/Collector",
			"Another_Case/Default",
			"Hello_World/Collector",
			"Hello_World/Default",
			"Omit_Collector/Default",
			"Using_Debug/Collector",
			"Using_Debug/Default",
			"Using_Defaults/Collector",
			"Using_Defaults/Default",
		}, visitedTestsSlice)
	})
}

func TestGetRancherOptions(t *testing.T) {
	o := GetRancherOptions()
	assert.NotNil(t, o, "RancherOptions should not be nil")
	if t.Failed() {
		return
	}
	assert.NotNil(t, o.HelmLint, "RancherOptions.HelmLint should not be nil")
	if t.Failed() {
		return
	}
	assert.True(t, o.HelmLint.Rancher.Enabled, "RancherOptions.HelmLint.Rancher.Enabled is false")
	if t.Failed() {
		return
	}
}
