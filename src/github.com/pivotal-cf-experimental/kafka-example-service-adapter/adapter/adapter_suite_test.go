package adapter_test

import (
	"bytes"
	"io"
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/kafka-example-service-adapter/adapter"
	"github.com/pivotal-cf-experimental/kafka-example-service-adapter/adapter/fake_command_runner"

	"testing"
)

var (
	manifestGenerator   *adapter.ManifestGenerator
	binder              *adapter.Binder
	expectedCommandName = "command-to-create-topic"
	fakeCommandRunner   *fake_command_runner.FakeCommandRunner
	stderr              bytes.Buffer
)

var _ = BeforeEach(func() {
	fakeCommandRunner = new(fake_command_runner.FakeCommandRunner)
	manifestGenerator = &adapter.ManifestGenerator{
		StderrLogger: log.New(io.MultiWriter(GinkgoWriter, &stderr), "", log.LstdFlags),
	}
	binder = &adapter.Binder{
		TopicCreatorCommand: expectedCommandName,
		CommandRunner:       fakeCommandRunner,
		StderrLogger:        log.New(io.MultiWriter(GinkgoWriter, &stderr), "", log.LstdFlags),
	}
})

func TestAdapter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Adapter Suite")
}
