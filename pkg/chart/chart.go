package chart

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rancher/helm-locker/pkg/objectset/parser"
	"github.com/rancher/wrangler/pkg/objectset"
	helmChart "helm.sh/helm/v3/pkg/chart"
	helmLoader "helm.sh/helm/v3/pkg/chart/loader"
	helmChartUtil "helm.sh/helm/v3/pkg/chartutil"
	helmEngine "helm.sh/helm/v3/pkg/engine"
)

type Chart interface {
	GetPath() string
	GetHelmChart() *helmChart.Chart
	RenderTemplate(opts *TemplateOptions) (Template, error)
}

type chart struct {
	*helmChart.Chart

	Path string
}

func NewChart(path string) (Chart, error) {
	c := &chart{
		Path: path,
	}
	var err error
	c.Chart, err = helmLoader.Load(c.Path)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *chart) GetPath() string {
	return c.Path
}

func (c *chart) GetHelmChart() *helmChart.Chart {
	return c.Chart
}

func (c *chart) RenderTemplate(opts *TemplateOptions) (Template, error) {
	opts, err := opts.setDefaults(c.Metadata.Name)
	if err != nil {
		return nil, err
	}
	values, err := opts.ValuesOptions.MergeValues(nil)
	if err != nil {
		return nil, err
	}
	renderValues, err := helmChartUtil.ToRenderValues(c.Chart, values, opts.Release, opts.Capabilities)
	if err != nil {
		return nil, err
	}
	e := helmEngine.Engine{LintMode: true}
	templateYamls, err := e.Render(c.Chart, renderValues)
	if err != nil {
		return nil, err
	}
	files := make(map[string]string)
	objectsets := map[string]*objectset.ObjectSet{
		"": objectset.NewObjectSet(),
	}
	for source, manifestString := range templateYamls {
		if filepath.Ext(source) != ".yaml" {
			continue
		}
		source := strings.SplitN(source, string(filepath.Separator), 2)[1]
		manifestString := fmt.Sprintf("---\n%s", manifestString)
		manifestOs, err := parser.Parse(manifestString)
		if err != nil {
			return nil, err
		}
		files[source] = manifestString
		objectsets[source] = manifestOs
		objectsets[""] = objectsets[""].Add(manifestOs.All()...)
	}
	t := &template{
		Options:    opts,
		Files:      files,
		ObjectSets: objectsets,
		Values:     renderValues,
	}
	t.Chart = c
	return t, nil
}
