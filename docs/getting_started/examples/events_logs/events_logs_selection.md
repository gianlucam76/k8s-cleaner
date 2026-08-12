---
title: k8s-cleaner - Event and log-based resource selection
description: Use recent Kubernetes Events and container log tails to gate resource selection in k8s-cleaner
tags:
    - Kubernetes
    - Controller
    - Kubernetes Resources
    - Events
    - Logs
    - Identify
    - Remove
authors:
    - Gianluca Mardente
---

## Introduction

Besides the resource's own spec/status (`obj`) and Prometheus metrics
(`metrics`), k8s-cleaner can fetch two more live signals for each candidate
resource before it is evaluated: the Kubernetes Events involving that
resource, and — for Pods — the tail of its container logs. Both are opt-in
and both are scoped to the `evaluate` script only (not `transform`, not the
`ResourcePolicySet`-level aggregated selection function).

Set `includeEvents: true` to fetch recent Events for each candidate resource,
exposed as the global `events` table. Set `logSource` (only meaningful when
the `ResourceSelector` targets Pods) to fetch container log tails, exposed as
the `logs`/`logsByContainer` globals (and `logsPrevious`/
`logsPreviousByContainer` when `logSource.previous` is set).

```yaml
resourceSelectors:
- kind: Pod
  group: ""
  version: v1
  includeEvents: true
  logSource:
    tailLines: 100
  evaluate: |
    function evaluate(obj)
      hs = {}
      hs.matching = false
      if logs ~= nil and string.find(logs, "OutOfMemoryError") ~= nil then
        hs.matching = true
      end
      return hs
    end
```

Both fetches happen once per *candidate* resource — after the
`kind`/`group`/`version`/`namespace`/`labelFilters` selection has already
narrowed down the list, not once per scan. A fetch failure for one resource
(e.g. a container that has since been removed) is logged and treated as "no
data available" for that resource; it does not abort evaluation of the other
candidates in the same `ResourceSelector`.

---

## The `events` Global

`events` is an array, most recent first, of:

| Field | Description |
|---|---|
| `reason` | Machine-readable reason, e.g. `FailedMount`, `BackOff` |
| `message` | Human-readable event message |
| `type` | `Normal` or `Warning` |
| `count` | Number of occurrences the apiserver has deduplicated into this Event |
| `lastTimestamp` | RFC 3339 timestamp of the most recent occurrence |

This includes Events raised through the newer `events.k8s.io/v1` API — the
apiserver converts between the two representations, so nothing raised by a
modern controller (or by k8s-cleaner itself) is missed.

## The `logs` / `logsByContainer` Globals

`logs` concatenates every selected container's tail into one string — headered
with `==> <container> <==` only when the Pod has more than one selected
container, so `string.find(logs, "pattern")` works whether the Pod has one
container or five, and the script never has to name a container. When you do
need to know which container matched, `logsByContainer["<name>"]` gives the
same content split per container.

`logSource.containers` restricts which containers are fetched; leaving it
empty (the default) fetches every container in the Pod. `logSource.tailLines`
defaults to 50. Setting `logSource.previous: true` additionally populates
`logsPrevious`/`logsPreviousByContainer` for every selected container that has
actually restarted (`RestartCount > 0`) — useful for inspecting why a
container in `CrashLoopBackOff` last exited, since its current instance's log
may be empty or unrelated to the crash.

All four globals are always defined — empty string / empty table when
`logSource` is unset, or when the `ResourceSelector`'s `kind` is not `Pod` —
so a script never needs a `nil` check before using them.

---

## Example — Delete Pods With Repeated FailedMount Events

The following Cleaner deletes Pods that have accumulated more than 5
`FailedMount` Events — a Pod stuck unable to mount a volume is unlikely to
recover on its own.

!!! example ""

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: delete-pods-with-repeated-failedmount
    spec:
      schedule: "*/10 * * * *"
      action: Delete
      resourcePolicySet:
        resourceSelectors:
        - kind: Pod
          group: ""
          version: v1
          includeEvents: true
          evaluate: |
            function evaluate(obj)
              hs = {}
              hs.matching = false
              for _, e in ipairs(events) do
                if e.reason == "FailedMount" and e.count > 5 then
                  hs.matching = true
                  hs.message = "repeated FailedMount events: " .. e.count
                  break
                end
              end
              return hs
            end
    ```

---

## Example — Delete Crash-Looping Pods Whose Previous Log Shows an OOM

The following Cleaner deletes Pods whose `app` container has restarted and
whose *previous* instance's log (the one that actually crashed) mentions an
out-of-memory error — the current instance's log, right after a fresh
restart, would not show it yet.

!!! example ""

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: delete-oom-crashlooping-pods
    spec:
      schedule: "*/5 * * * *"
      action: Delete
      resourcePolicySet:
        resourceSelectors:
        - kind: Pod
          group: ""
          version: v1
          logSource:
            containers: ["app"]
            tailLines: 100
            previous: true
          evaluate: |
            function evaluate(obj)
              hs = {}
              hs.matching = logsPrevious ~= nil and
                string.find(logsPrevious, "OutOfMemoryError") ~= nil
              return hs
            end
    ```

---

## RBAC

Reading Events (`get`/`list`/`watch` on `events`) and container logs
(`get` on `pods/log`) is already granted to k8s-cleaner's default
`ClusterRole` — no extra permissions are needed to use `includeEvents` or
`logSource`.
