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

## Selecting Namespaces

Each entry in the `resourceSelectors` list can limit the namespaces it looks at. Two fields control this.

| Field | Description |
|---|---|
| `namespace` | A single namespace, by name. |
| `namespaceSelector` | A Kubernetes label selector. Every namespace in the cluster whose labels match is considered. |

If **both** are set, the two results are combined: the Cleaner considers the namespaces matched by `namespaceSelector` **plus** the one named in `namespace`.

If **neither** is set, the Cleaner considers resources in **all** namespaces.

### A Single Namespace

Restrict the Cleaner to the `foo` namespace.

```yaml
resourceSelectors:
- namespace: foo
  kind: Pod
  group: ""
  version: v1
```

### Namespaces by Label

`namespaceSelector` accepts any valid Kubernetes label selector. To act only on namespaces labelled `environment=dev`:

```yaml
resourceSelectors:
- namespaceSelector: "environment=dev"
  kind: Pod
  group: ""
  version: v1
```

## Excluding Namespaces

A common requirement is the inverse: run a Cleaner across the whole cluster, but never touch protected namespaces such as `kube-system`. This is worth stating explicitly, because it is easy to assume a dedicated "exclude" field is needed. It is not.

`namespaceSelector` is parsed as a full Kubernetes label selector, so it supports **set-based operators**, including `notin`. Kubernetes also labels every namespace automatically with `kubernetes.io/metadata.name`, whose value is the namespace's own name. No manual labelling is required.

Put those together and you can exclude namespaces **by name**:

```yaml
namespaceSelector: "kubernetes.io/metadata.name notin (kube-system,cattle-system,cert-manager)"
```

### Example - Delete Failed Pods, Except in Protected Namespaces

This Cleaner deletes Pods that terminated with an error, in every namespace except the protected ones. Note that the `evaluate` function contains no namespace check at all: the namespace filtering happens in `namespaceSelector`, which keeps the exclusion list declarative and in one place.

```yaml
---
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: error-state-pods
spec:
  schedule: "0 * * * *"
  action: Delete
  resourcePolicySet:
    resourceSelectors:
    - kind: Pod
      group: ""
      version: v1
      namespaceSelector: "kubernetes.io/metadata.name notin (kube-system,cattle-system,cert-manager,my-production-db)"
      evaluate: |
        function evaluate()
          hs = {}
          hs.matching = false
          if obj.status.phase == "Failed" then
            if obj.status.containerStatuses then
              for _, containerStatus in ipairs(obj.status.containerStatuses) do
                if containerStatus.state and containerStatus.state.terminated and
                   containerStatus.state.terminated.reason == "Error" and
                   containerStatus.state.terminated.exitCode ~= 0 then
                  hs.matching = true
                  break
                end
              end
            end
          end
          return hs
        end
```

!!! warning "Make sure the selector matches at least one namespace"
    A `namespaceSelector` that matches **no** namespace does not narrow the Cleaner down to nothing. The Cleaner falls back to considering **all** namespaces, exactly as if no namespace filtering had been set.

    This matters most when `action` is `Delete`. Before applying an exclusion selector, confirm it selects the namespaces you expect:

    ```bash
    kubectl get namespaces -l 'kubernetes.io/metadata.name notin (kube-system,cattle-system,cert-manager)'
    ```

    Running the Cleaner in [dry-run mode](../dryrun/dryrun.md) first is the safest way to review what it would act on.

!!! note "Things to keep in mind"
    - `namespaceSelector` is set **per `resourceSelector`**. A Cleaner that selects several kinds must repeat the selector on each entry.
    - The automatic `kubernetes.io/metadata.name` label is applied by Kubernetes from **v1.22** onwards. On older clusters, label the namespaces yourself and select on your own label.
    - Namespace selection does not apply to **cluster-scoped** resources, which do not belong to a namespace.