package test

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/checker"
)

type Suite struct {
	ChartPath     string
	ValuesStruct  interface{}
	DefaultChecks []checker.Check
	Cases         []Case
}

func GetRancherOptions() *SuiteOptions {
	return &SuiteOptions{
		HelmLint: &chart.HelmLintOptions{
			Rancher: chart.RancherHelmLintOptions{
				Enabled: true,
			},
		},
	}
}

type SuiteOptions struct {
	HelmLint                      *chart.HelmLintOptions
	DoNotModifyChartSchemaInPlace bool
}

func (o *SuiteOptions) setDefaults() *SuiteOptions {
	if o == nil {
		o = &SuiteOptions{}
	}
	if o.HelmLint == nil {
		o.HelmLint = &chart.HelmLintOptions{}
	}
	return o
}

type Case struct {
	Name            string
	TemplateOptions *chart.TemplateOptions
	Checks          []checker.Check
}

func (s *Suite) Run(t *testing.T, opts *SuiteOptions) {
	opts = opts.setDefaults()
	if s.ValuesStruct == nil {
		s.ValuesStruct = struct{}{}
	}
	c, err := chart.NewChart(s.ChartPath)
	if err != nil {
		t.Error(err)
		return
	}
	t.Run("SchemaMustMatchStruct", func(t *testing.T) {
		c.SchemaMustMatchStruct(t, s.ValuesStruct, !opts.DoNotModifyChartSchemaInPlace)
	})
	for _, tc := range s.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			template, err := c.RenderTemplate(tc.TemplateOptions)
			if err != nil {
				t.Error(err)
				return
			}
			t.Run("HelmLint", func(t *testing.T) {
				template.HelmLint(t, opts.HelmLint)
			})
			t.Run("YamlLint", template.YamlLint)
			for _, check := range s.DefaultChecks {
				t.Run(check.Name, func(t *testing.T) {
					template.Check(t, check.Options, check.Func)
				})
			}
			for _, check := range tc.Checks {
				t.Run(check.Name, func(t *testing.T) {
					template.Check(t, check.Options, check.Func)
				})
			}
		})
	}
}
