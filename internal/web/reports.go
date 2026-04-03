// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package web

import (
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// reportResponse is the JSON representation of a Report for the API.
type reportResponse struct {
	Name      string         `json:"name"`
	Action    string         `json:"action"`
	Resources []resourceItem `json:"resources"`
}

// resourceItem is a single flagged resource within a report.
type resourceItem struct {
	Kind       string `json:"kind"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name"`
	APIVersion string `json:"apiVersion"`
	Message    string `json:"message,omitempty"`
}

// ListReportsHandler returns all reports, with optional filtering.
// Query params: ?cleaner=X&namespace=Y&kind=Z
func ListReportsHandler(c client.Client, log logr.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		reports, err := listReports(ctx, c)
		if err != nil {
			log.Error(err, "failed to list reports")
			respondError(w, http.StatusInternalServerError, "failed to list reports")
			return
		}

		// Filters
		filterCleaner := r.URL.Query().Get("cleaner")
		filterNamespace := r.URL.Query().Get("namespace")
		filterKind := r.URL.Query().Get("kind")

		result := make([]reportResponse, 0, len(reports))
		for i := range reports {
			// Filter by cleaner name (Report name == Cleaner name)
			if filterCleaner != "" && !strings.EqualFold(reports[i].Name, filterCleaner) {
				continue
			}

			resp := toReportResponse(&reports[i], filterNamespace, filterKind)
			result = append(result, resp)
		}

		respondJSON(w, http.StatusOK, result)
	}
}

// GetReportHandler returns a single report by name.
// Query params: ?namespace=Y&kind=Z (filter resources within the report)
func GetReportHandler(c client.Client, log logr.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		name := r.PathValue("name")

		var report appsv1alpha1.Report
		if err := c.Get(ctx, client.ObjectKey{Name: name}, &report); err != nil {
			if client.IgnoreNotFound(err) == nil {
				respondError(w, http.StatusNotFound, "report not found")
				return
			}
			log.Error(err, "failed to get report", "name", name)
			respondError(w, http.StatusInternalServerError, "failed to get report")
			return
		}

		filterNamespace := r.URL.Query().Get("namespace")
		filterKind := r.URL.Query().Get("kind")

		resp := toReportResponse(&report, filterNamespace, filterKind)
		respondJSON(w, http.StatusOK, resp)
	}
}

// toReportResponse converts a Report CR to the API response format.
// Optionally filters resources by namespace and kind.
func toReportResponse(report *appsv1alpha1.Report, filterNS, filterKind string) reportResponse {
	resp := reportResponse{
		Name:   report.Name,
		Action: string(report.Spec.Action),
	}

	resources := make([]resourceItem, 0, len(report.Spec.ResourceInfo))
	for _, ri := range report.Spec.ResourceInfo {
		ref := ri.Resource

		if filterNS != "" && !strings.EqualFold(ref.Namespace, filterNS) {
			continue
		}
		if filterKind != "" && !strings.EqualFold(ref.Kind, filterKind) {
			continue
		}

		resources = append(resources, resourceItem{
			Kind:       ref.Kind,
			Namespace:  ref.Namespace,
			Name:       ref.Name,
			APIVersion: ref.APIVersion,
			Message:    ri.Message,
		})
	}
	resp.Resources = resources

	return resp
}
