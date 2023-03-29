package tpl

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	multierr "github.com/hashicorp/go-multierror"
	"github.com/rancher/hull/pkg/chart"
	"github.com/rancher/hull/pkg/tpl/parse"
	"github.com/rancher/hull/pkg/tpl/utils"
	helmChart "helm.sh/helm/v3/pkg/chart"
)

type TemplateUsage struct {
	Files          map[string]*parse.Result
	NamedTemplates map[string]*parse.Result
}

func (t *TemplateUsage) GetWarnings() error {
	var multiErr error
	for filename, result := range t.Files {
		if !result.EmitWarning {
			continue
		}
		multiErr = multierr.Append(multiErr, fmt.Errorf("file %s cannot be fully captured by coverage", filename))
	}
	for namedTemplate, result := range t.NamedTemplates {
		if !result.EmitWarning {
			continue
		}
		multiErr = multierr.Append(multiErr, fmt.Errorf("template %s cannot be fully captured by coverage", namedTemplate))
	}
	return multiErr
}

func CollectTemplateUsage(c chart.Chart) (*TemplateUsage, error) {
	// get helm chart
	ch := c.GetHelmChart()
	fileTemplates, namedTemplates, err := CollectAllTemplates(ch)
	if err != nil {
		return nil, err
	}
	result := &TemplateUsage{}
	for _, t := range fileTemplates {
		if t == nil {
			panic("expected non-nil templates to be collected from Helm chart")
		}
		name := t.Name()
		if strings.HasPrefix(filepath.Base(name), "_") {
			// ignore files like _helpers.tpl
			continue
		}
		if filepath.Ext(name) != ".yml" && filepath.Ext(name) != ".yaml" {
			// ignore files like NOTES.txt
			continue
		}
		if result.Files == nil {
			result.Files = make(map[string]*parse.Result)
		}
		result.Files[name] = parse.Template(t)
	}
	for _, t := range namedTemplates {
		if t == nil {
			panic("expected non-nil named templates to be collected from Helm chart")
		}
		name := t.Name()
		if result.NamedTemplates == nil {
			result.NamedTemplates = make(map[string]*parse.Result)
		}
		result.NamedTemplates[name] = parse.Template(t)
	}
	return result, nil
}

func CollectAllTemplates(c *helmChart.Chart) ([]*template.Template, []*template.Template, error) {
	var fileTemplates []*template.Template
	var namedTemplates []*template.Template
	var multiErr error
	funcMap := utils.GetNoopHelmFuncMap()
	for _, tpl := range CollectAllTemplateFiles(c) {
		if tpl == nil {
			panic("expected *helmChart.Chart to contain non-nil templates")
		}
		t, err := template.New(tpl.Name).Funcs(funcMap).Parse(string(tpl.Data))
		if err != nil {
			multiErr = multierr.Append(multiErr, err)
			continue
		}
		for _, t := range t.Templates() {
			if t.Name() == tpl.Name {
				fileTemplates = append(fileTemplates, t)
			} else {
				namedTemplates = append(namedTemplates, t)
			}
		}
	}
	return fileTemplates, namedTemplates, multiErr
}

func CollectAllTemplateFiles(c *helmChart.Chart) []*helmChart.File {
	var collectAllTemplateFiles func(c *helmChart.Chart, pathRelativeToRoot string) []*helmChart.File
	collectAllTemplateFiles = func(c *helmChart.Chart, pathRelativeToRoot string) []*helmChart.File {
		var allTemplateFiles []*helmChart.File
		for _, f := range c.Templates {
			allTemplateFiles = append(allTemplateFiles, &helmChart.File{
				Name: filepath.Join(pathRelativeToRoot, f.Name),
				Data: f.Data,
			})
		}
		for _, dep := range c.Dependencies() {
			depPath := filepath.Join(pathRelativeToRoot, "charts", dep.Name())
			allTemplateFiles = append(allTemplateFiles, collectAllTemplateFiles(dep, depPath)...)
		}
		return allTemplateFiles
	}
	return collectAllTemplateFiles(c, "")
}
