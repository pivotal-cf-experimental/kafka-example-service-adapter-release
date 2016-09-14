package adapter_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"
)

var _ = Describe("Create", func() {
	var (
		boshVMs map[string][]string

		binding           serviceadapter.Binding
		requestParameters map[string]interface{}
		bindErr           error
	)

	JustBeforeEach(func() {
		binding, bindErr = binder.CreateBinding("binding-id", boshVMs, bosh.BoshManifest{}, requestParameters)
	})

	Context("when there are vms", func() {
		BeforeEach(func() {
			requestParameters = map[string]interface{}{}
			boshVMs = map[string][]string{"kafka_server": {"foo", "bar"}, "zookeeper_server": {"baz", "joe"}}
		})

		It("returns no error", func() {
			Expect(bindErr).NotTo(HaveOccurred())
		})

		It("returns bootstrap_servers in credentials", func() {
			Expect(binding).To(Equal(
				serviceadapter.Binding{
					Credentials: map[string]interface{}{"bootstrap_servers": []interface{}{"foo:9092", "bar:9092"}},
				},
			))
		})

		It("creates a topic with binding id as name", func() {
			Expect(fakeCommandRunner.RunCallCount()).To(Equal(1))
			actualName, acutalArgs := fakeCommandRunner.RunArgsForCall(0)
			Expect(actualName).To(Equal(expectedCommandName))
			Expect(acutalArgs).To(ConsistOf("baz,joe", "binding-id"))
		})

		Context("topic creation fails", func() {
			BeforeEach(func() {
				fakeCommandRunner.RunReturns([]byte{}, []byte{}, fmt.Errorf("a message for the operator"))
			})

			It("returns an error", func() {
				Expect(bindErr).To(MatchError(ContainSubstring("")))
			})

			It("logs an error for the operator", func() {
				Expect(stderr.String()).To(ContainSubstring("a message for the operator"))
			})
		})

		Context("topic creation fails with existing topic error", func() {
			BeforeEach(func() {
				fakeCommandRunner.RunReturns([]byte{}, []byte(`ERROR kafka.common.TopicExistsException: Topic "binding-id" already exists.
	at kafka.admin.AdminUtils$.createOrUpdateTopicPartitionAssignmentPathInZK(AdminUtils.scala:253)
	at kafka.admin.AdminUtils$.createTopic(AdminUtils.scala:237)
	at kafka.admin.TopicCommand$.createTopic(TopicCommand.scala:105)
	at kafka.admin.TopicCommand$.main(TopicCommand.scala:60)
	at kafka.admin.TopicCommand.main(TopicCommand.scala)`), fmt.Errorf(""))
			})

			It("returns an error to the cli user", func() {
				Expect(bindErr).To(Equal(serviceadapter.NewBindingAlreadyExistsError()))
			})

			It("logs an error for the operator", func() {
				Expect(stderr.String()).To(ContainSubstring("topic 'binding-id' already exists"))
			})
		})

		Context("topic is given in binding params", func() {
			BeforeEach(func() {
				requestParameters = map[string]interface{}{"parameters": map[string]interface{}{"topic": "foo"}}
			})

			It("calls the topic command, with topic name and zookeepers", func() {
				Expect(fakeCommandRunner.RunCallCount()).To(Equal(2))
				actualName, acutalArgs := fakeCommandRunner.RunArgsForCall(1)
				Expect(actualName).To(Equal(expectedCommandName))
				Expect(acutalArgs).To(ConsistOf("baz,joe", "foo"))
			})

			Context("when the command fails", func() {
				BeforeEach(func() {
					callCount := 0
					fakeCommandRunner.RunStub = func(name string, arg ...string) ([]byte, []byte, error) {
						if callCount == 0 {
							callCount++
							return []byte{}, []byte{}, nil
						} else {
							return []byte{}, []byte{}, fmt.Errorf("Everything failed again")
						}
					}
				})

				It("returns an error to the cli user", func() {
					Expect(bindErr.Error()).To(Equal(""))
				})
				It("logs an error for the operator", func() {
					Expect(stderr.String()).To(ContainSubstring("Everything failed again"))
				})
			})
		})
	})

	Context("when there are no vms for kafka_server", func() {
		BeforeEach(func() {
			requestParameters = map[string]interface{}{}
			boshVMs = map[string][]string{"baz": {"foo", "bar"}, "zookeeper_server": {"baz", "joe"}}
		})

		It("returns an error to the cli user", func() {
			Expect(bindErr).To(MatchError(""))
		})

		It("logs an error for the operator", func() {
			Expect(stderr.String()).To(ContainSubstring("no VMs for instance group kafka_server"))
		})
	})

	Context("when there are no vms for zookeeper_server", func() {
		BeforeEach(func() {
			requestParameters = map[string]interface{}{}
			boshVMs = map[string][]string{"kafka_server": {"foo", "bar"}, "baz": {"baz", "joe"}}
		})

		It("returns an error", func() {
			Expect(bindErr).To(MatchError(""))
		})

		It("logs an error for the operator", func() {
			Expect(stderr.String()).To(ContainSubstring("no VMs for job zookeeper_server"))
		})
	})
})
