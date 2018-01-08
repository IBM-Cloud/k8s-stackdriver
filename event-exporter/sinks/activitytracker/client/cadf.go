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

package client

import (
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker/client/pb"
	"github.com/golang/glog"
)

// Write Writes log entries to Activity Tracker
func (service *Service) Write(req *WriteLogEntriesRequest) *WriteLogEntriesResponse {

	// Generate a list of CADF events to send to Activity Tracker
	for i, entry := range req.Entries {
		glog.Infof("processing event %d: %s %s %s", i, entry.Timestamp, entry.Outcome, string(entry.JSONPayload))

		// Serialize the event set
		eventsetJSON, err := service.FromEntry(entry).SerializeJSON()
		if err != nil {
			return &WriteLogEntriesResponse{
				HTTPStatusCode: http.StatusInternalServerError,
				Message:        err.Error(),
			}
		}
		service.Logger.Info(eventsetJSON)
	}

	// Event processing complete
	return &WriteLogEntriesResponse{
		HTTPStatusCode: http.StatusOK,
		Message:        "success",
	}
}

// FromEntry converts an entry to CADF event
func (service *Service) FromEntry(event *LogEntry) *pb.CadfEventWK {

	// Create one CADF event
	cadfEvent := &pb.CadfEvent{
		TypeURI:   atCadfTypeURI,
		Action:    fmt.Sprintf("%s.%s", event.ResourceType, event.Reason),
		EventTime: event.Timestamp,
		Outcome:   event.Outcome,
		Initiator: &pb.CadfResource{
			Id:        service.ClusterID,
			Name:      fmt.Sprintf("%s.%s", atClusterType, event.SourceComponent),
			ProjectId: service.UserAccountID,
			TypeURI:   GetCadfTypeURI(event.SourceComponent),
		},
		Target: &pb.CadfResource{
			Id:        event.ResourceID,
			Name:      fmt.Sprintf("%s.%s", atClusterType, event.ResourceType),
			ProjectId: service.UserAccountID,
			TypeURI:   GetCadfTypeURI(event.ResourceType),
		},
		RequestData: string(event.JSONPayload),
	}

	// Create metadata for this event
	// This describes:
	//   - which service generate this trailing event
	//   - what is this service's provider space id
	//   - which region is this trail event from (ng, eu-gb and etc.)
	//   - which user space being trailed generated this event
	cadfMeta := &pb.Meta{
		ServiceProviderName:      service.ServiceName,
		ServiceProviderProjectId: service.ServiceAccountID,
		ServiceProviderRegion:    service.Location,
		UserAccountIds:           []string{service.UserAccountID},
	}

	// Create one well-known event to hold event and above metadata
	return &pb.CadfEventWK{
		Meta:    cadfMeta,
		Payload: cadfEvent,
	}
}

// GetCadfTypeURI generate a valid type URI for given resource
func GetCadfTypeURI(resourceType string) string {

	// TODO: Determine what resource URI to prepend here
	return resourceType
}
