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

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
)

const (
	outputDirEnvVar = "TEST_OUTPUT_DIR"
)

var (
	outputWriterLock sync.Mutex

	// ensure multiple runs are in new files
	outputFile = fmt.Sprintf("test-%d.md", time.Now().Unix())
)

type outputWriter struct {
	Name    string
	Source  string
	Command string
	Raw     string

	outputFs billy.Filesystem
}

func getWriterName(t *testing.T) string {
	name := t.Name()
	name = strings.ReplaceAll(name, "/", " | ")
	name = strings.ReplaceAll(name, "_", " ")
	return name
}

func NewOutputWriter(t *testing.T, source, command, raw string) io.Writer {
	w := &outputWriter{
		Name:    getWriterName(t),
		Source:  source,
		Command: command,
		Raw:     raw,
	}
	w.SetOutputDir(os.Getenv(outputDirEnvVar))
	return w
}

func (w *outputWriter) SetOutputDir(outputDir string) {
	if outputDir == "" {
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		// do not set anything
		return
	}
	outputDir = filepath.Join(wd, outputDir)
	w.outputFs = osfs.New(outputDir)
}

func (w *outputWriter) Write(out []byte) (n int, err error) {
	if w.outputFs == nil {
		return 0, nil
	}

	outputWriterLock.Lock()
	defer outputWriterLock.Unlock()

	f, err := w.outputFs.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var markdownType string
	switch filepath.Ext(w.Source) {
	case ".yaml":
		markdownType = "yaml"
	case ".json":
		markdownType = "json"
	}

	_, err = f.Write([]byte(fmt.Sprintf("## %s\n", w.Name)))

	if len(w.Raw) > 0 && len(w.Source) > 0 {
		_, err := f.Write([]byte("\n### Raw\n\n"))
		if err != nil {
			return 0, err
		}
		sourceString := fmt.Sprintf("**Source:** `%s`\n\n", w.Source)
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

	if len(w.Command) > 0 && len(out) > 0 {
		_, err = f.Write([]byte("\n### Output\n\n"))
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
	}

	return len(out), nil
}
