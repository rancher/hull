package metadata

import (
	"testing"

	"github.com/rancher/hull/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestMustGetAnnotation(t *testing.T) {
	chartPath := utils.MustGetPathFromModuleRoot("testdata", "charts", "example-chart")

	testCases := []struct {
		Name          string
		ChartPath     string
		Annotation    string
		ExpectedValue string
		ShouldPanic   bool
	}{
		{
			Name:        "Non-Existent Chart",
			ChartPath:   "",
			ShouldPanic: true,
		},
		{
			Name:        "Non-Existent Annotation",
			ChartPath:   chartPath,
			Annotation:  "hull.cattle.io",
			ShouldPanic: true,
		},
		{
			Name:          "Release Name",
			ChartPath:     chartPath,
			Annotation:    CattleReleaseNameAnnotation,
			ExpectedValue: "example-chart",
		},
		{
			Name:          "Release Name",
			ChartPath:     chartPath,
			Annotation:    CattleReleaseNamespaceAnnotation,
			ExpectedValue: "cattle-hull-system",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			defer func() {
				// recover from panic if one occured. Set err to nil otherwise.
				err := recover()
				if err != nil && !tc.ShouldPanic {
					t.Error(err)
				}
			}()
			val := MustHaveAnnotation(tc.ChartPath, tc.Annotation)
			assert.Equal(t, tc.ExpectedValue, val)
			assert.False(t, tc.ShouldPanic, "expected panic")
		})
	}
}
