package metadata

import (
	"fmt"

	"github.com/aiyengar2/hull/pkg/chart"
)

const (
	CattleReleaseNameAnnotation      = "catalog.cattle.io/release-name"
	CattleReleaseNamespaceAnnotation = "catalog.cattle.io/namespace"
)

func MustHaveAnnotation(chartPath string, annotation string) string {
	a, err := GetAnnotation(chartPath, annotation)
	if err != nil {
		panic(fmt.Errorf("could not find chart at %s: %s", chartPath, err))
	}
	return a
}

func GetAnnotation(chartPath string, annotation string) (string, error) {
	c, err := chart.NewChart(chartPath)
	if err != nil {
		return "", err
	}
	cMeta := c.GetHelmChart().Metadata
	if cMeta == nil {
		return "", fmt.Errorf("chart does not have metadata")
	}
	annotations := cMeta.Annotations
	if annotations == nil {
		return "", fmt.Errorf("chart does not have annotations")
	}
	val, ok := annotations[annotation]
	if !ok {
		return "", fmt.Errorf("chart does not have annotation %s", annotation)
	}
	return val, nil
}
