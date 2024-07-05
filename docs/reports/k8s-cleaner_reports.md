---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Welcome to the k8s-cleaner reports page
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

## Introduction to k8s-cleaner Reports

Users have the ability to instruct the k8s-cleaner to generate a report with all the resources deleted or modified.

The k8s-cleaner will create a Report instance based on the name of the Report name.

## Example - Report Defintion

!!! example "Cleaner Definition with Notifications set to type CleanerReport"

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
    name: cleaner-with-report
    spec:
    schedule: "0 * * * *"
    action: Delete # Delete matching resources
    resourcePolicySet:
        resourceSelectors:
        - namespace: test
        kind: Deployment
        group: "apps"
        version: v1
    notifications:
    - name: report
        type: CleanerReport
    ```

### Validation

```bash
$ kubectl get report           
NAME              AGE
cleaner-sample3   51m
```

### Report Output

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Report
metadata:
  creationTimestamp: "2023-12-17T17:05:00Z"
  generation: 2
  name: cleaner-with-report
  resourceVersion: "1625"
  uid: dda9a231-9a51-4133-aeb5-f0520feb8746
spec:
  action: Delete
  message: 'time: 2023-12-17 17:07:00.394736089 +0000 UTC m=+129.172023518'
  resources:
  - apiVersion: apps/v1
    kind: Deployment
    name: my-nginx-deployment
    namespace: test
```