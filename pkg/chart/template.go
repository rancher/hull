package chart

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/aiyengar2/hull/pkg/checker"
	"github.com/aiyengar2/hull/pkg/writer"
	multierr "github.com/hashicorp/go-multierror"
	"github.com/rancher/wrangler/pkg/objectset"
	helmLintSupport "helm.sh/helm/v3/pkg/lint/support"
)

type Template interface {
	checker.Checker

	GetChart() Chart
	GetOptions() *TemplateOptions
	GetFiles() map[string]string
	GetObjectSets() map[string]*objectset.ObjectSet
	GetValues() map[string]interface{}

	YamlLint(t *testing.T)
	HelmLint(t *testing.T)
}

type template struct {
	Chart   *chart
	Options *TemplateOptions

	Files      map[string]string
	ObjectSets map[string]*objectset.ObjectSet
	Values     map[string]interface{}
}

func (t *template) GetChart() Chart {
	return t.Chart
}

func (t *template) GetOptions() *TemplateOptions {
	return t.Options
}

func (t *template) GetFiles() map[string]string {
	return t.Files
}

func (t *template) GetObjectSets() map[string]*objectset.ObjectSet {
	return t.ObjectSets
}

func (t *template) GetValues() map[string]interface{} {
	return t.Values
}

func (t *template) YamlLint(tT *testing.T) {
	for templateFile := range t.ObjectSets {
		if len(templateFile) == 0 {
			continue
		}
		t.yamlLint(tT, templateFile)
	}
}

func (t *template) yamlLint(tT *testing.T, templateFile string) {
	objectSet, ok := t.ObjectSets[templateFile]
	if !ok || objectSet.Len() == 0 {
		// no objects to lint
		return
	}
	raw, ok := t.Files[templateFile]
	if !ok {
		// objectset cannot exist without template file
		tT.Errorf("could not find raw file associated with templateFile %s", templateFile)
		return
	}

	cmd := exec.Command("yamllint", "-")
	cmd.Stdin = strings.NewReader(raw)
	out, err := cmd.CombinedOutput()
	if err != nil {
		tT.Errorf("[%s@%s] %s failed lint checks against %s", t.Chart.Metadata.Name, t.Chart.Metadata.Version, templateFile, t.Options)
		w := writer.NewChartPathWriter(
			tT,
			t.Chart.Metadata.Name,
			t.Chart.Metadata.Version,
			templateFile,
			t.Options.String(),
			raw,
		)
		w = io.MultiWriter(w, os.Stderr)
		if _, err := w.Write(out); err != nil {
			tT.Error(err)
		}
	}
}

func (t *template) HelmLint(tT *testing.T) {
	// Run helm linting
	l := &helmLintSupport.Linter{
		ChartDir: t.Chart.Path,
	}
	t.runDefaultRules(l)
	t.runCustomRules(l)

	// log errors
	errMap := map[string]error{}
	for _, msg := range l.Messages {
		switch msg.Severity {
		case helmLintSupport.UnknownSev:
			continue
		case helmLintSupport.InfoSev:
			tT.Log(msg.Error())
			continue
		case helmLintSupport.WarningSev:
			tT.Log(msg.Error())
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
	for _, f := range t.Chart.Raw {
		rawFiles[f.Name] = string(f.Data)
	}

	command := "helm lint --strict"
	valArgs := toValuesArgs(t.Options.ValuesOptions)
	if len(valArgs) > 0 {
		command += " " + valArgs
	}
	for path, err := range errMap {
		tT.Error(err)
		w := writer.NewChartPathWriter(
			tT,
			t.Chart.Metadata.Name,
			t.Chart.Metadata.Version,
			path,
			command,
			rawFiles[path],
		)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			tT.Error(err)
		}
	}
}

func (t *template) Check(tT *testing.T, opts *checker.Options, objStructFunc interface{}) {
	if t == nil || t.ObjectSets == nil {
		return
	}
	check := checker.NewChecker(t.ObjectSets)
	check.Check(tT, opts, objStructFunc)
}
