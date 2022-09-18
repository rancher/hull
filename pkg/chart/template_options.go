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

func (o *TemplateOptions) SetKubeVersion(version string) *TemplateOptions {
	kubeVersion, err := helmChartUtil.ParseKubeVersion(version)
	if err != nil {
		return nil
	}
	if o.Capabilities == nil {
		o.Capabilities = &helmChartUtil.Capabilities{}
	}
	o.Capabilities.KubeVersion = *kubeVersion
	return o
}

func (o *TemplateOptions) SetValue(key, value string) *TemplateOptions {
	if o.ValuesOptions == nil {
		o.ValuesOptions = &helmValues.Options{}
	}
	o.ValuesOptions.Values = append(o.ValuesOptions.Values, fmt.Sprintf("%s=%s", key, value))
	return o
}

func (o *TemplateOptions) IsUpgrade(isUpgrade bool) *TemplateOptions {
	o.Release.IsUpgrade = isUpgrade
	return o
}

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
	args := "helm template"
	if len(o.Release.Namespace) > 0 {
		args += " -n " + o.Release.Namespace
	}
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
	if len(o.Release.Name) > 0 {
		args += fmt.Sprintf(" %s", o.Release.Name)
	}
	args += " <path-to-chart>"
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
	return fmt.Sprintf("--kube-version %s", capOpts.KubeVersion.Version)
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
