package writer

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"os"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/header_format.md
var headerFmt string

//go:embed testdata/raw_format.md
var rawFmt string

//go:embed testdata/output_format.md
var outputFmt string

func TestGetOutputFsFromEnv(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("No Output Directory Set", func(t *testing.T) {
		os.Unsetenv(outputDirEnvVar)

		ioW := NewOutputWriter(t, "", "", "")
		w := ioW.(*outputWriter)

		assert.NotNil(t, w)
		if t.Failed() {
			return
		}
		assert.Nil(t, w.outputFs, "expected writer to have no outputFs set")
	})

	t.Run("Output Directory Set By Environment", func(t *testing.T) {
		outputDirName := uuid.New().String()
		os.Setenv(outputDirEnvVar, outputDirName)

		ioW := NewOutputWriter(t, "", "", "")
		w := ioW.(*outputWriter)

		assert.NotNil(t, w)
		if t.Failed() {
			return
		}

		assert.NotNil(t, w.outputFs, "expected writer's outputFs to match env var value on creation")
		if t.Failed() {
			return
		}

		assert.Equal(t, w.outputFs.Root(), filepath.Join(wd, outputDirName), "expected writer to be set to write to outputDir from env var")
		if t.Failed() {
			return
		}
	})

	t.Run("Output Directory Set By Call", func(t *testing.T) {
		os.Unsetenv(outputDirEnvVar)

		ioW := NewOutputWriter(t, "", "", "")
		w := ioW.(*outputWriter)
		assert.NotNil(t, w)
		if t.Failed() {
			return
		}
		assert.Nil(t, w.outputFs, "expected writer to have no outputFs set")

		outputDirName := uuid.New().String()
		w.SetOutputDir(outputDirName)
		assert.Nil(t, err)
		if t.Failed() {
			return
		}

		assert.NotNil(t, w.outputFs, "expected writer to not be nil since the outputDir has been manually set")
		if t.Failed() {
			return
		}
		assert.Equal(t, w.outputFs.Root(), filepath.Join(wd, outputDirName), "expected writer to be set to write to outputDir from Set call")
		if t.Failed() {
			return
		}
	})
}

func TestOutputWriter(t *testing.T) {

	testCases := []struct {
		Name string

		Source string
		Raw    string

		Command string
		Out     string
	}{
		{
			Name: "No Values",
		},
		{
			Name: "Basic YAML",

			Source: "templates/mytemplate.yaml",
			Raw:    "hello-world",

			Command: "echo hello-world",
			Out:     "hello-world",
		},
		{
			Name: "Basic JSON",

			Source: "templates/mytemplate.json",
			Raw:    "hello-world",

			Command: "echo hello-world",
			Out:     "hello-world",
		},
		{
			Name: "Source Omitted",

			Raw: "hello-world",

			Command: "echo hello-world",
			Out:     "hello-world",
		},
		{
			Name: "Raw Omitted",

			Source: "templates/mytemplate.yaml",

			Command: "echo hello-world",
			Out:     "hello-world",
		},
		{
			Name: "Command Omitted",

			Source: "templates/mytemplate.yaml",
			Raw:    "hello-world",

			Out: "hello-world",
		},
		{
			Name: "Out Omitted",

			Source: "templates/mytemplate.yaml",
			Raw:    "hello-world",

			Out: "hello-world",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			outputFs := memfs.New()

			w := NewOutputWriter(t, tc.Source, tc.Command, tc.Raw)
			cpw := w.(*outputWriter)
			cpw.outputFs = outputFs

			_, err := cpw.Write([]byte(tc.Out))
			assert.Nil(t, err)
			if t.Failed() {
				return
			}

			f, err := cpw.outputFs.OpenFile(outputFile, os.O_RDONLY, os.ModePerm)
			assert.Nil(t, err)
			if t.Failed() {
				return
			}
			defer f.Close()

			lstat, err := cpw.outputFs.Lstat(outputFile)
			assert.Nil(t, err)
			if t.Failed() {
				return
			}
			assert.NotZero(t, lstat.Size(), "output file does not contain anything")
			if t.Failed() {
				return
			}

			outputFileContents := make([]byte, lstat.Size())
			n, err := f.Read(outputFileContents)
			assert.Nil(t, err)
			if t.Failed() {
				return
			}
			assert.NotZero(t, n, "output file does not contain anything")
			if t.Failed() {
				return
			}

			ext := strings.TrimPrefix(filepath.Ext(tc.Source), ".")
			expectedOutput := fmt.Sprintf(headerFmt, getWriterName(t))
			if len(tc.Source) > 0 && len(tc.Raw) > 0 {
				expectedOutput += "\n" + fmt.Sprintf(rawFmt, tc.Source, ext, tc.Raw)
			}
			if len(tc.Command) > 0 && len(tc.Out) > 0 {

				expectedOutput += "\n" + fmt.Sprintf(outputFmt, tc.Command, ext, tc.Out)
			}
			assert.Equal(t, expectedOutput, string(outputFileContents))
		})
	}
}
