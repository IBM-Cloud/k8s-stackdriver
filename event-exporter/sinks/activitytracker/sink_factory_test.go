package activitytracker

import (
	"os"

	"github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks"
	atClient "github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SinkFactory", func() {

	var factory sinks.SinkFactory
	BeforeEach(func() {
		factory = NewAtSinkFactory()
		Expect(factory).NotTo(BeNil())
	})

	It("creates a valid AT sink factory", func() {
		atSinkFactory := factory.(*atSinkFactory)
		Expect(atSinkFactory).NotTo(BeNil())
	})

	It("creates a new sink factory with default options", func() {
		By("creating a sink with no options and checking for default values")
		sink, err := factory.CreateNew([]string{})
		Expect(err).NotTo(HaveOccurred())
		Expect(*sink.(*atSink).GetConfig()).To(Equal(atSinkConfig{
			FlushDelay:     5000000000,
			MaxBufferSize:  100,
			MaxConcurrency: 10,
			LogName:        "accounts/ACCOUNT_ID-VAL/events",
			Resource: &atClient.MonitoredResource{
				Type: "ibm_cluster",
				Labels: map[string]string{
					"account_id": "ACCOUNT_ID-VAL",
					"cluster_id": "CLUSTER_ID-VAL",
					"location":   "DATACENTER-VAL",
				},
			},
		}))
	})

	It("creates a new sink factory with custom options", func() {
		By("creating a sink with custom options and checking for expected values")
		sink, err := factory.CreateNew([]string{"-max-buffer-size", "1"})
		Expect(err).NotTo(HaveOccurred())
		Expect(*sink.(*atSink).GetConfig()).To(Equal(atSinkConfig{
			FlushDelay:     5000000000,
			MaxBufferSize:  1,
			MaxConcurrency: 10,
			LogName:        "accounts/ACCOUNT_ID-VAL/events",
			Resource: &atClient.MonitoredResource{
				Type: "ibm_cluster",
				Labels: map[string]string{
					"account_id": "ACCOUNT_ID-VAL",
					"cluster_id": "CLUSTER_ID-VAL",
					"location":   "DATACENTER-VAL",
				},
			},
		}))
	})

	Context("when an invalid flag is specified", func() {
		It("fails with an error", func() {
			sink, err := factory.CreateNew([]string{"-invalid-parameter", "1"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse sink opts"))
			Expect(sink).To(BeNil())
		})
	})

	Context("when not running on an IBM container service cluster", func() {
		BeforeEach(func() {
			// Remove cluster ID which is required to be set on IBM Container Service
			os.Unsetenv("CLUSTER_ID")
		})

		It("fails with an error", func() {
			sink, err := factory.CreateNew([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to build config"))
			Expect(sink).To(BeNil())
		})
	})

	Context("when an invalid log root is specified", func() {
		BeforeEach(func() {
			// Specify an impossible place to write logs
			os.Setenv("CLUSTER_ID", "/../../dev/bad")
		})

		It("fails with an error", func() {
			sink, err := factory.CreateNew([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to initialize Activity Tracker service"))
			Expect(sink).To(BeNil())
		})
	})
})
