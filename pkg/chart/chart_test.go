package chart

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewChart(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal("cannot find repository root at ../..")
	}

	testCases := []struct {
		Name             string
		ChartPath        string
		ShouldThrowError bool
	}{
		{
			Name:             "Valid Chart",
			ChartPath:        filepath.Join("..", "..", "testdata", "charts", "example-chart"),
			ShouldThrowError: false,
		},
		{
			Name:             "Valid Chart From Absolute Path",
			ChartPath:        filepath.Join(repoRoot, "testdata", "charts", "example-chart"),
			ShouldThrowError: false,
		},
		{
			Name:             "Invalid Chart",
			ChartPath:        filepath.Join("..", "..", "testdata", "charts", "does-not-exist"),
			ShouldThrowError: true,
		},
		{
			Name:             "Invalid Glob Path",
			ChartPath:        filepath.Join("..", "..", "testdata", "charts", "*"),
			ShouldThrowError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			c, err := NewChart(tc.ChartPath)
			if tc.ShouldThrowError {
				if err == nil {
					t.Errorf("expected error to be thrown")
				}
				return
			}
			expectedChartPath, err := filepath.Abs(tc.ChartPath)
			if err != nil {
				t.Fatal("test case is invalid, chartPath provided is not a valid path")
			}
			assert.Equal(t, expectedChartPath, c.GetPath())
			assert.NotNil(t, c.GetHelmChart(), "did not load underlying helm chart")
		})
	}
}
