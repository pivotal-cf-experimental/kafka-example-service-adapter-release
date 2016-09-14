package main

import (
	"log"
	"os"

	"github.com/pivotal-cf-experimental/kafka-example-service-adapter/adapter"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"
)

func main() {
	topicCreatorCommand := os.Getenv("TOPIC_CREATOR_COMMAND")
	if topicCreatorCommand == "" {
		topicCreatorCommand = "/var/vcap/packages/topic_manager/topic_creator"
	}
	topicDeleterCommand := os.Getenv("TOPIC_DELETER_COMMAND")
	if topicDeleterCommand == "" {
		topicDeleterCommand = "/var/vcap/packages/topic_manager/topic_deleter"
	}
	stderrLogger := log.New(os.Stderr, "[kafka-service-adapter] ", log.LstdFlags)
	manifestGenerator := &adapter.ManifestGenerator{
		StderrLogger: stderrLogger,
	}
	binder := &adapter.Binder{
		CommandRunner:       adapter.ExternalCommandRunner{},
		TopicCreatorCommand: topicCreatorCommand,
		TopicDeleterCommand: topicDeleterCommand,
		StderrLogger:        stderrLogger,
	}
	serviceadapter.HandleCommandLineInvocation(os.Args, manifestGenerator, binder, &adapter.DashboardUrlGenerator{})
}
