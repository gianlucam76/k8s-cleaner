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

var _ = Describe("Library", func() {
	It("should list every curated library entry with title, description and schedule", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/library", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))

		var entries []librarySummaryResponse
		Expect(json.NewDecoder(w.Body).Decode(&entries)).To(Succeed())
		Expect(entries).To(HaveLen(len(libraryManifest)))
		for i := range entries {
			Expect(entries[i].ID).ToNot(BeEmpty())
			Expect(entries[i].Group).ToNot(BeEmpty())
			Expect(entries[i].Title).ToNot(BeEmpty())
			Expect(entries[i].Description).ToNot(BeEmpty())
			Expect(entries[i].Schedule).ToNot(BeEmpty())
			Expect(entries[i].Selectors).ToNot(BeEmpty())
		}
	})

	It("should keep same-group entries contiguous within each category", func() {
		// The dashboard renders one labeled section per (category, group) pair
		// in manifest order, with no re-sorting - so a group must never be
		// interrupted by a different group within the same category.
		type categoryGroup struct{ category, group string }
		seen := make(map[categoryGroup]bool)
		var lastKey categoryGroup

		for i := range libraryManifest {
			entry := &libraryManifest[i]
			key := categoryGroup{entry.Category, entry.Group}
			if key == lastKey {
				continue
			}
			Expect(seen[key]).To(BeFalse(),
				"group %q in category %q is split across non-contiguous entries", entry.Group, entry.Category)
			seen[key] = true
			lastKey = key
		}
	})

	It("should have a unique ID for every manifest entry and load without error", func() {
		seen := make(map[string]bool, len(libraryManifest))
		for i := range libraryManifest {
			entry := &libraryManifest[i]
			Expect(seen[entry.ID]).To(BeFalse(), "duplicate library ID: %s", entry.ID)
			seen[entry.ID] = true

			cleaner, err := loadLibraryCleaner(entry)
			Expect(err).ToNot(HaveOccurred(), "entry %s failed to load", entry.ID)
			Expect(cleaner.Spec.Action).To(Equal(appsv1alpha1.ActionScan),
				"curated entry %s must be Action: Scan in the source example", entry.ID)
		}
	})

	It("should return a library entry's detail including its Lua script and selectors", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/library/"+unusedClusterRolesRecipeID, http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))

		var resp libraryDetailResponse
		Expect(json.NewDecoder(w.Body).Decode(&resp)).To(Succeed())
		Expect(resp.ID).To(Equal(unusedClusterRolesRecipeID))
		Expect(resp.LuaScript).ToNot(BeEmpty())
		Expect(resp.Selectors).ToNot(BeEmpty())
	})

	It("should return 404 for an unknown library entry", func() {
		c := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		handler := testHandler(c, false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/library/nonexistent", http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusNotFound))
	})
})
