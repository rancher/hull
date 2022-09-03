package chart

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/aiyengar2/hull/pkg/utils"
	multierr "github.com/hashicorp/go-multierror"
	helmLintRules "helm.sh/helm/v3/pkg/lint/rules"
	helmLintSupport "helm.sh/helm/v3/pkg/lint/support"
)

func (m *Manifest) helmLint(t *testing.T) {
	if err := m.load(); err != nil {
		t.Error(err)
		return
	}
	// Run helm linting
	l := &helmLintSupport.Linter{
		ChartDir: filepath.Join(utils.GetRepoRoot(), m.Path),
	}
	m.runDefaultRules(l)
	m.runCustomRules(l)

	// log errors
	m.logHelmLintErrors(t, l)
}

func (m *Manifest) logHelmLintErrors(t *testing.T, linter *helmLintSupport.Linter) {
	errMap := map[string]error{}
	for _, msg := range linter.Messages {
		switch msg.Severity {
		case helmLintSupport.UnknownSev:
			continue
		case helmLintSupport.InfoSev:
			t.Log(msg.Error())
			continue
		case helmLintSupport.WarningSev:
			t.Log(msg.Error())
			continue
		case helmLintSupport.ErrorSev:
			err := errMap[msg.Path]
			errMap[msg.Path] = multierr.Append(err, fmt.Errorf(msg.Error()))
		}
	}
	if len(errMap) == 0 {
		return
	}
	rawFiles := map[string]string{}
	for _, f := range m.Chart.Raw {
		rawFiles[f.Name] = string(f.Data)
	}

	command := "helm lint --strict"
	valArgs := toValuesArgs(m.Configuration.ValuesOptions)
	if len(valArgs) > 0 {
		command += " " + valArgs
	}
	for path, err := range errMap {
		t.Error(err)
		w := NewChartPathWriter(
			t,
			m.Chart.Metadata.Name,
			m.Chart.Metadata.Version,
			filepath.Join(m.Path, path),
			command,
			rawFiles[path],
		)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			t.Error(err)
		}
	}
}

// Custom Helm Lint Rules

func (m *Manifest) runDefaultRules(linter *helmLintSupport.Linter) {
	helmLintRules.Chartfile(linter)
	helmLintRules.ValuesWithOverrides(linter, m.Values)
	helmLintRules.Templates(linter, m.Values, m.Configuration.Release.Namespace, true)
	helmLintRules.Dependencies(linter)
}

func (m *Manifest) runCustomRules(linter *helmLintSupport.Linter) {
	linter.RunLinterRule(helmLintSupport.ErrorSev, "values.schema.json", m.validateValuesSchemaExists())
}

func (m *Manifest) validateValuesSchemaExists() error {
	if m.Chart.Schema == nil {
		return errors.New("no values.schema.json found")
	}
	return nil
}
