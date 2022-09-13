package unit

import (
	"path/filepath"
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/checker"
	"github.com/aiyengar2/hull/pkg/checker/resource"
	"github.com/stretchr/testify/assert"
	helmValues "helm.sh/helm/v3/pkg/cli/values"
)

func TestLint(t *testing.T) {
	c, err := chart.NewChart(
		filepath.Join("..", "testdata", "charts", "example-chart"),
	)
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		Name            string
		TemplateOptions *chart.TemplateOptions
	}{
		{
			Name: "Default",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			template, err := c.RenderTemplate(tc.TemplateOptions)
			if err != nil {
				t.Error(err)
				return
			}
			t.Run("HelmLint", template.HelmLint)
			t.Run("YamlLint", template.YamlLint)
		})
	}
}

type testCase struct {
	Name            string
	TemplateOptions *chart.TemplateOptions
	Checks          []check
}

type check struct {
	Name    string
	Func    interface{}
	Options *checker.Options
}

func TestUnit(t *testing.T) {
	c, err := chart.NewChart(
		filepath.Join("..", "testdata", "charts", "example-chart"),
	)
	if err != nil {
		t.Fatal(err)
	}
	testCases := []testCase{
		{
			Name: "Check .Values.data",
			TemplateOptions: &chart.TemplateOptions{
				ValuesOptions: &helmValues.Options{
					Values: []string{
						"data.hello=world",
					},
				},
			},
			Checks: []check{
				{
					Name: "Can Set Data",
					Func: checkIfConfigMapsHaveData(map[string]string{
						"config": "hello: world",
					}),
					Options: nil,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			template, err := c.RenderTemplate(tc.TemplateOptions)
			if err != nil {
				t.Fatal(err)
			}
			for _, check := range tc.Checks {
				t.Run(check.Name, func(t *testing.T) {
					template.Check(t, check.Options, check.Func)
				})
			}
		})
	}
}

type configmaps struct {
	ConfigMaps resource.ConfigMaps
}

func checkIfConfigMapsHaveData(data map[string]string) interface{} {
	return func(t *testing.T, cms configmaps) {
		for _, cm := range cms.ConfigMaps {
			assert.Equal(t, cm.Data, data)
		}
	}
}
