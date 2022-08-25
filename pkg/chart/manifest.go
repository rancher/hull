package chart

import (
	"sort"
	"sync"

	"github.com/rancher/helm-locker/pkg/objectset/parser"
	"github.com/rancher/wrangler/pkg/objectset"
	"gopkg.in/yaml.v2"
	helmChart "helm.sh/helm/v3/pkg/chart"
)

type Manifest struct {
	ChartMetadata *helmChart.Metadata

	Path          string
	Configuration *ManifestConfiguration

	templateManifests map[string]*TemplateManifest

	lock sync.Mutex
	raw  string
	os   *objectset.ObjectSet
}

func (m *Manifest) Raw() string {
	_ = m.load()
	return m.raw
}

func (m *Manifest) ToObjectSet() (*objectset.ObjectSet, error) {
	err := m.load()
	return m.os, err
}

func (m *Manifest) load() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.os != nil && len(m.raw) > 0 {
		return nil
	}

	m.os = objectset.NewObjectSet()
	m.raw = ""
	for _, tm := range m.templateManifests {
		tmOs, err := tm.ToObjectSet()
		if err != nil {
			return err
		}
		m.os.Add(tmOs.All()...)
	}

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

func (m *Manifest) sorted() []*TemplateManifest {
	var tms []*TemplateManifest
	for _, tm := range m.templateManifests {
		tms = append(tms, tm)
	}
	sort.SliceStable(tms, func(i, j int) bool {
		return tms[i].TemplateFile > tms[j].TemplateFile
	})
	return tms
}

type TemplateManifest struct {
	ChartMetadata *helmChart.Metadata
	TemplateFile  string
	raw           string

	ManifestConfiguration *ManifestConfiguration

	lock sync.Mutex
	os   *objectset.ObjectSet
}

func (m *TemplateManifest) Raw() string {
	return m.raw
}

func (m *TemplateManifest) ToObjectSet() (*objectset.ObjectSet, error) {
	err := m.load()
	return m.os, err
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
