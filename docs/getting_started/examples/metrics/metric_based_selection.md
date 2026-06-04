---
title: k8s-cleaner - Metric-based resource selection
description: Use Prometheus metrics to gate resource selection in k8s-cleaner
tags:
    - Kubernetes
    - Controller
    - Kubernetes Resources
    - Prometheus
    - Metrics
    - Identify
    - Update
    - Remove
authors:
    - Gianluca Mardente
---

## Introduction

k8s-cleaner can query a Prometheus-compatible endpoint before evaluating each
resource. The results are exposed to the Lua `evaluate` script through a global
`metrics` table, so you can combine resource properties with live metric values
to decide whether a resource should be acted on.

Configure `metricSource` with the URL of a Prometheus-compatible endpoint and
list the PromQL queries you need under `metricQueries`. Each query result is a
scalar float accessible as `metrics["<name>"]` inside the `evaluate` function.

```yaml
resourceSelectors:
- kind: Deployment
  group: apps
  version: v1
  metricSource:
    url: http://prometheus-server.monitoring.svc:9090
  metricQueries:
  - name: errorRate
    query: >-
      sum(rate(http_requests_errors_total[5m]))
      /
      sum(rate(http_requests_total[5m]))
  evaluate: |
    function evaluate()
      hs = {}
      hs.matching = false
      if metrics["errorRate"] ~= nil and metrics["errorRate"] > 0.05 then
        hs.matching = true
      end
      return hs
    end
```

The `metricSource` endpoint must be reachable from the cluster where
k8s-cleaner is running — a plain in-cluster Kubernetes Service URL works.

If a query returns no data (`metrics["name"]` is `nil`), the script should
handle that case explicitly to avoid Lua comparison errors on `nil`.

---

## Example — Scale Down Deployments With a High Error Rate

The following Cleaner scales Deployments in the `my-app` namespace to zero
replicas when the HTTP error rate reported by Prometheus exceeds 5%.

The `metricSource` points to the in-cluster Prometheus server. The
`metricQueries` entry runs a PromQL aggregation and exposes the result as
`metrics["errorRate"]`.

!!! example ""

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: scale-down-high-error-rate-deployments
    spec:
      schedule: "*/15 * * * *"
      action: Transform
      transform: |
        function transform()
          hs = {}
          obj.spec.replicas = 0
          hs.resource = obj
          return hs
        end
      resourcePolicySet:
        resourceSelectors:
        - kind: Deployment
          group: apps
          version: v1
          namespace: my-app
          metricSource:
            url: http://prometheus-server.monitoring.svc:9090
          metricQueries:
          - name: errorRate
            query: >-
              sum(rate(http_requests_errors_total{namespace="my-app"}[5m]))
              /
              sum(rate(http_requests_total{namespace="my-app"}[5m]))
          evaluate: |
            function evaluate()
              hs = {}
              hs.matching = false
              if metrics["errorRate"] == nil then
                return hs
              end
              if metrics["errorRate"] > 0.05 then
                hs.matching = true
                hs.message = "error rate above 5%: " .. tostring(metrics["errorRate"])
              end
              return hs
            end
    ```

---

## Example — Delete Pods Exceeding a Memory Threshold

The following Cleaner deletes Pods whose memory working set (reported by
Prometheus) exceeds 90% of their configured limit. Deleting the Pod lets the
controller recreate it with a clean memory state.

Two queries are used: `memUsed` (current working set in bytes) and `memLimit`
(configured limit in bytes). The ratio is computed inside the Lua script so the
PromQL expressions stay simple.

!!! example ""

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: restart-memory-saturated-pods
    spec:
      schedule: "*/10 * * * *"
      action: Delete
      resourcePolicySet:
        resourceSelectors:
        - kind: Pod
          group: ""
          version: v1
          namespace: my-app
          metricSource:
            url: http://prometheus-server.monitoring.svc:9090
          metricQueries:
          - name: memUsed
            query: >-
              sum(container_memory_working_set_bytes{namespace="my-app",container!=""})
          - name: memLimit
            query: >-
              sum(kube_pod_container_resource_limits{namespace="my-app",resource="memory",container!=""})
          evaluate: |
            function evaluate()
              hs = {}
              hs.matching = false
              if metrics["memUsed"] == nil or metrics["memLimit"] == nil then
                return hs
              end
              if metrics["memLimit"] == 0 then
                return hs
              end
              local ratio = metrics["memUsed"] / metrics["memLimit"]
              if ratio > 0.9 then
                hs.matching = true
                hs.message = string.format("memory usage at %.0f%% of limit", ratio * 100)
              end
              return hs
            end
    ```

---

## Authenticating Against a Protected Endpoint

If the Prometheus endpoint requires authentication, reference a Secret via
`metricSource.secretRef`. The Secret must contain either a `token` key (Bearer
auth) or both `username` and `password` keys (Basic auth).

!!! example ""

    ```yaml
    metricSource:
      url: https://prometheus.example.com
      secretRef:
        namespace: monitoring
        name: prometheus-credentials
    ```

    ```yaml
    # Bearer token
    apiVersion: v1
    kind: Secret
    metadata:
      name: prometheus-credentials
      namespace: monitoring
    stringData:
      token: "<bearer-token>"
    ```

    ```yaml
    # Basic auth
    apiVersion: v1
    kind: Secret
    metadata:
      name: prometheus-credentials
      namespace: monitoring
    stringData:
      username: "<user>"
      password: "<password>"
    ```
