package unit

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
)

func TestUnit(t *testing.T) {
	charts := chart.CollectFromEnv(t, true)
	for _, c := range charts {
		c.GetManifest(t, nil)
	}

}

func TestUnitCI(t *testing.T) {
	charts := chart.CollectFromEnv(t, true)
	for _, c := range charts {
		for _, mc := range c.ManifestConfigurationsFromCI(t) {
			c.GetManifest(t, mc)
		}
	}
}
