---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Cleanup unused PersistentVolumes
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

## Introduction

The k8s-cleaner is able to delete unsused `PersistentVolumes`. The below Cleaner instance will find any `PersistentVolume` resources with the Phase set to anything **but** `Bound`.

## Example - Cleaner Instance

!!! example ""

    ```yaml
    ---
	apiVersion: apps.projectsveltos.io/v1alpha1
	kind: Cleaner
	metadata:
	  name: unbound-peristent-volumes
	spec:
	  schedule: "* 0 * * *"
	  resourcePolicySet:
		resourceSelectors:
		- kind: PersistentVolume
		  group: ""
		  version: v1
		  evaluate: |
			function evaluate()
			  hs = {}
			  hs.matching = false
			  if obj.status ~= nil and obj.status.phase ~= "Bound" then
				hs.matching = true
			  end
			  return hs
			end
	  action: Delete
    ```
