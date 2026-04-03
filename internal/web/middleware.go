// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package web

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

// applyMiddleware wraps a handler with the full middleware stack.
func applyMiddleware(h http.Handler, readOnly bool, log logr.Logger) http.Handler {
	h = withMaxBody(h, 1<<20)
	h = withCaching(h)
	h = withCompression(h)
	h = withSecurityHeaders(h)
	if readOnly {
		h = withReadOnlyGuard(h)
	}
	h = withRequestLog(log, h)
	return h
}

func withRequestLog(log logr.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		sw := &responseCapture{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.V(1).Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.statusCode,
			"latency", time.Since(t).String(),
		)
	})
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com")
		next.ServeHTTP(w, r)
	})
}

func withReadOnlyGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"read-only mode enabled"}`))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func withCaching(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}
		next.ServeHTTP(w, r)
	})
}

func withMaxBody(next http.Handler, limit int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
		}
		next.ServeHTTP(w, r)
	})
}

func withCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		defer gz.Close()
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length")
		next.ServeHTTP(&compressedWriter{ResponseWriter: w, gzWriter: gz}, r)
	})
}

type responseCapture struct {
	http.ResponseWriter
	statusCode int
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.statusCode = code
	rc.ResponseWriter.WriteHeader(code)
}

type compressedWriter struct {
	http.ResponseWriter
	gzWriter io.Writer
}

func (cw *compressedWriter) Write(b []byte) (int, error) {
	return cw.gzWriter.Write(b)
}
