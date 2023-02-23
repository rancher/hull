package test

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
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
			}
		})
	}
}
