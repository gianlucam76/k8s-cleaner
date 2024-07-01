---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Scale up and down Kubernetes Resources
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

## Manage and Automate Resource Operations

The k8s-cleaner does not just help users identify **unused** or **unhealthy** resources; it can also **automate** various operations to enhance the cluster's efficiency and management. For example, scale down **Deployments/Daemonsets/Statefulsets** based on a specified annotation.

The example below demonstrates how to automatically scale down Deployments, DaemonSets, and StatefulSets with a specified annotation at a desired time (e.g., 8 PM nightly). Before scaling down, the replica count is stored in another annotation for later retrieval. At the configured scale-up time (e.g., 8 AM), resources are restored, ensuring efficient resource utilization during off-peak hours.

### Pause YAML Definition

!!! example "Example - Pause"

    ```yaml
    # The defintion:
    # - runs at 8PM every day
    # - finds all Deployments/StatefulSet/DaemonSet with
    # annotation "pause-resume"
    #
    # For matched resources: 
    # - stores current replicas in the annotation "previous-replicas"
    # - sets their replicas to zero (scale down and pause)
    #
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
    name: scale-down-deployment-statefulset-daemonset
    spec:
    schedule: "* 20 * * *"
    action: Transform
    transform: |
        -- Set replicas to 0
        function transform()
        hs = {}

        if obj.metadata.annotations == nil then
            obj.metadata.annotations = {}
        end
        -- store in the annotation current replicas value
        obj.metadata.annotations["previous-replicas"] = tostring(obj.spec.replicas)
        
        -- reset replicas to 0
        obj.spec.replicas = 0
        
        hs.resource = obj
        return hs
        end  
    resourcePolicySet:
        resourceSelectors:
        - kind: Deployment
        group: apps
        version: v1
        - kind: StatefulSet
        group: "apps"
        version: v1
        - kind: DaemonSet
        group: "apps"
        version: v1
        aggregatedSelection: |
        function evaluate()
            local hs = {}

            -- returns true if object has annotaiton "pause-resume" 
            function hasPauseAnnotation(obj)
            if obj.metadata.annotations ~= nil then
                if obj.metadata.annotations["pause-resume"] then
                return true
                end

                return false
            end

            return
            end

            local resourceToPause = {}

            for _, resource in ipairs(resources) do
            if hasPauseAnnotation(resource) then
                table.insert(resourceToPause, {resource = resource})
            end
            end

            if #resourceToPause > 0 then
            hs.resources = resourceToPause
            end
            return hs
        end   
    ```

### Resume YAML Definition

As we defined the pause action for specific resources, during peak times, we would like to scale the Kubernetes resources back to their initial state. To do that, we will use the `resume` YAML definition found below.

!!! example "Example - Resume"

    ```yaml
    # The cleaner:
    # - runs at 8AM every day
    # - finds all Deployments/StatefulSet/DaemonSet with
    # annotation "pause-resume"
    #
    # For matched resources: 
    # - gets the old replicas in the annotation "previous-replicas"
    # - sets the replicas to such value found above (scale deployment/statefulset/daemonset up)
    #
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
    name: scale-up-deployment-statefulset-daemonset
    spec:
    schedule: "* 8 * * *"
    action: Transform
    transform: |
        -- Set replicas to 0
        function transform()
        hs = {}
        if obj.metadata.annotations == nil then
            return
        end
        if not obj.metadata.annotations["previous-replicas"] then
            return
        end
        -- reset replicas
        obj.spec.replicas = tonumber(obj.metadata.annotations["previous-replicas"])
        hs.resource = obj
        return hs
        end  
    resourcePolicySet:
        resourceSelectors:
        - kind: Deployment
        group: apps
        version: v1
        - kind: StatefulSet
        group: "apps"
        version: v1
        - kind: DaemonSet
        group: "apps"
        version: v1
        aggregatedSelection: |
        function evaluate()
            local hs = {}

            -- returns true if object has annotaiton "pause-resume" 
            function hasPauseAnnotation(obj)
            if obj.metadata.annotations ~= nil then
                if obj.metadata.annotations["pause-resume"] then
                return true
                end

                return false
            end

            return false
            end

            local resourceToUnPause = {}

            for _, resource in ipairs(resources) do
            if hasPauseAnnotation(resource) then
                table.insert(resourceToUnPause, {resource = resource})
            end
            end

            if #resourceToUnPause > 0 then
            hs.resources = resourceToUnPause
            end
            return hs
        end
    ```
