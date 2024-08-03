---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Store Resource Yaml
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

## Store Resource Yaml

This is a small section describing how to store resources before the k8s-cleaner deletes or modifies them. The k8s-cleaner has an optional fiel called `StoreResourcePath`.

When this option is set, the k8s-cleaner will dump all the maching resources before any modification (update and/or deletion) is performed.

The maching resource will be stored in the below directory.

```bash
/<__StoreResourcePath__ value>/<Cleaner name>/<resourceNamespace>/<resource Kind>/<resource Name>.yaml
```
## Example

!!! example ""

    ```yaml
    ```