---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Cleanup unused Completed Jobs and Long Running Pods
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

There are a number of Kubernetes installations that leave complete `Jobs` behind. This could be problematic as a new Helm chart installation might fail because completed `Jobs` are not cleanered from a previous uninstallation.

## Example Completed Jobs

The Cleaner instance definition will find any `Jobs` with the below specifications.

- `status.completionTime set`
- `status.succeeded` set to a value greater than zero
- The `Job` has no running or pending pods and will instruct Cleaner to delete it.


!!! example ""

    ```yaml
    ---
	apiVersion: apps.projectsveltos.io/v1alpha1
	kind: Cleaner
	metadata:
	  name: completed-jobs
	spec:
	  schedule: "* 0 * * *"
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

## Example - Long Running Pods

The Cleaner instance definition will find any `Pod` with the below specifications.

- Has been running for longer than one hour (3600 seconds)
- Was created by a Job

The Cleaner instance will delete the Pod but do not the Job.


!!! example ""

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
    name: pods-from-job
    spec:
    schedule: "* 0 * * *"
    resourcePolicySet:
        resourceSelectors:
        - kind: Pod
        group: ""
        version: v1
        evaluate: |
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

            function evaluate()
            hs = {}
            hs.matching = false

            currentTime = os.time()

            creationTimestamp = convertTimestampString(obj.metadata.creationTimestamp)

            hs.message = creationTimestamp
            print('creationTimestamp: ' .. creationTimestamp)
            print('currentTime: ' .. currentTime)

            timeDifference = os.difftime(currentTime, tonumber(creationTimestamp))

            print('timeDifference: ' .. timeDifference)

            -- if pod has been running for over an hour
            if timeDifference > 3600 then
                if obj.metadata.ownerReferences ~= nil then
                for _, owner in ipairs(obj.metadata.ownerReferences) do
                    if owner.kind == "Job" and owner.apiVersion == "batch/v1" then
                    hs.matching = true
                    end
                end
                end
            end


            return hs
            end
    action: Delete
    ```
