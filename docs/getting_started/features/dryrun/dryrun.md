---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Dry Run
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

## Introduction to Dry Run

To preview which resources match the Cleaner's criteria, set the **Action** field to **Scan**. The k8s-cleaner will still execute its logic but will **not** delete and/or update any resources.

To identify matching resources, search the controller logs for the message `resource is a match for cleaner`.


## Example - Dry Run

The example below, provides a definition of eliminating **Deployments** in the `test` namespace with both the `serving=api` and the `environment!=production` labels set. 
!!! example ""

	```yaml
	---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: cleaner-sample1
    spec:
      schedule: "* 0 * * *" # Runs every day at midnight
      resourcePolicySet:
        resourceSelectors:
        - namespace: test
          kind: Deployment
          group: "apps"
          version: v1
          labelFilters:
          - key: serving
            operation: Equal
            value: api # Match deployments with the "serving" label set to "api"
          - key: environment
            operation: Different
            value: prouction # Match deployments with the "environment" label different from "production"
      action: Scan
	```

By setting the **Action** field to **Scan**, we can safely test the Cleaner's filtering logic without affecting your actual deployment configurations. Once we are confident in the filtering criteria, you can set the **Action** to **delete** or **modify**.