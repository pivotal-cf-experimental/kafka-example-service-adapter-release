package adapter

import (
	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"
)

type DashboardUrlGenerator struct {
}

func (a *DashboardUrlGenerator) DashboardUrl(instanceID string, plan serviceadapter.Plan, manifest bosh.BoshManifest) (serviceadapter.DashboardUrl, error) {
	return serviceadapter.DashboardUrl{DashboardUrl: "http://example_dashboard.com/" + instanceID}, nil
}
