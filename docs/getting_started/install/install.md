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
1. kubectl CLI installed
1. kubeconfig for authentication

## Installation

The k8s-cleaner can be installed in any Kubernetes cluster independent if it is in an on-prem or in a Cloud environment. The installation is pretty simple.

```bash
$ export KUBECONFIG=<directory to the kubeconfig file>
$ kubectl apply -f https://raw.githubusercontent.com/gianlucam76/k8s-cleaner/main/manifest/manifest.yaml
```

!!! note
    The above command will create a new namespace with the name `projectsveltos` and install the Kubernetes cleaner controller there.