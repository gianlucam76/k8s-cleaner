---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Resource Update
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

## Introduction to Resource Update

Beyond removing **stale** resources, the k8s-cleaner can also facilitate in the dynamic update of existing resource configurations.

The capability allows users to modify resource specifications based on **specific criteria**, ensuring alignment with evolving requirements and maintaining resource consistency.


## Example - Resource Update

Consider the scenario where we want to update **Service** objects in the `foo` namespace to use **version2** apps.

1. The **evaluate** function allows users to select resources, Services in the `foo` namespace pointing to **version1** apps.
2. The **trasnform** function will change the matching resources, by updating the `obj.spec.selector["app"]` to **version2**.

!!! example ""

	```yaml
	---
	apiVersion: apps.projectsveltos.io/v1alpha1
	kind: Cleaner
	metadata:
	  name: cleaner-sample3
	spec:
      schedule: "* 0 * * *"
      resourcePolicySet:
        resourceSelectors:
    	- namespace: foo
      	  kind: Service
          group: ""
          version: v1
          evaluate: |
            -- Define how resources will be selected 
            function evaluate()
            hs = {}
            hs.matching = false
            if obj.spec.selector ~= nil then
              if obj.spec.selector["app"] == "version1" then
                hs.matching = true
              end
            end
            return hs
            end
      action: Transform # Update matching resources
      transform: |
          -- Define how resources will be updated
          function transform()
          hs = {}
          obj.spec.selector["app"] = "version2"
          hs.resource = obj
          return hs
          end
	```