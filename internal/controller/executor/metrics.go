/*
Copyright 2025. projectsveltos.io. All rights reserved.

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

package executor

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// --- Metric Constants ---

const (
	deletedCounterName = "k8s_cleaner_deleted_resources_total"
	deletedCounterHelp = "The cumulative count of resources successfully deleted by the cleaner."

	updatedCounterName = "k8s_cleaner_updated_resources_total"
	updatedCounterHelp = "The cumulative count of resources successfully updated by the cleaner."

	scanCounterName = "k8s_cleaner_scan_resources_total"
	scanCounterHelp = "The cumulative count of resources successfully found by the cleaner during a scan."
)

var (
	deletedResourceCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: os.Getenv("NAMESPACE"),
			Name:      deletedCounterName,
			Help:      deletedCounterHelp,
		},
		[]string{"cleaner_instance", "resource_apiversion", "resource_type"},
	)

	updatedResourceCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: os.Getenv("NAMESPACE"),
			Name:      updatedCounterName,
			Help:      updatedCounterHelp,
		},
		[]string{"cleaner_instance", "resource_apiversion", "resource_type"},
	)

	scanResourceCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: os.Getenv("NAMESPACE"),
			Name:      scanCounterName,
			Help:      scanCounterHelp,
		},
		[]string{"cleaner_instance", "resource_apiversion", "resource_type"},
	)
)

//nolint:gochecknoinits // forced pattern for prometheus and controller-runtime
func init() {
	// Register custom metrics with the controller-runtime's global registry.
	// This ensures it is exported alongside default controller-runtime metrics.
	metrics.Registry.MustRegister(deletedResourceCounter, updatedResourceCounter, scanResourceCounter)
}

// getDeletedResourcesCounterVec returns the singleton CounterVec instance.
func getDeletedResourcesCounterVec() *prometheus.CounterVec {
	return deletedResourceCounter
}

func reportDeletionEvent(cleanerName, resourceAPIVersion, resourceKind string) {
	// Get the global CounterVec instance.
	counterVec := getDeletedResourcesCounterVec()

	// Increment the counter for the specific cleaner instance and resource kind.
	counterVec.WithLabelValues(cleanerName, resourceAPIVersion, resourceKind).Inc()
}

// getUpdatedResourcesCounterVec returns the singleton CounterVec instance.
func getUpdatedResourcesCounterVec() *prometheus.CounterVec {
	return deletedResourceCounter
}

func reportUpdateEvent(cleanerName, resourceAPIVersion, resourceKind string) {
	// Get the global CounterVec instance.
	counterVec := getUpdatedResourcesCounterVec()

	// Increment the counter for the specific cleaner instance and resource kind.
	counterVec.WithLabelValues(cleanerName, resourceAPIVersion, resourceKind).Inc()
}

// getScanResourcesCounterVec returns the singleton CounterVec instance.
func getScanResourcesCounterVec() *prometheus.CounterVec {
	return scanResourceCounter
}

func reportScanEvent(cleanerName, resourceAPIVersion, resourceKind string) {
	// Get the global CounterVec instance.
	counterVec := getScanResourcesCounterVec()

	// Increment the counter for the specific cleaner instance and resource kind.
	counterVec.WithLabelValues(cleanerName, resourceAPIVersion, resourceKind).Inc()
}
