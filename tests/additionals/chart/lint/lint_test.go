package lint

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
)

func TestChartHasValuesSchema(t *testing.T) {
	charts := chart.CollectFromEnv(t, true)
	for _, c := range charts {
		if c.Schema == nil {
			t.Errorf("no values.schema.json found at %s/values.schema.json", c.Path)
		}
	}
}

func TestLintChart(t *testing.T) {
	charts := chart.CollectFromEnv(t, true)
	for _, c := range charts {
		c.Lint(t, nil)
	}
}

func TestLintChartCI(t *testing.T) {
	charts := chart.CollectFromEnv(t, true)
	for _, c := range charts {
		for _, mc := range c.ManifestConfigurationsFromCI(t) {
			c.Lint(t, mc)
		}
	}
}

// func TestLintManifest(t *testing.T) {
// 	charts := chart.CollectFromEnv(t, true)
// 	for _, c := range charts {
// 		c.GetManifest(t, nil).Lint(t)
// 	}
// }

// func TestLintManifestCI(t *testing.T) {
// 	charts := chart.CollectFromEnv(t, true)
// 	for _, c := range charts {
// 		for _, mc := range c.ManifestConfigurationsFromCI(t) {
// 			c.GetManifest(t, mc).Lint(t)
// 		}
// 	}
// }
