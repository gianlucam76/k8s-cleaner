/*
Copyright 2026. projectsveltos.io. All rights reserved.

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

package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

var _ = Describe("Summary", func() {
	It("should calculate dashboard summary correctly", func() {
		now := metav1.NewTime(time.Now())
		cleaner1 := newTestCleaner("cleaner-a", "0 * * * *")
		cleaner1.Status.LastRunTime = &now
		cleaner2 := newTestCleaner("cleaner-b", "5 * * * *")

		report := newTestReport("cleaner-a", []appsv1alpha1.ResourceInfo{
			{Resource: corev1.ObjectReference{Kind: "ConfigMap", Name: "old"}, Message: "orphaned"},
			{Resource: corev1.ObjectReference{Kind: "ConfigMap", Name: "stale"}, Message: "orphaned"},
		})

		c := fake.NewClientBuilder().WithScheme(newTestScheme()).
			WithObjects(cleaner1, cleaner2, report).
			Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/summary", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))

		var summary dashboardSummary
		Expect(json.NewDecoder(w.Body).Decode(&summary)).To(Succeed())
		Expect(summary.TotalCleaners).To(Equal(2))
		Expect(summary.TotalFlaggedResources).To(Equal(2))
		Expect(summary.CleanersWithFindings).To(Equal(1))
		Expect(summary.LastScanTime).ToNot(BeNil())
	})

	It("should handle empty cluster", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/summary", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var summary dashboardSummary
		Expect(json.NewDecoder(w.Body).Decode(&summary)).To(Succeed())
		Expect(summary.TotalCleaners).To(Equal(0))
		Expect(summary.LastScanTime).To(BeNil())
	})
})
