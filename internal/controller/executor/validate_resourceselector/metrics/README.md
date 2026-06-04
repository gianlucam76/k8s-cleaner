# Metric-based resource selection

These examples show how to use `metricSource` and `metricQueries` on a
`ResourceSelector` to gate resource matching on live Prometheus data.

Before each resource is evaluated, k8s-cleaner queries the configured
Prometheus-compatible endpoint and populates a global `metrics` table in
the Lua script. The script can then combine resource properties with metric
values to decide whether a resource matches.

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
      sum(rate(http_requests_errors_total{namespace="my-app"}[5m]))
      /
      sum(rate(http_requests_total{namespace="my-app"}[5m]))
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

> **Note:** These examples require a live Prometheus endpoint and cannot be
> validated with `make test` alone. The `cleaner.yaml` files are illustrative
> — deploy them against a cluster that has a compatible metrics endpoint.
