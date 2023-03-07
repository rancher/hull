package chart

import (
	"fmt"

	helmChartUtil "helm.sh/helm/v3/pkg/chartutil"
)

type Capabilities helmChartUtil.Capabilities

func toCapabilitiesArgs(capOpts *Capabilities) string {
	if capOpts == nil || capOpts == (*Capabilities)(helmChartUtil.DefaultCapabilities) {
		return ""
	}
	return fmt.Sprintf("--kube-version '%s'", capOpts.KubeVersion.Version)
}
