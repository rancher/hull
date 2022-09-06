package test

import (
	"fmt"
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/test/internal"
	"github.com/rancher/helm-locker/pkg/objectset/parser"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	Scheme = runtime.NewScheme()
)

func FromManifest(m *chart.Manifest) (*Runner, error) {
	if m == nil {
		return nil, fmt.Errorf("cannot create runner from nil manifest")
	}
	osMap, err := m.ToObjectSetMap()
	return &Runner{
		templateToObjectSet: osMap,
	}, err
}

func FromObjectSet(os *objectset.ObjectSet) (*Runner, error) {
	if os == nil {
		return nil, fmt.Errorf("cannot create runner from nil objects")
	}
	return &Runner{
		templateToObjectSet: map[string]*objectset.ObjectSet{
			"": os,
		},
	}, nil
}

func FromString(rawYaml string) (*Runner, error) {
	os, err := parser.Parse(rawYaml)
	if err != nil {
		return nil, err
	}
	return FromObjectSet(os)
}

type Runner struct {
	templateToObjectSet map[string]*objectset.ObjectSet
}

func (r *Runner) Run(t *testing.T, opts *RunnerOptions, objStructFunc interface{}) {
	if r == nil || r.templateToObjectSet == nil {
		return
	}
	if opts == nil {
		opts = &RunnerOptions{}
	}
	if len(r.templateToObjectSet) == 0 {
		t.Error("no templates to execute")
		return
	}
	doFunc := internal.WrapFunc(objStructFunc, &internal.ParseOptions{
		Scheme: Scheme,
	})
	if !opts.PerTemplateManifest {
		fullManifestOs, ok := r.templateToObjectSet[""]
		if !ok {
			t.Errorf("runner does not have valid template to execute tests")
			return
		}
		doFunc(t, fullManifestOs.All())
		return
	}
	if len(r.templateToObjectSet) <= 1 {
		// per template objectsets not provided
		t.Error("no templates to execute")
		return
	}
	for path, os := range r.templateToObjectSet {
		if path == "" {
			continue
		}
		t.Run(path, func(t *testing.T) {
			doFunc(t, os.All())
		})
	}
}

// Per test contexts:
// run on each templatemanifest, not overall (single resource tests)
// exclude a set of manifests or exclude those that don't belong in a certain set (windows, linux)

type RunnerOptions struct {
	PerTemplateManifest bool
}
