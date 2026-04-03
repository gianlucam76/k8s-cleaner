// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package web

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// SPAHandler returns an http.Handler that serves the embedded SPA.
// Requests for existing files are served directly; all other paths
// receive index.html so the client-side router can handle them.
func SPAHandler(embedded fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := strings.TrimPrefix(r.URL.Path, "/")
		if reqPath == "" {
			reqPath = "index.html"
		}

		if strings.Contains(reqPath, "..") {
			http.NotFound(w, r)
			return
		}

		fullPath := path.Join("dist", reqPath)
		if _, err := fs.Stat(embedded, fullPath); err != nil {
			// File not found: serve index.html for SPA routing
			fullPath = path.Join("dist", "index.html")
		}

		http.ServeFileFS(w, r, embedded, fullPath)
	})
}
