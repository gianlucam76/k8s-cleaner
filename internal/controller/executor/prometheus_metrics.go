/*
Copyright 2023. projectsveltos.io. All rights reserved.

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

package executor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"

	logs "github.com/projectsveltos/libsveltos/lib/logsettings"
)

const (
	defaultPrometheusPath = "api/v1/query"
	metricsHTTPTimeout    = 30 * time.Second
)

type prometheusResponse struct {
	Status string         `json:"status"`
	Error  string         `json:"error,omitempty"`
	Data   prometheusData `json:"data"`
}

type prometheusData struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result"`
}

// fetchMetrics queries the MetricSource defined on the ResourceSelector and
// returns a map from query name to scalar float value. Returns nil when no
// MetricSource or MetricQueries are configured.
func fetchMetrics(ctx context.Context, sr *appsv1alpha1.ResourceSelector, logger logr.Logger) (map[string]float64, error) {
	if sr.MetricSource == nil || len(sr.MetricQueries) == 0 {
		return nil, nil
	}

	source := sr.MetricSource

	authHeader, err := buildAuthHeader(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("reading metric credentials: %w", err)
	}

	path := source.Path
	if path == "" {
		path = defaultPrometheusPath
	}
	path = strings.TrimLeft(path, "/")

	endpoint := strings.TrimRight(source.URL, "/") + "/" + path

	httpClient := &http.Client{Timeout: metricsHTTPTimeout}

	results := make(map[string]float64, len(sr.MetricQueries))
	for i := range sr.MetricQueries {
		q := &sr.MetricQueries[i]
		logger.V(logs.LogDebug).Info("querying metric", "name", q.Name, "query", q.Query)
		value, err := queryMetricEndpoint(ctx, httpClient, endpoint, q.Query, authHeader)
		if err != nil {
			return nil, fmt.Errorf("metric query %q: %w", q.Name, err)
		}
		results[q.Name] = value
	}
	return results, nil
}

func queryMetricEndpoint(ctx context.Context, httpClient *http.Client,
	endpoint, query, authHeader string) (float64, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return 0, fmt.Errorf("building request: %w", err)
	}

	params := req.URL.Query()
	params.Set("query", query)
	req.URL.RawQuery = params.Encode()

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("querying Prometheus: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("metric source returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var promResp prometheusResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return 0, fmt.Errorf("decoding Prometheus response: %w", err)
	}
	if promResp.Status != "success" {
		return 0, fmt.Errorf("metric query failed: %s", promResp.Error)
	}

	return extractScalar(promResp.Data)
}

// extractScalar pulls a single float64 from a Prometheus instant-query result.
// Accepts scalar and single-element vector result types.
func extractScalar(data prometheusData) (float64, error) {
	switch data.ResultType {
	case "scalar":
		var pair [2]json.RawMessage
		if err := json.Unmarshal(data.Result, &pair); err != nil {
			return 0, fmt.Errorf("parsing scalar result: %w", err)
		}
		return parsePromValue(pair[1])
	case "vector":
		var results []struct {
			Value [2]json.RawMessage `json:"value"`
		}
		if err := json.Unmarshal(data.Result, &results); err != nil {
			return 0, fmt.Errorf("parsing vector result: %w", err)
		}
		if len(results) == 0 {
			return 0, fmt.Errorf("metric query returned no data")
		}
		if len(results) > 1 {
			return 0, fmt.Errorf("metric query returned %d series; use an aggregation (e.g. sum()) to produce a single value",
				len(results))
		}
		return parsePromValue(results[0].Value[1])
	default:
		return 0, fmt.Errorf("unsupported Prometheus result type %q; use an instant query returning scalar or vector",
			data.ResultType)
	}
}

func parsePromValue(raw json.RawMessage) (float64, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return 0, fmt.Errorf("parsing value token: %w", err)
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("converting %q to float64: %w", s, err)
	}
	return v, nil
}

// buildAuthHeader fetches credentials from the Secret referenced by source.SecretRef
// and returns the appropriate Authorization header value, or empty string when
// no SecretRef is set.
func buildAuthHeader(ctx context.Context, source *appsv1alpha1.MetricSource) (string, error) {
	if source.SecretRef == nil {
		return "", nil
	}

	secret := &corev1.Secret{}
	if err := k8sClient.Get(ctx, types.NamespacedName{
		Namespace: source.SecretRef.Namespace,
		Name:      source.SecretRef.Name,
	}, secret); err != nil {
		return "", fmt.Errorf("getting Secret %s/%s: %w",
			source.SecretRef.Namespace, source.SecretRef.Name, err)
	}

	if token, ok := secret.Data["token"]; ok {
		return "Bearer " + string(token), nil
	}

	username, hasUser := secret.Data["username"]
	password, hasPass := secret.Data["password"]
	if hasUser && hasPass {
		encoded := base64.StdEncoding.EncodeToString(
			[]byte(string(username) + ":" + string(password)))
		return "Basic " + encoded, nil
	}

	return "", fmt.Errorf("secret %s/%s must contain key \"token\" or keys \"username\" and \"password\"",
		source.SecretRef.Namespace, source.SecretRef.Name)
}
