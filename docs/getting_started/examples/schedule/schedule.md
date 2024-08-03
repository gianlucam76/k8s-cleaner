---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Schedule
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

## Introduction to Schedule

The schedule field specifies when the k8s-cleaner should run its logic to identify and potentially delete or update matching resources.

It adheres to the `Cron syntax`, which is a widely adopted scheduling language for tasks and events.

The Cron syntax consists of five fields, separated by spaces, each representing a specific part of the scheduling period: minute, hour, day of month, month and day of week, in that order.

The k8s-cleaner is able to accept the below schedule formats.

- **Standard crontab** specs, e.g. "* * * * ?"
- **Descriptors**, e.g. "@midnight", "@every 1h30m"

## Example

!!! example ""

    ```yaml
    ```