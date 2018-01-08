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
	"github.com/prometheus/client_golang/prometheus"
)

const (
	retryDelay = 10 * time.Second
)

var (
	requestCountAT = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "request_count",
			Help:      "Number of requests, issued to Activity Tracker API",
			Subsystem: "activitytracker_sink",
		},
		[]string{"code"},
	)
)

type atWriter interface {
	Write([]*atClient.LogEntry, string, *atClient.MonitoredResource) int
	GetService() *atClient.Service
}

type atWriterImpl struct {
	service *atClient.Service
}

func newAtWriter(service *atClient.Service) atWriter {
	return &atWriterImpl{
		service: service,
	}
}

func (w atWriterImpl) GetService() *atClient.Service {
	return w.service
}

func (w atWriterImpl) Write(entries []*atClient.LogEntry, logName string, resource *atClient.MonitoredResource) int {
	w.service.Write(&atClient.WriteLogEntriesRequest{
		Entries:  entries,
		LogName:  logName,
		Resource: resource,
	})
	return len(entries)
}
