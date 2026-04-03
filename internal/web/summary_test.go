// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// TestSummaryEndpoint verifies the dashboard summary calculation.
func TestSummaryEndpoint(t *testing.T) {
	now := metav1.NewTime(time.Now())
	cleaner1 := newTestCleaner("cleaner-a", "0 * * * *", appsv1alpha1.ActionScan)
	cleaner1.Status.LastRunTime = &now
	cleaner2 := newTestCleaner("cleaner-b", "5 * * * *", appsv1alpha1.ActionScan)

	report := newTestReport("cleaner-a", appsv1alpha1.ActionScan, []appsv1alpha1.ResourceInfo{
		{Resource: corev1.ObjectReference{Kind: "ConfigMap", Name: "old"}, Message: "orphaned"},
		{Resource: corev1.ObjectReference{Kind: "ConfigMap", Name: "stale"}, Message: "orphaned"},
	})

	c := fake.NewClientBuilder().WithScheme(newTestScheme()).
		WithObjects(cleaner1, cleaner2, report).
		Build()
	handler := testHandler(c, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/summary", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var summary dashboardSummary
	json.NewDecoder(w.Body).Decode(&summary)

	if summary.TotalCleaners != 2 {
		t.Fatalf("expected 2 cleaners, got %d", summary.TotalCleaners)
	}
	if summary.TotalFlaggedResources != 2 {
		t.Fatalf("expected 2 flagged, got %d", summary.TotalFlaggedResources)
	}
	if summary.CleanersWithFindings != 1 {
		t.Fatalf("expected 1 with findings, got %d", summary.CleanersWithFindings)
	}
	if summary.LastScanTime == nil {
		t.Fatal("expected lastScanTime to be set")
	}
}

// TestSummaryEmptyCluster verifies summary with no cleaners.
func TestSummaryEmptyCluster(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
	handler := testHandler(c, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/summary", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var summary dashboardSummary
	json.NewDecoder(w.Body).Decode(&summary)

	if summary.TotalCleaners != 0 {
		t.Fatalf("expected 0 cleaners, got %d", summary.TotalCleaners)
	}
	if summary.LastScanTime != nil {
		t.Fatal("expected lastScanTime to be nil for empty cluster")
	}
}
