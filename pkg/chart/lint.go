package chart

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aiyengar2/hull/pkg/utils"
	multierr "github.com/hashicorp/go-multierror"
	helmLintRules "helm.sh/helm/v3/pkg/lint/rules"
	helmLintSupport "helm.sh/helm/v3/pkg/lint/support"
)

func (c *Chart) Lint(t *testing.T, conf *ManifestConfiguration) {
	if err := c.Load(); err != nil {
		t.Error(err)
		return
	}

	conf, err := conf.setDefaults(c.Metadata.Name)
	if err != nil {
		t.Error(err)
		return
	}
	values, err := conf.ValuesOptions.MergeValues(nil)
	if err != nil {
		t.Error(err)
		return
	}
	l := &helmLintSupport.Linter{
		ChartDir: filepath.Join(utils.GetRepoRoot(), c.Path),
	}
	helmLintRules.Chartfile(l)
	helmLintRules.ValuesWithOverrides(l, values)
	helmLintRules.Templates(l, values, conf.Release.Namespace, true)
	helmLintRules.Dependencies(l)

	errMap := map[string]error{}
	for _, msg := range l.Messages {
		err := errMap[msg.Path]
		errMap[msg.Path] = multierr.Append(err, fmt.Errorf(msg.Error()))
	}
	if len(errMap) == 0 {
		return
	}
	rawFiles := map[string]string{}
	for _, f := range c.Raw {
		rawFiles[f.Name] = string(f.Data)
	}

	command := "helm lint --strict"
	valArgs := toValuesArgs(conf.ValuesOptions)
	if len(valArgs) > 0 {
		command += " " + valArgs
	}
	for path, err := range errMap {
		t.Error(err)
		w := NewChartPathWriter(
			c.Metadata.Name,
			c.Metadata.Version,
			filepath.Join(c.Path, path),
			command,
			rawFiles[path],
		)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			t.Error(err)
		}
	}
}

func (m *Manifest) Lint(t *testing.T) {
	if m == nil {
		return
	}
	for _, tm := range m.sorted() {
		tm.Lint(t)
	}
}

func (m *TemplateManifest) Lint(t *testing.T) error {
	objs, err := m.ToObjectSet()
	if err != nil {
		return err
	}
	if objs.Len() == 0 {
		return nil
	}

	cmd := exec.Command("yamllint", "-")
	cmd.Stdin = strings.NewReader(m.Raw())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("[%s@%s] %s failed lint checks against %s", m.ChartMetadata.Name, m.ChartMetadata.Version, m.TemplateFile, m.ManifestConfiguration)

		w := NewChartPathWriter(
			m.ChartMetadata.Name,
			m.ChartMetadata.Version,
			strings.TrimPrefix(m.TemplateFile, m.ChartMetadata.Name+string(os.PathSeparator)),
			m.ManifestConfiguration.String(),
			m.Raw(),
		)
		w = io.MultiWriter(w, os.Stderr)
		if _, err := w.Write(out); err != nil {
			t.Error(err)
		}
	}
	return nil
}
