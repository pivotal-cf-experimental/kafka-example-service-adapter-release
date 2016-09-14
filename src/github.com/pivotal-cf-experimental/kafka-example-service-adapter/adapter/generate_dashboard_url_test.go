package adapter_test

import (
	"github.com/pivotal-cf-experimental/kafka-example-service-adapter/adapter"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Adapter/GenerateDashboardUrl", func() {
	It("generates a arbitrary dashboard url", func() {
		generator := &adapter.DashboardUrlGenerator{}
		Expect(generator.DashboardUrl("instanceID", serviceadapter.Plan{}, bosh.BoshManifest{})).To(Equal(serviceadapter.DashboardUrl{DashboardUrl: "http://example_dashboard.com/instanceID"}))
	})
})
