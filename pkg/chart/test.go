package chart

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/checker"
)

type TestSuite struct {
	ChartPath    string
	ValuesStruct interface{}
	Cases        []TestCase
}

type TestCase struct {
	Name            string
	TemplateOptions *TemplateOptions
	Checks          []checker.Check
}

func (s *TestSuite) Run(t *testing.T, helmLintOpts *HelmLintOptions) {
	c, err := NewChart(s.ChartPath)
	if err != nil {
		t.Error(err)
		return
	}
	t.Run("StructMatchesValuesSchema", func(t *testing.T) {
		c.MatchesValuesSchema(t, s.ValuesStruct)
	})
	for _, tc := range s.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			template, err := c.RenderTemplate(tc.TemplateOptions)
			if err != nil {
				t.Error(err)
				return
			}
			t.Run("HelmLint", func(t *testing.T) {
				template.HelmLint(t, helmLintOpts)
			})
			t.Run("YamlLint", template.YamlLint)
			for _, check := range tc.Checks {
				t.Run(check.Name, func(t *testing.T) {
					template.Check(t, check.Options, check.Func)
				})
			}
		})
	}
}
