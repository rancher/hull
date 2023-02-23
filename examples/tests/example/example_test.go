package example

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/test"
)

func TestChart(t *testing.T) {
	suite.Run(t, test.GetRancherOptions())
}
