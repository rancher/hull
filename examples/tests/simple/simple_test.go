package simple

import (
	"testing"

	"github.com/rancher/hull/pkg/test"
)

func TestChart(t *testing.T) {
	opts := test.GetRancherOptions()
	opts.YAMLLint.Enabled = true
	suite.Run(t, opts)
}
