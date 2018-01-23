package activitytracker

import (
	"fmt"
	"time"

	atClient "github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/util/wait"

	"os"
	"testing"
)

const (
	defaultTestFlushDelay     = 10 * time.Millisecond
	defaultTestMaxConcurrency = 10
	defaultTestMaxBufferSize  = 10

	bufferSizeParamName = "buffersize"
	flushDelayParamName = "flushdelay"
	blockingParamName   = "blocking"
)

var envVars = []string{
	"ACCOUNT_ID",
	"CLUSTER_ID",
	"CLUSTER_NAME",
	"DATACENTER",
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

func TestActivitytracker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Activitytracker Suite")
}

type fakeAtWriter struct {
	writeFunc func([]*atClient.LogEntry, string, *atClient.MonitoredResource) int
}

func (w fakeAtWriter) Write(entries []*atClient.LogEntry, logName string, resource *atClient.MonitoredResource) int {
	if w.writeFunc != nil {
		return w.writeFunc(entries, logName, resource)
	}
	return 0
}

func (w fakeAtWriter) GetService() *atClient.Service {
	return nil
}

func waitWritesCount(q chan struct{}, want int) int {

	// Wait until the queue has the desired number of items
	wait.Poll(10*time.Millisecond, 10000*time.Millisecond, func() (bool, error) {
		return len(q) == want, nil
	})

	// Wait for some more time to ensure that the number is not greater.
	time.Sleep(100 * time.Millisecond)
	return len(q)
}
