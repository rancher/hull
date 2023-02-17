package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testIndex         = "testdata/index.yaml"
	doesNotExistIndex = "testdata/doesnotexist.yaml"
	absTestDataPath   = MustGetPathFromModuleRoot("testdata")
)

func TestMustGetPathFromModuleRoot(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	testPath := filepath.Join("hello", "world")
	testCases := []struct {
		Name             string
		FromDir          string
		Expect           string
		ShouldThrowError bool
	}{
		{
			Name:    "Current",
			FromDir: ".",
			Expect:  filepath.Join(filepath.Dir(filepath.Dir(wd)), testPath),
		},
		{
			Name:             "Outside Module",
			FromDir:          filepath.Join("..", "..", ".."),
			ShouldThrowError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			defer func() {
				if err := os.Chdir(wd); err != nil {
					t.Error(err)
					return
				}
			}()
			defer func() {
				err := recover()
				if err != nil {
					assert.True(t, tc.ShouldThrowError, "unexpected error: %s", err)
				}
			}()
			if err := os.Chdir(filepath.Join(wd, tc.FromDir)); err != nil {
				t.Error(err)
				return
			}
			modulePath := MustGetPathFromModuleRoot(testPath)
			assert.False(t, tc.ShouldThrowError, "expected error to be thrown, found modulePath %s", modulePath)
			if t.Failed() {
				return
			}
			assert.Equal(t, tc.Expect, modulePath, "did not find expected path")
		})
	}
}

func TestMustGetLatestChartVersionPathFromIndex(t *testing.T) {
	testCases := []struct {
		Name              string
		IndexFile         string
		ChartName         string
		IncludePrerelease bool

		Expect           string
		ShouldThrowError bool
	}{
		{
			Name:              "Nonexistent Index File",
			IndexFile:         doesNotExistIndex,
			ChartName:         "test-chart",
			IncludePrerelease: true,

			ShouldThrowError: true,
		},
		{
			Name:              "With Prerelease",
			IndexFile:         testIndex,
			ChartName:         "test-chart",
			IncludePrerelease: true,

			Expect: filepath.Join(absTestDataPath, "charts/test-chart/1.0.1-rc1"),
		},
		{
			Name:              "Without Prerelease",
			IndexFile:         testIndex,
			ChartName:         "test-chart",
			IncludePrerelease: false,

			Expect: filepath.Join(absTestDataPath, "charts/test-chart/1.0.0"),
		},
		{
			Name:              "Nonexistent",
			IndexFile:         testIndex,
			ChartName:         "does-not-exist",
			IncludePrerelease: false,

			ShouldThrowError: true,
		},
		{
			Name:              "No Stable Release With Prereleases",
			IndexFile:         testIndex,
			ChartName:         "experimental-chart",
			IncludePrerelease: true,

			Expect: filepath.Join(absTestDataPath, "charts/experimental-chart/0.0.0-rc1"),
		},
		{
			Name:              "No Stable Release Without Prereleases",
			IndexFile:         testIndex,
			ChartName:         "experimental-chart",
			IncludePrerelease: false,

			ShouldThrowError: true,
		},
		{
			Name:      "No URL",
			IndexFile: testIndex,
			ChartName: "no-url-chart",

			ShouldThrowError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			defer func() {
				err := recover()
				if err != nil {
					assert.True(t, tc.ShouldThrowError, "unexpected error: %s", err)
				}
			}()
			chartVersionPath := MustGetLatestChartVersionPathFromIndex(tc.IndexFile, tc.ChartName, tc.IncludePrerelease)
			assert.False(t, tc.ShouldThrowError, "expected error to be thrown, found chartVersionPath %s", chartVersionPath)
			if t.Failed() {
				return
			}
			assert.Equal(t, tc.Expect, chartVersionPath, "did not find expected path")
		})
	}
}
