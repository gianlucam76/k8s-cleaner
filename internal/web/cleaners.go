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
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// cleanerResponse is the JSON representation of a Cleaner for the API.
type cleanerResponse struct {
	Name             string             `json:"name"`
	Schedule         string             `json:"schedule"`
	Action           string             `json:"action"`
	LastRunTime      *time.Time         `json:"lastRunTime"`
	NextScheduleTime *time.Time         `json:"nextScheduleTime"`
	FailureMessage   string             `json:"failureMessage,omitempty"`
	FlaggedCount     int                `json:"flaggedCount"`
	Selectors        []selectorInfo     `json:"selectors"`
	Notifications    []notificationInfo `json:"notifications"`
	LuaScript        string             `json:"luaScript,omitempty"`
}

// selectorInfo is a summary of a ResourceSelector.
type selectorInfo struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

// notificationInfo is a summary of a Notification.
type notificationInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
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

	resp.Selectors = selectorInfosFromCleaner(cleaner)
	resp.Notifications = notificationInfosFromCleaner(cleaner)

	if includeDetails {
		resp.LuaScript = primaryLuaScript(cleaner)
	}

	return resp
}

// selectorInfosFromCleaner summarizes a Cleaner's ResourceSelectors for API responses.
func selectorInfosFromCleaner(cleaner *appsv1alpha1.Cleaner) []selectorInfo {
	selectors := make([]selectorInfo, 0, len(cleaner.Spec.ResourcePolicySet.ResourceSelectors))
	for i := range cleaner.Spec.ResourcePolicySet.ResourceSelectors {
		rs := &cleaner.Spec.ResourcePolicySet.ResourceSelectors[i]
		selectors = append(selectors, selectorInfo{
			Group:   rs.Group,
			Version: rs.Version,
			Kind:    rs.Kind,
		})
	}
	return selectors
}

// notificationInfosFromCleaner summarizes a Cleaner's Notifications for API responses.
func notificationInfosFromCleaner(cleaner *appsv1alpha1.Cleaner) []notificationInfo {
	notifications := make([]notificationInfo, 0, len(cleaner.Spec.Notifications))
	for i := range cleaner.Spec.Notifications {
		n := &cleaner.Spec.Notifications[i]
		notifications = append(notifications, notificationInfo{
			Name: n.Name,
			Type: string(n.Type),
		})
	}
	return notifications
}

// primaryLuaScript returns the Lua script most relevant to show a user: the
// ResourcePolicySet's AggregatedSelection when set, otherwise the first
// non-empty per-selector Evaluate script.
func primaryLuaScript(cleaner *appsv1alpha1.Cleaner) string {
	if cleaner.Spec.ResourcePolicySet.AggregatedSelection != "" {
		return cleaner.Spec.ResourcePolicySet.AggregatedSelection
	}
	for i := range cleaner.Spec.ResourcePolicySet.ResourceSelectors {
		rs := &cleaner.Spec.ResourcePolicySet.ResourceSelectors[i]
		if rs.Evaluate != "" {
			return rs.Evaluate
		}
	}
	return ""
}
