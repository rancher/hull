package chart

import (
	helmChartUtil "helm.sh/helm/v3/pkg/chartutil"
)

type Release helmChartUtil.ReleaseOptions

func toReleaseArgs(relOpts Release) string {
	if relOpts.IsUpgrade {
		return "--is-upgrade"
	}
	return ""
}
