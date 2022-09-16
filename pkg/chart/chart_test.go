package chart

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	helmValues "helm.sh/helm/v3/pkg/cli/values"
)

func TestNewChart(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal("cannot find repository root at ../..")
	}

	testCases := []struct {
		Name             string
		ChartPath        string
		ShouldThrowError bool
	}{
		{
			Name:             "Valid Chart",
			ChartPath:        filepath.Join("..", "..", "testdata", "charts", "example-chart"),
			ShouldThrowError: false,
		},
		{
			Name:             "Valid Chart From Absolute Path",
			ChartPath:        filepath.Join(repoRoot, "testdata", "charts", "example-chart"),
			ShouldThrowError: false,
		},
		{
			Name:             "Invalid Chart",
			ChartPath:        filepath.Join("..", "..", "testdata", "charts", "does-not-exist"),
			ShouldThrowError: true,
		},
		{
			Name:             "Invalid Glob Path",
			ChartPath:        filepath.Join("..", "..", "testdata", "charts", "*"),
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
				t.Fatal("test case is invalid, chartPath provided is not a valid path")
			}
			assert.Equal(t, expectedChartPath, c.GetPath())
			assert.NotNil(t, c.GetHelmChart(), "did not load underlying helm chart")
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	c, err := NewChart(filepath.Join("..", "..", "testdata", "charts", "example-chart"))
	if err != nil {
		t.Error("unable to construct chart from chart path")
		return
	}
	if c == nil {
		t.Errorf("received nil chart")
		return
	}

	badC, err := NewChart(filepath.Join("..", "..", "testdata", "charts", "bad-chart"))
	if err != nil {
		t.Error("unable to construct chart from chart path")
		return
	}
	if badC == nil {
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
		{
			Name:  "Does Not Match Schema",
			Chart: c,
			Opts: &TemplateOptions{
				ValuesOptions: &helmValues.Options{
					Values: []string{"hello=hi"},
				},
			},
			ShouldThrowError: true,
		},
		{
			Name:             "Bad Chart",
			Chart:            badC,
			Opts:             nil,
			ShouldThrowError: false,
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

func TestMatchesValuesSchema(t *testing.T) {
	testCases := []struct {
		Name             string
		ChartPath        string
		Struct           interface{}
		ShouldThrowError bool
	}{
		{
			Name:      "Example Chart With Valid Schema Struct",
			ChartPath: filepath.Join("..", "..", "testdata", "charts", "example-chart"),
			Struct: struct {
				Data map[string]interface{} `jsonschema:"description=Data to be inserted into a ConfigMap"`
			}{},
			ShouldThrowError: false,
		},
		{
			Name:      "No Schema With Valid Schema Struct",
			ChartPath: filepath.Join("..", "..", "testdata", "charts", "no-schema"),
			Struct: struct {
				Data map[string]interface{} `jsonschema:"description=Data to be inserted into a ConfigMap"`
			}{},
			ShouldThrowError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			c, err := NewChart(tc.ChartPath)
			if err != nil {
				t.Error("unable to construct chart from chart path")
				return
			}
			if c == nil {
				t.Errorf("received nil chart")
				return
			}
			if tc.ShouldThrowError {
				fakeT := &testing.T{}
				c.MatchesValuesSchema(fakeT, tc.Struct)
				if !fakeT.Failed() {
					t.Errorf("expected error to be thrown")
				}
				return
			}
			c.MatchesValuesSchema(t, tc.Struct)
		})
	}
}
