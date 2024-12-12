---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Welcome to the k8s-cleaner installation page
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

## What is k8s-cleaner?

The Kubernetes controller Cleaner (k8s-cleaner) **identifies**, **removes**, or **updates** stale/orphaned or unhealthy resources to maintain a **clean** and **efficient** Kubernetes cluster.

It is designed to handle any Kubernetes resource types (including custom Kubernetes resources) and provides sophisticated filtering capabilities, including label-based selection and custom Lua-based criteria.

## Pre-requisites
To work with the k8s-cleaner, ensure you have the below points covered.

1. A Kubernetes cluster
1. [kubectl CLI](https://kubernetes.io/releases/download/) installed
1. [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) for authentication

## Installation

### Kubernetes Manifest

The k8s-cleaner can be installed in any Kubernetes cluster independent if it is in an on-prem or in a Cloud environment. The k8s-cleaner can be deployed down the clusters with your favourite Continious Deployment tool! The installation is pretty simple.

!!! example ""
    ```bash
    $ export KUBECONFIG=<directory to the kubeconfig file>

    $ kubectl apply -f https://raw.githubusercontent.com/gianlucam76/k8s-cleaner/v0.11.0/manifest/manifest.yaml
    ```

!!! note
    The above command will create a new namespace with the name `projectsveltos` and install the Kubernetes cleaner controller there.

### Helm Chart

There is the option to install the k8s-cleaner with a Helm chart. To do so, simply follow the commands listed below.

!!! example ""
    ```bash
    $ helm install k8s-cleaner oci://ghcr.io/gianlucam76/charts/k8s-cleaner \
        --version 0.10.0 \
        --namespace k8s-cleaner \
        --create-namespace #(1)
    ```

    1. It will create the namespace k8s-cleaner and deploy everything in the namespace

#### Validation

```bash
$ kubectl get namespace
NAME              STATUS   AGE
default           Active   6h11m
k8s-cleaner       Active   34s
kube-node-lease   Active   6h11m
kube-public       Active   6h11m
kube-system       Active   6h11m

$ kubectl get all -n k8s-cleaner
NAME                               READY   STATUS    RESTARTS   AGE
pod/k8s-cleaner-78b9d794c5-jpp76   2/2     Running   0          43s

NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/k8s-cleaner-metrics   ClusterIP   10.43.149.237   <none>        8081/TCP   43s

NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/k8s-cleaner   1/1     1            1           43s

NAME                                     DESIRED   CURRENT   READY   AGE
replicaset.apps/k8s-cleaner-78b9d794c5   1         1         1       43s
```

!!! tip
    Before getting started with the k8s-cleaner, have a look at the **Features** section. Familiarise with the [label filters](../features/label_filters/label_filters.md), [store resource](../features/store_resources/store_resource_yaml.md), [update resources](../features/update_resources/update_resources.md), [resource selector](../features/resourceselector/resourceselector.md), and [schedule](../features/schedule/schedule.md) sections. Use the examples provided and familiarise with the syntax and the capabilities provided
