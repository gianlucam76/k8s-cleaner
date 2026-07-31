---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Blast Radius Limit
tags:
    - Kubernetes
    - Controller
    - Kubernetes Resources
    - Identify
    - Update
    - Remove
authors:
    - Eleni Grosdouli
---

## Introduction to Blast Radius Limit

`blastRadiusLimit` is an optional safety net for the `Delete` and `Transform` actions. It aborts a Cleaner run before any resource is deleted or updated if the number of matching resources looks too large, for example because of a typo in a label filter or a bug in an `evaluate`/`aggregatedSelection` Lua script.

Two, independently optional, thresholds are supported. If both are set, exceeding **either** one aborts the run.

- **maxCount**: aborts the run if more than this many resources match.
- **maxPercentage**: aborts the run if the matching resources are more than this percentage of all resources considered by the `resourceSelectors`, i.e. before label/Lua filtering narrows them down. For instance, if a selector considers 100 ConfigMaps in a namespace and 40 of them match, that is 40%.

`blastRadiusLimit` has no effect when `action` is set to `Scan`, since Scan never deletes or updates anything.

When the limit is exceeded:

- No resource is deleted or updated.
- Notifications and, if configured, [`storeResourcePath`](../store_resources/store_resource_yaml.md) still report the full set of matching resources, exactly as they would for a successful run, so the operator can inspect what would have happened.
- The abort reason is reported in the Cleaner's `status.failureMessage`, the same field used to surface any other run failure.

## Example - Cap the Number of Deleted Resources

The example below deletes unused ConfigMaps, but aborts if more than 5 would be deleted in a single run.

!!! example ""

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: unused-configmaps
    spec:
      schedule: "* 0 * * *" # Runs every day at midnight
      action: Delete
      blastRadiusLimit:
        maxCount: 5
      resourcePolicySet:
        resourceSelectors:
        - namespace: test
          kind: ConfigMap
          group: ""
          version: v1
          evaluate: |
            function evaluate()
              hs = {}
              hs.matching = true
              return hs
            end
    ```

## Example - Cap the Percentage of Deleted Resources

The example below deletes unhealthy Pods, but aborts if the matches are more than 30% of all Pods considered by the selector.

!!! example ""

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: unhealthy-pods
    spec:
      schedule: "*/5 * * * *" # Runs every 5 minutes
      action: Delete
      blastRadiusLimit:
        maxPercentage: 30
      resourcePolicySet:
        resourceSelectors:
        - namespace: test
          kind: Pod
          group: ""
          version: v1
          evaluate: |
            function evaluate()
              hs = {}
              hs.matching = false
              if obj.status ~= nil and obj.status.phase == "Failed" then
                hs.matching = true
              end
              return hs
            end
    ```

### Validation

If a run trips the limit, the Cleaner's status reports why, and no resource is touched:

```bash
$ kubectl get cleaner unused-configmaps -o yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: unused-configmaps
spec:
  action: Delete
  blastRadiusLimit:
    maxCount: 5
  ...
status:
  failureMessage: 'blast radius limit exceeded: 12 resources matched (max count 5)'
```
