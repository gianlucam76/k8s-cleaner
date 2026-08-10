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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

const (
	apiCleanersPath     = "/api/v1/cleaners"
	testCreateName      = "my-unused-clusterroles"
	testCreateSchedule  = "0 3 * * *"
	testCustomNotifName = "custom-notif"
)

func postJSON(handler http.Handler, body any) *httptest.ResponseRecorder {
	data, err := json.Marshal(body)
	Expect(err).ToNot(HaveOccurred())

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, apiCleanersPath, bytes.NewReader(data))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

var _ = Describe("CreateCleaner", func() {
	It("should create a Cleaner from a library recipe, Scan action, default CleanerReport notification", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		w := postJSON(handler, createCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Name:            testCreateName,
			Schedule:        testCreateSchedule,
		})

		Expect(w.Code).To(Equal(http.StatusCreated))

		var resp cleanerResponse
		Expect(json.NewDecoder(w.Body).Decode(&resp)).To(Succeed())
		Expect(resp.Name).To(Equal(testCreateName))
		Expect(resp.Schedule).To(Equal(testCreateSchedule))
		Expect(resp.Action).To(Equal(string(appsv1alpha1.ActionScan)))

		var created appsv1alpha1.Cleaner
		Expect(c.Get(context.Background(), client.ObjectKey{Name: testCreateName}, &created)).To(Succeed())
		Expect(created.Spec.Notifications).To(HaveLen(1))
		Expect(created.Spec.Notifications[0].Type).To(Equal(appsv1alpha1.NotificationTypeCleanerReport))
		Expect(created.Spec.ResourcePolicySet.ResourceSelectors).ToNot(BeEmpty())
	})

	It("should preserve caller-provided notifications instead of the default", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		w := postJSON(handler, createCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Name:            testCustomNotifName,
			Schedule:        testCreateSchedule,
			Notifications: []appsv1alpha1.Notification{
				{Name: "slack", Type: appsv1alpha1.NotificationTypeSlack},
			},
		})

		Expect(w.Code).To(Equal(http.StatusCreated))

		var created appsv1alpha1.Cleaner
		Expect(c.Get(context.Background(), client.ObjectKey{Name: testCustomNotifName}, &created)).To(Succeed())
		Expect(created.Spec.Notifications).To(HaveLen(1))
		Expect(created.Spec.Notifications[0].Type).To(Equal(appsv1alpha1.NotificationTypeSlack))
	})

	It("should allow overriding Action to Delete", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		w := postJSON(handler, createCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Name:            "delete-clusterroles",
			Schedule:        testCreateSchedule,
			Action:          string(appsv1alpha1.ActionDelete),
		})

		Expect(w.Code).To(Equal(http.StatusCreated))

		var resp cleanerResponse
		Expect(json.NewDecoder(w.Body).Decode(&resp)).To(Succeed())
		Expect(resp.Action).To(Equal(string(appsv1alpha1.ActionDelete)))
	})

	It("should reject an invalid action", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		w := postJSON(handler, createCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Name:            "bad-action",
			Schedule:        testCreateSchedule,
			Action:          string(appsv1alpha1.ActionTransform),
		})

		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})

	It("should reject a name that already exists", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).
			WithObjects(newTestCleaner("taken", "0 * * * *")).
			Build()
		handler := testHandler(c, false)

		w := postJSON(handler, createCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Name:            "taken",
			Schedule:        testCreateSchedule,
		})

		Expect(w.Code).To(Equal(http.StatusConflict))
	})

	It("should reject an unknown sourceLibraryId", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		w := postJSON(handler, createCleanerRequest{
			SourceLibraryID: unknownLibraryID,
			Name:            "whatever",
			Schedule:        testCreateSchedule,
		})

		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})

	It("should require a name", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		w := postJSON(handler, createCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Schedule:        testCreateSchedule,
		})

		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})

	It("should require a schedule", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		w := postJSON(handler, createCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Name:            "no-schedule",
		})

		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})

	It("should be blocked in read-only mode", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, true)

		w := postJSON(handler, createCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Name:            "blocked",
			Schedule:        testCreateSchedule,
		})

		Expect(w.Code).To(Equal(http.StatusForbidden))
	})
})
