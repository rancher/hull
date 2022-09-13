package checker

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/checker/internal"
	"github.com/rancher/helm-locker/pkg/objectset/parser"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	Scheme = runtime.NewScheme()
)

type Checker interface {
	Check(t *testing.T, opts *Options, objStructFunc interface{})
}

func NewChecker(objectsets map[string]*objectset.ObjectSet) Checker {
	return &checker{
		ObjectSets: objectsets,
	}
}

func NewCheckerFromObjectSet(os *objectset.ObjectSet) Checker {
	if os == nil {
		return nil
	}
	return &checker{
		ObjectSets: map[string]*objectset.ObjectSet{
			"": os,
		},
	}
}

func NewCheckerFromString(rawYaml string) (Checker, error) {
	os, err := parser.Parse(rawYaml)
	if err != nil {
		return nil, err
	}
	return NewCheckerFromObjectSet(os), nil
}

type checker struct {
	ObjectSets map[string]*objectset.ObjectSet
}

func (c *checker) Check(t *testing.T, opts *Options, objStructFunc interface{}) {
	if c == nil || c.ObjectSets == nil || len(c.ObjectSets) == 0 {
		return
	}
	if opts == nil {
		opts = &Options{}
	}
	doFunc := internal.WrapFunc(objStructFunc, &internal.ParseOptions{
		Scheme: Scheme,
	})
	if !opts.PerTemplateManifest {
		fullManifestOs, ok := c.ObjectSets[""]
		if !ok {
			t.Errorf("runner does not have valid template to execute tests")
			return
		}
		doFunc(t, fullManifestOs.All())
		return
	}
	if len(c.ObjectSets) <= 1 {
		// per template objectsets not provided
		t.Error("no templates to execute")
		return
	}
	for path, os := range c.ObjectSets {
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

type Options struct {
	PerTemplateManifest bool
}
