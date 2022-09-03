package chart

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"testing"

	"github.com/aiyengar2/hull/pkg/utils"
	"github.com/google/uuid"
	helmChart "helm.sh/helm/v3/pkg/chart"
)

const (
	outputDirEnvVar = "TEST_OUTPUT_DIR"
)

var (
	chartPathWriters    map[string]*chartPathWriter
	chartPathWriterLock sync.Mutex

	// ensure multiple runs are in new directories
	chartPathNamespace = fmt.Sprintf("test-run-%s", uuid.New())
)

type chartPathWriter struct {
	ChartMetadata helmChart.Metadata

	ChartName    string
	ChartVersion string
	ChartPath    string
	Command      string
	Raw          string
}

func NewChartPathWriter(t *testing.T, chart, version, path, command, raw string) io.Writer {
	chartPathWriterLock.Lock()
	defer chartPathWriterLock.Unlock()

	if chartPathWriters == nil {
		chartPathWriters = make(map[string]*chartPathWriter)
	}

	key := fmt.Sprintf("%s-%s-%s", chart, version, path)
	w, ok := chartPathWriters[key]
	if !ok {
		chartPathWriters[key] = &chartPathWriter{
			ChartName:    chart,
			ChartVersion: version,
			ChartPath:    path,
			Command:      command,
			Raw:          raw,
		}
		w = chartPathWriters[key]
	}
	return w
}

func (w *chartPathWriter) Write(out []byte) (n int, err error) {
	repoFs := utils.GetRepoFs()
	outputDir := os.Getenv(outputDirEnvVar)
	if len(outputDir) == 0 {
		return len(out), nil
	}
	outputDir = filepath.Join(outputDir, chartPathNamespace)

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
