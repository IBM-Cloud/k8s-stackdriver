package client_test

import (
	"net/http"
	"os"

	. "github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Models", func() {

	var client *http.Client
	BeforeEach(func() {
		client = &http.Client{}
	})

	It("generates a new service object", func() {
		service, err := NewService()
		Expect(err).NotTo(HaveOccurred())
		Expect(service).NotTo(BeNil())
		Expect(service.LogPath).To(Equal("/tmp/event-exporter/CLUSTER_ID-VAL-events.log"))
		Expect(service.LogPath).To(BeAnExistingFile())
		Expect(service.Logger).NotTo(BeNil())
	})

	It("returns the expected JSON payload of a log entry object", func() {

		By("returning the JSON payload when available")
		entry := &LogEntry{
			JSONPayload: []byte("jsonpayload"),
			TextPayload: "textpayload",
		}
		Expect(entry.GetPayload()).To(Equal("jsonpayload"))
	})

	It("returns the text payload of a log entry object if JSON payload is missing", func() {
		By("returning the JSON payload when available")
		entry := &LogEntry{
			TextPayload: "textpayload",
		}
		Expect(entry.GetPayload()).To(Equal("textpayload"))
	})

	Context("when an invalid logging directory is specified", func() {

		BeforeEach(func() {
			os.Setenv("ACTIVITY_TRACKER_LOG_ROOT", "/dev/bad")
		})

		It("fails to create a service object with an error", func() {
			service, err := NewService()
			Expect(err).To(HaveOccurred())
			Expect(service).To(BeNil())
		})
	})

	Context("when an invalid cluster name is specified", func() {

		BeforeEach(func() {
			os.Setenv("CLUSTER_ID", "/../../dev/bad")
		})

		AfterEach(func() {
			os.Unsetenv("CLUSTER_ID")
		})

		It("fails to create a service object with an error", func() {
			service, err := NewService()
			Expect(err).To(HaveOccurred())
			Expect(service).To(BeNil())
		})
	})
})
