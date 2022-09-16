package chart

import (
	"fmt"
	"strings"

	helmChartUtil "helm.sh/helm/v3/pkg/chartutil"
	helmValues "helm.sh/helm/v3/pkg/cli/values"
)

func NewTemplateOptions(name, namespace string) *TemplateOptions {
	o := &TemplateOptions{
		Release: helmChartUtil.ReleaseOptions{
			Name:      name,
			Namespace: namespace,
		},
	}
	return o
}

type TemplateOptions struct {
	ValuesOptions *helmValues.Options
	Release       helmChartUtil.ReleaseOptions
	Capabilities  *helmChartUtil.Capabilities
}

func (o *TemplateOptions) SetKubeVersion(version string) error {
	kubeVersion, err := helmChartUtil.ParseKubeVersion(version)
	if err != nil {
		return err
	}
	if o.Capabilities == nil {
		o.Capabilities = &helmChartUtil.Capabilities{}
	}
	o.Capabilities.KubeVersion = *kubeVersion
	return nil
}

func (o *TemplateOptions) SetValue(key, value string) *TemplateOptions {
	if o.ValuesOptions == nil {
		o.ValuesOptions = &helmValues.Options{}
	}
	o.ValuesOptions.Values = append(o.ValuesOptions.Values, fmt.Sprintf("%s=%s", key, value))
	return o
}

// func (o *TemplateOptions) SetValueSlice(key, values []string) {
// 	if o.ValuesOptions == nil {
// 		o.ValuesOptions = &helmValues.Options{}
// 	}
// 	valuesStr := "{"
// 	for _, value := range values {
// 		if len(valuesStr) > 1 {
// 			valuesStr += ", "
// 		}
// 		valuesStr += value
// 	}
// 	valuesStr += "}"
// 	o.ValuesOptions.Values = append(o.ValuesOptions.Values, fmt.Sprintf("%s=%s", key, valuesStr))
// }

// func (o *TemplateOptions) SetValueObject(key, values map[string]interface{}) []string {
// 	var setValues []string
// 	var helper func(string, interface{})
// 	helper = func(prefix string, value interface{}) {
// 		switch val := value.(type) {
// 		case map[string]interface{}:
// 			for k, v := range val {
// 				if len(prefix) > 0 {
// 					prefix = prefix + "."
// 				}
// 				helper(fmt.Sprintf("%s%s", prefix, k), v)
// 			}
// 		case []interface{}:
// 			for i, v := range val {
// 				helper(fmt.Sprintf("%s[%d]", prefix, i), v)
// 			}
// 		default:
// 			setString := fmt.Sprintf("%s=%s", prefix, val)
// 			setValues = append(setValues, setString)
// 		}
// 	}
// 	helper("", values)
// 	return setValues
// }

func (o *TemplateOptions) setDefaults(chart string) *TemplateOptions {
	if o == nil {
		o = &TemplateOptions{}
	}
	if len(o.Release.Name) == 0 {
		o.Release.Name = chart
	}
	if len(o.Release.Namespace) == 0 {
		o.Release.Namespace = "default"
	}
	if !o.Release.IsInstall && !o.Release.IsUpgrade {
		o.Release.IsInstall = true
	}
	if o.Capabilities == nil {
		o.Capabilities = helmChartUtil.DefaultCapabilities
	}
	if o.ValuesOptions == nil {
		o.ValuesOptions = &helmValues.Options{}
	}
	return o
}

func (o TemplateOptions) String() string {
	args := fmt.Sprintf("helm template -n %s", o.Release.Namespace)
	relArgs := toReleaseArgs(o.Release)
	if len(relArgs) > 0 {
		args += " " + relArgs
	}
	capArgs := toCapabilitiesArgs(o.Capabilities)
	if len(capArgs) > 0 {
		args += " " + capArgs
	}
	valArgs := toValuesArgs(o.ValuesOptions)
	if len(valArgs) > 0 {
		args += " " + valArgs
	}
	args += fmt.Sprintf(" %s <path-to-chart>", o.Release.Name)
	return args
}

func toReleaseArgs(relOpts helmChartUtil.ReleaseOptions) string {
	if relOpts.IsUpgrade {
		return "--is-upgrade"
	}
	return ""
}

func toCapabilitiesArgs(capOpts *helmChartUtil.Capabilities) string {
	if capOpts == nil || capOpts == helmChartUtil.DefaultCapabilities {
		return ""
	}
	return fmt.Sprintf("--kube-version %s", capOpts.KubeVersion)
}

func toValuesArgs(valOpts *helmValues.Options) string {
	if valOpts == nil {
		return ""
	}
	var args string
	if len(valOpts.ValueFiles) > 0 {
		args += fmt.Sprintf(" -f %s", strings.Join(valOpts.ValueFiles, ","))
	}
	if len(valOpts.Values) > 0 {
		args += fmt.Sprintf(" --set %s", strings.Join(valOpts.Values, ","))
	}
	if len(valOpts.StringValues) > 0 {
		args += fmt.Sprintf(" --set-string %s", strings.Join(valOpts.StringValues, ","))
	}
	if len(valOpts.FileValues) > 0 {
		args += fmt.Sprintf(" --set-file %s", strings.Join(valOpts.FileValues, ","))
	}
	return strings.TrimPrefix(args, " ")
}
