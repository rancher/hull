package chart

import (
	"fmt"
	"strings"

	helmChartUtil "helm.sh/helm/v3/pkg/chartutil"
	helmValues "helm.sh/helm/v3/pkg/cli/values"
)

type ManifestConfiguration struct {
	Name          string
	ValuesOptions *helmValues.Options
	Release       helmChartUtil.ReleaseOptions
	Capabilities  *helmChartUtil.Capabilities
}

func (c *ManifestConfiguration) setDefaults(chart string) (*ManifestConfiguration, error) {
	if c == nil {
		c = &ManifestConfiguration{}
	}
	if len(c.Release.Name) == 0 {
		c.Release.Name = chart
	}
	if len(c.Release.Namespace) == 0 {
		c.Release.Namespace = "default"
	}
	if !c.Release.IsInstall && !c.Release.IsUpgrade {
		c.Release.IsInstall = true
	}
	if c.Capabilities == nil {
		c.Capabilities = helmChartUtil.DefaultCapabilities
	}
	if c.ValuesOptions == nil {
		c.ValuesOptions = &helmValues.Options{}
	}
	return c, nil
}

func (c ManifestConfiguration) String() string {
	args := fmt.Sprintf("helm template -n %s", c.Release.Namespace)
	relArgs := toReleaseArgs(c.Release)
	if len(relArgs) > 0 {
		args += " " + relArgs
	}
	capArgs := toCapabilitiesArgs(c.Capabilities)
	if len(capArgs) > 0 {
		args += " " + capArgs
	}
	valArgs := toValuesArgs(c.ValuesOptions)
	if len(valArgs) > 0 {
		args += " " + valArgs
	}
	args += fmt.Sprintf(" %s <path-to-chart>", c.Release.Name)
	if len(c.Name) > 0 {
		args = fmt.Sprintf("%s (%s)", c.Name, args)
	}
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
	return fmt.Sprintf("--kube-version ", capOpts.KubeVersion)
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
