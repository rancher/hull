package coverage

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/tpl"
	"github.com/aiyengar2/hull/pkg/tpl/parse"
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
				FieldUsage: map[string]TemplateTracker{
					".Values.hello": {
						"configmap.yaml": false,
						"complex.yaml":   false,
					},
					".Values.world": {
						"deployment.yaml": false,
						"complex.yaml":    false,
					},
					".Values.cattle": {
						"example-chart.name : system_default_registry : complex.yaml": false,
					},
					".Values.rancher": {
						"system_default_registry : complex.yaml": false,
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
					Covers:          []string{"configmap.yaml"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("world", "hello"),
					Covers:          []string{"deployment.yaml"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("rancher", "hello"),
					Covers:          []string{"complex.yaml"},
				},
			},

			Expect: &Tracker{
				FieldUsage: map[string]TemplateTracker{
					".Values.hello": {
						"configmap.yaml": true,
						"complex.yaml":   false,
					},
					".Values.world": {
						"deployment.yaml": true,
						"complex.yaml":    false,
					},
					".Values.cattle": {
						"example-chart.name : system_default_registry : complex.yaml": false,
					},
					".Values.rancher": {
						"system_default_registry : complex.yaml": true,
					},
				},
			},
			Coverage: float64(3) / float64(6),
		},
		{
			Name: "Usage With Partial Coverage With Globs",
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
					Covers:          []string{"*"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("world", "hello"),
					Covers:          []string{"*"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("rancher", "hello"),
					Covers:          []string{"*"},
				},
			},

			Expect: &Tracker{
				FieldUsage: map[string]TemplateTracker{
					".Values.hello": {
						"configmap.yaml": true,
						"complex.yaml":   true,
					},
					".Values.world": {
						"deployment.yaml": true,
						"complex.yaml":    true,
					},
					".Values.cattle": {
						"example-chart.name : system_default_registry : complex.yaml": false,
					},
					".Values.rancher": {
						"system_default_registry : complex.yaml": true,
					},
				},
			},
			Coverage: float64(5) / float64(6),
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
					Covers:          []string{"*"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("world", "hello"),
					Covers:          []string{"*"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("rancher", "hello"),
					Covers:          []string{"*"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("doesNot", "exist"),
					Covers:          []string{"*"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("doesNot", "exist"),
					Covers:          nil,
				},
				{
					TemplateOptions: &chart.TemplateOptions{},
					Covers:          []string{"*"},
				},
				{
					TemplateOptions: nil,
					Covers:          []string{"*"},
				},
			},

			Expect: &Tracker{
				FieldUsage: map[string]TemplateTracker{
					".Values.hello": {
						"configmap.yaml": true,
						"complex.yaml":   true,
					},
					".Values.world": {
						"deployment.yaml": true,
						"complex.yaml":    true,
					},
					".Values.cattle": {
						"example-chart.name : system_default_registry : complex.yaml": false,
					},
					".Values.rancher": {
						"system_default_registry : complex.yaml": true,
					},
				},
			},
			Coverage: float64(5) / float64(6),
		},
		{
			Name: "Usage With Full Coverage Without Globs",
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
					Covers:          []string{"configmap.yaml", "complex.yaml"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("world", "hello"),
					Covers:          []string{"complex.yaml", "deployment.yaml"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("rancher", "hello"),
					Covers:          []string{"complex.yaml"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("cattle", "world"),
					Covers:          []string{"complex.yaml"},
				},
			},

			Expect: &Tracker{
				FieldUsage: map[string]TemplateTracker{
					".Values.hello": {
						"configmap.yaml": true,
						"complex.yaml":   true,
					},
					".Values.world": {
						"deployment.yaml": true,
						"complex.yaml":    true,
					},
					".Values.cattle": {
						"example-chart.name : system_default_registry : complex.yaml": true,
					},
					".Values.rancher": {
						"system_default_registry : complex.yaml": true,
					},
				},
			},
			Coverage: 1,
		},
		{
			Name: "Usage With Full Coverage With Globs",
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
					Covers:          []string{"*"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("world", "hello"),
					Covers:          []string{"*"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("rancher", "hello"),
					Covers:          []string{"*"},
				},
				{
					TemplateOptions: chart.NewTemplateOptions("example-chart", "default").SetValue("cattle", "world"),
					Covers:          []string{"*"},
				},
			},

			Expect: &Tracker{
				FieldUsage: map[string]TemplateTracker{
					".Values.hello": {
						"configmap.yaml": true,
						"complex.yaml":   true,
					},
					".Values.world": {
						"deployment.yaml": true,
						"complex.yaml":    true,
					},
					".Values.cattle": {
						"example-chart.name : system_default_registry : complex.yaml": true,
					},
					".Values.rancher": {
						"system_default_registry : complex.yaml": true,
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
			Expect:  &Tracker{},
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
