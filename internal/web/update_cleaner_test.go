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
	testUpdateSchedule = "0 5 * * *"
	testEventNotifName = "alerts"
)

func putJSON(handler http.Handler, url string, body any) *httptest.ResponseRecorder {
	data, err := json.Marshal(body)
	Expect(err).ToNot(HaveOccurred())

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, url, bytes.NewReader(data))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

var _ = Describe("UpdateCleaner", func() {
	It("should resync selectors, action, schedule and notifications from the library", func() {
		existing := newTestCleaner("my-cleaner", "0 * * * *") // selectors=[ConfigMap], action=Scan by default
		existing.Spec.Action = appsv1alpha1.ActionDelete
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).WithObjects(existing).Build()
		handler := testHandler(c, false)

		w := putJSON(handler, "/api/v1/cleaners/my-cleaner", updateCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Schedule:        testUpdateSchedule,
			Notifications: []appsv1alpha1.Notification{
				{Name: testEventNotifName, Type: appsv1alpha1.NotificationTypeEvent},
			},
		})

		Expect(w.Code).To(Equal(http.StatusOK))

		var resp cleanerResponse
		Expect(json.NewDecoder(w.Body).Decode(&resp)).To(Succeed())
		Expect(resp.Schedule).To(Equal(testUpdateSchedule))
		Expect(resp.Action).To(Equal(string(appsv1alpha1.ActionScan)))
		Expect(resp.Notifications).To(Equal([]notificationInfo{
			{Name: testEventNotifName, Type: string(appsv1alpha1.NotificationTypeEvent)},
		}))

		var updated appsv1alpha1.Cleaner
		Expect(c.Get(context.Background(), client.ObjectKey{Name: "my-cleaner"}, &updated)).To(Succeed())
		Expect(updated.Spec.ResourcePolicySet.ResourceSelectors).To(HaveLen(3)) // ClusterRole/ClusterRoleBinding/RoleBinding
		kinds := make([]string, len(updated.Spec.ResourcePolicySet.ResourceSelectors))
		for i, rs := range updated.Spec.ResourcePolicySet.ResourceSelectors {
			kinds[i] = rs.Kind
		}
		Expect(kinds).To(ContainElement("ClusterRole"))
		Expect(kinds).ToNot(ContainElement(kindConfigMap))
	})

	It("should default to a CleanerReport notification when none is provided", func() {
		existing := newTestCleaner("my-cleaner", "0 * * * *")
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).WithObjects(existing).Build()
		handler := testHandler(c, false)

		w := putJSON(handler, "/api/v1/cleaners/my-cleaner", updateCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Schedule:        testUpdateSchedule,
		})

		Expect(w.Code).To(Equal(http.StatusOK))

		var resp cleanerResponse
		Expect(json.NewDecoder(w.Body).Decode(&resp)).To(Succeed())
		Expect(resp.Notifications).To(Equal([]notificationInfo{
			{Name: defaultNotificationName, Type: string(appsv1alpha1.NotificationTypeCleanerReport)},
		}))
	})

	It("should allow overriding Action to Delete", func() {
		existing := newTestCleaner("my-cleaner", "0 * * * *")
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).WithObjects(existing).Build()
		handler := testHandler(c, false)

		w := putJSON(handler, "/api/v1/cleaners/my-cleaner", updateCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Schedule:        testUpdateSchedule,
			Action:          string(appsv1alpha1.ActionDelete),
		})

		Expect(w.Code).To(Equal(http.StatusOK))

		var resp cleanerResponse
		Expect(json.NewDecoder(w.Body).Decode(&resp)).To(Succeed())
		Expect(resp.Action).To(Equal(string(appsv1alpha1.ActionDelete)))
	})

	It("should reject an invalid action", func() {
		existing := newTestCleaner("my-cleaner", "0 * * * *")
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).WithObjects(existing).Build()
		handler := testHandler(c, false)

		w := putJSON(handler, "/api/v1/cleaners/my-cleaner", updateCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Schedule:        testUpdateSchedule,
			Action:          string(appsv1alpha1.ActionTransform),
		})

		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})

	It("should return 404 for a nonexistent cleaner", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		w := putJSON(handler, "/api/v1/cleaners/nonexistent", updateCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Schedule:        testUpdateSchedule,
		})

		Expect(w.Code).To(Equal(http.StatusNotFound))
	})

	It("should require a schedule", func() {
		existing := newTestCleaner("my-cleaner", "0 * * * *")
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).WithObjects(existing).Build()
		handler := testHandler(c, false)

		w := putJSON(handler, "/api/v1/cleaners/my-cleaner", updateCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
		})

		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})

	It("should reject an unknown sourceLibraryId", func() {
		existing := newTestCleaner("my-cleaner", "0 * * * *")
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).WithObjects(existing).Build()
		handler := testHandler(c, false)

		w := putJSON(handler, "/api/v1/cleaners/my-cleaner", updateCleanerRequest{
			SourceLibraryID: unknownLibraryID,
			Schedule:        testUpdateSchedule,
		})

		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})

	It("should be blocked in read-only mode", func() {
		existing := newTestCleaner("my-cleaner", "0 * * * *")
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).WithObjects(existing).Build()
		handler := testHandler(c, true)

		w := putJSON(handler, "/api/v1/cleaners/my-cleaner", updateCleanerRequest{
			SourceLibraryID: unusedClusterRolesRecipeID,
			Schedule:        testUpdateSchedule,
		})

		Expect(w.Code).To(Equal(http.StatusForbidden))
	})
})
