package chart

import (
	"path/filepath"

	"github.com/Masterminds/semver/v3"
)

type ChartRepository map[string]*Chart

func (c ChartRepository) removeSubcharts() {
	if c == nil {
		return
	}
	for chartPath := range c {
		chartPathDir := chartPath
		for {
			chartPathDir = filepath.Dir(chartPathDir)
			if chartPathDir == "." {
				break
			}
			_, ok := c[chartPathDir]
			if ok {
				// Identified a subchart
				delete(c, chartPath)
				break
			}
		}
	}
}

func (c ChartRepository) onlyLatest() error {
	latestVersions := make(map[string]*semver.Version)
	for _, chart := range c {
		if err := chart.Load(); err != nil {
			return err
		}
		v, err := semver.NewVersion(chart.Metadata.Version)
		if err != nil {
			return err
		}
		currVer, ok := latestVersions[chart.Metadata.Name]
		if !ok || v.GreaterThan(currVer) {
			latestVersions[chart.Metadata.Name] = v
		}
	}
	for path, chart := range c {
		v, err := semver.NewVersion(chart.Metadata.Version)
		if err != nil {
			return err
		}
		if v.LessThan(latestVersions[chart.Metadata.Name]) {
			delete(c, path)
		}
	}
	return nil
}
