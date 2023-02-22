package chart

import (
	"path/filepath"
	"testing"

	"github.com/aiyengar2/hull/pkg/utils"
	"github.com/stretchr/testify/assert"
	helmValues "helm.sh/helm/v3/pkg/cli/values"
)

func TestNewChart(t *testing.T) {
	testCases := []struct {
		Name             string
		ChartPath        string
		ShouldThrowError bool
	}{
		{
			Name:             "Valid Chart",
			ChartPath:        utils.MustGetPathFromModuleRoot("testdata", "charts", "example-chart"),
			ShouldThrowError: false,
		},
		{
			Name:             "Valid Chart From Absolute Path",
			ChartPath:        utils.MustGetPathFromModuleRoot("testdata", "charts", "example-chart"),
			ShouldThrowError: false,
		},
		{
			Name:             "Invalid Chart",
			ChartPath:        utils.MustGetPathFromModuleRoot("testdata", "charts", "does-not-exist"),
			ShouldThrowError: true,
		},
		{
			Name:             "Invalid Glob Path",
			ChartPath:        utils.MustGetPathFromModuleRoot("testdata", "charts", "*"),
			ShouldThrowError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			c, err := NewChart(tc.ChartPath)
			if tc.ShouldThrowError {
				if err == nil {
					t.Errorf("expected error to be thrown")
				}
				return
			}
			if c == nil {
				t.Errorf("received nil chart")
				return
			}
			expectedChartPath, err := filepath.Abs(tc.ChartPath)
			if err != nil {
				t.Error("test case is invalid, chartPath provided is not a valid path")
				return
			}
			assert.Equal(t, expectedChartPath, c.GetPath())
			assert.NotNil(t, c.GetHelmChart(), "did not load underlying helm chart")
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	chartPath := utils.MustGetPathFromModuleRoot("testdata", "charts", "with-schema")
	c, err := NewChart(chartPath)
	if err != nil {
		t.Errorf("unable to construct chart from chart path %s: %s", chartPath, err)
		return
	}
	if c == nil {
		t.Errorf("received nil chart")
		return
	}

	testCases := []struct {
		Name             string
		Chart            Chart
		Opts             *TemplateOptions
		ShouldThrowError bool
	}{
		{
			Name:             "Nil Options",
			Chart:            c,
			Opts:             nil,
			ShouldThrowError: false,
		},
		{
			Name:  "Bad Values",
			Chart: c,
			Opts: &TemplateOptions{
				ValuesOptions: &helmValues.Options{
					Values: []string{"i-am-a-bad-option#2@"},
				},
			},
			ShouldThrowError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			template, err := c.RenderTemplate(tc.Opts)
			if tc.ShouldThrowError {
				if err == nil {
					t.Errorf("expected error to be thrown")
				}
				return
			}
			if err != nil {
				t.Error(err)
				return
			}
			assert.NotNil(t, template)
		})
	}
}
