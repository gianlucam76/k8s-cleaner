---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Schedule
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

## Introduction to Schedule

The schedule field specifies when the k8s-cleaner should run its logic to identify and potentially delete or update matching resources.

It adheres to the `Cron syntax`, which is a widely adopted scheduling language for tasks and events.

The Cron syntax consists of five fields, separated by spaces, each representing a specific part of the scheduling period: minute, hour, day of month, month and day of week, in that order.

The k8s-cleaner is able to accept the below schedule formats.

- **Standard crontab** specs, e.g. "* * * * ?"
- **Descriptors**, e.g. "@midnight", "@every 1h30m"

## Example

!!! example ""

    ```yaml
    ---
    # This Cleaner instance finds any Jobs that:
    # - have status.completionTime set
    # - have status.succeeded set to a value greater than zero
    # - have no running or pending pods
    # Cleaner will delete the resources every 1h30m.
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: completed-jobs
    spec:
      schedule: "@every 1h30m"
      resourcePolicySet:
        resourceSelectors:
        - kind: Job
            group: "batch"
            version: v1
            evaluate: |
              function evaluate()
                hs = {}
                hs.matching = false
                if obj.status ~= nil then
                  if obj.status.completionTime ~= nil and obj.status.succeeded > 0 then
                    hs.matching = true
                  end
                end
                return hs
              end
      action: Delete    
    ```

### Validation

```bash
$ kubectl apply -f "cleaner.yaml" 
cleaner.apps.projectsveltos.io/completed-jobs created

$ kubectl get cleaner -n projectsveltos
NAME             AGE
completed-jobs   7s

$ kubectl get cleaner completed-jobs -n projectsveltos -o yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"apps.projectsveltos.io/v1alpha1","kind":"Cleaner","metadata":{"annotations":{},"name":"completed-jobs"},"spec":{"action":"Delete","resourcePolicySet":{"resourceSelectors":[{"evaluate":"function evaluate()\n  hs = {}\n  hs.matching = false\n  if obj.status ~= nil then\n    if obj.status.completionTime ~= nil and obj.status.succeeded \u003e 0 then\n      hs.matching = true\n    end\n  end\n  return hs\nend\n","group":"batch","kind":"Job","version":"v1"}]},"schedule":"@every 1h30m"}}
  creationTimestamp: "2024-08-03T14:09:49Z"
  finalizers:
  - projectsveltos.io/cleaner-finalizer
  generation: 1
  name: completed-jobs
  resourceVersion: "11334709"
  uid: 205c20e1-cb3a-43ee-96f9-9ccb2b2cdf40
spec:
  action: Delete
  resourcePolicySet:
    resourceSelectors:
    - evaluate: |
        function evaluate()
          hs = {}
          hs.matching = false
          if obj.status ~= nil then
            if obj.status.completionTime ~= nil and obj.status.succeeded > 0 then
              hs.matching = true
            end
          end
          return hs
        end
      group: batch
      kind: Job
      version: v1
  schedule: '@every 1h30m'
status:
  nextScheduleTime: "2024-08-03T15:39:49Z"
```

From the above example, we can observe the next schedule to be **nextScheduleTime: "2024-08-03T15:39:49Z"**.