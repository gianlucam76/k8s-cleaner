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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

func fullResourceFor(namespace, name string, data map[string]string) []byte {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion(apiVersionV1)
	u.SetKind(kindConfigMap)
	u.SetNamespace(namespace)
	u.SetName(name)
	if data != nil {
		Expect(unstructured.SetNestedStringMap(u.Object, data, "data")).To(Succeed())
	}
	b, err := json.Marshal(u)
	Expect(err).To(BeNil())
	return b
}

var _ = Describe("Rollback", func() {
	It("recreates a resource deleted by a Delete action", func() {
		cmName := resourceNameOld
		resourceInfo := appsv1alpha1.ResourceInfo{
			Resource: corev1.ObjectReference{
				Kind: kindConfigMap, APIVersion: apiVersionV1, Namespace: namespaceDefault, Name: cmName,
			},
			FullResource: fullResourceFor(namespaceDefault, cmName, map[string]string{"k": "v"}),
		}

		c := fake.NewClientBuilder().WithScheme(newTestScheme()).
			WithObjects(newTestReportWithAction("cleaner-a", appsv1alpha1.ActionDelete,
				[]appsv1alpha1.ResourceInfo{resourceInfo})).
			Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/reports/cleaner-a/rollback", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))

		var results []map[string]any
		Expect(json.NewDecoder(w.Body).Decode(&results)).To(Succeed())
		Expect(results).To(HaveLen(1))
		Expect(results[0]["success"]).To(Equal(true))

		restored := &corev1.ConfigMap{}
		Expect(c.Get(context.TODO(),
			types.NamespacedName{Namespace: namespaceDefault, Name: cmName}, restored)).To(Succeed())
		Expect(restored.Data).To(Equal(map[string]string{"k": "v"}))
	})

	It("returns 404 when no report exists for the cleaner", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/reports/nonexistent/rollback", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusNotFound))
	})

	It("returns 400 when the last execution's action was Scan", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).
			WithObjects(newTestReport("cleaner-a", nil)).
			Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/reports/cleaner-a/rollback", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})

	It("is blocked in read-only mode", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).
			WithObjects(newTestReport("cleaner-a", nil)).
			Build()
		handler := testHandler(c, true)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/reports/cleaner-a/rollback", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusForbidden))
	})
})
