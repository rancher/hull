package unit

import (
	"path/filepath"
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
)

func TestUnit(t *testing.T) {
	charts := chart.CollectFromEnv(t, true)
	for _, c := range charts {
		m := c.GetManifest(t, nil)
		m.EnforcePolicies(t, false, filepath.Join("policy", "single"))
	}
}

// func TestUnitCI(t *testing.T) {
// 	charts := chart.CollectFromEnv(t, true)
// 	for _, c := range charts {
// 		for _, mc := range c.ManifestConfigurationsFromCI(t) {
// 			c.GetManifest(t, mc)
// 		}
// 	}
// }
