package unit

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
)

func TestUnit(t *testing.T) {
	chart.TestManifests(t, testManifest)
}

func testManifest(t *testing.T, m *chart.Manifest) {
	// TODO: implement
	//
	// output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/deployment.yaml"})
	// Note: the above line using "github.com/gruntwork-io/terratest/modules/helm" is equivalent to m.raw

	// Per test contexts:
	// run on each templatemanifest, not overall (single resource tests)
	// exclude a set of manifests or exclude those that don't belong in a certain set (windows, linux)

	// convert to map[string]*runtime.ObjectSet where the key is the path where the object is found for writing out
	// if string is "", that means that it is the overall manifest. This is always expected to be set

	// FromString(output) takes in the raw manifest and converts it into map[string]*runtime.Object{"": parser.Parse(output)}
	// FromManifest(m *chart.Manifest) uses m.ToObjectSetMap()

	t.Error(m.Path)
}
