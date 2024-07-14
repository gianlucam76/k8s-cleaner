---
title: k8s-cleaner - Kubernetes Controller that identifies, removes, or updates stale/orphaned or unhealthy resources
description: Welcome to the k8s-cleaner notifications page
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

## Introduction to Notifications

Notifications is an easy way of k8s-cleaner to keep users in the loop about relevant updates. Each notification contains a list of successfully deleted or modified resources by the k8s-cleaner.

The below notifications are available.
- **Slack**
- **Webex**
- **Discord**
- **Teams**

## Slack Notifications Example

### Kubernetes Secret

To allow the k8s-cleaner to write messages or upload files to a channel, we need to create a Kubernetes secret

```bash
$ kubectl create secret generic slack --from-literal=SLACK_TOKEN=<YOUR TOKEN> --from-literal=SLACK_CHANNEL_ID=<YOUR CHANNEL ID>
```


!!! example "Slack Notifications Defintion"

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: cleaner-with-slack-notifications
    spec:
      schedule: "0 * * * *"
      action: Delete # Delete matching resources
      resourcePolicySet:
        resourceSelectors:
        - namespace: test
          kind: Deployment
          group: "apps"
          version: v1
      notifications:
      - name: slack
        type: Slack
        notificationRef:
          apiVersion: v1
          kind: Secret
          name: slack
          namespace: default
    ```

Anytime this Cleaner instance is processed, a Slack message is sent containing all the resources that were deleted by k8s-cleaner.

## Webex Notifications Example

### Kubernetes Secret

To allow the k8s-cleaner to write messages or upload files to a channel, we need to create a Kubernetes secret

```bash
$ kubectl create secret generic webex --from-literal=WEBEX_TOKEN=<YOUR TOKEN> --from-literal=WEBEX_ROOM_ID=<YOUR WEBEX CHANNEL ID>
```


!!! example "Webex Notifications Defintion"

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: cleaner-with-webex-notifications
    spec:
      schedule: "0 * * * *"
      action: Delete # Delete matching resources
      resourcePolicySet:
        resourceSelectors:
        - namespace: test
          kind: Deployment
          group: "apps"
          version: v1
      notifications:
      - name: webex
        type: Webex
        notificationRef:
          apiVersion: v1
          kind: Secret
          name: webex
          namespace: default
    ```

## Discord Notifications Example

### Kubernetes Secret

To allow the k8s-cleaner to write messages or upload files to a channel, we need to create a Kubernetes secret

```bash
$ kubectl create secret generic discord --from-literal=DISCORD_TOKEN=<YOUR TOKEN> --from-literal=DISCORD_CHANNEL_ID=<YOUR DISCORD CHANNEL ID>
```


!!! example "Discord Notifications Defintion"

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: cleaner-with-discord-notifications
    spec:
      schedule: "0 * * * *"
      action: Delete # Delete matching resources
      resourcePolicySet:
        resourceSelectors:
        - namespace: test
          kind: Deployment
          group: "apps"
          version: v1
      notifications:
      - name: discord
        type: Discord
        notificationRef:
          apiVersion: v1
          kind: Secret
          name: discord
          namespace: default
    ```

## Teams Notifications Example

### Kubernetes Secret

To allow the k8s-cleaner to write messages or upload files to a channel, we need to create a Kubernetes secret

```bash
$ kubectl create secret generic teams --from-literal=TEAMS_WEBHOOK_URL="<your URL>"
```


!!! example "Teams Notifications Defintion"

    ```yaml
    ---
    apiVersion: apps.projectsveltos.io/v1alpha1
    kind: Cleaner
    metadata:
      name: cleaner-with-teams-notifications
    spec:
      schedule: "0 * * * *"
      action: Delete # Delete matching resources
      resourcePolicySet:
        resourceSelectors:
        - namespace: test
          kind: Deployment
          group: "apps"
          version: v1
      notifications:
      - name: teams
        type: Teams
        notificationRef:
          apiVersion: v1
          kind: Secret
          name: teams
          namespace: default
    ```
