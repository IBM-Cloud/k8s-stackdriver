package client_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var envVars = []string{
	"ACCOUNT_ID",
	"CLUSTER_ID",
	"CLUSTER_LOCATION",
	"SERVICE_PROVIDER_ACCOUNT_ID",
	"SERVICE_PROVIDER_NAME",
	"SERVICE_PROVIDER_TOKEN",
}

var _ = BeforeEach(func() {
	for _, e := range envVars {
		os.Setenv(e, fmt.Sprintf("%s-VAL", e))
	}
	os.Setenv("ACTIVITY_TRACKER_LOG_ROOT", "/tmp")
})

var _ = AfterEach(func() {
	for _, e := range envVars {
		os.Unsetenv(e)
	}
	os.Unsetenv("ACTIVITY_TRACKER_LOG_ROOT")
})

var _ = AfterSuite(func() {
	err := os.Remove("/tmp/event-exporter/CLUSTER_ID-VAL-events.log")
	Expect(err).NotTo(HaveOccurred())
})

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Suite")
}
