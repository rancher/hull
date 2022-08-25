package chart

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aiyengar2/hull/pkg/utils"
)

func VerifyPolicies(t *testing.T, policyPaths ...string) {
	args := []string{"verify", "--no-color"}
	for _, pp := range policyPaths {
		args = append(args, "-p")
		args = append(args, filepath.Join(utils.GetRepoRoot(), pp))
	}
	cmd := exec.Command("conftest", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, err := os.Stderr.Write(out); err != nil {
			t.Error(err)
		}
	}
}

func (m *Manifest) EnforcePolicies(t *testing.T, combine bool, policyPaths ...string) {
	if m == nil {
		return
	}
	objs, err := m.ToObjectSet()
	if err != nil {
		t.Error(err)
		return
	}
	if objs.Len() == 0 {
		t.Errorf("[%s@%s] chart %s has no content", m.ChartMetadata.Name, m.ChartMetadata.Version, m.Path)
		return
	}
	VerifyPolicies(t, policyPaths...)

	args := []string{"test", "-", "--no-color"}
	if combine {
		args = append(args, "--combine")
	}
	for _, pp := range policyPaths {
		args = append(args, "-p")
		args = append(args, filepath.Join(utils.GetRepoRoot(), pp))
	}
	cmd := exec.Command("conftest", args...)
	cmd.Stdin = strings.NewReader(m.Raw())
	out, err := cmd.CombinedOutput()
	if err != nil {
		command := "conftest " + strings.Join(args, " ")
		command = strings.Replace(command, "-", "<"+m.Configuration.String()+" : "+m.Path+">", 1)
		t.Errorf("[%s@%s] chart %s failed policy enforcement against %s", m.ChartMetadata.Name, m.ChartMetadata.Version, m.Path, command)
		w := NewChartPathWriter(
			m.ChartMetadata.Name,
			m.ChartMetadata.Version,
			"manifest.yaml",
			command,
			m.Raw(),
		)
		w = io.MultiWriter(w, os.Stderr)
		if _, err := w.Write(out); err != nil {
			t.Error(err)
		}
	}

	if combine {
		return
	}
	for _, tm := range m.sorted() {
		tm.EnforcePolicies(t, policyPaths...)
	}
}

func (m *TemplateManifest) EnforcePolicies(t *testing.T, policyPaths ...string) {
	if m == nil {
		return
	}
	objs, err := m.ToObjectSet()
	if err != nil {
		t.Error(err)
		return
	}
	if objs.Len() == 0 {
		return
	}
	VerifyPolicies(t, policyPaths...)

	args := []string{"test", "-", "--no-color"}
	for _, pp := range policyPaths {
		args = append(args, "-p")
		args = append(args, filepath.Join(utils.GetRepoRoot(), pp))
	}
	cmd := exec.Command("conftest", args...)
	cmd.Stdin = strings.NewReader(m.Raw())
	out, err := cmd.CombinedOutput()
	if err != nil {
		command := "conftest " + strings.Join(args, " ")
		command = strings.Replace(command, "-", "<"+m.ManifestConfiguration.String()+" : "+m.TemplateFile+">", 1)
		t.Errorf("[%s@%s] %s failed policy enforcement against %s", m.ChartMetadata.Name, m.ChartMetadata.Version, m.TemplateFile, command)
		w := NewChartPathWriter(
			m.ChartMetadata.Name,
			m.ChartMetadata.Version,
			m.TemplateFile,
			command,
			m.Raw(),
		)
		w = io.MultiWriter(w, os.Stderr)
		if _, err := w.Write(out); err != nil {
			t.Error(err)
		}
	}
}
