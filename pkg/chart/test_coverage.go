package chart

import (
	"reflect"
	"testing"

	"github.com/aiyengar2/hull/pkg/chart/internal"
	helmValues "helm.sh/helm/v3/pkg/cli/values"
)

func Coverage(t *testing.T, valuesStruct interface{}, opts ...*TemplateOptions) float64 {
	coverage, err := CoverageE(valuesStruct, opts...)
	if err != nil {
		t.Error(err)
		return 0
	}
	return coverage
}

func CoverageE(valuesStruct interface{}, opts ...*TemplateOptions) (float64, error) {
	valueOpts := make([]helmValues.Options, len(opts))
	for i, opt := range opts {
		if opts == nil {
			continue
		}
		if opt.ValuesOptions == nil {
			continue
		}
		valueOpts[i] = *opt.ValuesOptions
	}
	values, err := mergeValues(valueOpts...).MergeValues(nil)
	if err != nil {
		return 0, err
	}
	return internal.CalculateCoverage(values, reflect.TypeOf(valuesStruct)), nil
}

func mergeValues(opts ...helmValues.Options) *helmValues.Options {
	valueOpts := &helmValues.Options{}
	for _, opt := range opts {
		valueOpts.FileValues = append(valueOpts.FileValues, opt.FileValues...)
		valueOpts.StringValues = append(valueOpts.StringValues, opt.StringValues...)
		valueOpts.ValueFiles = append(valueOpts.ValueFiles, opt.ValueFiles...)
		valueOpts.Values = append(valueOpts.Values, opt.Values...)

	}
	return valueOpts
}
