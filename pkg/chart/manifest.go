package chart

import (
	"sort"
	"sync"

	"github.com/rancher/helm-locker/pkg/objectset/parser"
	"github.com/rancher/wrangler/pkg/objectset"
	helmChart "helm.sh/helm/v3/pkg/chart"
)

type Manifest struct {
	ChartMetadata *helmChart.Metadata

	Configuration *ManifestConfiguration

	templateManifests map[string]*TemplateManifest
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
	m.os, err = parser.Parse(m.raw)
	if err != nil {
		return err
	}
	return nil
}
