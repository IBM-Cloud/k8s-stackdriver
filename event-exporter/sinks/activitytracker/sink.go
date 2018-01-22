/*
Copyright 2017 IBM Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package activitytracker

import (
	"time"

	atClient "github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker/client"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/apimachinery/pkg/util/clock"
	api_v1 "k8s.io/client-go/pkg/api/v1"
)

var (
	receivedEntryCountAT = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "received_entry_count",
			Help:      "Number of entries, recieved by the Activity Tracker sink",
			Subsystem: "activitytracker_sink",
		},
		[]string{"component"},
	)

	successfullySentEntryCountAT = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name:      "successfully_sent_entry_count",
			Help:      "Number of entries, successfully ingested by Activity Tracker",
			Subsystem: "activitytracker_sink",
		},
	)
)

type atSink struct {
	logEntryChannel chan *atClient.LogEntry
	config          *atSinkConfig
	logEntryFactory *atLogEntryFactory
	writer          atWriter
	logName         string

	currentBuffer   []*atClient.LogEntry
	timer           *time.Timer
	fakeTimeChannel chan time.Time

	// Channel for controlling how many requests are being sent at the same
	// time. It's empty initially, each request adds an object at the start
	// and takes it out upon completion. Channel's capacity is set to the
	// maximum level of parallelism, so any extra request will lock on addition.
	concurrencyChannel chan struct{}

	beforeFirstList bool
}

func newAtSink(writer atWriter, clock clock.Clock, config *atSinkConfig) *atSink {
	return &atSink{
		logEntryChannel:    make(chan *atClient.LogEntry, config.MaxBufferSize),
		config:             config,
		logEntryFactory:    newAtLogEntryFactory(clock),
		writer:             writer,
		logName:            config.LogName,
		currentBuffer:      []*atClient.LogEntry{},
		timer:              nil,
		fakeTimeChannel:    make(chan time.Time),
		concurrencyChannel: make(chan struct{}, config.MaxConcurrency),
		beforeFirstList:    true,
	}
}

func (s *atSink) GetConfig() *atSinkConfig {
	return s.config
}

func (s *atSink) OnAdd(event *api_v1.Event) {
	receivedEntryCountAT.WithLabelValues(event.Source.Component).Inc()
	s.logEntryChannel <- s.logEntryFactory.FromEvent(event)
}

func (s *atSink) OnUpdate(oldEvent *api_v1.Event, newEvent *api_v1.Event) {
	var oldCount int32
	if oldEvent != nil {
		oldCount = oldEvent.Count
	}

	if newEvent.Count != oldCount+1 {
		glog.Infof("Event count has increased by %d != 1.\n"+
			"\tOld event: %+v\n\tNew event: %+v", newEvent.Count-oldCount, oldEvent, newEvent)
	}

	receivedEntryCountAT.WithLabelValues(newEvent.Source.Component).Inc()

	logEntry := s.logEntryFactory.FromEvent(newEvent)
	s.logEntryChannel <- logEntry
}

func (s *atSink) OnDelete(*api_v1.Event) {
	// Nothing to do here
}

func (s *atSink) OnList(list *api_v1.EventList) {
	if s.beforeFirstList {
		s.logEntryChannel <- s.logEntryFactory.FromMessage("Event exporter started watching. " +
			"Some events may have been lost up to this point.")
		s.beforeFirstList = false
	}
}

func (s *atSink) Run(stopCh <-chan struct{}) {
	glog.Info("Starting Activity Tracker sink")
	for {
		select {
		case entry := <-s.logEntryChannel:
			s.currentBuffer = append(s.currentBuffer, entry)
			if len(s.currentBuffer) >= s.config.MaxBufferSize {
				s.flushBuffer()
			} else if len(s.currentBuffer) == 1 {
				s.setTimer()
			}
			break
		case <-s.getTimerChannel():
			s.flushBuffer()
			break
		case <-stopCh:
			glog.Info("Activity Tracker sink recieved stop signal, waiting for all requests to finish")
			for i := 0; i < s.config.MaxConcurrency; i++ {
				s.concurrencyChannel <- struct{}{}
			}
			glog.Info("All requests to Activity Tracker finished, exiting Activity Tracker sink")
			return
		}
	}
}

func (s *atSink) flushBuffer() {
	entries := s.currentBuffer
	s.currentBuffer = nil
	s.concurrencyChannel <- struct{}{}
	go s.sendEntries(entries)
}

func (s *atSink) sendEntries(entries []*atClient.LogEntry) {

	glog.Infof("Sending %d entries to Activity Tracker", len(entries))
	written := s.writer.Write(entries, s.logName, s.config.Resource)
	successfullySentEntryCountAT.Add(float64(written))

	<-s.concurrencyChannel
	glog.Infof("Successfully sent %d entries to Activity Tracker", len(entries))
}

func (s *atSink) getTimerChannel() <-chan time.Time {
	if s.timer == nil {
		return s.fakeTimeChannel
	}
	return s.timer.C
}

func (s *atSink) setTimer() {
	if s.timer == nil {
		s.timer = time.NewTimer(s.config.FlushDelay)
	} else {
		s.timer.Reset(s.config.FlushDelay)
	}
}
