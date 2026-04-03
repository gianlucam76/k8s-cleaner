// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package web

import (
	"net/http"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"gianlucam76/k8s-cleaner/internal/controller/executor"
)

// triggerResponse is returned by trigger endpoints.
type triggerResponse struct {
	Message   string `json:"message"`
	Cleaner   string `json:"cleaner,omitempty"`
	Triggered int    `json:"triggered,omitempty"`
}

// TriggerHandler triggers an on-demand scan for a single cleaner.
// It calls executor.GetClient().Process() directly (no annotation patching).
func TriggerHandler(c client.Client, log logr.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		name := r.PathValue("name")

		// Verify the cleaner exists
		var cleaner appsv1alpha1.Cleaner
		if err := c.Get(ctx, client.ObjectKey{Name: name}, &cleaner); err != nil {
			if client.IgnoreNotFound(err) == nil {
				respondError(w, http.StatusNotFound, "cleaner not found")
				return
			}
			log.Error(err, "failed to get cleaner", "name", name)
			respondError(w, http.StatusInternalServerError, "failed to get cleaner")
			return
		}

		executorClient := executor.GetClient()
		if executorClient == nil {
			respondError(w, http.StatusServiceUnavailable, "executor not initialized")
			return
		}

		executorClient.Process(ctx, name)
		log.Info("triggered scan", "cleaner", name)

		respondJSON(w, http.StatusAccepted, triggerResponse{
			Message: "scan triggered",
			Cleaner: name,
		})
	}
}

// TriggerAllHandler triggers on-demand scans for all cleaners.
func TriggerAllHandler(c client.Client, log logr.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		cleaners, err := listCleaners(ctx, c)
		if err != nil {
			log.Error(err, "failed to list cleaners")
			respondError(w, http.StatusInternalServerError, "failed to list cleaners")
			return
		}

		executorClient := executor.GetClient()
		if executorClient == nil {
			respondError(w, http.StatusServiceUnavailable, "executor not initialized")
			return
		}

		triggered := 0
		for i := range cleaners {
			executorClient.Process(ctx, cleaners[i].Name)
			triggered++
		}

		log.Info("triggered all scans", "count", triggered)

		respondJSON(w, http.StatusAccepted, triggerResponse{
			Message:   "all scans triggered",
			Triggered: triggered,
		})
	}
}
