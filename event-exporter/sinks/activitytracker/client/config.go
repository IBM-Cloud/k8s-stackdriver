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
	"os"
	"strconv"
	"time"
)

/*
 * Environment variables required for IBM Activity Tracker:
 *
 *   ACCOUNT_ID
 *   ACTIVITY_TRACKER_LOG_ROOT
 *   CLUSTER_ID
 *   CLUSTER_LOCATION
 *   SERVICE_PROVIDER_ACCOUNT_ID
 *   SERVICE_PROVIDER_NAME
 */

const (
	atLogRootKey  = "ACTIVITY_TRACKER_LOG_ROOT"
	atCadfTypeURI = "http://schemas.dmtf.org/cloud/audit/1.0/event"
	atClusterType = "kubernetes_cluster"

	userLocationKey = "CLUSTER_LOCATION"
	userClusterKey  = "CLUSTER_ID"
	userAccountKey  = "ACCOUNT_ID"

	serviceProviderAccountKey = "SERVICE_PROVIDER_ACCOUNT_ID"
	serviceProviderName       = "SERVICE_PROVIDER_NAME"
	octetStream               = "application/octet-stream"

	// EventTypeError indicates an error event
	EventTypeError = "FAILURE"
	// EventTypeSuccess indicates a successful event
	EventTypeSuccess = "SUCCESS"
	// EventTypeInfo indicates an informational event
	EventTypeInfo = "INFO"
)

// GetClusterLocation location key of the IBM datacenter
func GetClusterLocation() string {
	return os.Getenv(userLocationKey)
}

// GetFormattedTimestamp formatted timestamp used to generate encrypted token
func GetFormattedTimestamp() string {
	return strconv.FormatInt(time.Now().UTC().Unix(), 10)
}

// GetClusterID unique ID for the cluster
func GetClusterID() string {
	return os.Getenv(userClusterKey)
}

// GetServiceAccountID service provider account ID
func GetServiceAccountID() string {
	return os.Getenv(serviceProviderAccountKey)
}

// GetServiceName service provider name
func GetServiceName() string {
	return os.Getenv(serviceProviderName)
}

// GetUserAccountID account ID associated with the cluster
func GetUserAccountID() string {
	return os.Getenv(userAccountKey)
}

// GetAtLogRoot directory where logs should be stored
func GetAtLogRoot() string {
	return os.Getenv(atLogRootKey)
}

// OnIBM determines whether the expected environment variables are present
// in the running container to support the AT sink implementation.
func OnIBM() bool {
	return GetClusterID() != "" &&
		GetClusterLocation() != "" &&
		GetUserAccountID() != "" &&
		GetServiceAccountID() != "" &&
		GetServiceName() != "" &&
		GetAtLogRoot() != ""
}
