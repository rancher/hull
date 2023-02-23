package checker

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/checker/internal"
	"github.com/aiyengar2/hull/pkg/parser"
	"github.com/rancher/wrangler/pkg/objectset"
)

type Checker interface {
	Check(t *testing.T, objStructFunc CheckFunc)
}

func NewChecker(osMap map[string]*objectset.ObjectSet) (Checker, error) {
	if osMap == nil {
		return nil, nil
	}
	if len(osMap) == 0 {
		return nil, nil
	}
	osMapCopy := make(map[string]*objectset.ObjectSet)
	rootOs := objectset.NewObjectSet()
	for osPath, os := range osMap {
		if osPath == "" {
			continue
		}
		if os == nil {
			continue
		}
		if os.Len() == 0 {
			continue
		}
		osMapCopy[osPath] = os
		rootOs.Add(os.All()...)
	}
	osMapCopy[""] = rootOs
	return &checker{
		ObjectSets: osMapCopy,
	}, nil
}

func NewCheckerFromObjectSet(os *objectset.ObjectSet, name string) (Checker, error) {
	if os == nil {
		return nil, nil
	}
	if os.Len() == 0 {
		return nil, nil
	}
	return NewChecker(map[string]*objectset.ObjectSet{
		name: os,
	})
}

func NewCheckerFromString(rawYaml string, name string) (Checker, error) {
	os, err := parser.Parse(rawYaml)
	if err != nil {
		return nil, err
	}
	if os == nil {
		return nil, nil
	}
	return NewCheckerFromObjectSet(os, name)
}

type checker struct {
	ObjectSets map[string]*objectset.ObjectSet
}

func (c *checker) Check(t *testing.T, objStructFunc CheckFunc) {
	if objStructFunc == nil {
		return
	}
	doFunc := internal.WrapFunc(objStructFunc, &internal.ParseOptions{
		Scheme: Scheme,
	})
	doFunc(t, c.ObjectSets[""].All())
}
