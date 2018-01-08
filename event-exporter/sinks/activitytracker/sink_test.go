package activitytracker

import (
	atClient "github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker/client"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/clock"
	api_v1 "k8s.io/client-go/pkg/api/v1"
)

var _ = Describe("Sink", func() {

	var sink *atSink
	var done chan struct{}
	var q chan struct{}
	BeforeEach(func() {

		// Create the sink object (to be tested below)
		sinkConfig := &atSinkConfig{
			Resource:       nil,
			FlushDelay:     defaultTestFlushDelay,
			LogName:        "logname",
			MaxConcurrency: defaultTestMaxConcurrency,
			MaxBufferSize:  defaultTestMaxBufferSize,
		}
		done = make(chan struct{})
		q = make(chan struct{}, 2*defaultMaxBufferSize)
		sink = newAtSink(fakeAtWriter{
			writeFunc: func(entries []*atClient.LogEntry, param string, r *atClient.MonitoredResource) int {
				for i := 0; i < len(entries); i++ {
					// increment number of written entries
					q <- struct{}{}
				}
				if param == blockingParamName {
					<-done
				}
				return len(entries)
			},
		}, clock.NewFakeClock(time.Time{}), sinkConfig)

		// start the sink
		go sink.Run(done)
	})

	AfterEach(func() {

		// shutdown the sink
		close(done)
	})

	It("generates a new activity tracker sink object with empty timer", func() {
		Expect(sink).NotTo(BeNil())
		Expect(sink.timer).To(BeNil())
	})

	It("creates a new timer when it is set for the first time", func() {

		By("getting a fake channel")
		fakeChannel := sink.getTimerChannel()
		Expect(fakeChannel).NotTo(BeNil())

		By("setting the timer")
		sink.setTimer()
		Expect(sink.timer).NotTo(BeNil())

		timerChannel := sink.getTimerChannel()
		Expect(timerChannel).NotTo(BeNil())
		Expect(timerChannel).NotTo(Equal(fakeChannel))
	})

	It("creates a single initial message on first list attempt", func() {
		for i := 0; i < 3; i++ {
			sink.OnList(&api_v1.EventList{})
			got := waitWritesCount(q, 1)
			Expect(got).To(Equal(1)) // only the single initial write should ever occur
		}
	})

	It("processes a single event when added", func() {

		By("adding an event to the sink")
		sink.OnAdd(&api_v1.Event{})

		By("waiting for the single event to be processed")
		got := waitWritesCount(q, 1)
		Expect(got).To(Equal(1))
	})

	It("processes multiple events when added less than concurrency limit", func() {
		numEntriesToWrite := defaultMaxConcurrency - 1
		for i := 0; i < numEntriesToWrite; i++ {
			sink.OnAdd(&api_v1.Event{})
		}
		got := waitWritesCount(q, numEntriesToWrite)
		Expect(got).To(Equal(numEntriesToWrite))
	})

	It("processes multiple events when added greater than concurrency limit", func() {
		numEntriesToWrite := defaultMaxConcurrency + 1
		for i := 0; i < numEntriesToWrite; i++ {
			sink.OnAdd(&api_v1.Event{})
		}
		got := waitWritesCount(q, numEntriesToWrite)
		Expect(got).To(Equal(numEntriesToWrite))
	})

	It("flushes the buffer without blocking when it is exceeded", func() {
		numEntriesToWrite := defaultMaxBufferSize + 1
		for i := 0; i < numEntriesToWrite; i++ {
			sink.OnAdd(&api_v1.Event{})
		}
		got := waitWritesCount(q, numEntriesToWrite)
		Expect(got).To(Equal(numEntriesToWrite))
	})

	It("processes event updates", func() {
		sink.OnUpdate(&api_v1.Event{}, &api_v1.Event{})
		got := waitWritesCount(q, 1)
		Expect(got).To(Equal(1))
	})
})
