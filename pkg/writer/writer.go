package writer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
)

const (
	outputDirEnvVar = "TEST_OUTPUT_DIR"
)

var (
	chartPathWriterLock sync.Mutex

	// ensure multiple runs are in new files
	outputFile = fmt.Sprintf("test-%d.md", time.Now().Unix())
)

type chartPathWriter struct {
	Name         string
	ChartName    string
	ChartVersion string
	ChartPath    string
	Command      string
	Raw          string
}

func NewChartPathWriter(t *testing.T, chart, version, path, command, raw string) io.Writer {
	name := t.Name()
	name = strings.ReplaceAll(name, "/", " | ")
	name = strings.ReplaceAll(name, "_", " ")
	return &chartPathWriter{
		Name:         name,
		ChartName:    chart,
		ChartVersion: version,
		ChartPath:    path,
		Command:      command,
		Raw:          raw,
	}
}

func (w *chartPathWriter) Write(out []byte) (n int, err error) {
	chartPathWriterLock.Lock()
	defer chartPathWriterLock.Unlock()

	wd, err := os.Getwd()
	if err != nil {
		return 0, err
	}
	outputDir := os.Getenv(outputDirEnvVar)
	if len(outputDir) == 0 {
		return len(out), nil
	}
	outputDir = filepath.Join(wd, outputDir)
	outputFs := osfs.New(outputDir)

	f, err := outputFs.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var markdownType string
	switch filepath.Ext(w.ChartPath) {
	case ".yaml":
		markdownType = "yaml"
	case ".json":
		markdownType = "json"
	}

	_, err = f.Write([]byte(fmt.Sprintf("## %s\n\n", w.Name)))

	if len(w.Raw) > 0 {
		_, err := f.Write([]byte("### Raw\n"))
		if err != nil {
			return 0, err
		}
		sourceString := fmt.Sprintf("**Source:** `%s`\n\n", filepath.Join(w.ChartName, w.ChartVersion, w.ChartPath))
		_, err = f.Write([]byte(sourceString))
		if err != nil {
			return 0, err
		}
		_, err = f.Write([]byte(fmt.Sprintf("```%s\n", markdownType)))
		if err != nil {
			return 0, err
		}
		_, err = f.Write([]byte(w.Raw))
		if err != nil {
			return 0, err
		}
		_, err = f.Write([]byte("\n```\n"))
		if err != nil {
			return 0, err
		}
	}

	_, err = f.Write([]byte("### Output\n"))
	if err != nil {
		return 0, err
	}
	commandString := fmt.Sprintf("**Command:** `%s`\n\n", w.Command)
	_, err = f.Write([]byte(commandString))
	if err != nil {
		return 0, err
	}
	_, err = f.Write([]byte(fmt.Sprintf("```%s\n", markdownType)))
	if err != nil {
		return 0, err
	}
	_, err = f.Write(out)
	if err != nil {
		return 0, err
	}
	_, err = f.Write([]byte("\n```\n"))
	if err != nil {
		return 0, err
	}

	return len(out), nil
}
