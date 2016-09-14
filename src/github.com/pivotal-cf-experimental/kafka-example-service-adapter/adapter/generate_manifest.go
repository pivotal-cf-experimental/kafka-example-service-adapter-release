package adapter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"
)

const OnlyStemcellAlias = "only-stemcell"

func defaultDeploymentInstanceGroupsToJobs() map[string][]string {
	return map[string][]string{
		"kafka_server":     []string{"kafka_server"},
		"zookeeper_server": []string{"zookeeper_server"},
		"smoke_tests":      []string{"smoke_tests"},
	}
}

func (a *ManifestGenerator) GenerateManifest(serviceDeployment serviceadapter.ServiceDeployment,
	servicePlan serviceadapter.Plan,
	requestParams serviceadapter.RequestParameters,
	previousManifest *bosh.BoshManifest,
	previousPlan *serviceadapter.Plan,
) (bosh.BoshManifest, error) {

	if previousPlan != nil {
		prev := instanceCounts(*previousPlan)
		current := instanceCounts(servicePlan)
		if (prev["kafka_server"] > current["kafka_server"]) || (prev["zookeeper_server"] > current["zookeeper_server"]) {
			a.StderrLogger.Println("cannot migrate to a smaller plan")
			return bosh.BoshManifest{}, errors.New("")
		}
	}

	var releases []bosh.Release

	loggingRaw, ok := servicePlan.Properties["logging"]
	includeMetron := false
	if ok {
		includeMetron = true
	}

	for _, serviceRelease := range serviceDeployment.Releases {
		releases = append(releases, bosh.Release{
			Name:    serviceRelease.Name,
			Version: serviceRelease.Version,
		})
	}

	deploymentInstanceGroupsToJobs := defaultDeploymentInstanceGroupsToJobs()
	if includeMetron {
		for instanceGroup, jobs := range deploymentInstanceGroupsToJobs {
			deploymentInstanceGroupsToJobs[instanceGroup] = append(jobs, "metron_agent")
		}
	}

	err := checkInstanceGroupsPresent([]string{"kafka_server", "zookeeper_server"}, servicePlan.InstanceGroups)
	if err != nil {
		a.StderrLogger.Println(err.Error())
		return bosh.BoshManifest{}, errors.New("Contact your operator, service configuration issue occurred")
	}

	instanceGroups, err := InstanceGroupMapper(servicePlan.InstanceGroups, serviceDeployment.Releases, OnlyStemcellAlias, deploymentInstanceGroupsToJobs)
	if err != nil {
		a.StderrLogger.Println(err.Error())
		return bosh.BoshManifest{}, errors.New("")
	}

	kafkaBrokerInstanceGroup := &instanceGroups[0]

	if len(kafkaBrokerInstanceGroup.Networks) != 1 {
		a.StderrLogger.Println(fmt.Sprintf("expected 1 network for %s, got %d", kafkaBrokerInstanceGroup.Name, len(kafkaBrokerInstanceGroup.Networks)))
		return bosh.BoshManifest{}, errors.New("")
	}

	autoCreateTopics := true
	arbitraryParameters := requestParams.ArbitraryParams()

	if arbitraryVal, ok := arbitraryParameters["auto_create_topics"]; ok {
		autoCreateTopics = arbitraryVal.(bool)
	} else if previousVal, previousOk := getPreviousManifestProperty("auto_create_topics", previousManifest); previousOk {
		autoCreateTopics = previousVal.(bool)
	} else if planVal, ok := servicePlan.Properties["auto_create_topics"]; ok {
		autoCreateTopics = planVal.(bool)
	}

	defaultReplicationFactor := 3
	if val, ok := servicePlan.Properties["default_replication_factor"]; ok {
		defaultReplicationFactor = int(val.(float64))
	}

	if kafkaBrokerJob, ok := getJobFromInstanceGroup("kafka_server", kafkaBrokerInstanceGroup); ok {
		kafkaBrokerJob.Properties = map[string]interface{}{
			"default_replication_factor": defaultReplicationFactor,
			"auto_create_topics":         autoCreateTopics,
			"network":                    kafkaBrokerInstanceGroup.Networks[0].Name,
		}
	}

	manifestProperties := map[string]interface{}{}

	if includeMetron {
		logging := loggingRaw.(map[string]interface{})
		manifestProperties["syslog_daemon_config"] = map[interface{}]interface{}{
			"address": logging["syslog_address"],
			"port":    logging["syslog_port"],
		}
		manifestProperties["metron_agent"] = map[interface{}]interface{}{
			"zone":       "",
			"deployment": serviceDeployment.DeploymentName,
		}
		manifestProperties["loggregator"] = map[interface{}]interface{}{
			"etcd": map[interface{}]interface{}{
				"machines": logging["loggregator_etcd_addresses"].([]interface{}),
			},
		}
		manifestProperties["metron_endpoint"] = map[interface{}]interface{}{
			"shared_secret": logging["loggregator_shared_secret"],
		}
	}

	var updateBlock = bosh.Update{
		Canaries:        1,
		MaxInFlight:     10,
		CanaryWatchTime: "30000-240000",
		UpdateWatchTime: "30000-240000",
		Serial:          boolPointer(false),
	}

	if servicePlan.Update != nil {
		updateBlock = bosh.Update{
			Canaries:        servicePlan.Update.Canaries,
			MaxInFlight:     servicePlan.Update.MaxInFlight,
			CanaryWatchTime: servicePlan.Update.CanaryWatchTime,
			UpdateWatchTime: servicePlan.Update.UpdateWatchTime,
			Serial:          servicePlan.Update.Serial,
		}
	}

	return bosh.BoshManifest{
		Name:     serviceDeployment.DeploymentName,
		Releases: releases,
		Stemcells: []bosh.Stemcell{{
			Alias:   OnlyStemcellAlias,
			OS:      serviceDeployment.Stemcell.OS,
			Version: serviceDeployment.Stemcell.Version,
		}},
		InstanceGroups: instanceGroups,
		Properties:     manifestProperties,
		Update:         updateBlock,
	}, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getPreviousManifestProperty(name string, manifest *bosh.BoshManifest) (interface{}, bool) {
	if manifest != nil {
		if val, ok := manifest.Properties["auto_create_topics"]; ok {
			return val, true
		}
	}
	return nil, false
}

func getJobFromInstanceGroup(name string, instanceGroup *bosh.InstanceGroup) (*bosh.Job, bool) {
	for index, job := range instanceGroup.Jobs {
		if job.Name == name {
			return &instanceGroup.Jobs[index], true
		}
	}
	return &bosh.Job{}, false
}

func instanceCounts(plan serviceadapter.Plan) map[string]int {
	val := map[string]int{}
	for _, instanceGroup := range plan.InstanceGroups {
		val[instanceGroup.Name] = instanceGroup.Instances
	}
	return val
}

func boolPointer(b bool) *bool {
	return &b
}

func checkInstanceGroupsPresent(names []string, instanceGroups []serviceadapter.InstanceGroup) error {
	var missingNames []string

	for _, name := range names {
		if !containsInstanceGroup(name, instanceGroups) {
			missingNames = append(missingNames, name)
		}
	}

	if len(missingNames) > 0 {
		return fmt.Errorf("Invalid instance group configuration: expected to find: '%s' in list: '%s'",
			strings.Join(missingNames, ", "),
			strings.Join(getInstanceGroupNames(instanceGroups), ", "))
	}
	return nil
}

func getInstanceGroupNames(instanceGroups []serviceadapter.InstanceGroup) []string {
	var instanceGroupNames []string
	for _, instanceGroup := range instanceGroups {
		instanceGroupNames = append(instanceGroupNames, instanceGroup.Name)
	}
	return instanceGroupNames
}

func containsInstanceGroup(name string, instanceGroups []serviceadapter.InstanceGroup) bool {
	for _, instanceGroup := range instanceGroups {
		if instanceGroup.Name == name {
			return true
		}
	}

	return false
}
