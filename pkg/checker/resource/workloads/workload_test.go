package workloads

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/checker/internal"
	"github.com/davecgh/go-spew/spew"
)

func TestValidStructs(t *testing.T) {
	doFunc := internal.WrapFunc(testResources, nil)
	doFunc(t, nil)
}

type resources struct {
	CronJobs
	DaemonSets
	Deployments
	Jobs
	StatefulSets
	ConfigMaps
	Secrets
	PodSecurityPolicies
	HorizontalPodAutoscalers
}

func testResources(t *testing.T, resources resources) {
	// log contents to be able to inspect if it is able to pick up the resources in each category
	t.Log(spew.Sdump(resources))
}
