package adapter

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"
)

func (b *Binder) CreateBinding(bindingId string, boshVMs bosh.BoshVMs, manifest bosh.BoshManifest, requestParams serviceadapter.RequestParameters) (serviceadapter.Binding, error) {
	params := requestParams.ArbitraryParams()

	var invalidParams []string
	for paramKey, _ := range params {
		if paramKey != "topic" {
			invalidParams = append(invalidParams, paramKey)
		}
	}

	if len(invalidParams) > 0 {
		sort.Strings(invalidParams)
		errorMessage := fmt.Sprintf("unsupported parameter(s) for this service: %s", strings.Join(invalidParams, ", "))
		b.StderrLogger.Println(errorMessage)
		return serviceadapter.Binding{}, errors.New(errorMessage)
	}

	kafkaHosts := boshVMs["kafka_server"]
	if len(kafkaHosts) == 0 {
		b.StderrLogger.Println("no VMs for instance group kafka_server")
		return serviceadapter.Binding{}, errors.New("")
	}

	var kafkaAddresses []interface{}
	for _, kafkaHost := range kafkaHosts {
		kafkaAddresses = append(kafkaAddresses, fmt.Sprintf("%s:9092", kafkaHost))
	}

	zookeeperServers := boshVMs["zookeeper_server"]
	if len(zookeeperServers) == 0 {
		b.StderrLogger.Println("no VMs for job zookeeper_server")
		return serviceadapter.Binding{}, errors.New("")
	}

	if _, errorStream, err := b.Run(b.TopicCreatorCommand, strings.Join(zookeeperServers, ","), bindingId); err != nil {
		if strings.Contains(string(errorStream), "kafka.common.TopicExistsException") {
			b.StderrLogger.Println(fmt.Sprintf("topic '%s' already exists", bindingId))
			return serviceadapter.Binding{}, serviceadapter.NewBindingAlreadyExistsError()
		}
		b.StderrLogger.Println("Error creating topic: " + err.Error())
		return serviceadapter.Binding{}, errors.New("")
	}

	if params["topic"] != nil {
		if _, _, err := b.Run(b.TopicCreatorCommand, strings.Join(zookeeperServers, ","), params["topic"].(string)); err != nil {
			b.StderrLogger.Println("Error creating topic: " + err.Error())
			return serviceadapter.Binding{}, errors.New("")
		}
	}

	return serviceadapter.Binding{
		Credentials: map[string]interface{}{
			"bootstrap_servers": kafkaAddresses,
		},
	}, nil
}

//go:generate counterfeiter -o fake_command_runner/fake_command_runner.go . CommandRunner
type CommandRunner interface {
	Run(name string, arg ...string) ([]byte, []byte, error)
}

type ExternalCommandRunner struct{}

func (c ExternalCommandRunner) Run(name string, arg ...string) ([]byte, []byte, error) {
	cmd := exec.Command(name, arg...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.Output()
	return stdout, stderr.Bytes(), err
}
