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
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// createCleanerRequest is the body of POST /api/v1/cleaners. A Cleaner can
// only be created from a library recipe: the ResourceSelectors/Lua always
// come from the trusted, embedded library entry identified by
// SourceLibraryID, never from the request body.
type createCleanerRequest struct {
	SourceLibraryID string                      `json:"sourceLibraryId"`
	Name            string                      `json:"name"`
	Schedule        string                      `json:"schedule"`
	Action          string                      `json:"action"`
	Notifications   []appsv1alpha1.Notification `json:"notifications"`
}

// CreateCleanerHandler creates a new Cleaner from a library recipe. The caller may
// customize the name, schedule, action, and notifications; the ResourceSelectors/Lua
// always come from the library entry.
func CreateCleanerHandler(c client.Client, log logr.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req createCleanerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Name == "" {
			respondError(w, http.StatusBadRequest, "name is required")
			return
		}
		if req.Schedule == "" {
			respondError(w, http.StatusBadRequest, "schedule is required")
			return
		}
		action, err := resolveAction(req.Action)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		base, ok := loadLibraryEntryOrRespond(w, log, req.SourceLibraryID)
		if !ok {
			return
		}

		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{Name: req.Name},
			Spec:       buildCleanerSpec(base, req.Schedule, action, req.Notifications),
		}

		if err := c.Create(ctx, cleaner); err != nil {
			if apierrors.IsAlreadyExists(err) {
				respondError(w, http.StatusConflict, "a cleaner with this name already exists")
				return
			}
			if apierrors.IsInvalid(err) {
				respondError(w, http.StatusBadRequest, err.Error())
				return
			}
			log.Error(err, "failed to create cleaner", "name", req.Name)
			respondError(w, http.StatusInternalServerError, "failed to create cleaner")
			return
		}

		log.Info("created cleaner from library", "name", req.Name, "sourceLibraryId", req.SourceLibraryID)
		respondJSON(w, http.StatusCreated, toCleanerResponse(cleaner, nil, true))
	}
}

// updateCleanerRequest is the body of PUT /api/v1/cleaners/{name}. Like create, the
// ResourceSelectors/Lua always come from the trusted library entry, never the request
// body - an update fully resyncs the existing Cleaner's spec to match the library
// recipe plus the caller's schedule/notifications, discarding any manual edits made
// to it outside the dashboard (e.g. via kubectl).
type updateCleanerRequest struct {
	SourceLibraryID string                      `json:"sourceLibraryId"`
	Schedule        string                      `json:"schedule"`
	Action          string                      `json:"action"`
	Notifications   []appsv1alpha1.Notification `json:"notifications"`
}

// UpdateCleanerHandler resyncs an existing Cleaner's spec from a library recipe:
// ResourceSelectors and Lua come from the library entry; Schedule, Action, and
// Notifications come from the caller. Intended to resolve a name conflict from
// CreateCleanerHandler once the caller has reviewed what would change.
func UpdateCleanerHandler(c client.Client, log logr.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		name := r.PathValue("name")

		var req updateCleanerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Schedule == "" {
			respondError(w, http.StatusBadRequest, "schedule is required")
			return
		}
		action, err := resolveAction(req.Action)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		base, ok := loadLibraryEntryOrRespond(w, log, req.SourceLibraryID)
		if !ok {
			return
		}

		var existing appsv1alpha1.Cleaner
		if err := c.Get(ctx, client.ObjectKey{Name: name}, &existing); err != nil {
			if client.IgnoreNotFound(err) == nil {
				respondError(w, http.StatusNotFound, "cleaner not found")
				return
			}
			log.Error(err, "failed to get cleaner", "name", name)
			respondError(w, http.StatusInternalServerError, "failed to get cleaner")
			return
		}

		existing.Spec = buildCleanerSpec(base, req.Schedule, action, req.Notifications)

		if err := c.Update(ctx, &existing); err != nil {
			if apierrors.IsConflict(err) {
				respondError(w, http.StatusConflict, "cleaner was modified concurrently, please retry")
				return
			}
			if apierrors.IsInvalid(err) {
				respondError(w, http.StatusBadRequest, err.Error())
				return
			}
			log.Error(err, "failed to update cleaner", "name", name)
			respondError(w, http.StatusInternalServerError, "failed to update cleaner")
			return
		}

		log.Info("updated cleaner from library", "name", name, "sourceLibraryId", req.SourceLibraryID)
		respondJSON(w, http.StatusOK, toCleanerResponse(&existing, nil, true))
	}
}

// loadLibraryEntryOrRespond resolves a library entry by ID and loads its Cleaner,
// writing the appropriate error response and returning ok=false on failure.
func loadLibraryEntryOrRespond(w http.ResponseWriter, log logr.Logger, sourceLibraryID string,
) (base *appsv1alpha1.Cleaner, ok bool) {

	entry := findLibraryEntry(sourceLibraryID)
	if entry == nil {
		respondError(w, http.StatusBadRequest, "unknown sourceLibraryId")
		return nil, false
	}

	base, err := loadLibraryCleaner(entry)
	if err != nil {
		log.Error(err, "failed to load library recipe", "id", sourceLibraryID)
		respondError(w, http.StatusInternalServerError, "failed to load library recipe")
		return nil, false
	}
	return base, true
}

const (
	// defaultNotificationName is the name given to the CleanerReport notification
	// injected when a create/update request doesn't specify any notifications.
	defaultNotificationName = "report"
)

// resolveAction validates a caller-supplied action string, defaulting to Scan when
// empty. Transform is deliberately not allowed: no library recipe ships a transform
// script, so a Cleaner left as Transform would just error at run time.
func resolveAction(raw string) (appsv1alpha1.Action, error) {
	if raw == "" {
		return appsv1alpha1.ActionScan, nil
	}
	switch appsv1alpha1.Action(raw) {
	case appsv1alpha1.ActionScan, appsv1alpha1.ActionDelete:
		return appsv1alpha1.Action(raw), nil
	case appsv1alpha1.ActionTransform:
		// explicitly listed (not just caught by the fallback below) so this
		// switch stays exhaustive over the Action enum
	}
	return "", fmt.Errorf("action must be %q or %q", appsv1alpha1.ActionScan, appsv1alpha1.ActionDelete)
}

// buildCleanerSpec assembles a CleanerSpec from a library recipe plus caller-supplied
// schedule/action/notifications. ResourceSelectors/Lua always come from base.
func buildCleanerSpec(base *appsv1alpha1.Cleaner, schedule string, action appsv1alpha1.Action,
	notifications []appsv1alpha1.Notification) appsv1alpha1.CleanerSpec {

	if len(notifications) == 0 {
		notifications = []appsv1alpha1.Notification{
			{Name: defaultNotificationName, Type: appsv1alpha1.NotificationTypeCleanerReport},
		}
	}

	return appsv1alpha1.CleanerSpec{
		ResourcePolicySet: base.Spec.ResourcePolicySet,
		Action:            action,
		Schedule:          schedule,
		Notifications:     notifications,
	}
}
