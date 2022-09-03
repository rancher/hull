package chart

import (
	"fmt"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
)

type ChartRepository map[string]*Chart

func (c ChartRepository) Exclude(chart, version string) {
	if c == nil {
		return
	}
	for chartPath, ch := range c {
		if ch.Metadata.Name == chart && ch.Metadata.Version == version {
			delete(c, chartPath)
		}
	}
}

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
		key := fmt.Sprintf("%s/%s.%s.x", chart.Metadata.Name, fmt.Sprint(v.Major()), fmt.Sprint(v.Minor()))
		currVer, ok := latestVersions[key]
		if !ok || v.GreaterThan(currVer) {
			latestVersions[key] = v
		}
	}
	for path, chart := range c {
		v, err := semver.NewVersion(chart.Metadata.Version)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%s.%s.x", chart.Metadata.Name, fmt.Sprint(v.Major()), fmt.Sprint(v.Minor()))
		if v.LessThan(latestVersions[key]) {
			delete(c, path)
		}
	}
	return nil
}
