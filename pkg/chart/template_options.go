package chart

import (
	"fmt"
	"strings"

	helmChartUtil "helm.sh/helm/v3/pkg/chartutil"
	helmValues "helm.sh/helm/v3/pkg/cli/values"
)

type TemplateOptions struct {
	ValuesOptions *helmValues.Options
	Release       helmChartUtil.ReleaseOptions
	Capabilities  *helmChartUtil.Capabilities
}

func (o *TemplateOptions) setDefaults(chart string) (*TemplateOptions, error) {
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
	return o, nil
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
