---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Cleanup unused Secrets
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

There are many Kubernetes installations either as a manifest or a Helm chart that after undeployment leave Kubernetes `secret` behind.

The below Cleaner instance finds all the `Secret` instances in **all** the namespaces which are **orphaned**.

### Orphaned Secrets

By orphaned we refer to a Secret that is not used by:

- Pod through volumes (pod.spec.volumes)
- Pod through environment variables (pod.spec.containers.env and pod.spec.containers.envFrom)
- Pod for image pulls (pod.spec.imagePullSecrets)
- Ingress TLS (ingress.spec.tls)
- ServiceAccounts

## Example - Cleaner Instance

!!! example ""

    ```yaml
    ---
	apiVersion: apps.projectsveltos.io/v1alpha1
	kind: Cleaner
	metadata:
	  name: unused-secrets
	spec:
	  schedule: "* 0 * * *"
	  action: Delete
	  resourcePolicySet:
		resourceSelectors:
		- kind: Pod
		  group: ""
		  version: v1
		- kind: Secret
		  group: ""
		  version: v1
		- kind: Ingress
		  group: "networking.k8s.io"
		  version: v1      
		aggregatedSelection: |
		  function getKey(namespace, name)
			return namespace .. ":" .. name
		  end

		  function secretsUsedByPods(pods)
			local podSecrets = {}

			for _, pod in ipairs(pods) do
			  if pod.spec.containers ~= nil then
				for _, container in ipairs(pod.spec.containers) do
				  
				  if container.env ~= nil then
					for _, env in ipairs(container.env) do
					  if env.valueFrom ~= nil and env.valueFrom.secretKeyRef ~= nil then
						key = getKey(pod.metadata.namespace, env.valueFrom.secretKeyRef.name)
						podSecrets[key] = true
					  end
					end
				  end

				  if  container.envFrom ~= nil then
					for _, envFrom in ipairs(container.envFrom) do
					  if envFrom.secretRef ~= nil then
						key = getKey(pod.metadata.namespace, envFrom.secretRef.name)
						podSecrets[key] = true
					  end
					end
				  end  
				end
			  end

			  if pod.spec.initContainers ~= nil then
				for _, initContainer in ipairs(pod.spec.initContainers) do
				  if initContainer.env ~= nil then
					for _, env in ipairs(initContainer.env) do
					  if env.valueFrom ~= nil and env.valueFrom.secretKeyRef ~= nil then
						key = getKey(pod.metadata.namespace, env.valueFrom.secretKeyRef.name)
						podSecrets[key] = true
					  end
					end
				  end
				end
			  end

			  if pod.spec.imagePullSecrets ~= nil then
				for _, secret in ipairs(pod.spec.imagePullSecrets) do
				  key = getKey(pod.metadata.namespace, secret.name)
				  podSecrets[key] = true
				end
			  end          

			  if pod.spec.volumes ~= nil then  
				for _, volume in ipairs(pod.spec.volumes) do
				  if volume.secret ~= nil then
					key = getKey(pod.metadata.namespace, volume.secret.secretName)
					podSecrets[key] = true
				  end

				  if volume.projected ~= nil and volume.projected.sources ~= nil then
					for _, projectedResource in ipairs(volume.projected.sources) do
					  if projectedResource.secret ~= nil then
						key = getKey(pod.metadata.namespace, projectedResource.secret.name)
						podSecrets[key] = true
					  end
					end
				  end
				end
			  end
			end  

			return podSecrets
		  end

		  function secretsUsedByIngresses(ingresses)
			local ingressSecrets = {}
			for _, ingress in ipairs(ingresses) do
			  if ingress.spec.tls ~= nil  then
				for _, tls in ipairs(ingress.spec.tls) do
				  key = getKey(ingress.metadata.namespace, tls.secretName)
				  ingressSecrets[key] = true
				end
			  end
			end
			
			return ingressSecrets
		  end

		  function evaluate()
			local hs = {}
			hs.message = ""

			local pods = {}
			local secrets = {}
			local ingresses = {}
			local unusedSecrets = {}

			-- Separate secrets and pods and ingresses from the resources
			for _, resource in ipairs(resources) do
				local kind = resource.kind
				if kind == "Secret" then
				  table.insert(secrets, resource)
				elseif kind == "Pod" then
				  table.insert(pods, resource)
				elseif kind == "Ingress" then
				  table.insert(ingresses, resource)
				end
			end

			podSecrets = secretsUsedByPods(pods)
			ingressSecrets = secretsUsedByIngresses(ingresses)

			for _, secret in ipairs(secrets) do
			  if secret.type ~= "kubernetes.io/service-account-token" then
				key = getKey(secret.metadata.namespace, secret.metadata.name)
				if not podSecrets[key] and not ingressSecrets[key] then
				  table.insert(unusedSecrets, {resource = secret})
				end
			  end
			end

			if #unusedSecrets > 0 then
			  hs.resources = unusedSecrets
			end
			return hs
		  end
    ```
