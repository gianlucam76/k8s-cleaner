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
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Dashboard is the embedded web dashboard server.
// Implements manager.Runnable for integration with controller-runtime.
type Dashboard struct {
	listenPort int
	readOnly   bool
	version    string
	k8sClient  client.Client
	logger     logr.Logger
}

// NewDashboard creates a dashboard server instance.
func NewDashboard(port int, readOnly bool, version string, c client.Client, log logr.Logger) *Dashboard {
	return &Dashboard{
		listenPort: port,
		readOnly:   readOnly,
		version:    version,
		k8sClient:  c,
		logger:     log,
	}
}

// Start runs the HTTP server until the context is canceled.
func (d *Dashboard) Start(ctx context.Context) error {
	mux := setupRoutes(d.k8sClient, d.readOnly, d.version, d.logger)
	handler := applyMiddleware(mux, d.readOnly, d.logger)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", d.listenPort),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	const shutdownTimeout = 10 * time.Second

	go func() { //nolint:gosec // background ctx is intentional for shutdown after parent ctx is canceled
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			d.logger.Error(err, "dashboard shutdown error")
		}
	}()

	d.logger.Info("starting dashboard", "port", d.listenPort, "readOnly", d.readOnly)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("dashboard server failed: %w", err)
	}
	return nil
}
