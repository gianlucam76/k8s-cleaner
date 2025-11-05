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
	deletedCounterHelp = "The cumulative count of resources successfully deleted by cleaner."

	updatedCounterName = "k8s_cleaner_updated_resources_total"
	updatedCounterHelp = "The cumulative count of resources successfully updated by cleaner."

	scanCounterName = "k8s_cleaner_scan_resources_total"
	scanCounterHelp = "The cumulative count of resources successfully found by cleaner during a scan."

	errorCounterName = "k8s_cleaner_error_resources_total"
	errorCounterHelp = "The cumulative count of erros encountered by cleaner during a scan."
)

const (
	deletedGaugeName = "k8s_cleaner_current_deleted_resources_count"
	deletedGaugeHelp = "The current total count of resources deleted by the cleaner (snapshot of the counter)."

	updatedGaugeName = "k8s_cleaner_current_updated_resources_count"
	updatedGaugeHelp = "The current total count of resources updated by the cleaner (snapshot of the counter)."

	scanGaugeName = "k8s_cleaner_current_resources_count"
	scanGaugeHelp = "The current count of resources found in the latest scan."

	errorGaugeName = "k8s_cleaner_current_error_resources_count"
	errorGaugeHelp = "The current total count of errors encountered (snapshot of the counter)."
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

	errorResourceCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: os.Getenv("NAMESPACE"),
			Name:      errorCounterName,
			Help:      errorCounterHelp,
		},
		[]string{"cleaner_instance", "resource_apiversion", "resource_type"},
	)
)

var (
	// New Gauge Definitions
	deletedResourceGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: os.Getenv("NAMESPACE"),
			Name:      deletedGaugeName,
			Help:      deletedGaugeHelp,
		},
		[]string{"cleaner_instance"},
	)

	updatedResourceGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: os.Getenv("NAMESPACE"),
			Name:      updatedGaugeName,
			Help:      updatedGaugeHelp,
		},
		[]string{"cleaner_instance"},
	)

	scanResourceGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: os.Getenv("NAMESPACE"),
			Name:      scanGaugeName,
			Help:      scanGaugeHelp,
		},
		[]string{"cleaner_instance"},
	)

	errorResourceGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: os.Getenv("NAMESPACE"),
			Name:      errorGaugeName,
			Help:      errorGaugeHelp,
		},
		[]string{"cleaner_instance"},
	)
)

//nolint:gochecknoinits // forced pattern for prometheus and controller-runtime
func init() {
	// Register custom metrics with the controller-runtime's global registry.
	// This ensures it is exported alongside default controller-runtime metrics.
	metrics.Registry.MustRegister(deletedResourceCounter, updatedResourceCounter,
		scanResourceCounter, errorResourceCounter,
		deletedResourceGauge, updatedResourceGauge,
		scanResourceGauge, errorResourceGauge)
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

// getErrorResourcesCounterVec returns the singleton CounterVec instance.
func getErrorResourcesCounterVec() *prometheus.CounterVec {
	return errorResourceCounter
}

func reportErrorEvent(cleanerName, resourceAPIVersion, resourceKind string) {
	// Get the global CounterVec instance.
	counterVec := getErrorResourcesCounterVec()

	// Increment the counter for the specific cleaner instance and resource kind.
	counterVec.WithLabelValues(cleanerName, resourceAPIVersion, resourceKind).Inc()
}

// Helper functions for Gauges to report an absolute count

// getDeletedResourcesGaugeVec returns the singleton GaugeVec instance.
func getDeletedResourcesGaugeVec() *prometheus.GaugeVec {
	return deletedResourceGauge
}

// reportDeletedCount sets the gauge to the absolute total count of deleted resources
// for the given cleaner instance and resource kind.
func reportDeletedCount(cleanerName string, count float64) {
	gaugeVec := getDeletedResourcesGaugeVec()

	// Set the gauge to the specific count value
	gaugeVec.WithLabelValues(cleanerName).Set(count)
}

// ---

// getUpdatedResourcesGaugeVec returns the singleton GaugeVec instance.
func getUpdatedResourcesGaugeVec() *prometheus.GaugeVec {
	return updatedResourceGauge
}

// reportUpdatedCount sets the gauge to the absolute total count of updated resources
// for the given cleaner instance and resource kind.
func reportUpdatedCount(cleanerName string, count float64) {
	gaugeVec := getUpdatedResourcesGaugeVec()

	// Set the gauge to the specific count value
	gaugeVec.WithLabelValues(cleanerName).Set(count)
}

// ---

// getScanResourcesGaugeVec returns the singleton GaugeVec instance.
func getScanResourcesGaugeVec() *prometheus.GaugeVec {
	return scanResourceGauge
}

// reportScanCount sets the gauge to the absolute total count of resources
// found in the latest scan for the given cleaner instance and resource kind.
func reportScanCount(cleanerName string, count float64) {
	gaugeVec := getScanResourcesGaugeVec()

	// Set the gauge to the specific count value
	gaugeVec.WithLabelValues(cleanerName).Set(count)
}

// ---

// getErrorResourcesGaugeVec returns the singleton GaugeVec instance.
func getErrorResourcesGaugeVec() *prometheus.GaugeVec {
	return errorResourceGauge
}

// reportErrorCount sets the gauge to the absolute total count of errors
// for the given cleaner instance and resource kind.
func reportErrorCount(cleanerName string, count float64) {
	gaugeVec := getErrorResourcesGaugeVec()

	// Set the gauge to the specific count value
	gaugeVec.WithLabelValues(cleanerName).Set(count)
}
