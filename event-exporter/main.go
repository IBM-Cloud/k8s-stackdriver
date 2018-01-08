/*
Copyright 2017 Google Inc.

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

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks"
	"github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/activitytracker"
	"github.com/GoogleCloudPlatform/k8s-stackdriver/event-exporter/sinks/stackdriver"
)

var (
	resyncPeriod       = flag.Duration("resync-period", 1*time.Minute, "Reflector resync period")
	sinkOpts           = flag.String("sink-opts", "", "Parameters for configuring sink")
	sinkProvider       = flag.String("sink-provider", sinks.SinkProviderGKE, "Provider for event storage: IBM or GKE")
	prometheusEndpoint = flag.String("prometheus-endpoint", ":80", "Endpoint on which to "+
		"expose Prometheus http handler")
)

func newSystemStopChannel() chan struct{} {
	ch := make(chan struct{})
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		sig := <-c
		glog.Infof("Recieved signal %s, terminating", sig.String())

		ch <- struct{}{}
	}()

	return ch
}

func newKubernetesClient() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %v", err)
	}

	return kubernetes.NewForConfig(config)
}

func main() {
	flag.Set("logtostderr", "true")
	defer glog.Flush()
	flag.Parse()

	// Choose the sink provider
	var sink sinks.Sink
	var err error
	switch *sinkProvider {
	case sinks.SinkProviderGKE:
		// GKE Stackdriver
		sink, err = stackdriver.NewSdSinkFactory().CreateNew(strings.Split(*sinkOpts, " "))
		if err != nil {
			glog.Fatalf("Failed to initialize sink: %v", err)
		}
	case sinks.SinkProviderIBM:
		// IBM Activity Tracker
		sink, err = activitytracker.NewAtSinkFactory().CreateNew(strings.Split(*sinkOpts, " "))
		if err != nil {
			glog.Fatalf("Failed to initialize sink: %v", err)
		}
	default:
		// Fail on other stack providers
		glog.Fatalf("Unsupported sink provider: %s", *sinkProvider)
	}

	// Prepare the Kubernetes client
	client, err := newKubernetesClient()
	if err != nil {
		glog.Fatalf("Failed to initialize kubernetes client: %v", err)
	}
	eventExporter := newEventExporter(client, sink, *resyncPeriod)

	// Expose the Prometheus http endpoint
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		glog.Fatalf("Prometheus monitoring failed: %v", http.ListenAndServe(*prometheusEndpoint, nil))
	}()

	stopCh := newSystemStopChannel()
	eventExporter.Run(stopCh)
}
