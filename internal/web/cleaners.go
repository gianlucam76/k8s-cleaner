// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package web

import (
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// cleanerResponse is the JSON representation of a Cleaner for the API.
type cleanerResponse struct {
	Name             string         `json:"name"`
	Schedule         string         `json:"schedule"`
	Action           string         `json:"action"`
	LastRunTime      *time.Time     `json:"lastRunTime"`
	NextScheduleTime *time.Time     `json:"nextScheduleTime"`
	FailureMessage   string         `json:"failureMessage,omitempty"`
	FlaggedCount     int            `json:"flaggedCount"`
	Selectors        []selectorInfo `json:"selectors"`
	LuaScript        string         `json:"luaScript,omitempty"`
}

// selectorInfo is a summary of a ResourceSelector.
type selectorInfo struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

// ListCleanersHandler returns all cleaners with their report counts.
func ListCleanersHandler(c client.Client, log logr.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		cleaners, err := listCleaners(ctx, c)
		if err != nil {
			log.Error(err, "failed to list cleaners")
			respondError(w, http.StatusInternalServerError, "failed to list cleaners")
			return
		}

		reports, err := listReports(ctx, c)
		if err != nil {
			log.Error(err, "failed to list reports")
			respondError(w, http.StatusInternalServerError, "failed to list reports")
			return
		}

		reportMap := make(map[string]*appsv1alpha1.Report, len(reports))
		for i := range reports {
			reportMap[reports[i].Name] = &reports[i]
		}

		result := make([]cleanerResponse, 0, len(cleaners))
		for i := range cleaners {
			result = append(result, toCleanerResponse(&cleaners[i], reportMap[cleaners[i].Name], false))
		}

		respondJSON(w, http.StatusOK, result)
	}
}

// GetCleanerHandler returns a single cleaner with full details including Lua scripts.
func GetCleanerHandler(c client.Client, log logr.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		name := r.PathValue("name")

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

		// Look up associated report
		var report appsv1alpha1.Report
		var reportPtr *appsv1alpha1.Report
		if err := c.Get(ctx, client.ObjectKey{Name: name}, &report); err == nil {
			reportPtr = &report
		}

		resp := toCleanerResponse(&cleaner, reportPtr, true)
		respondJSON(w, http.StatusOK, resp)
	}
}

// toCleanerResponse converts a Cleaner CR to the API response format.
// When includeDetails is true, Lua scripts are included.
func toCleanerResponse(cleaner *appsv1alpha1.Cleaner, report *appsv1alpha1.Report, includeDetails bool) cleanerResponse {
	resp := cleanerResponse{
		Name:     cleaner.Name,
		Schedule: cleaner.Spec.Schedule,
		Action:   string(cleaner.Spec.Action),
	}

	if cleaner.Status.LastRunTime != nil {
		t := cleaner.Status.LastRunTime.Time
		resp.LastRunTime = &t
	}
	if cleaner.Status.NextScheduleTime != nil {
		t := cleaner.Status.NextScheduleTime.Time
		resp.NextScheduleTime = &t
	}
	if cleaner.Status.FailureMessage != nil {
		resp.FailureMessage = *cleaner.Status.FailureMessage
	}

	if report != nil {
		resp.FlaggedCount = len(report.Spec.ResourceInfo)
	}

	selectors := make([]selectorInfo, 0, len(cleaner.Spec.ResourcePolicySet.ResourceSelectors))
	for _, rs := range cleaner.Spec.ResourcePolicySet.ResourceSelectors {
		selectors = append(selectors, selectorInfo{
			Group:   rs.Group,
			Version: rs.Version,
			Kind:    rs.Kind,
		})
	}
	resp.Selectors = selectors

	if includeDetails {
		// Include the first non-empty Lua evaluate script
		for _, rs := range cleaner.Spec.ResourcePolicySet.ResourceSelectors {
			if rs.Evaluate != "" {
				resp.LuaScript = rs.Evaluate
				break
			}
		}
		// If there's an aggregated selection, prefer that
		if cleaner.Spec.ResourcePolicySet.AggregatedSelection != "" {
			resp.LuaScript = cleaner.Spec.ResourcePolicySet.AggregatedSelection
		}
	}

	return resp
}
