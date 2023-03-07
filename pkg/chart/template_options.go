package chart

import (
	"fmt"

	helmChartUtil "helm.sh/helm/v3/pkg/chartutil"
)

func NewTemplateOptions(name, namespace string) *TemplateOptions {
	o := &TemplateOptions{
		Release: Release{
			Name:      name,
			Namespace: namespace,
		},
	}
	return o
}

type TemplateOptions struct {
	Values       *Values
	Release      Release
	Capabilities *Capabilities
}

func (o *TemplateOptions) SetKubeVersion(version string) *TemplateOptions {
	kubeVersion, err := helmChartUtil.ParseKubeVersion(version)
	if err != nil {
		panic(fmt.Errorf("invalid kubeVersion %s provided: %s", version, err))
	}
	if o.Capabilities == nil {
		o.Capabilities = &Capabilities{}
	}
	o.Capabilities.KubeVersion = *kubeVersion
	return o
}

func (o *TemplateOptions) SetValue(key, value string) *TemplateOptions {
	o.Values = o.Values.SetValue(key, value)
	return o
}

func (o *TemplateOptions) Set(key, value interface{}) *TemplateOptions {
	o.Values = o.Values.Set(key, value)
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
		o.Capabilities = (*Capabilities)(helmChartUtil.DefaultCapabilities)
	}
	if o.Values == nil {
		o.Values = NewValues()
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
	valArgs := toValuesArgs(o.Values)
	if len(valArgs) > 0 {
		args += " " + valArgs
	}
	if len(o.Release.Name) > 0 {
		args += fmt.Sprintf(" %s", o.Release.Name)
	}
	args += " <path-to-chart>"
	return args
}
