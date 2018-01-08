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
	"flag"
	"fmt"
	"time"

	"github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks"
	atClient "github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker/client"
	"k8s.io/apimachinery/pkg/util/clock"
)

type atSinkFactory struct {
	flagSet        *flag.FlagSet
	flushDelay     *time.Duration
	maxBufferSize  *int
	maxConcurrency *int
}

// NewAtSinkFactory creates a new Activity Tracker sink factory
func NewAtSinkFactory() sinks.SinkFactory {
	fs := flag.NewFlagSet("activitytracker", flag.ContinueOnError)
	return &atSinkFactory{
		flagSet: fs,
		flushDelay: fs.Duration("flush-delay", defaultFlushDelay, "Delay after receiving "+
			"the first event in batch before sending the request to Activity Tracker, if batch"+
			"doesn't get sent before"),
		maxBufferSize: fs.Int("max-buffer-size", defaultMaxBufferSize, "Maximum number of events "+
			"in the request to Activity Tracker"),
		maxConcurrency: fs.Int("max-concurrency", defaultMaxConcurrency, "Maximum number of "+
			"concurrent requests to Activity Tracker"),
	}
}

func (f *atSinkFactory) CreateNew(opts []string) (sinks.Sink, error) {
	err := f.flagSet.Parse(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sink opts: %v", err)
	}

	config, err := f.createSinkConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %v", err)
	}

	service, err := atClient.NewService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Activity Tracker service: %v", err)
	}
	writer := newAtWriter(service)
	clk := clock.RealClock{}
	return newAtSink(writer, clk, config), nil
}

func (f *atSinkFactory) createSinkConfig() (*atSinkConfig, error) {
	config, err := newIbmAtSinkConfig()
	if err != nil {
		return nil, err
	}

	config.FlushDelay = *f.flushDelay
	config.MaxBufferSize = *f.maxBufferSize
	config.MaxConcurrency = *f.maxConcurrency

	return config, nil
}
