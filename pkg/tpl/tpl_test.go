package tpl

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCollectBuiltInObjectsUsage(t *testing.T) {
	testCases := []struct {
		Name             string
		ChartPath        string
		Expect           map[string]bool
		ShouldThrowError bool
	}{
		// {
		// 	Name:      "No Schema",
		// 	ChartPath: utils.MustGetPathFromModuleRoot("testdata", "charts", "no-schema"),
		// 	Expect:    nil,
		// },
		{
			Name:      "Example Chart",
			ChartPath: utils.MustGetPathFromModuleRoot("testdata", "charts", "example-chart"),
			Expect:    nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			builtInObjectUsage, err := CollectBuiltInObjectsUsage(tc.ChartPath)
			if err != nil {
				assert.True(t, tc.ShouldThrowError, "unexpected error: %s", err)
			}
			if err == nil {
				assert.False(t, tc.ShouldThrowError, "expected error to be thrown, found builtInObjectUsage %s", builtInObjectUsage)
			}
			if t.Failed() {
				return
			}
			assert.Equal(t, tc.Expect, builtInObjectUsage)
		})
	}
}
