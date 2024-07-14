---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Cleanup Unhealthy Deployments
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

There is an easy way to identify unhealthy Kubernetes Deployments. An unhealthy Deployment can be one that has Pods with may restarts.

## Example - Scale Down High Restart Deployment

The below Cleaner definition will find all unhealthy Deployments instances. 

A Deployment instance is considered unhealthy if each of its Pods have been restarted at least 50 times.

Any unhealthy deployment will be transformed by scaling its replicas to zero.

!!! example ""

    ```yaml
    ---
	apiVersion: apps.projectsveltos.io/v1alpha1
	kind: Cleaner
	metadata:
	  name: scale-down-high-restart-deployment
	spec:
	  schedule: "* 0 * * *"
	  action: Transform
	  transform: |
	    -- Set replicas to 0
	    function transform()
	      hs = {}
	      obj.spec.replicas = 0
	      hs.resource = obj
	      return hs
	    end  
	  resourcePolicySet:
	    resourceSelectors:
	    - kind: Deployment
	      group: apps
	      version: v1
	    - kind: Pod
	      group: ""
	      version: v1
	    aggregatedSelection: |
	      -- any pod restarted more than this value is considered unhealthy
	      min_restarts = 50

	      deployment_pods_map = {}

	      function getKey(namespace, name)
		return namespace .. ":" .. name
	      end

	      -- Given a ReplicaSet name, returns the deployment name
	      function getDeploymentName(replicaSetName)
		local reversed = string.reverse(replicaSetName)
		local firstDashIndex, _ = string.find(reversed, "-")
		if firstDashIndex then
		   -- Convert index from reversed string to original string
		  local lastDashIndex = #replicaSetName - firstDashIndex + 1
		  return replicaSetName:sub(1, lastDashIndex - 1)
		else
		  return replicaSetName
		end
	      end

	      -- If Pod's OwnerReference is a ReplicaSet
	      -- returns deployment name. Returns an empty string otherwise
	      function getDeployment(pod)
		if pod.metadata.ownerReferences ~= nil then
		  for _, owner in ipairs(pod.metadata.ownerReferences) do
		    if owner.kind == "ReplicaSet" and owner.apiVersion == "apps/v1" then
		      local ownerName = owner.name 
		      return getDeploymentName(ownerName)
		    end
		  end
		end
		return ""
	      end

	      -- Function to add a pod to a deployment's pod list
	      function add_pod_to_deployment(depl_key, pod)
		-- Get the existing pod list for the deployment (or create an empty list if it doesn't exist)
		pods = deployment_pods_map[depl_key] or {}

		-- Append the new pod to the list
		table.insert(pods, pod)
		-- Update the map with the modified pod list
		deployment_pods_map[depl_key] = pods
	      end

	      -- Returns true if 
	      function isPodUnhealthy(pod)
		if pod.status ~= nil and pod.status.containerStatuses ~= nil then
		  for _,container_status in ipairs (pod.status.containerStatuses) do
		    if container_status.restartCount > min_restarts then
		      return true
		    end
		  end
		end
		return false
	      end

	      function isDeploymentUnhealthy(deployment)
		depl_key = getKey(deployment.metadata.namespace, deployment.metadata.name)
		pods = deployment_pods_map[depl_key] or {}
		for _,pod in ipairs (pods) do
		  if not isPodUnhealthy(pod) then
		    return false
		  end
		end
	      
		return true
	      end

	      function evaluate()
		local hs = {}
		
		local deployments = {}
		
		-- Separate pods from deployments
		for _, resource in ipairs(resources) do
		  local kind = resource.kind
		  if kind == "Deployment" then
		    table.insert(deployments, resource)
		  else
		    deplName = getDeployment(resource)
		    if deplName ~= "" then
		      depl_key = getKey(resource.metadata.namespace, deplName)
		      add_pod_to_deployment(depl_key, resource)
		    end  
		  end
		end

		local unhealthy_deployments = {}

		for _, deployment in ipairs(deployments) do
		  isUnhealthy = isDeploymentUnhealthy(deployment, references)
		  if isUnhealthy then
		    table.insert(unhealthy_deployments, {resource = deployment})
		  end
		end

		if #unhealthy_deployments > 0 then
		  hs.resources = unhealthy_deployments
		end
		return hs
	      end   
    ```

