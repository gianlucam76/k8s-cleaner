---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Welcome to the k8s-cleaner docs page
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

<a class="github-button" href="https://github.com/gianlucam76/k8s-cleaner" data-icon="icon-park:star" target="_blank" data-show-count="true" aria-label="Star k8s-cleaner on GitHub">Star</a>

[<img src="assets/logo.png" width="200" alt="k8s-cleaner logo">](https://github.com/gianlucam76/k8s-cleaner "k8s-cleaner")


<h1>Kubernetes Controller Cleaner</h1>

## What is k8s-cleaner?

The Kubernetes controller Cleaner (k8s-cleaner) **identifies**, **removes**, or **updates** stale/orphaned or unhealthy resources to maintain a **clean** and **efficient** Kubernetes cluster.

It is designed to handle any Kubernetes resource types (including custom Kubernetes resources) and provides sophisticated filtering capabilities, including label-based selection and custom Lua-based criteria.

## Features and Capabilities

1️⃣ **Schedule**: Specify the frequency at which the k8s-cleaner should scan the cluster and identify stale resources. Utilise the Cron syntax to define recurring schedules.

2️⃣ **DryRun**: Enable safe testing of the k8s-cleaner filtering logic without affecting actual resource configurations. Resources matching the criteria will get identified, but no changes will get applied.

3️⃣ **Label Filtering**: Select resources based on user-defined labels, filtering out unwanted or outdated components. Refine the selection based on label key, operation (equal, different, etc.), and value.

4️⃣ **Lua-based Selection Criteria**: Leverage the [Lua](https://lua.org/) scripting language to create complex and dynamic selection criteria, catering to specific resource management needs. Define custom logic to identify and handle stale resources.

5️⃣ **Notifications**: Stay informed! The k8s-cleaner keeps users in the loop about every cleaned-up resource, whether removed or optimized. Get detailed notification lists and pick your preferred channel: Slack, Webex, Discord, Teams or reports.

## Benefits

💪 **Resource Removal**: Efficiently remove stale resources from your cluster, reclaiming unused resources and improving resource utilisation.

💪 **Resource Updates**: Update outdated resources to ensure they align with the latest configurations and maintain consistent functionality.

💪 **Reduced Resource Bloat**: Minimize resource bloat and maintain a clean and organized cluster, improving overall performance and stability.

By combining the **flexibility** of **scheduling**, the **accuracy** of **label filtering**, the **power** of **Lua-based criteria**, and the ability to **remove** or **update** stale resources, the k8s-cleaner empowers users to effectively manage Kubernetes environments and optimise resource usage.

## Support us

!!! tip ""
    If you like the project, please <a href="https://github.com/gianlucam76/k8s-cleaner" title="k8s-cleaner" target="_blank">give us a</a> <a href="https://github.com/gianlucam76/k8s-cleaner" title="k8s-cleaner" target="_blank" class="heart">:octicons-star-fill-24:</a> if you haven't done so yet. Your support means a lot to us. **Thank you :pray:.**


[:star: k8s-cleaner](https://github.com/gianlucam76/k8s-cleaner "k8s-cleaner"){:target="_blank" .md-button}