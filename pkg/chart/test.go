package chart

import (
	"fmt"
	"testing"
)

type ManifestTestFunc func(t *testing.T, m *Manifest)

func TestManifests(t *testing.T, mFunc ManifestTestFunc) {
	charts := CollectFromEnv(t, true)
	for _, c := range charts {
		mcs := c.ManifestConfigurationsFromCI(t)
		for _, mc := range append([]*ManifestConfiguration{nil}, mcs...) {
			m := c.GetManifest(t, nil)
			t.Run(
				testName(c, mc),
				func(t *testing.T) { mFunc(t, m) },
			)
		}
	}
}

func testName(c *Chart, mc *ManifestConfiguration) string {
	if c == nil {
		return ""
	}
	name := fmt.Sprintf("CHART (%s)", c.Path)
	if mc == nil {
		return name
	}
	name += fmt.Sprintf(" VALUES (%s)", mc.Name)
	return name
}
