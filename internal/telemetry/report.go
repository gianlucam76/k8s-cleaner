/*
Copyright 2024. projectsveltos.io. All rights reserved.

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

package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"

	logs "github.com/projectsveltos/libsveltos/lib/logsettings"
)

type Cluster struct {
	UUID             string `json:"uuid"`
	CleanerVersion   string `json:"cleanerVersion"`
	Nodes            int    `json:"nodes"`
	CleanerInstances int    `json:"cleanerInstances"`
}

type instance struct {
	version string
	client.Client
}

var (
	telemetryInstance *instance
	lock              = &sync.Mutex{}
)

const (
	contentTypeJSON = "application/json"
	domain          = "http://cleaner-telemetry.projectsveltos.io/"
)

func StartCollecting(ctx context.Context, c client.Client, sveltosVersion string) error {
	if telemetryInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if telemetryInstance == nil {
			telemetryInstance = &instance{
				Client:  c,
				version: sveltosVersion,
			}

			go telemetryInstance.reportData(ctx)
		}
	}

	return nil
}

// Collects telemetry data and send to to Sveltos telemetry server
func (m *instance) reportData(ctx context.Context) {
	// Data are collected twice times a day
	const twelve = 12
	ticker := time.NewTicker(twelve * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			telemetryInstance.collectAndSendData(ctx)
		}
	}
}

func (m *instance) retrieveUUID(ctx context.Context) (string, error) {
	var defaultNS corev1.Namespace
	if err := m.Get(ctx, types.NamespacedName{Name: "default"}, &defaultNS); err != nil {
		return "", errors.Wrap(err, "cannot start the telemetry controller")
	}

	return string(defaultNS.UID), nil
}

func (m *instance) collectAndSendData(ctx context.Context) {
	logger := log.FromContext(ctx)
	logger.V(logs.LogInfo).Info("collecting telemetry data")

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	uuid, err := telemetryInstance.retrieveUUID(ctx)
	if err != nil {
		return
	}

	payload, err := m.collectData(ctx, uuid)
	if err != nil {
		return
	}

	m.sendData(ctx, payload)
}

func (m *instance) collectData(ctx context.Context, uuid string) (*Cluster, error) {
	logger := log.FromContext(ctx)

	data := Cluster{
		UUID:           uuid,
		CleanerVersion: m.version,
	}

	var cleaners appsv1alpha1.CleanerList
	if err := m.List(ctx, &cleaners); err != nil {
		logger.V(logs.LogInfo).Info(fmt.Sprintf("failed to collect cleaner instances: %v", err))
		return nil, err
	}
	data.CleanerInstances = len(cleaners.Items)

	var nodes corev1.NodeList
	if err := m.List(ctx, &nodes); err != nil {
		logger.V(logs.LogInfo).Info(fmt.Sprintf("failed to collect nodes: %v", err))
		return nil, err
	}

	data.Nodes = len(nodes.Items)

	return &data, nil
}

func (m *instance) sendData(ctx context.Context, payload *Cluster) {
	logger := log.FromContext(ctx)

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, domain, bytes.NewBuffer(data))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", contentTypeJSON)
	req.Header.Set("User-Agent", "projectsveltos/sveltos-telemetry")

	// Create an HTTP client with follow redirects enabled
	c := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Follow redirect and set body
			newReq, err := http.NewRequestWithContext(ctx, http.MethodGet, domain, bytes.NewBuffer(data))
			req.Body = newReq.Body
			return err
		},
	}
	// Send the request
	//nolint: gosec // Just sending telemetry data.
	resp, err := c.Do(req)
	if err != nil {
		logger.V(logs.LogInfo).Info(fmt.Sprintf("error sending data: %v", err))
		return
	}
	defer resp.Body.Close()
	logger.V(logs.LogInfo).Info(fmt.Sprintf("Response status code: %d", resp.StatusCode))
}
