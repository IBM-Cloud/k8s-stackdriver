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
	go_json "encoding/json"
	"os"
	"time"

	atClient "github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker/client"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/kubernetes/scheme"
	api_v1 "k8s.io/client-go/pkg/api/v1"
)

var (
	// fieldBlacklist is a list of fields that should be excluded from the
	// json object sent to activity tracker.
	fieldBlacklist = []string{
		// Is unnecessary, because it's demuxed already
		"count",
		// Timestamp is in the logEntry's metadata
		"lastTimestamp",
		// Not relevant because of demuxing
		"firstTimestamp",
	}
)

type atLogEntryFactory struct {
	clock   clock.Clock
	encoder runtime.Encoder
}

func newAtLogEntryFactory(clock clock.Clock) *atLogEntryFactory {
	return &atLogEntryFactory{
		clock:   clock,
		encoder: newEncoder(),
	}
}

func (f *atLogEntryFactory) FromEvent(event *api_v1.Event) *atClient.LogEntry {
	payload, err := f.serializeEvent(event)
	if err != nil {
		glog.Warningf("Failed to encode event %+v: %v", event, err)
	}

	return &atClient.LogEntry{
		JSONPayload:     payload,
		Outcome:         f.parseOutcome(event),
		Timestamp:       event.LastTimestamp.Format(time.RFC3339Nano),
		Reason:          event.Reason,
		ResourceID:      event.InvolvedObject.Name,
		ResourceType:    event.InvolvedObject.Kind,
		SourceComponent: event.Source.Component,
	}
}

func (f *atLogEntryFactory) FromMessage(msg string) *atClient.LogEntry {
	// Retrieve hostname of the pod
	hostname, _ := os.Hostname()
	return &atClient.LogEntry{
		TextPayload:     msg,
		Outcome:         atClient.EventTypeSuccess,
		Timestamp:       f.clock.Now().Format(time.RFC3339Nano),
		Reason:          atClient.EventTypeInfo,
		ResourceID:      hostname,
		ResourceType:    atClient.GetServiceName(),
		SourceComponent: atClient.GetServiceName(),
	}
}

func (f *atLogEntryFactory) parseOutcome(event *api_v1.Event) string {
	if event.Type == "Warning" {
		return atClient.EventTypeError
	}
	return atClient.EventTypeSuccess
}

func (f *atLogEntryFactory) serializeEvent(event *api_v1.Event) ([]byte, error) {
	bytes, err := runtime.Encode(f.encoder, event)
	if err != nil {
		return nil, err
	}

	var obj map[string]interface{}
	err = go_json.Unmarshal(bytes, &obj)
	if err != nil {
		return nil, err
	}

	for _, field := range fieldBlacklist {
		delete(obj, field)
	}

	return go_json.Marshal(obj)
}

func newEncoder() runtime.Encoder {
	jsonSerializer := json.NewSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, false)
	directCodecFactory := serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
	return directCodecFactory.EncoderForVersion(jsonSerializer, api_v1.SchemeGroupVersion)
}
