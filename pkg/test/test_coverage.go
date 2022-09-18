package test

import (
	"reflect"
	"testing"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/test/internal"
	helmValues "helm.sh/helm/v3/pkg/cli/values"
)

func Coverage(t *testing.T, valuesStruct interface{}, opts ...*chart.TemplateOptions) (float64, string) {
	coverage, report, err := CoverageE(valuesStruct, opts...)
	if err != nil {
		t.Error(err)
		return 0, ""
	}
	return coverage, report
}

func CoverageE(valuesStruct interface{}, opts ...*chart.TemplateOptions) (float64, string, error) {
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
	mergedValueOpts := mergeValuesOpts(valueOpts...)
	values, err := mergedValueOpts.MergeValues(nil)
	if err != nil {
		return 0, "", err
	}
	coverage, report := internal.CalculateCoverage(values, reflect.TypeOf(valuesStruct))
	return coverage, report, nil
}

func mergeValuesOpts(opts ...helmValues.Options) helmValues.Options {
	valueOpts := helmValues.Options{}
	for _, opt := range opts {
		valueOpts.FileValues = append(valueOpts.FileValues, opt.FileValues...)
		valueOpts.StringValues = append(valueOpts.StringValues, opt.StringValues...)
		valueOpts.ValueFiles = append(valueOpts.ValueFiles, opt.ValueFiles...)
		valueOpts.Values = append(valueOpts.Values, opt.Values...)

	}
	return valueOpts
}
