---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Label Filters
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

## Introduction to Label Filters

The k8s-cleaner has the ability to select resources based on a label. This capability allows precise resource management.


## Example - Label Filters

The example below, provides a definition of eliminating **Deployments** in the `test` namespace with both the `serving=api` and the `environment!=production` labels set. 
!!! example ""

	```yaml
	---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: cleaner-sample1
    spec:
      schedule: "* 0 * * *" # Executes every day at midnight
      resourcePolicySet:
        resourceSelectors:
        - namespace: test
          kind: Deployment
          group: "apps"
          version: v1
          labelFilters:
          - key: serving
            operation: Equal
            value: api # Identifies Deployments with "serving" label set to "api"
          - key: environment
            operation: Different
            value: production # Identifies Deployments with "environment" label different from "production"
      action: Delete # Deletes matching Deployments
	```

By utilising the label filters capability, we can refine the scope of resource management, ensuring that only specific resources are targeted for removal and/or update. 

This approach helps maintain a **clean** and **organised** Kubernetes environment without affecting unintended resources.