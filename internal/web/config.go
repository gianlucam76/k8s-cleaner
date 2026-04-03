// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package web

import "net/http"

// configResponse is returned by GET /api/v1/config.
type configResponse struct {
	ReadOnly bool   `json:"readOnly"`
	Version  string `json:"version"`
}

// ConfigHandler returns the web server configuration.
func ConfigHandler(readOnly bool, version string) http.HandlerFunc {
	resp := configResponse{ReadOnly: readOnly, Version: version}
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, resp)
	}
}

// HealthHandler returns a simple health check.
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
