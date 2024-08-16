---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Cleanup unused ConfigMaps
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

There are many Kubernetes installations either as a manifest or a Helm chart that after undeployment leave `ConfigMap` behind.

The below Cleaner instance finds all the `ConfigMaps` instances in **all** the namespaces which are **orphaned** (namespaces starting with ```kube``` are excluded)

### Orphaned ConfigMap

By orphaned we refer to a ConfigMap that is not used by:
- Pod through volumes (pod.spec.volumes)
- Pod through environment variables (pod.spec.containers.env and pod.spec.containers.envFrom)

## Example - Cleaner Instance

!!! example ""

    ```yaml
    ---
	apiVersion: apps.projectsveltos.io/v1alpha1
	kind: Cleaner
	metadata:
	  name: unused-configmaps
	spec:
	  schedule: "* 0 * * *"
	  action: Delete
	  resourcePolicySet:
		resourceSelectors:
		- kind: Pod
		  group: ""
		  version: v1
		- kind: ConfigMap
		  group: ""
		  version: v1   
		aggregatedSelection: |
          function skipNamespace(namespace)
            return string.match(namespace, '^kube')
          end

		  function getKey(namespace, name)
			return namespace .. ":" .. name
		  end 

		  function configMapsUsedByPods(pods)
			local podConfigMaps = {}

			for _, pod in ipairs(pods) do
			  if pod.spec.containers ~= nil then
				for _, container in ipairs(pod.spec.containers) do
				  
				  if container.env ~= nil then
					for _, env in ipairs(container.env) do
					  if env.valueFrom ~= nil and env.valueFrom.configMapKeyRef ~= nil then
						key = getKey(pod.metadata.namespace, env.valueFrom.configMapKeyRef.name)
						podConfigMaps[key] = true
					  end
					end
				  end

				  if container.envFrom ~= nil then
					for _, envFrom in ipairs(container.envFrom) do
					  if envFrom.configMapRef ~= nil then
						key = getKey(pod.metadata.namespace, envFrom.configMapRef.name)
						podConfigMaps[key] = true
					  end
					end
				  end  
				end
			  end

			  if pod.spec.initContainers ~= nil then
				for _, initContainer in ipairs(pod.spec.initContainers) do

				  if initContainer.env ~= nil then
					for _, env in ipairs(initContainer.env) do
					  if env.valueFrom ~= nil and env.valueFrom.configMapKeyRef ~= nil then
						key = getKey(pod.metadata.namespace, env.valueFrom.configMapKeyRef.name)
						podConfigMaps[key] = true
					  end
					end
				  end

				  if initContainer.envFrom ~= nil then
					for _, envFrom in ipairs(initContainer.envFrom) do
					  if envFrom.configMapRef ~= nil then
						key = getKey(pod.metadata.namespace,envFrom.configMapRef.name)
						podConfigMaps[key] = true
					  end
					end
				  end  

				end
			  end    

			  if pod.spec.volumes ~= nil then  
				for _, volume in ipairs(pod.spec.volumes) do
				  if volume.configMap ~= nil then
					key = getKey(pod.metadata.namespace,volume.configMap.name)
					podConfigMaps[key] = true
				  end

				  if volume.projected ~= nil and volume.projected.sources ~= nil then
					for _, projectedResource in ipairs(volume.projected.sources) do
					  if projectedResource.configMap ~= nil then
						key = getKey(pod.metadata.namespace,projectedResource.configMap.name)
						podConfigMaps[key] = true
					  end
					end
				  end
				end
			  end
			end  

			return podConfigMaps
		  end

		  function evaluate()
			local hs = {}
			hs.message = ""

			local pods = {}
			local configMaps = {}
			local unusedConfigMaps = {}

			-- Separate configMaps and podsfrom the resources
			for _, resource in ipairs(resources) do
				local kind = resource.kind
				if kind == "ConfigMap" and not skipNamespace(resource.metadata.namespace) then
				  table.insert(configMaps, resource)
				elseif kind == "Pod" then
				  table.insert(pods, resource)
				end
			end

			podConfigMaps = configMapsUsedByPods(pods)

			for _, configMap in ipairs(configMaps) do
			  key = getKey(configMap.metadata.namespace,configMap.metadata.name)
			  if not podConfigMaps[key] then
				table.insert(unusedConfigMaps, {resource = configMap})
			  end
			end

			if #unusedConfigMaps > 0 then
			  hs.resources = unusedConfigMaps
			end
			return hs
		  end
    ```
