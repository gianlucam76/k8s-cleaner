---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Rollback
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

## Introduction to Rollback

`rollback` is an optional field on the Cleaner spec that captures the state of every resource right before a `Delete` or `Transform` action is applied to it, so the most recent execution can be reverted. Only the last execution can be rolled back: each run overwrites the previous one's captured data, the same way it overwrites the [Report](../../../reports/k8s-cleaner_reports.md) itself.

```yaml
spec:
  rollback:
    storage: Report
```

`storage` currently only supports `Report`, which stores the pre-action resource inline on the `Report` instance's `resourceInfo[].fullResource` field. No extra infrastructure (no PersistentVolume) is required.

### Requirements

Rollback data is only useful if it is durably persisted, and today the only place it is persisted is the `Report` instance. Because of that, a Cleaner with `rollback` set **must** also configure a `CleanerReport` [notification](../../../notifications/notifications.md). If it doesn't, the Cleaner refuses to run: no resource is scanned, matched, deleted, or transformed, and the reason is surfaced in `status.failureMessage`.

```bash
$ kubectl get cleaner unused-configmaps -o yaml
...
status:
  failureMessage: 'rollback is enabled but no CleanerReport notification is configured: rollback data would never be persisted'
```

### Ordering guarantee

Cleaner never deletes or transforms a resource before its pre-action state has been safely written to the `Report`. If persisting that snapshot fails for any reason, the run aborts before touching any resource, rather than risking a resource being changed with no way to revert it.

### What gets captured

- `rollback` has no effect when `action` is set to `Scan`: nothing is ever deleted or transformed, so there is nothing to revert.
- A single resource's captured state is capped at 256KB. Larger resources are skipped (rollback won't be available for them specifically), and a note is added to that resource's entry in the Report so it's clear why. This keeps the Report's total size bounded, together with [`blastRadiusLimit`](../blast_radius_limit/blast_radius_limit.md), which bounds how many resources a single run can affect.
- Captured resource bodies are never included in outgoing Slack/Webex/Discord/Teams/Telegram/SMTP notifications. They only ever live on the `Report` instance.

## Example - Rollback a Delete Action

!!! example "Cleaner Definition with Rollback and a CleanerReport Notification"

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: unused-configmaps
    spec:
      schedule: "0 * * * *"
      action: Delete
      rollback:
        storage: Report
      resourcePolicySet:
        resourceSelectors:
        - namespace: test
          kind: ConfigMap
          group: ""
          version: v1
          evaluate: |
            function evaluate()
              hs = {}
              hs.matching = true
              return hs
            end
      notifications:
      - name: report
        type: CleanerReport
    ```

## Example - Rollback a Transform Action

!!! example "Cleaner Definition with Rollback and a CleanerReport Notification"

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: scale-down-deployments
    spec:
      schedule: "0 * * * *"
      action: Transform
      rollback:
        storage: Report
      transform: |
        function transform()
          hs = {}
          obj.spec.replicas = 0
          hs.resource = obj
          return hs
        end
      resourcePolicySet:
        resourceSelectors:
        - namespace: test
          kind: Deployment
          group: "apps"
          version: v1
      notifications:
      - name: report
        type: CleanerReport
    ```

### Report Output

After either Cleaner runs, the Report instance carries the captured resource alongside the usual fields:

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Report
metadata:
  name: unused-configmaps
spec:
  action: Delete
  resourceInfo:
  - resource:
      apiVersion: v1
      kind: ConfigMap
      name: my-configmap
      namespace: test
    message: '. time: 2026-01-15 10:00:00 +0000 UTC'
    fullResource: <base64-encoded ConfigMap as it existed right before deletion>
```

### Triggering a Rollback

The easiest way is from the [web dashboard](../../install/install.md#web-dashboard): open the **Reports** page, expand a report whose action was `Delete` or `Transform`, and click **Rollback**. A confirmation dialog explains what will happen before anything is reverted. Like every other mutating action, the button is hidden when the dashboard is running in [read-only mode](../../install/install.md#web-dashboard).

The button calls the same API endpoint you can reach directly, by name of the Cleaner (the Report shares its name with the Cleaner that produced it):

```bash
$ kubectl port-forward -n projectsveltos svc/k8s-cleaner-web 9080:9080
$ curl -X POST http://localhost:9080/api/v1/reports/unused-configmaps/rollback
[
  {
    "kind": "ConfigMap",
    "namespace": "test",
    "name": "my-configmap",
    "success": true
  }
]
```

Each entry in the response reflects one captured resource: `success: true` means it was recreated (`Delete` action) or restored to its previous state (`Transform` action); `success: false` comes with a `message` explaining why, for instance when no rollback data was captured for that resource because it exceeded the size cap.
