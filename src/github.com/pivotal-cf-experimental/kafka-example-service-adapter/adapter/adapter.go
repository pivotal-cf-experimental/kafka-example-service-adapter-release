package adapter

import (
	"log"

	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"
)

type ManifestGenerator struct {
	StderrLogger *log.Logger
}

type Binder struct {
	TopicCreatorCommand string
	TopicDeleterCommand string
	CommandRunner
	StderrLogger *log.Logger
}

var InstanceGroupMapper = serviceadapter.GenerateInstanceGroupsWithNoProperties