## Example - Pods with High Restarts

The below Cleaner instance identifies and removes unhealthy Deployments in a Kubernetes cluster based on pod restarts.

Unhealthy is a Deployment that has Pods restarted at least 50 times.

When an unhealthy Deployment is found, it will be automatically **scaled down** to zero replicas.

!!! example ""

    ```yaml
    ---
	apiVersion: apps.projectsveltos.io/v1alpha1
	kind: Cleaner
	metadata:
	  name: scale-down-high-restart-deployment
	spec:
	  schedule: "* 0 * * *"
	  action: Transform
	  transform: |
	    -- Set replicas to 0
	    function transform()
	      hs = {}
	      obj.spec.replicas = 0
	      hs.resource = obj
	      return hs
	    end  
	  resourcePolicySet:
	    resourceSelectors:
	    - kind: Deployment
	      group: apps
	      version: v1
	    - kind: Pod
	      group: ""
	      version: v1
	    aggregatedSelection: |
	      -- any pod restarted more than this value is considered unhealthy
	      min_restarts = 50

	      deployment_pods_map = {}

	      function getKey(namespace, name)
		return namespace .. ":" .. name
	      end

	      -- Given a ReplicaSet name, returns the deployment name
	      function getDeploymentName(replicaSetName)
		local reversed = string.reverse(replicaSetName)
		local firstDashIndex, _ = string.find(reversed, "-")
		if firstDashIndex then
		   -- Convert index from reversed string to original string
		  local lastDashIndex = #replicaSetName - firstDashIndex + 1
		  return replicaSetName:sub(1, lastDashIndex - 1)
		else
		  return replicaSetName
		end
	      end

	      -- If Pod's OwnerReference is a ReplicaSet
	      -- returns deployment name. Returns an empty string otherwise
	      function getDeployment(pod)
		if pod.metadata.ownerReferences ~= nil then
		  for _, owner in ipairs(pod.metadata.ownerReferences) do
		    if owner.kind == "ReplicaSet" and owner.apiVersion == "apps/v1" then
		      local ownerName = owner.name 
		      return getDeploymentName(ownerName)
		    end
		  end
		end
		return ""
	      end

	      -- Function to add a pod to a deployment's pod list
	      function add_pod_to_deployment(depl_key, pod)
		-- Get the existing pod list for the deployment (or create an empty list if it doesn't exist)
		pods = deployment_pods_map[depl_key] or {}

		-- Append the new pod to the list
		table.insert(pods, pod)
		-- Update the map with the modified pod list
		deployment_pods_map[depl_key] = pods
	      end

	      -- Returns true if 
	      function isPodUnhealthy(pod)
		if pod.status ~= nil and pod.status.containerStatuses ~= nil then
		  for _,container_status in ipairs (pod.status.containerStatuses) do
		    if container_status.restartCount > min_restarts then
		      return true
		    end
		  end
		end
		return false
	      end

	      function isDeploymentUnhealthy(deployment)
		depl_key = getKey(deployment.metadata.namespace, deployment.metadata.name)
		pods = deployment_pods_map[depl_key] or {}
		for _,pod in ipairs (pods) do
		  if not isPodUnhealthy(pod) then
		    return false
		  end
		end
	      
		return true
	      end

	      function evaluate()
		local hs = {}
		
		local deployments = {}
		
		-- Separate pods from deployments
		for _, resource in ipairs(resources) do
		  local kind = resource.kind
		  if kind == "Deployment" then
		    table.insert(deployments, resource)
		  else
		    deplName = getDeployment(resource)
		    if deplName ~= "" then
		      depl_key = getKey(resource.metadata.namespace, deplName)
		      add_pod_to_deployment(depl_key, resource)
		    end  
		  end
		end

		local unhealthy_deployments = {}

		for _, deployment in ipairs(deployments) do
		  isUnhealthy = isDeploymentUnhealthy(deployment, references)
		  if isUnhealthy then
		    table.insert(unhealthy_deployments, {resource = deployment})
		  end
		end

		if #unhealthy_deployments > 0 then
		  hs.resources = unhealthy_deployments
		end
		return hs
	      end   
    ```