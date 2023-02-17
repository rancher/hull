package utils

import (
	"fmt"
	"os"
	"path/filepath"

	helmRepo "helm.sh/helm/v3/pkg/repo"
)

func MustGetPathFromModuleRoot(path ...string) string {
	modulePath, err := GetPathFromModuleRoot(path...)
	if err != nil {
		panic(err)
	}
	return modulePath
}

func GetPathFromModuleRoot(path ...string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for wd != string(filepath.Separator) {
		modFile := filepath.Join(wd, "go.mod")
		matches, err := filepath.Glob(modFile)
		if err != nil {
			return "", err
		}
		if matches != nil {
			return filepath.Join(wd, filepath.Join(path...)), nil
		}
		wd = filepath.Dir(wd)
	}
	return "", fmt.Errorf("path must exist within a go module")
}

func MustGetLatestChartVersionPathFromIndex(indexPath, chartName string, includePrerelease bool) string {
	latestChartVersionPath, err := GetLatestChartVersionPathFromIndex(indexPath, chartName, includePrerelease)
	if err != nil {
		panic(err)
	}
	return latestChartVersionPath
}

func GetLatestChartVersionPathFromIndex(indexPath, chartName string, includePrerelease bool) (string, error) {
	absIndexPath, err := GetPathFromModuleRoot(indexPath)
	if err != nil {
		return "", err
	}
	indexFile, err := helmRepo.LoadIndexFile(absIndexPath)
	if err != nil {
		return "", err
	}
	chartVersionPath, err := getLatestChartVersionPathFromIndexFile(indexFile, chartName, includePrerelease)
	if err != nil {
		return "", err
	}
	absChartVersionPath := filepath.Join(filepath.Dir(absIndexPath), chartVersionPath)
	return absChartVersionPath, nil
}

func getLatestChartVersionPathFromIndexFile(indexFile *helmRepo.IndexFile, chartName string, includePrerelease bool) (string, error) {
	indexFile.SortEntries()
	var versionSemver string
	if includePrerelease {
		versionSemver = ">= 0.0.0-0"
	}
	chartVersion, err := indexFile.Get(chartName, versionSemver)
	if err != nil {
		return "", err
	}
	if len(chartVersion.URLs) == 0 {
		return "", fmt.Errorf("could not find URL in index for chart %s-%s", chartName, chartVersion.Version)
	}
	return chartVersion.URLs[0], nil
}
