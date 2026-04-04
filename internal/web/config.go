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
)

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
