package coverage

import (
	"testing"

	"github.com/rancher/hull/pkg/chart"
	"github.com/rancher/hull/pkg/tpl"
	"github.com/rancher/hull/pkg/tpl/parse"
	"github.com/stretchr/testify/assert"
)

func TestTracker(t *testing.T) {
	type Record struct {
		TemplateOptions *chart.TemplateOptions
		Covers          []string
	}
	testCases := []struct {
		Name string

		Usage            *tpl.TemplateUsage
		IncludeSubcharts bool
		Records          []Record

		Expect   *Tracker
		Coverage float64
	}{
		{
			Name: "Nil Usage",
		},
		{
			Name: "Usage Without Coverage",
			Usage: &tpl.TemplateUsage{
				Files: map[string]*parse.Result{
					"configmap.yaml": {
						Fields: []string{".Capabilities.KubeVersion", ".Values.hello"},
					},
					"deployment.yaml": {
						Fields: []string{".Chart.Name", ".Release.Namespace", ".Values.world"},
					},
					"complex.yaml": {
						Fields:        []string{".Values.hello", ".Values.world"},
						TemplateCalls: []string{"system_default_registry"},
					},
				},
				NamedTemplates: map[string]*parse.Result{
					"system_default_registry": {
						Fields:        []string{".Values.rancher"},
						TemplateCalls: []string{"example-chart.name"},
					},
					"example-chart.name": {
						Fields: []string{".Values.cattle"},
					},
				},
			},

			Expect: &Tracker{
				FieldUsage: FieldTracker{
					".Values.hello": {
						Templates: []string{"complex.yaml", "configmap.yaml"},
					},
					".Values.world": {
						Templates: []string{"complex.yaml", "deployment.yaml"},
					},
					".Values.cattle : example-chart.name : system_default_registry": {
						Templates: []string{"complex.yaml"},
					},
					".Values.rancher : system_default_registry": {
						Templates: []string{"complex.yaml"},
					},
				},
			},
			Coverage: 0,
		},
		{
			Name: "Usage With Partial Coverage",
			Usage: &tpl.TemplateUsage{
				Files: map[string]*parse.Result{
					"configmap.yaml": {
						Fields: []string{".Capabilities.KubeVersion", ".Values.hello"},
					},
					"deployment.yaml": {
						Fields: []string{".Chart.Name", ".Release.Namespace", ".Values.world"},
					},
					"complex.yaml": {
						Fields:        []string{".Values.hello", ".Values.world"},
						TemplateCalls: []string{"system_default_registry"},
					},
				},
				NamedTemplates: map[string]*parse.Result{
					"system_default_registry": {
						Fields:        []string{".Values.rancher"},
						TemplateCalls: []string{"example-chart.name"},
					},
					"example-chart.name": {
						Fields: []string{".Values.cattle"},
					},
				},
			},
			Records: []Record{
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("hello", "world"),
					Covers:          []string{".Values.hello"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("rancher", "hello"),
					Covers:          []string{".Values.rancher"},
				},
			},

			Expect: &Tracker{
				FieldUsage: FieldTracker{
					".Values.hello": {
						Templates: []string{"complex.yaml", "configmap.yaml"},
						covered:   true,
					},
					".Values.world": {
						Templates: []string{"complex.yaml", "deployment.yaml"},
					},
					".Values.cattle : example-chart.name : system_default_registry": {
						Templates: []string{"complex.yaml"},
					},
					".Values.rancher : system_default_registry": {
						Templates: []string{"complex.yaml"},
						covered:   true,
					},
				},
			},
			Coverage: float64(3) / float64(6),
		},
		{
			Name: "Usage With Partial Coverage And Bad Tests",
			Usage: &tpl.TemplateUsage{
				Files: map[string]*parse.Result{
					"configmap.yaml": {
						Fields: []string{".Capabilities.KubeVersion", ".Values.hello"},
					},
					"deployment.yaml": {
						Fields: []string{".Chart.Name", ".Release.Namespace", ".Values.world"},
					},
					"complex.yaml": {
						Fields:        []string{".Values.hello", ".Values.world"},
						TemplateCalls: []string{"system_default_registry"},
					},
				},
				NamedTemplates: map[string]*parse.Result{
					"system_default_registry": {
						Fields:        []string{".Values.rancher"},
						TemplateCalls: []string{"example-chart.name"},
					},
					"example-chart.name": {
						Fields: []string{".Values.cattle"},
					},
				},
			},
			Records: []Record{
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("hello", "world"),
					Covers:          []string{".Values.rancher"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("world", "hello"),
					Covers:          []string{".Values.world"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("rancher", "hello"),
					Covers:          []string{".Values.hello"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("doesNot", "exist"),
					Covers:          []string{".Values.hello"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("doesNot", "exist"),
					Covers:          nil,
				},
				{
					TemplateOptions: &chart.TemplateOptions{},
					Covers:          []string{".Values.rancher"},
				},
				{
					TemplateOptions: nil,
					Covers:          []string{"random-string"},
				},
			},

			Expect: &Tracker{
				FieldUsage: FieldTracker{
					".Values.hello": {
						Templates: []string{"complex.yaml", "configmap.yaml"},
					},
					".Values.world": {
						Templates: []string{"complex.yaml", "deployment.yaml"},
						covered:   true,
					},
					".Values.cattle : example-chart.name : system_default_registry": {
						Templates: []string{"complex.yaml"},
					},
					".Values.rancher : system_default_registry": {
						Templates: []string{"complex.yaml"},
					},
				},
			},
			Coverage: float64(2) / float64(6),
		},
		{
			Name: "Usage With Full Coverage",
			Usage: &tpl.TemplateUsage{
				Files: map[string]*parse.Result{
					"configmap.yaml": {
						Fields: []string{".Capabilities.KubeVersion", ".Values.hello"},
					},
					"deployment.yaml": {
						Fields: []string{".Chart.Name", ".Release.Namespace", ".Values.world"},
					},
					"complex.yaml": {
						Fields:        []string{".Values.hello", ".Values.world"},
						TemplateCalls: []string{"system_default_registry"},
					},
				},
				NamedTemplates: map[string]*parse.Result{
					"system_default_registry": {
						Fields:        []string{".Values.rancher"},
						TemplateCalls: []string{"example-chart.name"},
					},
					"example-chart.name": {
						Fields: []string{".Values.cattle"},
					},
				},
			},
			Records: []Record{
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("hello", "world"),
					Covers:          []string{".Values.hello"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("world", "cattle"),
					Covers:          []string{".Values.world"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("cattle", "hello"),
					Covers:          []string{".Values.cattle"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("rancher", "cattle"),
					Covers:          []string{".Values.rancher"},
				},
			},

			Expect: &Tracker{
				FieldUsage: FieldTracker{
					".Values.hello": {
						Templates: []string{"complex.yaml", "configmap.yaml"},
						covered:   true,
					},
					".Values.world": {
						Templates: []string{"complex.yaml", "deployment.yaml"},
						covered:   true,
					},
					".Values.cattle : example-chart.name : system_default_registry": {
						Templates: []string{"complex.yaml"},
						covered:   true,
					},
					".Values.rancher : system_default_registry": {
						Templates: []string{"complex.yaml"},
						covered:   true,
					},
				},
			},
			Coverage: 1,
		},
		{
			Name: "Usage With Full Coverage With Templates",
			Usage: &tpl.TemplateUsage{
				Files: map[string]*parse.Result{
					"configmap.yaml": {
						Fields: []string{".Capabilities.KubeVersion", ".Values.hello"},
					},
					"deployment.yaml": {
						Fields: []string{".Chart.Name", ".Release.Namespace", ".Values.world"},
					},
					"complex.yaml": {
						Fields:        []string{".Values.hello", ".Values.world"},
						TemplateCalls: []string{"system_default_registry"},
					},
				},
				NamedTemplates: map[string]*parse.Result{
					"system_default_registry": {
						Fields:        []string{".Values.rancher"},
						TemplateCalls: []string{"example-chart.name"},
					},
					"example-chart.name": {
						Fields: []string{".Values.cattle"},
					},
				},
			},
			Records: []Record{
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("hello", "world"),
					Covers:          []string{".Values.hello"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("world", "cattle"),
					Covers:          []string{".Values.world"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("cattle", "hello"),
					Covers:          []string{"example-chart.name"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("rancher", "cattle"),
					Covers:          []string{"system_default_registry"},
				},
			},

			Expect: &Tracker{
				FieldUsage: FieldTracker{
					".Values.hello": {
						Templates: []string{"complex.yaml", "configmap.yaml"},
						covered:   true,
					},
					".Values.world": {
						Templates: []string{"complex.yaml", "deployment.yaml"},
						covered:   true,
					},
					".Values.cattle : example-chart.name : system_default_registry": {
						Templates: []string{"complex.yaml"},
						covered:   true,
					},
					".Values.rancher : system_default_registry": {
						Templates: []string{"complex.yaml"},
						covered:   true,
					},
				},
			},
			Coverage: 1,
		},
		{
			Name: "Only Excluded Subchart Without Coverage",
			Usage: &tpl.TemplateUsage{
				Files: map[string]*parse.Result{
					"charts/grafana/configmap.yaml": {
						Fields: []string{".Capabilities.KubeVersion", ".Values.hello"},
					},
				},
			},
			Records: []Record{},
			Expect: &Tracker{
				FieldUsage: NewFieldTracker(),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tracker := NewTracker(tc.Usage, tc.IncludeSubcharts)
			for _, record := range tc.Records {
				tracker.Record(record.TemplateOptions, record.Covers)
			}
			assert.Equal(t, tc.Expect, tracker)
			coverage, _ := tracker.CalculateCoverage()
			assert.Equal(t, tc.Coverage, coverage)
		})
	}
}
