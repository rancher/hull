package chart

import (
	"encoding/json"
	"fmt"
	"strings"

	helmValues "helm.sh/helm/v3/pkg/cli/values"
)

type Values helmValues.Options

func NewValues() *Values {
	return &Values{}
}

func (v *Values) SetValue(key, value string) *Values {
	if v == nil {
		v = &Values{}
	}
	v.Values = append(v.Values, fmt.Sprintf("%s=%s", key, value))
	return v
}

func (v *Values) Set(key, value interface{}) *Values {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Errorf("cannot marshall value %T (%s): %s", value, value, err))
	}
	if v == nil {
		v = &Values{}
	}
	v.JSONValues = append(v.JSONValues, fmt.Sprintf("%s=%s", key, jsonValue))
	return v
}

func (v *Values) MergeValues(values ...*Values) *Values {
	ret := &Values{}
	for _, value := range append([]*Values{v}, values...) {
		if value == nil {
			continue
		}
		ret.ValueFiles = append(ret.ValueFiles, value.ValueFiles...)
		ret.Values = append(ret.Values, value.Values...)
		ret.StringValues = append(ret.StringValues, value.StringValues...)
		ret.FileValues = append(ret.FileValues, value.FileValues...)
		ret.JSONValues = append(ret.JSONValues, value.JSONValues...)
	}
	return ret
}

func (v *Values) ToMap() (map[string]interface{}, error) {
	return (*helmValues.Options)(v).MergeValues(nil)
}

func toValuesArgs(valOpts *Values) string {
	if valOpts == nil {
		return ""
	}
	var args string
	if len(valOpts.ValueFiles) > 0 {
		for _, setArg := range valOpts.ValueFiles {
			args += fmt.Sprintf(" -f %s", setArg)
		}
	}
	if len(valOpts.Values) > 0 {
		for _, setArg := range valOpts.Values {
			args += fmt.Sprintf(" --set '%s'", setArg)
		}
	}
	if len(valOpts.StringValues) > 0 {
		for _, setArg := range valOpts.StringValues {
			args += fmt.Sprintf(" --set-string '%s'", setArg)
		}
	}
	if len(valOpts.FileValues) > 0 {
		for _, setArg := range valOpts.FileValues {
			args += fmt.Sprintf(" --set-file '%s'", setArg)
		}
	}
	if len(valOpts.JSONValues) > 0 {
		for _, setArg := range valOpts.JSONValues {
			args += fmt.Sprintf(" --set-json '%s'", setArg)
		}
	}
	return strings.TrimPrefix(args, " ")
}
