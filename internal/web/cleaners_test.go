// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// TestListCleaners verifies the cleaners list endpoint returns all cleaners.
func TestListCleaners(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newTestScheme()).
		WithObjects(
			newTestCleaner("unused-configmaps", "0 6-22 * * *", appsv1alpha1.ActionScan),
			newTestCleaner("pvc-scan", "15 6-22 * * *", appsv1alpha1.ActionScan),
		).
		Build()
	handler := testHandler(c, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cleaners", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var cleaners []cleanerResponse
	json.NewDecoder(w.Body).Decode(&cleaners)
	if len(cleaners) != 2 {
		t.Fatalf("expected 2 cleaners, got %d", len(cleaners))
	}
}

// TestGetCleanerDetail verifies the cleaner detail endpoint includes Lua script.
func TestGetCleanerDetail(t *testing.T) {
	cleaner := newTestCleaner("test-cleaner", "0 * * * *", appsv1alpha1.ActionScan)
	cleaner.Spec.ResourcePolicySet.ResourceSelectors[0].Evaluate = "function evaluate() return true end"

	c := fake.NewClientBuilder().WithScheme(newTestScheme()).
		WithObjects(cleaner).
		Build()
	handler := testHandler(c, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cleaners/test-cleaner", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp cleanerResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Name != "test-cleaner" {
		t.Fatalf("expected name test-cleaner, got %s", resp.Name)
	}
	if resp.LuaScript == "" {
		t.Fatal("expected Lua script in detail response")
	}
}

// TestGetCleanerNotFound verifies 404 for nonexistent cleaner.
func TestGetCleanerNotFound(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
	handler := testHandler(c, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cleaners/nonexistent", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// TestCleanerFlaggedCount verifies cleaners include the flagged count from reports.
func TestCleanerFlaggedCount(t *testing.T) {
	report := newTestReport("my-cleaner", appsv1alpha1.ActionScan, []appsv1alpha1.ResourceInfo{
		{Message: "orphaned"},
		{Message: "orphaned"},
	})

	c := fake.NewClientBuilder().WithScheme(newTestScheme()).
		WithObjects(
			newTestCleaner("my-cleaner", "0 * * * *", appsv1alpha1.ActionScan),
			report,
		).
		Build()
	handler := testHandler(c, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cleaners", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var cleaners []cleanerResponse
	json.NewDecoder(w.Body).Decode(&cleaners)
	if len(cleaners) != 1 {
		t.Fatalf("expected 1 cleaner, got %d", len(cleaners))
	}
	if cleaners[0].FlaggedCount != 2 {
		t.Fatalf("expected flaggedCount 2, got %d", cleaners[0].FlaggedCount)
	}
}
