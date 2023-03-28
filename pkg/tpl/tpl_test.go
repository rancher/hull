package tpl

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/rancher/hull/pkg/chart"
	"github.com/rancher/hull/pkg/tpl/parse"
	"github.com/rancher/hull/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetWarnings(t *testing.T) {
	exampleChart, err := chart.NewChart(utils.MustGetPathFromModuleRoot("testdata", "charts", "example-chart"))
	if err != nil {
		t.Fatal(err)
	}
	exampleTemplateUsage, err := CollectTemplateUsage(exampleChart)
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		Name             string
		TemplateUsage    *TemplateUsage
		Expect           []error
		ShouldThrowError bool
	}{
		{
			Name:          "No Template Usage",
			TemplateUsage: &TemplateUsage{},
			Expect:        nil,
		},
		{
			Name:          "Example Chart",
			TemplateUsage: exampleTemplateUsage,
			Expect:        nil,
		},
		{
			Name: "Has Warnings In Files",
			TemplateUsage: &TemplateUsage{
				Files: map[string]*parse.Result{
					"myfile.yaml": {
						EmitWarning: true,
					},
				},
			},
			Expect: []error{
				fmt.Errorf("file myfile.yaml cannot be fully captured by coverage"),
			},
		},
		{
			Name: "Has Warnings In NamedTemplates",
			TemplateUsage: &TemplateUsage{
				NamedTemplates: map[string]*parse.Result{
					"my-template": {
						EmitWarning: true,
					},
				},
			},
			Expect: []error{
				fmt.Errorf("template my-template cannot be fully captured by coverage"),
			},
		},
		{
			Name: "Has Warnings In Both",
			TemplateUsage: &TemplateUsage{
				Files: map[string]*parse.Result{
					"myfile.yaml": {
						EmitWarning: true,
					},
				},
				NamedTemplates: map[string]*parse.Result{
					"my-template": {
						EmitWarning: true,
					},
				},
			},
			Expect: []error{
				fmt.Errorf("file myfile.yaml cannot be fully captured by coverage"),
				fmt.Errorf("template my-template cannot be fully captured by coverage"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			multierr := tc.TemplateUsage.GetWarnings()
			if multierr == nil {
				assert.Nil(t, tc.Expect)
			} else {
				assert.Equal(t, tc.Expect, multierr.(*multierror.Error).WrappedErrors())
			}
		})
	}
}

func TestCollectTemplateUsage(t *testing.T) {
	testCases := []struct {
		Name             string
		ChartPath        string
		Expect           *TemplateUsage
		ShouldThrowError bool
	}{
		{
			Name:             "Nonexistent Chart",
			ChartPath:        utils.MustGetPathFromModuleRoot("testdata", "charts", "does-not-exist"),
			ShouldThrowError: true,
		},
		{
			Name:      "No Templates",
			ChartPath: utils.MustGetPathFromModuleRoot("testdata", "charts", "no-templates"),
			Expect:    &TemplateUsage{},
		},
		{
			Name:      "No Templates With Subchart",
			ChartPath: utils.MustGetPathFromModuleRoot("testdata", "charts", "no-templates-with-subchart"),
			Expect:    &TemplateUsage{},
		},
		{
			Name:             "Bad Templates",
			ChartPath:        utils.MustGetPathFromModuleRoot("testdata", "charts", "bad-templates"),
			ShouldThrowError: true,
		},
		{
			Name:      "Example Chart",
			ChartPath: utils.MustGetPathFromModuleRoot("testdata", "charts", "example-chart"),
			Expect: &TemplateUsage{
				Files: map[string]*parse.Result{
					"templates/clusterrole.yaml": {
						Fields: []string{
							".Values.global.rbac.create",
							".Values.global.rbac.userRoles.aggregateToDefaultRoles",
							".Values.global.rbac.userRoles.create",
						},
						TemplateCalls: []string{
							"example-chart.labels",
							"example-chart.name",
						},
					},
					"templates/deployment.yaml": {
						Fields: []string{
							".Release.Name",
							".Values.args",
							".Values.image.pullPolicy",
							".Values.image.repository",
							".Values.image.tag",
							".Values.nodeSelector",
							".Values.resources",
							".Values.securityContext",
							".Values.tolerations",
						},
						TemplateCalls: []string{
							"example-chart.labels",
							"example-chart.name",
							"example-chart.namespace",
							"linux-node-selector",
							"linux-node-tolerations",
							"system_default_registry",
						},
					},
					"templates/hardened.yaml": {
						Fields: []string{
							".Release.Namespace",
							".Values.global.kubectl.pullPolicy",
							".Values.global.kubectl.repository",
							".Values.global.kubectl.tag",
						},
						TemplateCalls: []string{
							"example-chart.labels",
							"example-chart.name",
							"linux-node-selector",
							"linux-node-tolerations",
							"system_default_registry",
						},
					},
					"templates/psp.yaml": {
						Fields: []string{
							".Capabilities.KubeVersion.GitVersion",
							".Values.global.cattle.psp.enabled",
							".Values.global.rbac.pspAnnotations",
						},
						TemplateCalls: []string{
							"example-chart.labels",
							"example-chart.name",
							"example-chart.namespace",
						},
					},
					"templates/rbac.yaml": {
						Fields: []string{
							".Values.global.imagePullSecrets",
						},
						TemplateCalls: []string{
							"example-chart.labels",
							"example-chart.name",
							"example-chart.namespace",
						},
					},
				},
				NamedTemplates: map[string]*parse.Result{
					"example-chart.chartref": {
						Fields: []string{
							".Chart.Name",
							".Chart.Version",
						},
					},
					"example-chart.labels": {
						Fields: []string{
							".Chart.Version",
							".Release.Name",
							".Release.Service",
							".Values.commonLabels",
						},
						TemplateCalls: []string{
							"example-chart.chartref",
							"example-chart.name",
						},
					},
					"example-chart.name": {
						Fields: []string{
							".Chart.Name",
							".Values.nameOverride",
						},
					},
					"example-chart.namespace": {
						Fields: []string{
							".Release.Namespace",
							".Values.namespaceOverride",
						},
					},
					"linux-node-selector": {
						Fields: []string{
							".Capabilities.KubeVersion.GitVersion",
						},
					},
					"linux-node-tolerations": {},
					"system_default_registry": {
						Fields: []string{
							".Values.global.cattle.systemDefaultRegistry",
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			c, err := chart.NewChart(tc.ChartPath)
			if err != nil {
				assert.True(t, tc.ShouldThrowError, "unexpected error: %s", err)
				return
			}
			templateUsage, err := CollectTemplateUsage(c)
			if err != nil {
				assert.True(t, tc.ShouldThrowError, "unexpected error: %s", err)
				return
			}
			if err == nil {
				assert.False(t, tc.ShouldThrowError, "expected error to be thrown, found templateUsage %v", templateUsage)
			}
			if t.Failed() {
				return
			}
			assert.Equal(t, tc.Expect, templateUsage)
		})
	}
}
