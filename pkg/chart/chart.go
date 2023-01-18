package chart

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aiyengar2/hull/pkg/parser"
	"github.com/aiyengar2/hull/pkg/schema"
	"github.com/aiyengar2/hull/pkg/writer"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/stretchr/testify/assert"
	helmChart "helm.sh/helm/v3/pkg/chart"
	helmLoader "helm.sh/helm/v3/pkg/chart/loader"
	helmChartUtil "helm.sh/helm/v3/pkg/chartutil"
	helmEngine "helm.sh/helm/v3/pkg/engine"
)

type Chart interface {
	GetPath() string
	GetHelmChart() *helmChart.Chart

	RenderTemplate(opts *TemplateOptions) (Template, error)

	SchemaMustMatchStruct(t *testing.T, schemaStruct interface{})
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

func (c *chart) RenderTemplate(opts *TemplateOptions) (Template, error) {
	opts = opts.setDefaults(c.Metadata.Name)
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
		Values:     values,
	}
	t.Chart = c
	return t, nil
}

func (c *chart) SchemaMustMatchStruct(t *testing.T, schemaStruct interface{}) {
	expectedSchemaBytes, err := schema.FromStructToSchemaBytes(schemaStruct)
	if err != nil {
		t.Error(err)
		return
	}

	var schema string
	if c.Chart.Schema != nil {
		schema = string(c.Chart.Schema)
	}

	// assert and print error
	assert.Equal(t, string(expectedSchemaBytes), schema)
	if !t.Failed() {
		return
	}

	// Write to output file
	w := writer.NewOutputWriter(
		t,
		filepath.Join(c.Chart.Metadata.Name, c.Chart.Metadata.Version, "values.schema.json"),
		fmt.Sprintf("jsonschema.Reflect(%T)", schemaStruct),
		string(c.Chart.Schema),
	)
	if _, err := w.Write(expectedSchemaBytes); err != nil {
		t.Error(err)
		return
	}
	return
}
