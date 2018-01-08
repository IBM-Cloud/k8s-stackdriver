package activitytracker

import (
	atClient "github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Writer", func() {

	var service *atClient.Service
	var err error
	BeforeEach(func() {
		service, err = atClient.NewService()
		Expect(err).NotTo(HaveOccurred())
		Expect(service).NotTo(BeNil())
	})

	It("generates a new activity tracker writer object", func() {
		atWriter := newAtWriter(service)
		Expect(atWriter.GetService()).To(Equal(service))
	})

	It("returns 0 if no entries are written", func() {
		atWriter := newAtWriter(service)
		numWritten := atWriter.Write([]*atClient.LogEntry{}, "noData", nil)
		Expect(numWritten).To(BeZero())
	})

	It("returns 1 if 1 log entry is written", func() {
		atWriter := newAtWriter(service)
		numWritten := atWriter.Write([]*atClient.LogEntry{
			&atClient.LogEntry{ResourceID: "resourceID"},
		}, "singleLog", &atClient.MonitoredResource{})
		Expect(numWritten).To(Equal(1))
	})
})
