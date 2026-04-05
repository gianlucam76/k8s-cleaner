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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

var _ = Describe("Reports", func() {
	It("should list all reports", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).
			WithObjects(
				newTestReport("report-a", nil),
				newTestReport("report-b", []appsv1alpha1.ResourceInfo{
					{Resource: corev1.ObjectReference{Kind: "ConfigMap", Namespace: "default", Name: "old-cm"}, Message: "orphaned"},
				}),
			).
			Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))

		var reports []reportResponse
		Expect(json.NewDecoder(w.Body).Decode(&reports)).To(Succeed())
		Expect(reports).To(HaveLen(2))
	})

	It("should filter reports by cleaner name", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).
			WithObjects(
				newTestReport("report-a", nil),
				newTestReport("report-b", nil),
			).
			Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports?cleaner=report-a", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var reports []reportResponse
		Expect(json.NewDecoder(w.Body).Decode(&reports)).To(Succeed())
		Expect(reports).To(HaveLen(1))
		Expect(reports[0].Name).To(Equal("report-a"))
	})

	It("should filter report resources by kind", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).
			WithObjects(
				newTestReport("report-with-cm", []appsv1alpha1.ResourceInfo{
					{Resource: corev1.ObjectReference{Kind: "ConfigMap", Namespace: "default", Name: "old"}, Message: "orphaned"},
				}),
				newTestReport("report-with-secret", []appsv1alpha1.ResourceInfo{
					{Resource: corev1.ObjectReference{Kind: "Secret", Namespace: "default", Name: "old"}, Message: "orphaned"},
				}),
			).
			Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports?kind=ConfigMap", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var reports []reportResponse
		Expect(json.NewDecoder(w.Body).Decode(&reports)).To(Succeed())
		// Both reports are returned, but only the ConfigMap report has resources
		Expect(reports).To(HaveLen(2))
		for _, r := range reports {
			if r.Name == "report-with-cm" {
				Expect(r.Resources).To(HaveLen(1))
			}
			if r.Name == "report-with-secret" {
				Expect(r.Resources).To(HaveLen(0))
			}
		}
	})

	It("should return 404 for nonexistent report", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/nonexistent", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusNotFound))
	})
})
