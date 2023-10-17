package chart

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	multierr "github.com/hashicorp/go-multierror"
	"github.com/rancher/hull/pkg/checker"
	"github.com/rancher/hull/pkg/writer"
	"github.com/rancher/wrangler/pkg/objectset"
	helmAction "helm.sh/helm/v3/pkg/action"
	helmLintSupport "helm.sh/helm/v3/pkg/lint/support"
)

//go:embed configuration/yamllint.yaml
var DefaultYamllintConf string

type Template interface {
	checker.Checker

	GetChart() Chart
	GetOptions() *TemplateOptions
	GetFiles() map[string]string
	GetObjectSets() map[string]*objectset.ObjectSet
	GetValues() map[string]interface{}

	YamlLint(t *testing.T, yamllintConf string)
	HelmLint(t *testing.T, opts *HelmLintOptions)
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
	if t.Options == nil {
		t.Options = &TemplateOptions{}
	}
	return t.Options.setDefaults(t.Chart.Name())
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

func (t *template) YamlLint(tT *testing.T, yamllintConf string) {
	for templateFile := range t.ObjectSets {
		if len(templateFile) == 0 {
			continue
		}
		t.yamlLint(tT, templateFile, yamllintConf)
	}
}

func (t *template) yamlLint(tT *testing.T, templateFile, yamllintConf string) {
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

	tempConfigFile, err := os.CreateTemp("", "")
	if err != nil {
		tT.Error(err)
		return
	}
	defer tempConfigFile.Close()
	defer os.RemoveAll(tempConfigFile.Name())
	_, err = tempConfigFile.Write([]byte(yamllintConf))
	if err != nil {
		tT.Error(err)
		return
	}
	cmd := exec.Command("yamllint", "-c", tempConfigFile.Name(), "-")
	cmd.Stdin = strings.NewReader(raw)
	out, err := cmd.CombinedOutput()
	if err != nil {
		tT.Errorf("[%s@%s] %s failed lint checks against %s: %s", t.Chart.Metadata.Name, t.Chart.Metadata.Version, templateFile, t.Options, err)
		w := writer.NewOutputWriter(
			tT,
			filepath.Join(t.Chart.Metadata.Name, t.Chart.Metadata.Version, templateFile),
			cmd.String(),
			raw,
		)
		w = io.MultiWriter(w, testErrorWriter{tT})
		if _, err := w.Write(out); err != nil {
			tT.Error(err)
		}
	}
}

type testErrorWriter struct {
	*testing.T
}

func (w testErrorWriter) Write(p []byte) (n int, err error) {
	w.T.Error(string(p))
	return len(p), nil
}

type HelmLintOptions struct {
	Rancher RancherHelmLintOptions
}

type RancherHelmLintOptions struct {
	Enabled bool

	IgnoreAnnotations []string
}

func (t *template) HelmLint(tT *testing.T, opts *HelmLintOptions) {
	if opts == nil {
		opts = &HelmLintOptions{}
	}
	// Construct linter
	lint := helmAction.NewLint()
	lint.Namespace = t.Options.Release.Namespace
	lint.Strict = true

	// Grab all subchart paths
	paths := []string{t.Chart.Path}
	filepath.Walk(filepath.Join(t.Chart.Path, "charts"), func(path string, info os.FileInfo, err error) error {
		if info != nil && info.Name() == "Chart.yaml" {
			paths = append(paths, filepath.Dir(path))
		}
		return nil
	})

	lintResult := lint.Run(paths, t.Values)

	// Add additional custom lints
	if err := t.validateRancherAnnotations(opts.Rancher); err != nil {
		msg := helmLintSupport.NewMessage(helmLintSupport.ErrorSev, "Chart.yaml", err)
		lintResult.Messages = append(lintResult.Messages, msg)
	}

	// log errors
	errMap := map[string]error{}
	for _, msg := range lintResult.Messages {
		switch msg.Severity {
		case helmLintSupport.InfoSev, helmLintSupport.WarningSev:
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
	valArgs := toValuesArgs(t.Options.Values)
	if len(valArgs) > 0 {
		command += " " + valArgs
	}
	for path, err := range errMap {
		tT.Error(err)
		w := writer.NewOutputWriter(
			tT,
			filepath.Join(t.Chart.Metadata.Name, t.Chart.Metadata.Version, path),
			command,
			rawFiles[path],
		)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			tT.Error(err)
		}
	}
}

func (t *template) Check(tT *testing.T, objStructFunc checker.CheckFunc) {
	if t.ObjectSets == nil {
		return
	}
	check, err := checker.NewChecker(t.ObjectSets)
	if err != nil {
		tT.Error(err)
		return
	}
	check.Check(tT, objStructFunc)
}
