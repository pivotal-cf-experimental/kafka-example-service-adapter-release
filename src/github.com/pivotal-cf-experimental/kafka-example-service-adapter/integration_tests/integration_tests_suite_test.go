package integration_tests

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var serviceAdapterBinPath string

var _ = BeforeSuite(func() {
	var err error
	serviceAdapterBinPath, err = gexec.Build("github.com/pivotal-cf-experimental/kafka-example-service-adapter/cmd/service-adapter")
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	os.Unsetenv("TEST_EXECUTABLE_SHOULD_FAIL")
})

func TestIntegrationTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}
