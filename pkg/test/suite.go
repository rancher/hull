package test

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/checker"
)

type Suite struct {
	ChartPath     string
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
	HelmLint *chart.HelmLintOptions
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
	c, err := chart.NewChart(s.ChartPath)
	if err != nil {
		t.Error(err)
		return
	}
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
					template.Check(t, check.Func)
				})
			}
			for _, check := range tc.Checks {
				t.Run(check.Name, func(t *testing.T) {
					template.Check(t, check.Func)
				})
			}
		})
	}
}
