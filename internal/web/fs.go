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
