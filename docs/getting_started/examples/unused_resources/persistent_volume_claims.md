---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Cleanup unused PersistentVolumeClaims
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

The k8s-cleaner is able to delete unsused `PersistentVolumeClaim`. The below Cleaner instance will find any `PersistentVolumeClaim` resources which are currently unused by any **Pods**. The example below will consider **all** namespaces.

## Example - Cleaner Instance

!!! example ""

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
    name: stale-persistent-volume-claim
    spec:
    schedule: "* 0 * * *"
    action: Delete # Delete matching resources
    resourcePolicySet:
        resourceSelectors:
        - kind: Pod
        group: ""
        version: v1
        - kind: PersistentVolumeClaim
        group: ""
        version: v1
        aggregatedSelection: |
        function isUsed(pvc, pods)
            if pods == nil then
            return false
            end
            for _, pod in ipairs(pods) do
            if pod.spec.volumes ~= nil then
                for _,volume in ipairs(pod.spec.volumes) do
                if volume.persistentVolumeClaim ~= nil and volume.persistentVolumeClaim.claimName == pvc.metadata.name then
                    return true
                end
                end
            end
            end
            return false
        end  

        function evaluate()
            local hs = {}
            hs.message = ""

            local pods = {}
            local pvcs = {}
            local unusedPVCs = {}

            -- Separate pods and pvcs from the resources
            -- Group those by namespace
            for _, resource in ipairs(resources) do
            local kind = resource.kind
            if kind == "Pod" then
                if not pods[resource.metadata.namespace] then
                pods[resource.metadata.namespace] = {}
                end
                table.insert(pods[resource.metadata.namespace], resource)
            elseif kind == "PersistentVolumeClaim" then
                if not pvcs[resource.metadata.namespace] then
                pvcs[resource.metadata.namespace] = {}
                end
                table.insert(pvcs[resource.metadata.namespace], resource)
            end
            end

            -- Iterate through each namespace and identify unused PVCs
            for namespace, perNamespacePVCs in pairs(pvcs) do
            for _, pvc in ipairs(perNamespacePVCs) do
                if not isUsed(pvc, pods[namespace]) then
                table.insert(unusedPVCs, {resource = pvc})
                end
            end
            end
            
            if #unusedPVCs > 0 then
            hs.resources = unusedPVCs
            end
            return hs
        end
    ```
