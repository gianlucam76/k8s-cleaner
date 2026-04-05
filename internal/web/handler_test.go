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
	_ = appsv1alpha1.AddToScheme(s)
	return s
}

func newTestCleaner(name, schedule string) *appsv1alpha1.Cleaner {
	return &appsv1alpha1.Cleaner{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appsv1alpha1.CleanerSpec{
			Schedule: schedule,
			Action:   appsv1alpha1.ActionScan,
			ResourcePolicySet: appsv1alpha1.ResourcePolicySet{
				ResourceSelectors: []appsv1alpha1.ResourceSelector{
					{Group: "", Version: "v1", Kind: "ConfigMap"},
				},
			},
		},
	}
}

func newTestReport(name string, resources []appsv1alpha1.ResourceInfo) *appsv1alpha1.Report {
	return &appsv1alpha1.Report{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appsv1alpha1.ReportSpec{
			Action:       appsv1alpha1.ActionScan,
			ResourceInfo: resources,
		},
	}
}

var _ = Describe("Handler", func() {
	It("should return 200 OK on health check", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))

		var resp map[string]string
		Expect(json.NewDecoder(w.Body).Decode(&resp)).To(Succeed())
		Expect(resp["status"]).To(Equal("ok"))
	})

	It("should return readOnly=false in read-write mode", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/config", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))

		var resp configResponse
		Expect(json.NewDecoder(w.Body).Decode(&resp)).To(Succeed())
		Expect(resp.ReadOnly).To(BeFalse())
	})

	It("should return readOnly=true in read-only mode", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, true)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/config", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))

		var resp configResponse
		Expect(json.NewDecoder(w.Body).Decode(&resp)).To(Succeed())
		Expect(resp.ReadOnly).To(BeTrue())
	})

	It("should block POST requests in read-only mode", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).
			WithObjects(newTestCleaner("test", "0 * * * *")).
			Build()
		handler := testHandler(c, true)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cleaners/test/trigger", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusForbidden))
	})

	It("should allow GET requests in read-only mode", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, true)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))
	})
})
