package test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	helmValues "helm.sh/helm/v3/pkg/cli/values"
)

func TestCoverage(t *testing.T) {
	type exampleStruct struct {
		First struct {
			Name int
		}
		Second struct {
			Third struct {
				NamePtr *float64
			}
		}
		Third struct {
			Fourth struct {
				Fifth struct {
					hidden string
				}
			}
		}
		Fourth struct {
			NameSlice []rune
			Fifth     struct {
				NameMap map[string]string
				Sixth   struct {
					NameSliceSlice [][]float64
					NameMapMap     map[string]map[string]int64
					Seventh        struct {
						NameSliceMap []map[string]byte
						NameMapSlice map[string][]interface{}
					}
				}
			}
		}
		Fifth []struct {
			NameHello string
			Sixth     map[string][]map[string]struct {
				Seventh map[string][]struct {
					NameWorld interface{}
				}
			}
		}
	}

	testCases := []struct {
		Name                 string
		Struct               interface{}
		TemplateOptionsSlice []*chart.TemplateOptions
		Coverage             float64
		ShouldThrowError     bool
	}{
		{
			Name:                 "No Options",
			Struct:               exampleStruct{},
			TemplateOptionsSlice: []*chart.TemplateOptions{},
			Coverage:             0,
			ShouldThrowError:     false,
		},
		{
			Name:   "Nil Options",
			Struct: exampleStruct{},
			TemplateOptionsSlice: []*chart.TemplateOptions{
				{
					ValuesOptions: nil,
				},
			},
			Coverage:         0,
			ShouldThrowError: false,
		},
		{
			Name:   "Bad Options",
			Struct: exampleStruct{},
			TemplateOptionsSlice: []*chart.TemplateOptions{
				{
					ValuesOptions: &helmValues.Options{
						Values: []string{"i-am-a-bad-option#2@"},
					},
				},
			},
			Coverage:         0,
			ShouldThrowError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.ShouldThrowError {
				fakeT := &testing.T{}
				Coverage(fakeT, tc.Struct, tc.TemplateOptionsSlice...)
				if !fakeT.Failed() {
					t.Error("expected error to be thrown")
				}
				return
			}
			coverage, report := Coverage(t, tc.Struct, tc.TemplateOptionsSlice...)

			if !t.Failed() {
				assert.Equal(t, tc.Coverage, coverage, report)
			}
		})
	}
}

func TestMergeValuesOpts(t *testing.T) {
	valueOptions := []helmValues.Options{}
	for i := 0; i < 10; i++ {
		randomArr := make([]string, 100)
		for j := 0; j < 100; j++ {
			randomArr[j] = uuid.NewString()
		}
		randomIndices := []int{rand.Intn(100), rand.Intn(100), rand.Intn(100)}
		sort.Ints(randomIndices)
		valueOptions = append(valueOptions, helmValues.Options{
			FileValues:   randomArr[:randomIndices[0]],
			StringValues: randomArr[randomIndices[0]:randomIndices[1]],
			ValueFiles:   randomArr[randomIndices[1]:randomIndices[2]],
			Values:       randomArr[randomIndices[2]:],
		})
	}
	values := mergeValuesOpts(valueOptions...)
	assert.Equal(t, 10*100, len(values.FileValues)+len(values.StringValues)+len(values.ValueFiles)+len(values.Values))
}
