package chart

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/aiyengar2/hull/pkg/utils"
	"github.com/go-git/go-billy/v5"
	"github.com/rancher/charts-build-scripts/pkg/filesystem"
	"github.com/rancher/charts-build-scripts/pkg/path"
)

func CollectFromEnv(t *testing.T, latest bool) ChartRepository {
	if latest {
		return CollectLatest(t, os.Getenv("CHART"))
	}
	return Collect(t, os.Getenv("CHART"), os.Getenv("VERSION"))
}

func Collect(t *testing.T, chart, version string) ChartRepository {
	var constraint *semver.Constraints
	if version == "" {
		constraint, _ = semver.NewConstraint("*")
	} else {
		var err error
		constraint, err = semver.NewConstraint(version)
		if err != nil {
			t.Error(err)
			t.FailNow()
			return nil
		}
	}
	cr := CollectAll(t)
	for path, c := range cr {
		if err := c.Load(); err != nil {
			t.Error(err)
			delete(cr, path)
			continue
		}
		if len(chart) > 0 && chart != c.Metadata.Name {
			delete(cr, path)
		}
		v, err := semver.NewVersion(c.Metadata.Version)
		if err != nil {
			t.Error(err)
			continue
		}
		if !constraint.Check(v) {
			delete(cr, path)
		}
	}
	return cr
}

func CollectLatest(t *testing.T, chart string) ChartRepository {
	c := Collect(t, chart, "")
	if c == nil {
		return c
	}
	err := c.onlyLatest()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return c
}

func CollectAll(t *testing.T) ChartRepository {
	// Collect all valid charts from charts directory
	charts := make(ChartRepository)
	collectAllValidCharts := func(fs billy.Filesystem, path string, isDir bool) error {
		if !isDir {
			return nil
		}
		c := CollectFromPath(t, path)
		if c != nil {
			charts[path] = c
		}
		return nil
	}

	repoFs := utils.GetRepoFs()
	if err := filesystem.WalkDir(repoFs, path.RepositoryChartsDir, collectAllValidCharts); err != nil {
		t.Error(err)
		t.FailNow()
		return nil
	}

	// Ensure you do not collect subcharts defined within charts if you are scanning the whole chart
	charts.removeSubcharts()
	return charts
}

func CollectFromPath(t *testing.T, chartPath string) *Chart {
	repoRoot := utils.GetRepoRoot()
	matches, err := filepath.Glob(filepath.Join(repoRoot, chartPath, "Chart.yaml"))
	if err != nil {
		t.Error(err)
		return nil
	}
	if len(matches) == 0 {
		return nil
	}

	return &Chart{
		Path: chartPath,
	}
}
