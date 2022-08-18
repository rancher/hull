package chart

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/aiyengar2/hull/pkg/utils"
	"github.com/rancher/charts-build-scripts/pkg/filesystem"
	helmChart "helm.sh/helm/v3/pkg/chart"
)

const (
	outputDirEnvVar = "TEST_OUTPUT_DIR"
)

var (
	deleteOutputDirOnce sync.Once
)

type chartPathWriter struct {
	ChartMetadata helmChart.Metadata

	ChartName    string
	ChartVersion string
	ChartPath    string
	Command      string
	Raw          string
}

func NewChartPathWriter(chart, version, path, command, raw string) io.Writer {
	return &chartPathWriter{
		ChartName:    chart,
		ChartVersion: version,
		ChartPath:    path,
		Command:      command,
		Raw:          raw,
	}
}

func (w *chartPathWriter) Write(out []byte) (n int, err error) {
	repoFs := utils.GetRepoFs()
	outputDir := os.Getenv(outputDirEnvVar)
	if len(outputDir) == 0 {
		return len(out), nil
	}
	deleteOutputDirOnce.Do(func() {
		filesystem.RemoveAll(repoFs, outputDir)
	})

	outputFile := filepath.Join(outputDir, w.ChartMetadata.Name, w.ChartMetadata.Version, w.ChartPath)
	f, err := repoFs.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	rawString := fmt.Sprintf("%s\n", w.Raw)
	_, err = f.Write([]byte(rawString))
	if err != nil {
		return 0, err
	}

	sourceString := fmt.Sprintf("#\n# Source: %s\n# Command: %s", w.ChartPath, w.Command)
	_, err = f.Write([]byte(sourceString))
	if err != nil {
		return 0, err
	}

	re := regexp.MustCompile("\n")
	outComment := re.ReplaceAllString("# "+string(out), "\n# ")
	outCommentString := fmt.Sprintf("\n#\n%s\n", outComment)
	_, err = f.Write([]byte(outCommentString))
	if err != nil {
		return 0, err
	}

	return len(out), nil
}
