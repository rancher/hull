package test

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/checker"
	"github.com/aiyengar2/hull/pkg/utils"
	"github.com/stretchr/testify/assert"
)

const (
	defaultReleaseName = "with-schema"
	defaultNamespace   = "default"
)

var (
	chartPath = utils.MustGetPathFromModuleRoot("testdata", "charts", "with-schema")
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
			Name: "With Schema",
			Suite: &Suite{
				ChartPath: chartPath,
				DefaultChecks: []checker.Check{
					{
						Name: "Noop Default",
						Func: func(*testing.T, struct{}) {},
					},
				},
				Cases: []Case{
					{
						Name:            "Using Defaults",
						TemplateOptions: chart.NewTemplateOptions(defaultReleaseName, defaultNamespace),
						Checks: []checker.Check{
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
