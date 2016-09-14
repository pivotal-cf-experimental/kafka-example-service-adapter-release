package integration_tests

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-binding subcommand", func() {
	var (
		bindingID     string
		boshManifest  string
		boshVMs       string
		requestParams string

		stdout   bytes.Buffer
		exitCode int
	)

	BeforeEach(func() {
		bindingID = "some-binding"
		boshManifest = "" // unused
		boshVMs = `{
			"kafka_server": ["broker1ip", "broker2ip"],
			"zookeeper_server": ["zookeeper1ip", "zookeeper2ip"]
		}`
		requestParams = "{}" // unused

		stdout = bytes.Buffer{}

		deleteTopicExecutable, err := gexec.Build("github.com/pivotal-cf-experimental/kafka-example-service-adapter/integration_tests/mock_executable")
		Expect(err).NotTo(HaveOccurred())
		os.Setenv("TOPIC_DELETER_COMMAND", deleteTopicExecutable)
		file, err := ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())
		file.Close()
		os.Setenv("TEST_PARAMS_FILE_NAME", file.Name())
	})

	AfterEach(func() {
		Expect(os.Remove(os.Getenv("TEST_PARAMS_FILE_NAME"))).To(Succeed())
	})

	JustBeforeEach(func() {
		cmd := exec.Command(serviceAdapterBinPath, "delete-binding", bindingID, boshVMs, boshManifest, requestParams)
		runningBin, err := gexec.Start(cmd, io.MultiWriter(GinkgoWriter, &stdout), GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(runningBin).Should(gexec.Exit())
		exitCode = runningBin.ExitCode()
	})

	Context("when the binding exists", func() {
		It("succeeds", func() {
			Expect(exitCode).To(Equal(0))
		})

		It("produces no output", func() {
			Expect(stdout.String()).To(Equal(""))
		})

		It("should call the delete topic binary", func() {
			Expect(paramsReceivedByExecutable()[1:]).To(ConsistOf("zookeeper1ip,zookeeper2ip", bindingID))
		})
	})

	Context("when the binding cannot be found", func() {
		BeforeEach(func() {
			os.Setenv("TEST_EXECUTABLE_SHOULD_FAIL", "Topic some-binding does not exist on ZK path")
		})

		It("exits with code 41", func() {
			Expect(exitCode).To(Equal(41))
		})

		It("produces no output", func() {
			Expect(stdout.String()).To(Equal(""))
		})
	})
})
