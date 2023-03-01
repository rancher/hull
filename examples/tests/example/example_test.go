package example

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/test"
)

func TestChart(t *testing.T) {
	opts := test.GetRancherOptions()
	// opts.Coverage.IncludeSubcharts = true
	opts.Coverage.Disabled = true
	opts.YAMLLint.Enabled = true
	suite.Run(t, opts)
}
