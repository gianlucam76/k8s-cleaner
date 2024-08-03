---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: resourceSelector
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

## Introduction to resourceSelector

It might be cases that operator need to examine resources of distinct types simultaneously.

Let's assume we would like to eliminate **all** Deployment instances that are not backed-up by an Autoscaler instance. The k8s-cleaner allows this action. By employing the `resourceSelector`, we can select **all** `Deployment` and `Autoscaler` instances.

As a next step, we have to define the `aggregatedSelection`. `AggregatedSelection` will be given all instances collected by the Cleaner using the `resourceSelector`. In this example, all Deployment and Autoscaler instances in the **foo** namespace.

## Example  - Deployment not Backed-up by Autoscaler

```yaml
---
# Find all Deployments not backed up by an Autoscaler. Those are a match.
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: cleaner-sample3
spec:
  schedule: "* 0 * * *"
  action: Delete # Delete matching resources
  resourcePolicySet:
    resourceSelectors:
    - namespace: foo
      kind: Deployment
      group: "apps"
      version: v1
    - namespace: foo
      kind: HorizontalPodAutoscaler
      group: "autoscaling"
      version: v2
    aggregatedSelection: |
      function evaluate()
        local hs = {}
        hs.valid = true
        hs.message = ""

        local deployments = {}
        local autoscalers = {}
        local deploymentWithNoAutoscaler = {}

        -- Separate deployments and services from the resources
        for _, resource in ipairs(resources) do
            local kind = resource.kind
                if kind == "Deployment" then
                    table.insert(deployments, resource)
                elseif kind == "HorizontalPodAutoscaler" then
                    table.insert(autoscalers, resource)
                end
        end

        -- Check for each deployment if there is a matching HorizontalPodAutoscaler
        for _, deployment in ipairs(deployments) do
            local deploymentName = deployment.metadata.name
            local matchingAutoscaler = false

            for _, autoscaler in ipairs(autoscalers) do
                if autoscaler.spec.scaleTargetRef.name == deployment.metadata.name then
                    matchingAutoscaler = true
                    break
                end
            end

            if not matchingAutoscaler then
                table.insert(deploymentWithNoAutoscaler, {resource = deployment})
                break
            end
        end

        if #deploymentWithNoAutoscaler > 0 then
          hs.resources = deploymentWithNoAutoscaler
        end
        return hs
      end
```