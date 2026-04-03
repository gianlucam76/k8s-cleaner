// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// testHandler creates a fully wrapped handler for testing (routes + middleware).
func testHandler(c client.Client, readOnly bool) http.Handler {
	log := zap.New(zap.UseDevMode(true))
	mux := setupRoutes(c, readOnly, "test", log)
	return applyMiddleware(mux, readOnly, log)
}

func newTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	appsv1alpha1.AddToScheme(s)
	return s
}

func newTestCleaner(name, schedule string, action appsv1alpha1.Action) *appsv1alpha1.Cleaner {
	return &appsv1alpha1.Cleaner{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appsv1alpha1.CleanerSpec{
			Schedule: schedule,
			Action:   action,
			ResourcePolicySet: appsv1alpha1.ResourcePolicySet{
				ResourceSelectors: []appsv1alpha1.ResourceSelector{
					{Group: "", Version: "v1", Kind: "ConfigMap"},
				},
			},
		},
	}
}

func newTestReport(name string, action appsv1alpha1.Action, resources []appsv1alpha1.ResourceInfo) *appsv1alpha1.Report {
	return &appsv1alpha1.Report{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appsv1alpha1.ReportSpec{
			Action:       action,
			ResourceInfo: resources,
		},
	}
}

// TestHealthEndpoint verifies the health check returns 200 OK.
func TestHealthEndpoint(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
	handler := testHandler(c, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %s", resp["status"])
	}
}

// TestConfigEndpoint verifies config returns readOnly state.
func TestConfigEndpoint(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()

	tests := []struct {
		name     string
		readOnly bool
	}{
		{"read-write", false},
		{"read-only", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := testHandler(c, tt.readOnly)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", w.Code)
			}

			var resp configResponse
			json.NewDecoder(w.Body).Decode(&resp)
			if resp.ReadOnly != tt.readOnly {
				t.Fatalf("expected readOnly=%v, got %v", tt.readOnly, resp.ReadOnly)
			}
		})
	}
}

// TestReadOnlyMiddlewareBlocksPOST verifies POST requests are rejected in read-only mode.
func TestReadOnlyMiddlewareBlocksPOST(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newTestScheme()).
		WithObjects(newTestCleaner("test", "0 * * * *", appsv1alpha1.ActionScan)).
		Build()
	handler := testHandler(c, true)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cleaners/test/trigger", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

// TestReadOnlyMiddlewareAllowsGET verifies GET requests pass in read-only mode.
func TestReadOnlyMiddlewareAllowsGET(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
	handler := testHandler(c, true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
