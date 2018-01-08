package client_test

import (
	"os"
	"strconv"

	. "github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	It("retrieves the correct AT log root from environment", func() {
		Expect(GetAtLogRoot()).To(Equal("/tmp"))
	})

	It("retrieves the correct account ID from environment", func() {
		Expect(GetUserAccountID()).To(Equal("ACCOUNT_ID-VAL"))
	})

	It("retrieves the correct cluster ID from environment", func() {
		Expect(GetClusterID()).To(Equal("CLUSTER_ID-VAL"))
	})

	It("retrieves the correct location from environment", func() {
		Expect(GetClusterLocation()).To(Equal("CLUSTER_LOCATION-VAL"))
	})

	It("retrieves the correct service account ID from environment", func() {
		Expect(GetServiceAccountID()).To(Equal("SERVICE_PROVIDER_ACCOUNT_ID-VAL"))
	})

	It("retrieves the correct service name from environment", func() {
		Expect(GetServiceName()).To(Equal("SERVICE_PROVIDER_NAME-VAL"))
	})

	It("indicates IBM cloud when all environment variables are set", func() {
		Expect(OnIBM()).To(BeTrue())
	})

	It("generates a valid timestamp", func() {
		Expect(GetFormattedTimestamp()).To(HaveLen(10))
		_, err := strconv.Atoi(GetFormattedTimestamp())
		Expect(err).NotTo(HaveOccurred())
	})

	Context("If an environment variable is not set", func() {
		BeforeEach(func() {
			os.Unsetenv("SERVICE_PROVIDER_NAME")
		})
		It("indicates NOT on IBM cloud", func() {
			Expect(OnIBM()).To(BeFalse())
		})
	})
})
