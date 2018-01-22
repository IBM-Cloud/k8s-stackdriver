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
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/qrtp/lumber"
	context "golang.org/x/net/context"
	gensupport "google.golang.org/api/gensupport"
)

// LogEntry defines a log entry
type LogEntry struct {
	JSONPayload     []byte
	Outcome         string
	Reason          string
	ResourceType    string
	ResourceID      string
	SourceComponent string
	TextPayload     string
	Timestamp       string
}

// GetPayload returns JSON payload if present, othewise text payload.
func (l *LogEntry) GetPayload() string {
	if l.JSONPayload != nil && len(l.JSONPayload) > 0 {
		return string(l.JSONPayload)
	}
	return l.TextPayload
}

// MonitoredResource defines a resource
type MonitoredResource struct {
	Type   string
	Labels map[string]string
}

// WriteLogEntriesRequest log entries to write
type WriteLogEntriesRequest struct {
	Entries  []*LogEntry
	LogName  string
	Resource *MonitoredResource
}

// WriteLogEntriesResponse result returned from WriteLogEntries. empty
type WriteLogEntriesResponse struct {
	HTTPStatusCode int    `json:"status_code"`
	Message        string `json:"message"`
}

// Service stub service struct
type Service struct {
	ClusterID        string
	Logger           *lumber.FileLogger
	LogPath          string
	Location         string
	ServiceName      string
	ServiceAccountID string
	UserAccountID    string
}

// EntriesWriteCall ...
type EntriesWriteCall struct {
	s                      *Service
	writelogentriesrequest *WriteLogEntriesRequest
	urlParams              gensupport.URLParams
	ctx                    context.Context
	header                 http.Header
}

// NewService creates a new service object use to reference AT
func NewService() (*Service, error) {

	// Initialize the log root directory for this cluster (if necessary)
	logFile := fmt.Sprintf("%s/event-exporter/%s-events.log", GetAtLogRoot(), GetClusterID())
	if err := os.MkdirAll(filepath.Dir(logFile), os.ModePerm); err != nil {
		glog.Errorf("unable to initialize log root for path %s", logFile)
		return nil, err
	}

	// Prepare the logger
	logger, err := lumber.NewFileLogger(logFile, lumber.INFO, lumber.ROTATE, 5000, 9, 100)
	if err != nil {
		glog.Errorf("error initializing logger: %s", err.Error())
		return nil, err
	}
	logger.TimeFormat("")

	// Initialize service configuration
	return &Service{
		ClusterID:        GetClusterID(),
		Logger:           logger,
		LogPath:          logFile,
		Location:         GetClusterLocation(),
		ServiceName:      GetServiceName(),
		ServiceAccountID: GetServiceAccountID(),
		UserAccountID:    GetUserAccountID(),
	}, nil
}
