// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package web

import (
	"context"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// dashboardSummary is the response for GET /api/v1/summary.
type dashboardSummary struct {
	TotalCleaners        int        `json:"totalCleaners"`
	TotalFlaggedResources int       `json:"totalFlaggedResources"`
	CleanersWithFindings int        `json:"cleanersWithFindings"`
	LastScanTime         *time.Time `json:"lastScanTime"`
}

// SummaryHandler returns a dashboard overview of all cleaners and reports.
func SummaryHandler(c client.Client, log logr.Logger) http.HandlerFunc {
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

		// Build report lookup by name
		reportMap := make(map[string]*appsv1alpha1.Report, len(reports))
		for i := range reports {
			reportMap[reports[i].Name] = &reports[i]
		}

		summary := dashboardSummary{}
		summary.TotalCleaners = len(cleaners)

		var lastScan time.Time
		for i := range cleaners {
			if cleaners[i].Status.LastRunTime != nil {
				t := cleaners[i].Status.LastRunTime.Time
				if t.After(lastScan) {
					lastScan = t
				}
			}
			if report, ok := reportMap[cleaners[i].Name]; ok {
				count := len(report.Spec.ResourceInfo)
				summary.TotalFlaggedResources += count
				if count > 0 {
					summary.CleanersWithFindings++
				}
			}
		}

		if !lastScan.IsZero() {
			summary.LastScanTime = &lastScan
		}

		respondJSON(w, http.StatusOK, summary)
	}
}

// listCleaners returns all Cleaner CRs in the cluster.
func listCleaners(ctx context.Context, c client.Client) ([]appsv1alpha1.Cleaner, error) {
	var list appsv1alpha1.CleanerList
	if err := c.List(ctx, &list); err != nil {
		return nil, err
	}
	return list.Items, nil
}

// listReports returns all Report CRs in the cluster.
func listReports(ctx context.Context, c client.Client) ([]appsv1alpha1.Report, error) {
	var list appsv1alpha1.ReportList
	if err := c.List(ctx, &list); err != nil {
		return nil, err
	}
	return list.Items, nil
}
