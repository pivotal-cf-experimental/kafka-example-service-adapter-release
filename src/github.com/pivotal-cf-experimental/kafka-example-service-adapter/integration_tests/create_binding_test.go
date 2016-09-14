package integration_tests

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"

	"github.com/onsi/gomega/gexec"
)

var _ = Describe("create-binding subcommand", func() {
	var (
		bindingID            string
		boshManifest         string
		boshVMs              string
		bindingRequestParams string

		stdout   bytes.Buffer
		stderr   bytes.Buffer
		exitCode int
	)

	BeforeEach(func() {
		bindingID = ""    // unused
		boshManifest = "" // unused
		boshVMs = `{
			"kafka_server": ["broker1ip", "broker2ip"],
			"zookeeper_server": ["zookeeper1ip", "zookeeper2ip"]
		}`
		bindingRequestParams = `{}`

		stdout = bytes.Buffer{}
		stderr = bytes.Buffer{}

		createTopicBinary, err := gexec.Build("github.com/pivotal-cf-experimental/kafka-example-service-adapter/integration_tests/mock_executable")
		Expect(err).NotTo(HaveOccurred())
		os.Setenv("TOPIC_CREATOR_COMMAND", createTopicBinary)
		file, err := ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())
		file.Close()
		os.Setenv("TEST_PARAMS_FILE_NAME", file.Name())
	})

	AfterEach(func() {
		Expect(os.Remove(os.Getenv("TEST_PARAMS_FILE_NAME"))).To(Succeed())
	})

	JustBeforeEach(func() {
		cmd := exec.Command(serviceAdapterBinPath, "create-binding", bindingID, boshVMs, boshManifest, bindingRequestParams)
		runningBin, err := gexec.Start(cmd, io.MultiWriter(GinkgoWriter, &stdout), io.MultiWriter(GinkgoWriter, &stderr))
		Expect(err).NotTo(HaveOccurred())
		Eventually(runningBin).Should(gexec.Exit())
		exitCode = runningBin.ExitCode()
	})

	Context("when the parameters are valid", func() {
		It("exits with 0", func() {
			Expect(exitCode).To(Equal(0))
		})

		It("prints a binding to stdout", func() {
			var binding serviceadapter.Binding
			Expect(json.Unmarshal(stdout.Bytes(), &binding)).To(Succeed())
			Expect(binding).To(Equal(
				serviceadapter.Binding{
					Credentials: map[string]interface{}{"bootstrap_servers": []interface{}{"broker1ip:9092", "broker2ip:9092"}},
				},
			))
		})

		Describe("creating topics", func() {
			BeforeEach(func() {
				bindingRequestParams = `{
					"parameters": {
						"topic": "foo"
					}
				}`
			})

			It("should call the create topic binary", func() {
				Expect(paramsReceivedByExecutable()[1:]).To(ConsistOf("zookeeper1ip,zookeeper2ip", "foo"))
			})
		})
	})

	Context("when there is an unknown arbitrary param", func() {
		BeforeEach(func() {
			bindingRequestParams = `{
				"parameters": {
					"unknown-param": "foo",
					"another-random-param": "bar",
					"topic": "baz"
				}
			}`
		})

		It("exits with 1", func() {
			Expect(exitCode).To(Equal(1))
		})

		It("outputs a user error message to stdout", func() {
			Expect(stdout.String()).To(ContainSubstring("unsupported parameter(s) for this service: another-random-param, unknown-param"))
		})

		It("outputs an operator error message to stderr", func() {
			Expect(stderr.String()).To(ContainSubstring("unsupported parameter(s) for this service: another-random-param, unknown-param"))
		})
	})
})

func paramsReceivedByExecutable() (params []string) {
	content, err := ioutil.ReadFile(os.Getenv("TEST_PARAMS_FILE_NAME"))
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Unmarshal(content, &params)).To(Succeed())
	return
}
