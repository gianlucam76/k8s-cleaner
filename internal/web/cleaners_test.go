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

	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

var _ = Describe("Cleaners", func() {
	It("should list all cleaners", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).
			WithObjects(
				newTestCleaner("unused-configmaps", "0 6-22 * * *"),
				newTestCleaner("pvc-scan", "15 6-22 * * *"),
			).
			Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cleaners", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))

		var cleaners []cleanerResponse
		Expect(json.NewDecoder(w.Body).Decode(&cleaners)).To(Succeed())
		Expect(cleaners).To(HaveLen(2))
	})

	It("should return cleaner detail with Lua script", func() {
		cleaner := newTestCleaner("test-cleaner", "0 * * * *")
		cleaner.Spec.ResourcePolicySet.ResourceSelectors[0].Evaluate = "function evaluate() return true end"

		c := fake.NewClientBuilder().WithScheme(newTestScheme()).
			WithObjects(cleaner).
			Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cleaners/test-cleaner", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))

		var resp cleanerResponse
		Expect(json.NewDecoder(w.Body).Decode(&resp)).To(Succeed())
		Expect(resp.Name).To(Equal("test-cleaner"))
		Expect(resp.LuaScript).ToNot(BeEmpty())
	})

	It("should return 404 for nonexistent cleaner", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cleaners/nonexistent", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusNotFound))
	})

	It("should include flagged count from reports", func() {
		report := newTestReport("my-cleaner", []appsv1alpha1.ResourceInfo{
			{Message: "orphaned"},
			{Message: "orphaned"},
		})

		c := fake.NewClientBuilder().WithScheme(newTestScheme()).
			WithObjects(
				newTestCleaner("my-cleaner", "0 * * * *"),
				report,
			).
			Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cleaners", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var cleaners []cleanerResponse
		Expect(json.NewDecoder(w.Body).Decode(&cleaners)).To(Succeed())
		Expect(cleaners).To(HaveLen(1))
		Expect(cleaners[0].FlaggedCount).To(Equal(2))
	})
})
