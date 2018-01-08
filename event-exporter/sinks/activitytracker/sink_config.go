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
	"errors"
	"fmt"
	"time"

	atClient "github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker/client"
)

const (
	defaultFlushDelay     = 5 * time.Second
	defaultMaxBufferSize  = 100
	defaultMaxConcurrency = 10
)

type atSinkConfig struct {
	FlushDelay     time.Duration
	MaxBufferSize  int
	MaxConcurrency int
	LogName        string
	Resource       *atClient.MonitoredResource
}

func newIbmAtSinkConfig() (*atSinkConfig, error) {

	// Validate running in IBM Container Service
	if !atClient.OnIBM() {
		return nil, errors.New("not running on IBM Container Service, which is not supported for Activity Tracker sink")
	}

	// Prepare the resource object
	accountID := atClient.GetUserAccountID()
	clusterID := atClient.GetClusterID()
	location := atClient.GetClusterLocation()
	logName := fmt.Sprintf("accounts/%s/events", accountID)
	resource := &atClient.MonitoredResource{
		Type: "ibm_cluster",
		Labels: map[string]string{
			"account_id": accountID,
			"cluster_id": clusterID,
			"location":   location,
		},
	}

	// Prepare the sink config object
	return &atSinkConfig{
		FlushDelay:     defaultFlushDelay,
		MaxBufferSize:  defaultMaxBufferSize,
		MaxConcurrency: defaultMaxConcurrency,
		LogName:        logName,
		Resource:       resource,
	}, nil
}
