package example

import (
	"fmt"
	"strings"

	"github.com/aiyengar2/hull/pkg/checker"
	"github.com/aiyengar2/hull/pkg/test"
	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	suite.NamedChecks = append(suite.NamedChecks, test.NamedCheck{
		Name:   "All Resources Have Helm Release Labels",
		Checks: AllResourcesHaveHelmReleaseLabels,
	})
}

// Check that all resources have expected Helm release labels
var AllResourcesHaveHelmReleaseLabels = test.Checks{
	checker.Once(func(tc *checker.TestContext) {
		chartName := checker.MustRenderValue[string](tc, ".Chart.Name")
		releaseName := checker.MustRenderValue[string](tc, ".Release.Name")
		normalizedChartVersion := strings.ReplaceAll(checker.MustRenderValue[string](tc, ".Chart.Version"), "+", "_")
		nameOverride, hasNameOverride := checker.RenderValue[string](tc, ".Values.nameOverride")
		checker.MapSet(tc, "Default Labels",
			"app.kubernetes.io/managed-by",
			"Helm",
		)
		checker.MapSet(tc, "Default Labels",
			"app.kubernetes.io/instance",
			releaseName,
		)
		checker.MapSet(tc, "Default Labels",
			"app.kubernetes.io/version",
			normalizedChartVersion,
		)
		if hasNameOverride && len(nameOverride) != 0 {
			checker.MapSet(tc, "Default Labels",
				"app.kubernetes.io/part-of",
				nameOverride,
			)
		} else {
			checker.MapSet(tc, "Default Labels",
				"app.kubernetes.io/part-of",
				chartName,
			)
		}
		checker.MapSet(tc, "Default Labels",
			"chart",
			fmt.Sprintf("%s-%s",
				chartName,
				normalizedChartVersion,
			),
		)
		checker.MapSet(tc, "Default Labels",
			"release",
			releaseName,
		)
		checker.MapSet(tc, "Default Labels",
			"heritage",
			"Helm",
		)
	}),
	checker.PerResource(func(tc *checker.TestContext, obj *unstructured.Unstructured) {
		expectedLabels, ok := checker.Get[string, map[string]string](tc, "Default Labels")
		if !ok {
			assert.True(tc.T, ok)
			return
		}
		objLabels := obj.GetLabels()
		relevantObjLabels := map[string]string{}
		for k := range expectedLabels {
			objVal, ok := objLabels[k]
			if !ok {
				continue
			}
			relevantObjLabels[k] = objVal
		}
		assert.Equal(tc.T, expectedLabels, relevantObjLabels, "%s %s's labels do not match expected labels", obj.GroupVersionKind().Kind, checker.Key(obj))
	}),
}
