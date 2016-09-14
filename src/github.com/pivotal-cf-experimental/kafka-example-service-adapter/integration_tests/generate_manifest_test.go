package integration_tests

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"
)

var _ = Describe("generate-manifest subcommand", func() {
	var boolPointer = func(b bool) *bool {
		return &b
	}

	const (
		deploymentName = "some-name"
	)

	var (
		plan              string
		previousPlan      string
		requestParams     string
		previousManifest  string
		serviceDeployment string
		stdout            bytes.Buffer
		stderr            bytes.Buffer
		exitCode          int
	)

	BeforeEach(func() {
		stdout = bytes.Buffer{}
		stderr = bytes.Buffer{}

		serviceDeployment = fmt.Sprintf(`{
			"deployment_name": "%s",
			"releases": [{
				"name": "kafka",
				"version": "0+dev.1",
				"jobs": ["kafka_server", "zookeeper_server", "smoke_tests"]
			}],
			"stemcell": {
				"stemcell_os": "Windows",
				"stemcell_version": "3.1"
			}
		}`, deploymentName)
		plan = `{
		  "instance_groups": [
		    {
		      "name": "kafka_server",
		      "vm_type": "small",
		      "persistent_disk_type": "ten",
		      "networks": [
		        "example-network"
		      ],
		      "azs": [
		        "example-az"
		      ],
		      "instances": 1
		    },
				{
		      "name": "zookeeper_server",
		      "vm_type": "medium",
		      "persistent_disk_type": "twenty",
		      "networks": [
		        "example-network"
		      ],
					"azs": [
						"example-az"
					],
		      "instances": 1
		    }
		  ],
		  "properties": {}
		}`
		requestParams = `{"parameters": {}}`
		previousManifest = ``
		previousPlan = "null"
	})

	JustBeforeEach(func() {
		cmd := exec.Command(serviceAdapterBinPath, "generate-manifest", serviceDeployment, plan, requestParams, previousManifest, previousPlan)
		runningBin, err := gexec.Start(cmd, io.MultiWriter(GinkgoWriter, &stdout), io.MultiWriter(GinkgoWriter, &stderr))
		Expect(err).NotTo(HaveOccurred())
		Eventually(runningBin).Should(gexec.Exit())
		exitCode = runningBin.ExitCode()
	})

	Context("when the parameters are valid", func() {
		It("exits with 0", func() {
			Expect(exitCode).To(Equal(0))
		})

		It("prints a manifest to stdout", func() {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			expectedManifest, err := ioutil.ReadFile(filepath.Join(cwd, "fixtures", "expected_manifest.yml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(MatchYAML(expectedManifest))
		})

		Context("plan migrations", func() {
			Context("when migrating to a plan with fewer kafka instances", func() {
				BeforeEach(func() {
					plan = `{
					"instance_groups": [
						{
							"name": "kafka_server",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						},
						{
							"name": "zookeeper_server",
							"vm_type": "medium",
							"persistent_disk_type": "twenty",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						}
					],
					"properties": {}
				}`
					previousPlan = `{
					"instance_groups": [
						{
							"name": "kafka_server",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 4
						},
						{
							"name": "zookeeper_server",
							"vm_type": "medium",
							"persistent_disk_type": "twenty",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						}
					],
					"properties": {}
				}`
				})

				It("exits with 1", func() {
					Expect(exitCode).To(Equal(1))
				})

				It("logs a message for the operator", func() {
					Expect(stderr.String()).To(ContainSubstring("cannot migrate to a smaller plan"))
				})
			})

			Context("when migrating to a plan with less zookeeper instances", func() {
				BeforeEach(func() {
					plan = `{
					"instance_groups": [
						{
							"name": "kafka_server",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						},
						{
							"name": "zookeeper_server",
							"vm_type": "medium",
							"persistent_disk_type": "twenty",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						}
					],
					"properties": {}
				}`
					previousPlan = `{
					"instance_groups": [
						{
							"name": "kafka_server",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						},
						{
							"name": "zookeeper_server",
							"vm_type": "medium",
							"persistent_disk_type": "twenty",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 2
						}
					],
					"properties": {}
				}`
				})

				It("exits with 1", func() {
					Expect(exitCode).To(Equal(1))
				})

				It("logs a message", func() {
					Expect(stderr.String()).To(ContainSubstring("cannot migrate to a smaller plan"))
				})
			})

			Context("when migrating to a plan with more zookeeper instances", func() {
				BeforeEach(func() {
					plan = `{
					"instance_groups": [
						{
							"name": "kafka_server",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 2
						},
						{
							"name": "zookeeper_server",
							"vm_type": "medium",
							"persistent_disk_type": "twenty",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 2
						}
					],
					"properties": {}
				}`
					previousPlan = `{
					"instance_groups": [
						{
							"name": "kafka_server",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						},
						{
							"name": "zookeeper_server",
							"vm_type": "medium",
							"persistent_disk_type": "twenty",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						}
					],
					"properties": {}
				}`
				})

				It("exits with 0", func() {
					Expect(exitCode).To(Equal(0))
				})
			})

		})

		Context("when auto_create_topics is set as a plan property", func() {
			BeforeEach(func() {
				plan = `{
				  "instance_groups": [
				    {
				      "name": "kafka_server",
				      "vm_type": "small",
				      "persistent_disk_type": "ten",
				      "networks": [
				        "example-network"
				      ],
				      "azs": [
				        "example-az"
				      ],
				      "instances": 1
				    },
						{
				      "name": "zookeeper_server",
				      "vm_type": "medium",
				      "persistent_disk_type": "twenty",
				      "networks": [
				        "example-network"
				      ],
							"azs": [
								"example-az"
							],
				      "instances": 1
				    }
				  ],
				  "properties": {
						"auto_create_topics": false
					}
				}`
			})

			It("prints a manifest to stdout", func() {
				var manifest bosh.BoshManifest
				Expect(yaml.Unmarshal(stdout.Bytes(), &manifest)).To(Succeed())
				Expect(manifest.InstanceGroups[0].Jobs[0].Properties).To(HaveKeyWithValue("auto_create_topics", false))
			})
		})

		Context("when auto_create_topics is set as an arbitrary param", func() {
			BeforeEach(func() {
				requestParams = `{
				"organization_guid": "an-org-guid",
				"plan_id": "a-plan",
				"service_id": "a-service-id",
				"space_guid": "a-space-guid",
				"parameters": {"auto_create_topics": false}
				}`
			})

			It("sets the arbitrary params for auto_create_topics", func() {
				var manifest bosh.BoshManifest
				Expect(yaml.Unmarshal(stdout.Bytes(), &manifest)).To(Succeed())
				Expect(manifest.InstanceGroups[0].Jobs[0].Properties).To(HaveKeyWithValue("auto_create_topics", false))
			})
		})

		Context("when auto_create_topics is set as a previous manifest property", func() {
			BeforeEach(func() {
				previousManifest = `
---
properties:
  auto_create_topics: false`
			})

			It("sets the previous manifest property for auto_create_topics", func() {
				var manifest bosh.BoshManifest
				Expect(yaml.Unmarshal(stdout.Bytes(), &manifest)).To(Succeed())
				Expect(manifest.InstanceGroups[0].Jobs[0].Properties).To(HaveKeyWithValue("auto_create_topics", false))
			})
		})

		Context("when auto_create_topics is set as a previous manifest property AND plan property", func() {
			BeforeEach(func() {
				previousManifest = `
---
properties:
  auto_create_topics: false`
				plan = `{
					"instance_groups": [
						{
							"name": "kafka_server",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						},
						{
							"name": "zookeeper_server",
							"vm_type": "medium",
							"persistent_disk_type": "twenty",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						}
					],
					"properties": {
						"auto_create_topics": true
					}
				}`
			})

			It("sets the previous manifest property for auto_create_topics", func() {
				var manifest bosh.BoshManifest
				Expect(yaml.Unmarshal(stdout.Bytes(), &manifest)).To(Succeed())
				Expect(manifest.InstanceGroups[0].Jobs[0].Properties).To(HaveKeyWithValue("auto_create_topics", false))
			})
		})

		Context("when auto_create_topics is set as an arbitrary param AND plan property", func() {
			BeforeEach(func() {
				plan = `{
					"instance_groups": [
						{
							"name": "kafka_server",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						},
						{
							"name": "zookeeper_server",
							"vm_type": "medium",
							"persistent_disk_type": "twenty",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						}
					],
					"properties": {
						"auto_create_topics": true
					}
				}`
				requestParams = `{"parameters": {"auto_create_topics": false}}`
			})

			It("sets the arbitrary param for auto_create_topics", func() {
				var manifest bosh.BoshManifest
				Expect(yaml.Unmarshal(stdout.Bytes(), &manifest)).To(Succeed())
				Expect(manifest.InstanceGroups[0].Jobs[0].Properties).To(HaveKeyWithValue("auto_create_topics", false))
			})
		})

		Context("when auto_create_topics is set as an arbitrary param AND previous manifest property AND plan property", func() {
			BeforeEach(func() {
				requestParams = `{"parameters": {"auto_create_topics": false}}`
				previousManifest = `
---
properties:
  auto_create_topics: false`
				plan = `{
					"instance_groups": [
						{
							"name": "kafka_server",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						},
						{
							"name": "zookeeper_server",
							"vm_type": "medium",
							"persistent_disk_type": "twenty",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						}
					],
					"properties": {
						"auto_create_topics": true
					}
				}`
			})

			It("sets the arbitrary param for auto_create_topics", func() {
				var manifest bosh.BoshManifest
				Expect(yaml.Unmarshal(stdout.Bytes(), &manifest)).To(Succeed())
				Expect(manifest.InstanceGroups[0].Jobs[0].Properties).To(HaveKeyWithValue("auto_create_topics", false))
			})
		})

		Context("when the default_replication_factor plan property is set", func() {
			BeforeEach(func() {
				plan = `{
					"instance_groups": [
						{
							"name": "kafka_server",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						},
						{
							"name": "zookeeper_server",
							"vm_type": "medium",
							"persistent_disk_type": "twenty",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						}
					],
					"properties": {
						"default_replication_factor": 43
					}
				}`
			})
			It("prints a manifest to stdout", func() {
				var manifest bosh.BoshManifest
				Expect(yaml.Unmarshal(stdout.Bytes(), &manifest)).To(Succeed())
				Expect(manifest.InstanceGroups[0].Jobs[0].Properties).To(HaveKeyWithValue("default_replication_factor", 43))
			})
		})

		Context("when an update block is provided for the plan", func() {
			Context("with all fields provided", func() {
				BeforeEach(func() {
					plan = `{
						"instance_groups": [
							{
								"name": "kafka_server",
								"vm_type": "small",
								"persistent_disk_type": "ten",
								"networks": [
									"example-network"
								],
								"azs": [
									"example-az"
								],
								"instances": 1
							},
							{
								"name": "zookeeper_server",
								"vm_type": "medium",
								"persistent_disk_type": "twenty",
								"networks": [
									"example-network"
								],
								"azs": [
									"example-az"
								],
								"instances": 1
							}
						],
						"properties": {
							"default_replication_factor": 43
						},
						"update": {
							"canaries": 2,
							"max_in_flight": 12,
							"canary_watch_time": "1000-3000",
							"update_watch_time": "1000-3000",
							"serial": true
						}
					}`
				})

				It("prints a manifest to stdout", func() {
					var manifest bosh.BoshManifest
					Expect(yaml.Unmarshal(stdout.Bytes(), &manifest)).To(Succeed())
					Expect(manifest.Update.Canaries).To(Equal(2))
					Expect(manifest.Update.MaxInFlight).To(Equal(12))
					Expect(manifest.Update.CanaryWatchTime).To(Equal("1000-3000"))
					Expect(manifest.Update.UpdateWatchTime).To(Equal("1000-3000"))
					Expect(manifest.Update.Serial).To(Equal(boolPointer(true)))
				})
			})

			Context("with only mandatory fields provided", func() {
				BeforeEach(func() {
					plan = `{
						"instance_groups": [
							{
								"name": "kafka_server",
								"vm_type": "small",
								"persistent_disk_type": "ten",
								"networks": [
									"example-network"
								],
								"azs": [
									"example-az"
								],
								"instances": 1
							},
							{
								"name": "zookeeper_server",
								"vm_type": "medium",
								"persistent_disk_type": "twenty",
								"networks": [
									"example-network"
								],
								"azs": [
									"example-az"
								],
								"instances": 1
							}
						],
						"properties": {
							"default_replication_factor": 43
						},
						"update": {
							"canaries": 2,
							"max_in_flight": 12,
							"canary_watch_time": "1000-3000",
							"update_watch_time": "1000-3000"
						}
					}`
				})

				It("prints a manifest to stdout", func() {
					var manifest bosh.BoshManifest
					Expect(yaml.Unmarshal(stdout.Bytes(), &manifest)).To(Succeed())
					Expect(manifest.Update.Canaries).To(Equal(2))
					Expect(manifest.Update.MaxInFlight).To(Equal(12))
					Expect(manifest.Update.CanaryWatchTime).To(Equal("1000-3000"))
					Expect(manifest.Update.UpdateWatchTime).To(Equal("1000-3000"))
					Expect(manifest.Update.Serial).To(BeNil())
				})
			})
		})

		Context("when a Metron release is included", func() {
			BeforeEach(func() {
				serviceDeployment = fmt.Sprintf(`{
					"deployment_name": "%s",
					"releases": [{
						"name": "kafka",
						"version": "0+dev.1",
						"jobs": ["kafka_server", "zookeeper_server"]
					},{
						"name": "some-metron-release",
						"version": "2000",
						"jobs": ["metron_agent"]
					}],
					"stemcell": {
						"stemcell_os": "Windows",
						"stemcell_version": "3.1"
					}
				}`, deploymentName)
			})

			Context("and Metron plan properties are included", func() {
				BeforeEach(func() {
					plan = `{
						"instance_groups": [
							{
								"name": "kafka_server",
								"vm_type": "small",
								"persistent_disk_type": "ten",
								"networks": [
									"example-network"
								],
								"azs": [
									"example-az"
								],
								"instances": 1
							},
							{
								"name": "zookeeper_server",
								"vm_type": "medium",
								"persistent_disk_type": "twenty",
								"networks": [
									"example-network"
								],
								"azs": [
									"example-az"
								],
								"instances": 1
							}
						],
						"properties": {
							"logging": {
								"syslog_address": "syslog_ip_adddress",
								"syslog_port": 1234,
								"loggregator_etcd_addresses": ["195.99.147.120"],
								"loggregator_shared_secret": "DonCossack"
							}
						}
					}`
				})

				It("puts the release in the generated manifest", func() {
					var manifest bosh.BoshManifest
					Expect(yaml.Unmarshal(stdout.Bytes(), &manifest)).To(Succeed())
					Expect(manifest.Releases).To(ConsistOf(
						bosh.Release{Name: "kafka", Version: "0+dev.1"},
						bosh.Release{Name: "some-metron-release", Version: "2000"},
					))
				})

				It("should add a metron_agent job to each instance group", func() {
					var manifest bosh.BoshManifest
					Expect(yaml.Unmarshal(stdout.Bytes(), &manifest)).To(Succeed())

					Expect(manifest.InstanceGroups[0].Jobs).To(ConsistOf(
						bosh.Job{Name: "kafka_server", Release: "kafka", Properties: map[string]interface{}{
							"default_replication_factor": 3, "auto_create_topics": true, "network": "example-network",
						}},
						bosh.Job{Name: "metron_agent", Release: "some-metron-release"},
					))
					Expect(manifest.InstanceGroups[1].Jobs).To(ConsistOf(
						bosh.Job{Name: "zookeeper_server", Release: "kafka"},
						bosh.Job{Name: "metron_agent", Release: "some-metron-release"},
					))
				})

				It("should add the Metron properties to the manifest", func() {
					var manifest bosh.BoshManifest
					Expect(yaml.Unmarshal(stdout.Bytes(), &manifest)).To(Succeed())
					Expect(manifest.Properties).To(HaveKeyWithValue("syslog_daemon_config", map[interface{}]interface{}{
						"address": "syslog_ip_adddress",
						"port":    1234,
					}))
					Expect(manifest.Properties).To(HaveKeyWithValue("metron_agent", map[interface{}]interface{}{
						"zone":       "",
						"deployment": deploymentName,
					}))
					Expect(manifest.Properties).To(HaveKeyWithValue("loggregator", map[interface{}]interface{}{
						"etcd": map[interface{}]interface{}{
							"machines": []interface{}{"195.99.147.120"},
						},
					}))
					Expect(manifest.Properties).To(HaveKeyWithValue("metron_endpoint", map[interface{}]interface{}{
						"shared_secret": "DonCossack",
					}))
				})
			})

			Context("and no Metron plan properties are included", func() {
				It("should not add a metron_agent job to each instance group", func() {
					var manifest bosh.BoshManifest
					Expect(yaml.Unmarshal(stdout.Bytes(), &manifest)).To(Succeed())

					Expect(manifest.InstanceGroups[0].Jobs).To(ConsistOf(
						bosh.Job{Name: "kafka_server", Release: "kafka", Properties: map[string]interface{}{
							"default_replication_factor": 3,
							"auto_create_topics":         true,
							"network":                    "example-network",
						}},
					))
					Expect(manifest.InstanceGroups[1].Jobs).To(ConsistOf(
						bosh.Job{Name: "zookeeper_server", Release: "kafka"},
					))
				})
			})
		})

		Context("when a logging properties block is provided but no Metron release is included", func() {
			BeforeEach(func() {
				plan = `{
						"instance_groups": [
							{
								"name": "kafka_server",
								"vm_type": "small",
								"persistent_disk_type": "ten",
								"networks": [
									"example-network"
								],
								"azs": [
									"example-az"
								],
								"instances": 1
							},
							{
								"name": "zookeeper_server",
								"vm_type": "medium",
								"persistent_disk_type": "twenty",
								"networks": [
									"example-network"
								],
								"azs": [
									"example-az"
								],
								"instances": 1
							}
						],
						"properties": {
							"logging": {
								"syslog_address": "syslog_ip_adddress",
								"syslog_port": 1234,
								"loggregator_etcd_addresses": ["195.99.147.120"],
								"loggregator_shared_secret": "DonCossack"
							}
						}
					}`
			})

			It("exits with non-zero status", func() {
				Expect(exitCode).ToNot(Equal(0))
			})
		})

		Context("when the smoke-test errand is included", func() {
			BeforeEach(func() {
				plan = `{
					"instance_groups": [
						{
							"name": "kafka_server",
							"vm_type": "small",
							"persistent_disk_type": "ten",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						},
						{
							"name": "zookeeper_server",
							"vm_type": "medium",
							"persistent_disk_type": "twenty",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1
						},
						{
							"name": "smoke_tests",
							"vm_type": "medium",
							"networks": [
								"example-network"
							],
							"azs": [
								"example-az"
							],
							"instances": 1,
							"lifecycle": "errand"
						}
					],
					"properties": {}
				}`
			})

			It("prints a manifest to stdout", func() {
				cwd, err := os.Getwd()
				Expect(err).NotTo(HaveOccurred())
				expectedManifest, err := ioutil.ReadFile(filepath.Join(cwd, "fixtures", "expected_manifest_with_errand.yml"))
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout.String()).To(MatchYAML(expectedManifest))
			})
		})
	})

	Context("when the service deployment parameter is invalid JSON", func() {
		BeforeEach(func() {
			serviceDeployment = "sadfsdl;fksajflasdf"
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})
	})

	Context("when the service deployment parameter contains too few elements", func() {
		BeforeEach(func() {
			serviceDeployment = `{}`
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})
	})

	Context("when there are no releases", func() {
		BeforeEach(func() {
			serviceDeployment = fmt.Sprintf(`{
				"deployment_name": "%s",
				"releases": [],
				"stemcell": {
					"stemcell_os": "Windows",
					"stemcell_version": "3.1"
				}
			}`, deploymentName)
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})

		It("prints an error message for the operator", func() {
			Expect(stderr.String()).To(ContainSubstring("job 'kafka_server' not provided"))
		})

		It("doesn't print an error message for the user", func() {
			Expect(stdout.String()).To(Equal(""))
		})
	})

	Context("when the plan parameter is invalid json", func() {
		BeforeEach(func() {
			plan = "afasfsafasfadsf"
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})
	})

	Context("when the plan contains no instance groups", func() {
		BeforeEach(func() {
			plan = `{
			  "instance_groups": [],
			  "properties": {}
			}`
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})

		It("outputs an error message for the CLI user to stdout", func() {
			Expect(stdout.String()).To(ContainSubstring("Contact your operator, service configuration issue occurred"))
		})

		It("outputs an error message to the operator to stderr", func() {
			Expect(stderr.String()).To(ContainSubstring("Invalid instance group configuration: expected to find: 'kafka_server, zookeeper_server' in list: ''"))
		})
	})

	Context("when the plan does not contain a kafka_server instance group", func() {
		BeforeEach(func() {
			plan = `{
			"instance_groups":
				[{
					"name": "zookeeper_server",
					"vm_type": "medium",
					"networks": [
						"example-network"
					],
					"azs": [
						"example-az"
					],
					"instances": 1
				}]
			}`
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})

		It("outputs an error message for the CLI user to stdout", func() {
			Expect(stdout.String()).To(ContainSubstring("Contact your operator, service configuration issue occurred"))
		})

		It("outputs an error message to the operator to stderr", func() {
			Expect(stderr.String()).To(ContainSubstring("Invalid instance group configuration: expected to find: 'kafka_server' in list: 'zookeeper_server'"))
		})
	})

	Context("when the plan does not contain a zookeeper_server instance group", func() {
		BeforeEach(func() {
			plan = `{
			"instance_groups":
				[{
					"name": "kafka_server",
					"vm_type": "small",
					"networks": [
						"example-network"
					],
					"azs": [
						"example-az"
					],
					"instances": 1
				}]
			}`
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})

		It("outputs an error message for the CLI user to stdout", func() {
			Expect(stdout.String()).To(ContainSubstring("Contact your operator, service configuration issue occurred"))
		})

		It("outputs an error message to the operator to stderr", func() {
			Expect(stderr.String()).To(ContainSubstring("Invalid instance group configuration: expected to find: 'zookeeper_server' in list: 'kafka_server'"))
		})
	})

	Context("release does not contain provided job", func() {
		BeforeEach(func() {
			serviceDeployment = fmt.Sprintf(`{
						"deployment_name": "%s",
						"releases": [{
							"name": "kafka",
							"version": "0+dev.1",
							"jobs": ["agshdj"]
						}],
						"stemcell": {
							"stemcell_os": "Windows",
							"stemcell_version": "3.1"
						}
					}`, deploymentName)
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})

		It("prints an error message for the operator", func() {
			Expect(stderr.String()).To(ContainSubstring("job 'kafka_server' not provided"))
		})

		It("doesn't print an error message for the user", func() {
			Expect(stdout.String()).To(Equal(""))
		})
	})

	Context("when there is no network defined for tha kafka_server job", func() {
		BeforeEach(func() {
			plan = `{
			  "instance_groups": [
			    {
			      "name": "kafka_server",
			      "vm_type": "small",
			      "persistent_disk_type": "ten",
			      "networks": [],
			      "azs": [
			        "example-az"
			      ],
			      "instances": 1
			    },
					{
			      "name": "zookeeper_server",
			      "vm_type": "medium",
			      "persistent_disk_type": "twenty",
			      "networks": [
			        "example-network"
			      ],
						"azs": [
							"example-az"
						],
			      "instances": 1
			    }
			  ],
			  "properties": {}
			}`
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})

		It("prints an error message for the operator", func() {
			Expect(stderr.String()).To(ContainSubstring("expected 1 network for kafka_server, got 0"))
		})

		It("doesn't print an error message for the user", func() {
			Expect(stdout.String()).To(Equal(""))
		})

	})

	Context("when there is no Metron, but Metron plan properties are included", func() {
		BeforeEach(func() {
			plan = `{
				"instance_groups": [
					{
						"name": "kafka_server",
						"vm_type": "small",
						"persistent_disk_type": "ten",
						"networks": [
							"example-network"
						],
						"azs": [
							"example-az"
						],
						"instances": 1
					},
					{
						"name": "zookeeper_server",
						"vm_type": "medium",
						"persistent_disk_type": "twenty",
						"networks": [
							"example-network"
						],
						"azs": [
							"example-az"
						],
						"instances": 1
					}
				],
				"properties": {
					"logging": {
						"syslog_address": "syslog_ip_adddress",
						"syslog_port": 1234
					}
				}
			}`
			previousPlan = `{
				"instance_groups": [
					{
						"name": "kafka_server",
						"vm_type": "small",
						"persistent_disk_type": "ten",
						"networks": [
							"example-network"
						],
						"azs": [
							"example-az"
						],
						"instances": 1
					},
					{
						"name": "zookeeper_server",
						"vm_type": "medium",
						"persistent_disk_type": "twenty",
						"networks": [
							"example-network"
						],
						"azs": [
							"example-az"
						],
						"instances": 1
					}
				],
				"properties": {
					"logging": {
						"syslog_address": "syslog_ip_adddress",
						"syslog_port": 1234
					}
				}
			}`
		})

		It("exits with non-zero status", func() {
			Expect(exitCode).ToNot(Equal(0))
		})
	})
})
