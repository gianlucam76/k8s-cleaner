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
	stdlog "log"
	"net/http"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	webfs "gianlucam76/k8s-cleaner/web"
)

// setupRoutes registers all API and SPA routes on a new ServeMux.
func setupRoutes(c client.Client, readOnly bool, version string, log logr.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("GET /api/v1/summary", SummaryHandler(c, log))
	mux.HandleFunc("GET /api/v1/cleaners", ListCleanersHandler(c, log))
	mux.HandleFunc("GET /api/v1/cleaners/{name}", GetCleanerHandler(c, log))
	mux.HandleFunc("GET /api/v1/reports", ListReportsHandler(c, log))
	mux.HandleFunc("GET /api/v1/reports/{name}", GetReportHandler(c, log))
	mux.HandleFunc("POST /api/v1/cleaners/{name}/trigger", TriggerHandler(c, log))
	mux.HandleFunc("POST /api/v1/trigger-all", TriggerAllHandler(c, log))
	mux.HandleFunc("GET /api/v1/config", ConfigHandler(readOnly, version))
	mux.HandleFunc("GET /api/v1/health", HealthHandler())

	// SPA static files
	mux.Handle("/", SPAHandler(webfs.StaticFS()))

	return mux
}

// ErrorResponse is the standard JSON error envelope.
type ErrorResponse struct {
	Error string `json:"error"`
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		stdlog.Printf("respondJSON encode error: %v", err)
	}
}

// respondError writes a JSON error response.
func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, ErrorResponse{Error: msg})
}
