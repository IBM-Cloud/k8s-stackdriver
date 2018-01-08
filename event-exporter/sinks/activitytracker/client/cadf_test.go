package client_test

import (
	"io/ioutil"
	"net/http"

	. "github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker/client"
	"github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker/client/pb"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cadf", func() {

	var kubeEventEntry *LogEntry
	var service *Service
	var err error

	BeforeEach(func() {
		kubeEventEntry = &LogEntry{
			JSONPayload:     []byte("json_payload"),
			Outcome:         "outcome",
			Reason:          "reason",
			ResourceID:      "resource_id",
			ResourceType:    "resource_type",
			SourceComponent: "source_component",
			Timestamp:       "timestamp",
		}

		service, err = NewService()
		Expect(err).NotTo(HaveOccurred())
	})

	It("generates a propert resource type URI", func() {
		// TODO - this test needs to be updated once we decide on type URI formatting
		Expect(GetCadfTypeURI("dummy")).To(Equal("dummy"))
	})

	It("serializes CADF to JSON format", func() {
		cadfEvent := service.FromEntry(kubeEventEntry)
		jsonString, err := cadfEvent.SerializeJSON()
		Expect(err).NotTo(HaveOccurred())
		Expect(jsonString).To(Equal("{\"meta\":{\"service_provider_name\":\"SERVICE_PROVIDER_NAME-VAL\",\"service_provider_region\":\"CLUSTER_LOCATION-VAL\",\"service_provider_project_id\":\"SERVICE_PROVIDER_ACCOUNT_ID-VAL\",\"user_account_ids\":[\"ACCOUNT_ID-VAL\"]},\"payload\":{\"typeURI\":\"http://schemas.dmtf.org/cloud/audit/1.0/event\",\"eventTime\":\"timestamp\",\"action\":\"resource_type.reason\",\"outcome\":\"outcome\",\"initiator\":{\"id\":\"CLUSTER_ID-VAL\",\"typeURI\":\"source_component\",\"name\":\"kubernetes_cluster.source_component\",\"project_id\":\"ACCOUNT_ID-VAL\"},\"target\":{\"id\":\"resource_id\",\"typeURI\":\"resource_type\",\"name\":\"kubernetes_cluster.resource_type\",\"project_id\":\"ACCOUNT_ID-VAL\"},\"requestData\":\"json_payload\"}}"))
	})

	It("translates a kubernetes event to CADF event", func() {

		cadfEvent := service.FromEntry(kubeEventEntry)
		Expect(cadfEvent).NotTo(BeNil())
		Expect(*cadfEvent).To(Equal(pb.CadfEventWK{
			Meta: &pb.Meta{
				ServiceProviderName:      "SERVICE_PROVIDER_NAME-VAL",
				ServiceProviderRegion:    "CLUSTER_LOCATION-VAL",
				ServiceProviderProjectId: "SERVICE_PROVIDER_ACCOUNT_ID-VAL",
				UserAccountIds:           []string{"ACCOUNT_ID-VAL"},
			},
			Payload: &pb.CadfEvent{
				TypeURI:   "http://schemas.dmtf.org/cloud/audit/1.0/event",
				EventTime: "timestamp",
				Action:    "resource_type.reason",
				Outcome:   "outcome",
				Initiator: &pb.CadfResource{
					Id:        "CLUSTER_ID-VAL",
					TypeURI:   "source_component",
					Name:      "kubernetes_cluster.source_component",
					ProjectId: "ACCOUNT_ID-VAL",
				},
				Target: &pb.CadfResource{
					Id:        "resource_id",
					TypeURI:   "resource_type",
					Name:      "kubernetes_cluster.resource_type",
					ProjectId: "ACCOUNT_ID-VAL",
				},
				RequestData: "json_payload",
			},
		}))
	})

	It("writes CADF eventset to expected file", func() {
		req := &WriteLogEntriesRequest{
			Entries: []*LogEntry{kubeEventEntry},
		}

		By("validating the response object")
		resp := service.Write(req)
		Expect(resp).NotTo(BeNil())
		Expect(*resp).To(Equal(WriteLogEntriesResponse{
			HTTPStatusCode: http.StatusOK,
			Message:        "success",
		}))

		By("closing the logger to force a filesystem sync")
		service.Logger.Close()

		By("validating the file contents")
		contents, err := ioutil.ReadFile("/tmp/event-exporter/CLUSTER_ID-VAL-events.log")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("{\"meta\":{\"service_provider_name\":\"SERVICE_PROVIDER_NAME-VAL\",\"service_provider_region\":\"CLUSTER_LOCATION-VAL\",\"service_provider_project_id\":\"SERVICE_PROVIDER_ACCOUNT_ID-VAL\",\"user_account_ids\":[\"ACCOUNT_ID-VAL\"]},\"payload\":{\"typeURI\":\"http://schemas.dmtf.org/cloud/audit/1.0/event\",\"eventTime\":\"timestamp\",\"action\":\"resource_type.reason\",\"outcome\":\"outcome\",\"initiator\":{\"id\":\"CLUSTER_ID-VAL\",\"typeURI\":\"source_component\",\"name\":\"kubernetes_cluster.source_component\",\"project_id\":\"ACCOUNT_ID-VAL\"},\"target\":{\"id\":\"resource_id\",\"typeURI\":\"resource_type\",\"name\":\"kubernetes_cluster.resource_type\",\"project_id\":\"ACCOUNT_ID-VAL\"},\"requestData\":\"json_payload\"}}\n"))
	})
})
