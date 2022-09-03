package lint

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
)

func TestLint(t *testing.T) {
	chart.TestManifests(t, func(t *testing.T, m *chart.Manifest) {
		m.Lint(t)
	})
}
