package chart

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/aiyengar2/hull/pkg/utils"
	"github.com/rancher/charts-build-scripts/pkg/filesystem"
	helmChart "helm.sh/helm/v3/pkg/chart"
	helmLoader "helm.sh/helm/v3/pkg/chart/loader"
	helmChartUtil "helm.sh/helm/v3/pkg/chartutil"
	helmValues "helm.sh/helm/v3/pkg/cli/values"
	helmEngine "helm.sh/helm/v3/pkg/engine"
)

type Chart struct {
	*helmChart.Chart

	loadLock sync.Mutex
	Path     string
}

func (c *Chart) Load() error {
	c.loadLock.Lock()
	defer c.loadLock.Unlock()

	if c.Chart != nil {
		return nil
	}
	chart, err := helmLoader.Load(filepath.Join(utils.GetRepoRoot(), c.Path))
	if err != nil {
		return err
	}
	c.Chart = chart
	return nil
}

func (c *Chart) ManifestConfigurationsFromCI(t *testing.T) []*ManifestConfiguration {
	valuesFiles := c.ValuesFilesFromCI(t)
	if valuesFiles == nil {
		return nil
	}
	mcs := make([]*ManifestConfiguration, len(valuesFiles))
	for i, v := range valuesFiles {
		withoutSuffix := strings.TrimSuffix(v, "-values.yaml")
		if v == withoutSuffix {
			withoutSuffix = "default"
		}
		mc := &ManifestConfiguration{
			Name: withoutSuffix,
			ValuesOptions: &helmValues.Options{
				ValueFiles: []string{v},
			},
		}
		mcs[i] = mc
	}
	return mcs
}

func (c *Chart) ValuesFilesFromCI(t *testing.T) []string {
	repoFs := utils.GetRepoFs()
	glob := filepath.Join(
		filesystem.GetAbsPath(repoFs, filepath.Join(c.Path)),
		"ci", "*-values.yaml")
	matches, err := filepath.Glob(glob)
	if err != nil {
		t.Error(err)
		return nil
	}
	return matches
}

func (c *Chart) GetManifest(t *testing.T, conf *ManifestConfiguration) *Manifest {
	if err := c.Load(); err != nil {
		t.Error(err)
		return nil
	}

	conf, err := conf.setDefaults(c.Metadata.Name)
	if err != nil {
		t.Error(err)
		return nil
	}
	renderedChart, err := renderChart(c.Chart, conf)
	if err != nil {
		t.Error(fmt.Errorf("[%s@%s] %s", c.Metadata.Name, c.Metadata.Version, err))
		return nil
	}
	manifest := Manifest{
		ChartMetadata: c.Metadata,
		Configuration: conf,
		Path:          c.Path,

		templateManifests: make(map[string]*TemplateManifest, 0),
	}
	for source, manifestString := range renderedChart {
		if filepath.Ext(source) != ".yaml" {
			delete(renderedChart, source)
		}
		templateFile, err := filesystem.MovePath(source, c.Metadata.Name, c.Path)
		if err != nil {
			t.Error(err)
			return nil
		}
		manifest.templateManifests[templateFile] = &TemplateManifest{
			ChartMetadata:         c.Metadata,
			TemplateFile:          templateFile,
			ManifestConfiguration: conf,
			raw:                   fmt.Sprintf("---\n%s", manifestString),
		}
	}
	return &manifest
}

func renderChart(chart *helmChart.Chart, conf *ManifestConfiguration) (map[string]string, error) {
	values, err := conf.ValuesOptions.MergeValues(nil)
	if err != nil {
		return nil, err
	}
	renderValues, err := helmChartUtil.ToRenderValues(chart, values, conf.Release, conf.Capabilities)
	if err != nil {
		return nil, err
	}
	e := helmEngine.Engine{LintMode: true}
	return e.Render(chart, renderValues)
}
