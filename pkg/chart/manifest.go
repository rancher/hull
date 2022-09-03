package chart

import (
	"sync"

	"github.com/rancher/charts-build-scripts/pkg/filesystem"
	"github.com/rancher/helm-locker/pkg/objectset/parser"
	"github.com/rancher/wrangler/pkg/objectset"
	"gopkg.in/yaml.v2"
	helmChart "helm.sh/helm/v3/pkg/chart"
)

type Manifest struct {
	Chart *helmChart.Chart

	Path          string
	Configuration *ManifestConfiguration
	Values        map[string]interface{}

	templateManifests map[string]*TemplateManifest
	lock              sync.Mutex
	raw               string
	os                *objectset.ObjectSet
	osMap             map[string]*objectset.ObjectSet
}

func (m *Manifest) Raw() string {
	_ = m.load()
	return m.raw
}

func (m *Manifest) ToObjectSet() (*objectset.ObjectSet, error) {
	err := m.load()
	// deep copy
	os := objectset.NewObjectSet(m.os.All()...)
	return os, err
}

func (m *Manifest) ToObjectSetMap() (map[string]*objectset.ObjectSet, error) {
	err := m.load()
	// deep copy
	osMap := make(map[string]*objectset.ObjectSet, len(m.osMap))
	for osPath, os := range m.osMap {
		osMap[osPath] = objectset.NewObjectSet(os.All()...)
	}
	return osMap, err
}

func (m *Manifest) load() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.os != nil && len(m.raw) > 0 {
		return nil
	}

	m.os = objectset.NewObjectSet()
	m.osMap = make(map[string]*objectset.ObjectSet)
	for tmPath, tm := range m.templateManifests {
		tmOs, err := tm.ToObjectSet()
		if err != nil {
			return err
		}
		m.os.Add(tmOs.All()...)
		tmRelativePath, err := filesystem.MovePath(tmPath, m.Path, "")
		if err != nil {
			return err
		}
		m.osMap[tmRelativePath] = tmOs
	}
	m.osMap[""] = m.os

	m.raw = ""
	for _, o := range m.os.All() {
		raw, err := yaml.Marshal(o)
		if err != nil {
			return err
		}
		if len(m.raw) != 0 {
			m.raw += "\n---\n"
		}
		m.raw += string(raw)
	}

	return nil
}

type TemplateManifest struct {
	ChartMetadata *helmChart.Metadata
	TemplateFile  string
	Values        map[string]interface{}

	ManifestConfiguration *ManifestConfiguration

	lock sync.Mutex
	os   *objectset.ObjectSet
	raw  string
}

func (m *TemplateManifest) Raw() string {
	return m.raw
}

func (m *TemplateManifest) ToObjectSet() (*objectset.ObjectSet, error) {
	err := m.load()
	// deep copy
	os := objectset.NewObjectSet(m.os.All()...)
	return os, err
}

func (m *TemplateManifest) load() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.os != nil {
		return nil
	}
	var err error
	m.os, err = parser.Parse(m.Raw())
	if err != nil {
		return err
	}
	return nil
}
