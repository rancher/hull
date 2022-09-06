package resource

import (
	"sync"
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/test"
	"github.com/davecgh/go-spew/spew"
)

func TestValidStructs(t *testing.T) {
	var done sync.Once
	chart.TestManifests(t, func(t *testing.T, m *chart.Manifest) {
		done.Do(func() {
			runner, err := test.FromManifest(m)
			if err != nil {
				t.Fatal(err)
				return
			}
			runner.Run(t, nil, func(t *testing.T, a Resources) {
				// log contents to be able to inspect if it is able to pick up the resources in each category
				t.Log(spew.Sdump(a))
			})
		})
	})
}
