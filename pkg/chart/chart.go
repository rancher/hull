package chart

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aiyengar2/hull/pkg/parser"
	"github.com/rancher/wrangler/pkg/objectset"
	helmChart "helm.sh/helm/v3/pkg/chart"
	helmLoader "helm.sh/helm/v3/pkg/chart/loader"
	helmChartUtil "helm.sh/helm/v3/pkg/chartutil"
	helmEngine "helm.sh/helm/v3/pkg/engine"
)

type Chart interface {
	GetPath() string
	GetHelmChart() *helmChart.Chart

	RenderValues(opts *TemplateOptions) (helmChartUtil.Values, error)
	RenderTemplate(opts *TemplateOptions) (Template, error)
}

type chart struct {
	*helmChart.Chart

	Path string
}

func NewChart(path string) (Chart, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	c := &chart{
		Path: absPath,
	}
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

func (c *chart) RenderValues(opts *TemplateOptions) (helmChartUtil.Values, error) {
	opts = opts.setDefaults(c.Metadata.Name)
	values, err := opts.Values.ToMap()
	if err != nil {
		return nil, err
	}
	return helmChartUtil.ToRenderValues(c.Chart, values, helmChartUtil.ReleaseOptions(opts.Release), (*helmChartUtil.Capabilities)(opts.Capabilities))
}

func (c *chart) RenderTemplate(opts *TemplateOptions) (Template, error) {
	opts = opts.setDefaults(c.Metadata.Name)
	values, err := opts.Values.ToMap()
	if err != nil {
		return nil, err
	}
	renderValues, err := c.RenderValues(opts)
	if err != nil {
		return nil, err
	}
	e := helmEngine.New(nil)
	e.LintMode = true
	templateYamls, err := e.Render(c.Chart, renderValues)
	if err != nil {
		return nil, err
	}
	files := make(map[string]string)
	objectsets := map[string]*objectset.ObjectSet{
		"": objectset.NewObjectSet(),
	}
	for source, manifestString := range templateYamls {
		source := strings.SplitN(source, string(filepath.Separator), 2)[1]
		manifestString := fmt.Sprintf("---\n%s", manifestString)
		manifestOs, err := parser.Parse(manifestString)
		if err != nil {
			return nil, err
		}
		if manifestOs == nil {
			// Note: the action taken here is to workaround a bug in wrangler:
			// https://github.com/rancher/wrangler/blob/5167c04fcdd50e24d9710813572382eeb3064805/pkg/objectset/objectset.go#L99
			// os.order is undefined, so without introducing at least on object os.All() will always fail
			continue
		}
		files[source] = manifestString
		objectsets[source] = manifestOs
		objectsets[""] = objectsets[""].Add(manifestOs.All()...)
	}
	t := &template{
		Options:    opts,
		Files:      files,
		ObjectSets: objectsets,
		Values:     values,
	}
	t.Chart = c
	return t, nil
}
