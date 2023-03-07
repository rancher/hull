package test

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/test/coverage"
	"github.com/aiyengar2/hull/pkg/tpl"
	"github.com/stretchr/testify/assert"

	helmValues "helm.sh/helm/v3/pkg/cli/values"
)

type Suite struct {
	ChartPath      string
	TemplateChecks []TemplateCheck
	Cases          []Case
}

type Case struct {
	Name            string
	TemplateOptions *chart.TemplateOptions
	ValueChecks     []ValueCheck
}

func (s *Suite) setDefaults() *Suite {
	for i := range s.Cases {
		if s.Cases[i].TemplateOptions == nil {
			s.Cases[i].TemplateOptions = &chart.TemplateOptions{}
		}
		if s.Cases[i].TemplateOptions.ValuesOptions == nil {
			s.Cases[i].TemplateOptions.ValuesOptions = &helmValues.Options{}
		}
	}
	return s
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
	YAMLLint YamlLintOptions
	Coverage CoverageOptions
}

type YamlLintOptions struct {
	Enabled bool
}

type CoverageOptions struct {
	IncludeSubcharts bool
	Disabled         bool
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

func (s *Suite) Run(t *testing.T, opts *SuiteOptions) {
	s = s.setDefaults()
	opts = opts.setDefaults()
	c, err := chart.NewChart(s.ChartPath)
	if err != nil {
		t.Error(err)
		return
	}
	templateUsage, err := tpl.CollectTemplateUsage(c)
	if err != nil {
		t.Error(err)
		return
	}
	if templateUsage == nil {
		t.Errorf("templateUsage is nil")
		return
	}
	coverageTracker := coverage.NewTracker(templateUsage, opts.Coverage.IncludeSubcharts)
	for _, tc := range s.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			template, err := c.RenderTemplate(tc.TemplateOptions)
			if err != nil {
				t.Errorf("failed to render template: %s", err)
				return
			}
			t.Run("HelmLint", func(t *testing.T) {
				template.HelmLint(t, opts.HelmLint)
			})
			if opts.YAMLLint.Enabled {
				t.Run("YamlLint", template.YamlLint)
			}
			for _, check := range s.TemplateChecks {
				// skip cases if necessary
				var skip bool
				for _, omitCase := range check.OmitCases {
					if tc.Name == omitCase {
						skip = true
					}
				}
				if skip {
					continue
				}
				t.Run(check.Name, func(t *testing.T) {
					template.Check(t, check.Func)
				})
			}
			for _, check := range tc.ValueChecks {
				t.Run(check.Name, func(t *testing.T) {
					template.Check(t, check.Func)
				})
				if err := coverageTracker.Record(tc.TemplateOptions, check.Covers); err != nil {
					t.Errorf("failed to track coverage: %s", err)
					// do not fail out, you should still continue with other checks
				}
			}
		})
	}
	if opts.Coverage.Disabled {
		return
	}
	t.Run("Coverage", func(t *testing.T) {
		coverage, report := coverageTracker.CalculateCoverage()
		assert.Equal(t, 1.00, coverage, report)
		if !t.Failed() {
			t.Log(report)
		}
		if err := templateUsage.GetWarnings(); err != nil {
			t.Log(err)
		}
	})
}
