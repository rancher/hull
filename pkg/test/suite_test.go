package test

import (
	"sort"
	"strings"
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/utils"
	"github.com/stretchr/testify/assert"
)

const (
	defaultReleaseName = "example-chart"
	defaultNamespace   = "default"
)

var (
	chartPath = utils.MustGetPathFromModuleRoot("testdata", "charts", "example-chart")
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
			Name: "Invalid Chart",
			Suite: &Suite{
				ChartPath: "",
			},
			ShouldThrowError: true,
		},
		{
			Name: "Example Chart",
			Suite: &Suite{
				ChartPath: chartPath,
				TemplateChecks: []TemplateCheck{
					{
						Name: "Noop Default",
						Func: func(*testing.T, struct{}) {},
					},
				},
				Cases: []Case{
					{
						Name:            "Using Defaults",
						TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace),
						ValueChecks: []ValueCheck{
							{
								Name: "Noop",
								Func: func(*testing.T, struct{}) {},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if !tc.ShouldThrowError {
				tc.Suite.Run(t, nil)
				return
			}
			fakeT := &testing.T{}
			tc.Suite.Run(fakeT, nil)
			assert.True(t, fakeT.Failed(), "expected error to be thrown")
		})
	}

	t.Run("OmitCases", func(t *testing.T) {
		visitedTests := map[string]bool{}
		collectTest := func(t *testing.T, _ struct{}) {
			name := strings.TrimPrefix(t.Name(), "TestRun/OmitCases/")
			visitedTests[name] = true
		}
		suite := &Suite{
			ChartPath: chartPath,
			TemplateChecks: []TemplateCheck{
				{
					Name: "Run On All",
					Func: collectTest,
				},
				{
					Name:      "Omit Debug",
					Func:      collectTest,
					OmitCases: []string{"Using Debug"},
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
			},
		}
		suite.Run(t, nil)
		var visitedTestsSlice []string
		for visitedTest := range visitedTests {
			visitedTestsSlice = append(visitedTestsSlice, visitedTest)
		}
		sort.Strings(visitedTestsSlice)
		assert.Equal(t, []string{
			"Another_Case/Omit_Debug",
			"Another_Case/Run_On_All",

			"Hello_World/Omit_Debug",
			"Hello_World/Run_On_All",

			"Using_Debug/Run_On_All",

			"Using_Defaults/Omit_Debug",
			"Using_Defaults/Run_On_All",
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
