---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Cleanup unhealthy Pod with Expired Certificates
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

There is an easy way to identify unhealthy Kubernetes resources with pods with expired certificates.

## Example - Pod with Outdated Secret Data

The below Cleaner instance finds all Pods in **all** namespaces mounting Secrets containing a Certificate issued by `cert-manager`.

The Cleaner instance identifies and reports any Pod that is using expired certificates.

A Pod is using an expired certificates if the secret with certificate have been modified since the Pod's creation.

!!! example ""

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
    name: list-pods-with-expired-certificates
    spec:
    action: Scan 
    schedule: "0 * * * *"
    notifications:
    - name: report
        type: CleanerReport
    resourcePolicySet:
        resourceSelectors:
        - kind: Pod
        group: ""
        version: v1
        - kind: Secret
        group: ""
        version: v1
        - kind: Certificate
        group: "cert-manager.io"
        version: "v1" 
        aggregatedSelection: |
        function getKey(namespace, name)
            return namespace .. ":" .. name
        end

        --  Convert creationTimestamp "2023-12-12T09:35:56Z"
        function convertTimestampString(timestampStr)
            local convertedTimestamp = string.gsub(
            timestampStr,
            '(%d+)-(%d+)-(%d+)T(%d+):(%d+):(%d+)Z',
            function(y, mon, d, h, mi, s)
                return os.time({
                year = tonumber(y),
                month = tonumber(mon),
                day = tonumber(d),
                hour = tonumber(h),
                min = tonumber(mi),
                sec = tonumber(s)
                })
            end
            )
            return convertedTimestamp
        end

        function getLatestTime(times)
            local latestTime = nil
            for _, time in ipairs(times) do
            if latestTime == nil or os.difftime(tonumber(time), tonumber(latestTime)) > 0 then
                latestTime = time
            end
            end
            return latestTime
        end

        function getSecretUpdateTime(secret)
            local times = {}
            if secret.metadata.managedFields ~= nil then
            for _, mf in ipairs(secret.metadata.managedFields) do
                if mf.time ~= nil then
                table.insert(times, convertTimestampString(mf.time))
                end
            end
            end

            return getLatestTime(times)
        end

        function isPodOlderThanSecret(podTimestamp, secretTimestamp)
            timeDifference = os.difftime(tonumber(podTimestamp), tonumber(secretTimestamp))
            return  timeDifference < 0
        end

        function getPodTimestamp(pod)
            if pod.status ~= nil and pod.status.conditions ~= nil then
            for _,condition in ipairs(pod.status.conditions) do
                if condition.type == "PodReadyToStartContainers" and condition.status == "True" then
                return convertTimestampString(condition.lastTransitionTime)
                end
            end
            end
            return convertTimestampString(pod.metadata.creationTimestamp)
        end

        -- secrets contains key:value where key identify a Secret with a Certificate and value
        -- if the time of latest update
        function hasOutdatedSecret(pod, secrets)
            podTimestamp = getPodTimestamp(pod)

            if pod.spec.volumes ~= nil then  
            for _, volume in ipairs(pod.spec.volumes) do
                if volume.secret ~= nil then
                key = getKey(pod.metadata.namespace, volume.secret.secretName)
                -- if secrets contains a certificate
                if secrets[key] ~= nil then
                    if isPodOlderThanSecret(podTimestamp, secrets[key]) then
                    return true, "secret " .. key .. " has been updated after pod creation"
                    end
                end  
                end

                if volume.projected ~= nil and volume.projected.sources ~= nil then
                for _, projectedResource in ipairs(volume.projected.sources) do
                    if projectedResource.secret ~= nil then
                    key = getKey(pod.metadata.namespace, projectedResource.secret.name)
                    -- if secrets contains a certificate
                    if secrets[key] ~= nil then
                        if isPodOlderThanSecret(podTimestamp, secrets[key]) then
                        return true, "secret " .. key .. " has been updated after pod creation"
                        end
                    end  
                    end
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
            local certificates = {}
            local secrets = {}

            -- Contains all Secrets containing a Certificate generated using cert-manager
            local certificateSecrets = {}

            -- Contains all Secrets containing a Certificate generated using cert-manager
            local degradedSecrets = {}

            -- Separate secrets, pods and certificates
            for _, resource in ipairs(resources) do
            local kind = resource.kind
            if kind == "Certificate" then
                table.insert(certificates, resource)
            elseif kind == "Secret" then
                key = getKey(resource.metadata.namespace, resource.metadata.name)
                updateTimestamp = getSecretUpdateTime(resource)
                secrets[key] = updateTimestamp
            elseif kind == "Pod" then
                table.insert(pods, resource)
            end
            end

            -- Find all secrets with certificate generated by cert-manager
            for _, certificate in ipairs(certificates) do
            key =  getKey(certificate.metadata.namespace, certificate.spec.secretName)
            certificateSecrets[key] = secrets[key]
            end

            local podsWithOutdatedSecret = {}

            for _, pod in ipairs(pods) do
            outdatedData, message = hasOutdatedSecret(pod, certificateSecrets)
            if outdatedData then
                table.insert(podsWithOutdatedSecret, {resource= pod, message = message})
            end
            end

            if #podsWithOutdatedSecret > 0 then
            hs.resources = podsWithOutdatedSecret
            end
            return hs
        end
    ```
