// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// TestListReports verifies the reports list endpoint.
func TestListReports(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newTestScheme()).
		WithObjects(
			newTestReport("report-a", appsv1alpha1.ActionScan, nil),
			newTestReport("report-b", appsv1alpha1.ActionScan, []appsv1alpha1.ResourceInfo{
				{Resource: corev1.ObjectReference{Kind: "ConfigMap", Namespace: "default", Name: "old-cm"}, Message: "orphaned"},
			}),
		).
		Build()
	handler := testHandler(c, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var reports []reportResponse
	json.NewDecoder(w.Body).Decode(&reports)
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}
}

// TestListReportsFilterByCleaner verifies filtering by cleaner name.
func TestListReportsFilterByCleaner(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newTestScheme()).
		WithObjects(
			newTestReport("report-a", appsv1alpha1.ActionScan, nil),
			newTestReport("report-b", appsv1alpha1.ActionScan, nil),
		).
		Build()
	handler := testHandler(c, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports?cleaner=report-a", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var reports []reportResponse
	json.NewDecoder(w.Body).Decode(&reports)
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].Name != "report-a" {
		t.Fatalf("expected report-a, got %s", reports[0].Name)
	}
}

// TestListReportsFilterByKind verifies kind filter removes non-matching resources within reports.
func TestListReportsFilterByKind(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newTestScheme()).
		WithObjects(
			newTestReport("report-with-cm", appsv1alpha1.ActionScan, []appsv1alpha1.ResourceInfo{
				{Resource: corev1.ObjectReference{Kind: "ConfigMap", Namespace: "default", Name: "old"}, Message: "orphaned"},
			}),
			newTestReport("report-with-secret", appsv1alpha1.ActionScan, []appsv1alpha1.ResourceInfo{
				{Resource: corev1.ObjectReference{Kind: "Secret", Namespace: "default", Name: "old"}, Message: "orphaned"},
			}),
		).
		Build()
	handler := testHandler(c, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports?kind=ConfigMap", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var reports []reportResponse
	json.NewDecoder(w.Body).Decode(&reports)
	// Both reports are returned, but only the ConfigMap report has resources
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}
	// Find the CM report and verify it has 1 resource
	for _, r := range reports {
		if r.Name == "report-with-cm" && len(r.Resources) != 1 {
			t.Fatalf("expected 1 resource in cm report, got %d", len(r.Resources))
		}
		if r.Name == "report-with-secret" && len(r.Resources) != 0 {
			t.Fatalf("expected 0 resources in secret report after kind filter, got %d", len(r.Resources))
		}
	}
}

// TestGetReportNotFound verifies 404 for nonexistent report.
func TestGetReportNotFound(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
	handler := testHandler(c, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/nonexistent", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
