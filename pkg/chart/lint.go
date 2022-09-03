package chart

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func (m *Manifest) Lint(t *testing.T) {
	if m == nil {
		return
	}
	m.helmLint(t)
	// for _, tm := range m.sorted() {
	// 	tm.Lint(t)
	// }
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
			t,
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
