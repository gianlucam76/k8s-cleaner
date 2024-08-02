
[![CI](https://github.com/gianlucam76/k8s-cleaner/actions/workflows/main.yaml/badge.svg)](https://github.com/gianlucam76/k8s-cleaner/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/gianlucam76/k8s-cleaner)](https://goreportcard.com/report/github.com/gianlucam76/k8s-cleaner)
[![Slack](https://img.shields.io/badge/join%20slack-%23projectsveltos-brighteen)](https://join.slack.com/t/projectsveltos/shared_invite/zt-1hraownbr-W8NTs6LTimxLPB8Erj8Q6Q)
[![License](https://img.shields.io/badge/license-Apache-blue.svg)](LICENSE)
[![Twitter Follow](https://img.shields.io/twitter/follow/projectsveltos?style=social)](https://twitter.com/projectsveltos)
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-2-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

<img src="https://raw.githubusercontent.com/gianlucam76/k8s-cleaner/main/assets/logo.png" width="200">

The Kubernetes controller __Cleaner__ identifies, removes, or updates stale/orphaned or unhealthy resources to maintain a clean and efficient Kubernetes cluster. It's designed to handle any Kubernetes resource types (including your own custom resources) and provides sophisticated filtering capabilities, including label-based selection and custom Lua-based criteria.
__Cleaner__ can also be used to identify unhealthy resources.

k8s-cleaner keeps you in the loop with handy notifications through:

1. <img src="assets/slack_logo.png" alt="Slack" width="30" />  [__Slack__](#slack-notification)
2. <img src="assets/webex_logo.png" alt="Webex" width="30" />  [__Webex__](#webex-notifications)
3. <img src="assets/discord_logo.png" alt="Discord" width="30" />  [__Discord__](#discord-notifications)
3. <img src="assets/teams_logo.svg" alt="Teams" width="30" />  [__Teams__](#teams-notifications)
4.  [__reports__](#cleaner-report)
  
Each notification contains list of all resources successfully deleted (or modified) by k8s-cleaner. Choose what works best for you!

- 👉 For feature requests and bugs, file an [issue](https://github.com/gianlucam76/k8s-cleaner/issues).
- 👉 To get updates [⭐️ star](https://github.com/gianlucam76/k8s-cleaner/stargazers) this repository.
- 👉 Working examples can be found in the [examples](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources) section.

Currently k8s-cleaner has rich set of working examples to identify and list unused:

- [ConfigMaps](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/configmaps)/[Secrets](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/secrets)
- [ClusterRoles](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/clusterroles)/[Roles](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/roles)
- [ServiceAccounts](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/service-accounts)
- [PersistentVolumes](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/persistent-volumes)/[PersistentVolumeClaims](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/persistent-volume-claims)
- [Deployments](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/deployments)/[StatefulSets](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/stateful-sets)
- [Identify resources based on annotation indicating the maximum lifespan or the expiration date](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/time_based_delete)
- many more 

There are also examples to identify unhealthy resources:

- [Pods Mounting Secrets with Old Content](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unhealthy-resources/pod-with-outdated-secrets): Detect pods that are not utilizing the most recent Secret data.
- [Pods Using Expired Certificates](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unhealthy-resources/pod-with-expired-certificates): Pinpoint pods that are operating with expired security certificates.

## Manage and Automate Resource Operations

K8s-cleaner doesn't just help you identify unused or unhealthy resources; it can also automate various operations to enhance your cluster's efficiency and management. The examples-operations directory showcases practical scripts and configurations covering key tasks:

#### Scaling Deployments/DaemonSets/StatefulSets with Nightly Downtime:

This [example](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-automated-operations/scheduled-scaling) demonstrates how to automatically scale down Deployments, DaemonSets, and StatefulSets with a specified annotation at a desired time (e.g., 8 PM nightly).
Before scaling down, the replica count is stored in another annotation for later retrieval.
At the configured scale-up time (e.g., 8 AM), resources are restored to their original replica count, ensuring efficient resource utilization during off-peak hours

# Contribute to Cleaner Examples

We encourage you to contribute to the _example_ directory by adding your own Cleaner configurations 💡. This will help the community benefit from your expertise and build a stronger knowledge base of Cleaner use cases.

To add an example, simply create a new file in the example directory with a descriptive name and put your Cleaner configuration within the file. Once you've added your example, feel free to submit a pull request to share it with the community.

🤝 together we can expand the range of Cleaner applications and make it an even more valuable tool for managing Kubernetes resources efficiently.

## Flexibility and Customization:

1️⃣ **Schedule**: Specify the frequency at which the Cleaner should scan the cluster and identify stale resources. Utilize the Cron syntax to define recurring schedules.

2️⃣ **DryRun**: Enable safe testing of the Cleaner's filtering logic without affecting actual resource configurations. Resources matching the criteria will be identified, but no changes will be applied. Set Action to __Scan__.

3️⃣ **Label Filtering**: Select resources based on user-defined labels, filtering out unwanted or outdated components. Refine the selection based on label key, operation (equal, different, etc.), and value.

4️⃣ **Lua-based Selection Criteria**: Leverage Lua scripting to create complex and dynamic selection criteria, catering to specific resource management needs. Define custom logic to identify and handle stale resources.

5️⃣ **Notifications**: Stay informed! k8s-cleaner keeps you in the loop about every cleaned-up resource, whether removed or optimized. Get detailed notification lists and pick your preferred channel: Slack, Webex, Discord, Teams or reports.

## Maintaining a Clean and Efficient Cluster:

💪 **Resource Removal**: Efficiently remove stale resources from your cluster, reclaiming unused resources and improving resource utilization.

💪 **Resource Updates**: Update outdated resources to ensure they align with the latest configurations and maintain consistent functionality.

💪 **Reduced Resource Bloat**: Minimize resource bloat and maintain a clean and organized cluster, improving overall performance and stability.

By combining the flexibility of scheduling, the accuracy of label filtering, the power of Lua-based criteria, and the ability to remove or update stale resources, Cleaner empowers users to effectively manage their Kubernetes environments and optimize resource usage.

## Deploying the K8s Cleaner
To deploy the __k8s-cleaner__ to your Kubernetes cluster, run the following command:

```
kubectl apply -f https://raw.githubusercontent.com/gianlucam76/k8s-cleaner/main/manifest/manifest.yaml
```

## Removing Unwanted Secrets

To remove all Secrets from the test namespace every day at 1 AM, use the following YAML configuration:

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: cleaner-sample
spec:
  schedule: "* 1 * * *" # Runs every day at 1 AM
  resourcePolicySet:
    resourceSelectors:
    - namespace: test
      kind: Secret
      group: ""
      version: v1
  action: Delete # Deletes matching Secrets
```

This configuration instructs the Cleaner to scan the test namespace every day at 1 AM, identify all Secrets, and effectively eliminate them, ensuring a clean and organized cluster.

## Selecting Resources with Label Filters

__Cleaner__ can select resources based on their labels, enabling precise resource management. For instance, to eliminate Deployments in the __test__ namespace with both ``serving=api`` and ``environment!=production`` labels, follow this YAML configuration:

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: cleaner-sample1
spec:
  schedule: "* 0 * * *" # Executes every day at midnight
  resourcePolicySet:
    resourceSelectors:
    - namespace: test
      kind: Deployment
      group: "apps"
      version: v1
      labelFilters:
      - key: serving
        operation: Equal
        value: api # Identifies Deployments with "serving" label set to "api"
      - key: environment
        operation: Different
        value: prouction # Identifies Deployments with "environment" label different from "production"
  action: Delete # Deletes matching Deployments
```

By utilizing label filters, you can refine the scope of resource management, ensuring that only specific resources are targeted for removal or update. This targeted approach helps maintain a clean and organized Kubernetes environment without affecting unintended resources.

## Namespace Selector

__Cleaner__ can select resources based on namespaces matching a label selector. For instance, to eliminate Deployments in all the namespaces with labels `env`:`qa` and `tag`: `dev` just set the `NamespaceSelector` field like following YAML configuration:

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: cleaner-sample1
spec:
  schedule: "* 0 * * *" # Executes every day at midnight
  resourcePolicySet:
    resourceSelectors:
    - kind: Deployment
      group: "apps"
      version: v1
      namespaceSelector: "env=qa, tag=dev" 
  action: Delete # Deletes matching Deployments
```

## Using Lua for Advanced Selection

__Cleaner__ extends its capabilities by enabling the use of __Lua__ scripts for defining advanced selection criteria. These Lua functions, named __evaluate__, receive the resource object as __obj__ and allow for complex and dynamic filtering rules.

 For example, the following YAML configuration utilizes a Lua script to select all Services in the __foo__ namespace that expose port __443__ or __8443__:

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: cleaner-sample2
spec:
  schedule: "* 0 * * *"
  resourcePolicySet:
    resourceSelectors:
    - namespace: foo
      kind: Service
      group: ""
      version: v1
      evaluate: |
        function evaluate()
          hs = {}
          hs.matching = false -- Initialize matching flag
          if obj.spec.ports ~= nil then
            for _,p in pairs(obj.spec.ports) do -- Iterate through the ports
              if p.port == 443 or p.port == 8443 then -- Check if port is 443 or 8443
                hs.matching = true -- Set matching flag to true
              end
            end
          end
          return hs
        end
  action: Delete
```

By leveraging Lua scripts, Cleaner empowers users to define complex and dynamic selection criteria, catering to specific resource management needs. This flexibility enables accurate and targeted identification of stale resources, ensuring effective resource utilization and maintenance of a clean Kubernetes environment.

Here is another example removing Pods in __completed__ state:

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: completed-pods
spec:
  schedule: "* 0 * * *"
  resourcePolicySet:
    resourceSelectors:
    - kind: Pod
      group: ""
      version: v1
      evaluate: |
        function evaluate()
          hs = {}
          hs.matching = false
          if obj.status.conditions ~= nil then
            for _, condition in ipairs(obj.status.conditions) do
              if condition.reason == "PodCompleted" and condition.status == "True" then
                hs.matching = true
              end
            end
          end
          return hs
        end
  action: Delete
```

## Updating Resources

Beyond removing stale resources, __Cleaner__ also facilitates the dynamic updating of existing resource configurations. This capability allows you to modify resource specifications based on specific criteria, ensuring alignment with evolving requirements and maintaining resource consistency.

Consider the scenario where you want to update Service objects in the foo namespace to use __version2__ apps. 

1. The __evaluate__ function allows you to select resources, Services in the __foo__ namespace pointing to ``version1``  apps. 
2. The __trasnform__ function will change any such a resources, by updating ``obj.spec.selector["app"]`` to ``version2``.

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: cleaner-sample3
spec:
  schedule: "* 0 * * *"
  resourcePolicySet:
    resourceSelectors:
    - namespace: foo
      kind: Service
      group: ""
      version: v1
      evaluate: |
        -- Define how resources will be selected 
        function evaluate()
          hs = {}
          hs.matching = false
          if obj.spec.selector ~= nil then
            if obj.spec.selector["app"] == "version1" then
              hs.matching = true
            end
          end
          return hs
          end
  action: Transform # Update matching resources
  transform: |
      -- Define how resources will be updated
      function transform()
        hs = {}
        obj.spec.selector["app"] = "version2"
        hs.resource = obj
        return hs
        end
```

## Remove resources based on their creation timestamp

The provided code snippet defines a Cleaner instance that targets Pods with a creation timestamp older than 24 hours and instructs
Cleaner to delete them automatically. This feature helps maintain resource cleanliness and avoid unnecessary resource usage in Kubernetes clusters.
By automatically deleting these resources at the end of the working day, you can significantly reduce the overall resource consumption and associated costs.

```yaml
# This Cleaner instance finds any Pod that:
# - has been running for longer than 24 hour
# and instruct Cleaner to delete this Pod.
#
# If you want to filter Pods based on namespace =>
#     - kind: Pod
#       group: ""
#       version: v1
#       namespace: <YOUR_NAMESPACE>
#
# If you want to filter Pods based on labels =>
#     - kind: Pod
#       group: ""
#       version: v1
#       labelFilters:
#       - key: app
#         operation: Equal
#         value: nginx 
#       - key: environment
#         operation: Different
#         value: prouction 
#
# If you need further filtering modify `function evaluate` you can access any
# field of obj
#
# If you want to remove any other resource including your own custom resources
# replace kind/group/version
#
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: pods-from-job
spec:
  schedule: "* 0 * * *"
  resourcePolicySet:
    resourceSelectors:
    - kind: Pod
      group: ""
      version: v1
      evaluate: |
        --  Convert creationTimestamp "2023-12-12T09:35:56Z"
        function convertTimestampString(timestampStr)
          local convertedTimestamp = string.gsub(
            timestampStr,
            '(%d+)-(%d+)-(%d+)T(%d+):(%d+):(%d+)Z',
            function(y, mon, d, h, mi, s)
              return os.time({
                year = tonumber(y),
                month = tonumber(mon),
                day = tonumber(d),
                hour = tonumber(h),
                min = tonumber(mi),
                sec = tonumber(s)
              })
            end
          )
          return convertedTimestamp
        end

        function evaluate()
          hs = {}
          hs.matching = false

          -- any resource older than this time will be removed 
          local removeAfterHour = 24

          currentTime = os.time()

          creationTimestamp = convertTimestampString(obj.metadata.creationTimestamp)

          hs.message = creationTimestamp
          print('creationTimestamp: ' .. creationTimestamp)
          print('currentTime: ' .. currentTime)

          timeDifference = os.difftime(currentTime, tonumber(creationTimestamp))

          print('timeDifference: ' .. timeDifference)

          -- if resource has been running for over 24 hours
          if timeDifference > removeAfterHour*60*60 then
            hs.matching = true
          end

          return hs
        end
  action: Delete
```

## Considering resources of different types together

Occasionally, it's necessary to examine resources of distinct types simultaneously. Imagine wanting to eliminate all Deployment instances that aren't backed by an Autoscaler instance. Cleaner allows you to do this. By employing __resourceSelector__, you can select all Deployment and Autoscaler instances.

Next, define __aggregatedSelection__. AggregatedSelection will be given all instances collected by Cleaner using resourceSelector, in this situation, all Deployment and Autoscaler instances in the foo namespace.

 aggregatedSelection will then refine this set further. The filtered resources will then be subjected to Cleaner's action.

```yaml
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

## Delete Kubernetes resources on a configured time to live or expiration date

Finds resources that have the __cleaner/ttl__ annotation, which specifies their maximum lifespan. Deletes resources that have lived longer than their specified TTL.
YAML can be found [here](https://github.com/gianlucam76/k8s-cleaner/blob/main/examples-unused-resources/time_based_delete/delete_resource_based_on_ttl_annotation.yaml).

Find resources that have the __cleaner_expires__ annotation, which specifies their expiration date. Deletes resources that have that have surpassed their expiration date.
YAML can be found [here](https://github.com/gianlucam76/k8s-cleaner/blob/main/examples-unused-resources/time_based_delete/delete_resource_based_on_expire_date.yaml).

## DryRun 

To preview which resources match the __Cleaner__'s criteria, set the __Action__ field to _Scan_. The Cleaner will still execute its logic but will not actually delete or update any resources. To identify matching resources, search the controller logs for the message "resource is a match for cleaner".

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: cleaner-sample1
spec:
  schedule: "* 0 * * *" # Runs every day at midnight
  resourcePolicySet:
    resourceSelectors:
    - namespace: test
      kind: Deployment
      group: "apps"
      version: v1
      labelFilters:
      - key: serving
        operation: Equal
        value: api # Match deployments with the "serving" label set to "api"
      - key: environment
        operation: Different
        value: prouction # Match deployments with the "environment" label different from "production"
  action: Scan
```

By setting __Action__ to _Scan_, you can safely test the Cleaner's filtering logic without affecting your actual deployment configurations. Once you're confident in the filtering criteria, you can set _Action_ to delete or modify.

## Schedule

The __schedule__ field specifies when the __Cleaner__ should run its logic to identify and potentially delete or update matching resources. It adheres to the Cron syntax, which is a widely adopted scheduling language for tasks and events.

The Cron syntax consists of five fields, separated by spaces, each representing a specific part of the scheduling period: minute, hour, day of month, month and day of week, in that order. 

It also accepts

- Standard crontab specs, e.g. "* * * * ?"
- Descriptors, e.g. "@midnight", "@every 1h30m"

## Notifications

k8s-cleaner keeps you in the loop with handy notifications through Slack, Webex, Discord, Teams or reports. Choose what works best for you!

### Slack Notification

Your app needs permission to:

1. write messages to a channel
2. upload files to a channel

Create a Kubernetes Secret:

```bash
kubectl create secret generic slack --from-literal=SLACK_TOKEN=<YOUR TOKEN> --from-literal=SLACK_CHANNEL_ID=<YOUR CHANNEL ID>                           
```

Set then the notifications field of a Cleaner instance

```yaml
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

### Webex Notifications

```bash
 kubectl create secret generic webex --from-literal=WEBEX_TOKEN=<YOUR TOKEN> --from-literal=WEBEX_ROOM_ID=<YOUR WEBEX CHANNEL ID>
 ```

Set then the notifications field of a Cleaner instance

```yaml
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

### Discord Notifications

```bash
 kubectl create secret generic discord --from-literal=DISCORD_TOKEN=<YOUR TOKEN> --from-literal=DISCORD_CHANNEL_ID=<YOUR DISCORD CHANNEL ID>
 ```

Set then the notifications field of a Cleaner instance

```yaml
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

### Teams Notifications

```bash
 kubectl create secret generic teams --from-literal=TEAMS_WEBHOOK_URL="<your URL>"
```

Set then the notifications field of a Cleaner instance

```yaml
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

### Cleaner Report

To instruct k8s-cleaner to generate a report with all resources deleted (or modified) set the notification fields:

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: cleaner-with-report
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
  - name: report
    type: CleanerReport
```

k8s-cleaner will create a __Report__ instance (name of the __Report__ instance is same as the Cleaner instance)

```
kubectl get report           
NAME              AGE
cleaner-sample3   51m
```

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Report
metadata:
  creationTimestamp: "2023-12-17T17:05:00Z"
  generation: 2
  name: cleaner-with-report
  resourceVersion: "1625"
  uid: dda9a231-9a51-4133-aeb5-f0520feb8746
spec:
  action: Delete
  message: 'time: 2023-12-17 17:07:00.394736089 +0000 UTC m=+129.172023518'
  resources:
  - apiVersion: apps/v1
    kind: Deployment
    name: my-nginx-deployment
    namespace: test
```

### Store Resource YAML

Sometimes it is convenient to store resources before Cleaner deletes/modifies those.

Cleaner has an optional field __StoreResourcePath__. When set, Cleaner will dump all matching resources before any modification
(delete or updated) was done. 

Matching resources will be stored

```
/<__StoreResourcePath__ value>/<Cleaner name>/<resourceNamespace>/<resource Kind>/<resource Name>.yaml
```

## Validate Your Cleaner Configuration 

To verify the correctness of your __Cleaner__ configuration, follow the comprehensive instructions provided in the documentation: [here](https://github.com/gianlucam76/k8s-cleaner/blob/main/internal/controller/executor/validate_transform/README.md) and [here](https://github.com/gianlucam76/k8s-cleaner/blob/main/internal/controller/executor/validate_transform/README.md).

In essence, you'll need to provide your Cleaner YAML file, along with YAML files representing matching and non-matching resources, and execute the simple ```make ut``` command. This will validate whether your configuration correctly identifies and manages the desired resources.

## Code of Conduct

This project adheres to the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md)

By participating, you are expected to honor this code.

## Contributors

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://projectsveltos.github.io/sveltos/"><img src="https://avatars.githubusercontent.com/u/52940363?v=4?s=100" width="100px;" alt="Gianluca Mardente"/><br /><sub><b>Gianluca Mardente</b></sub></a><br /><a href="https://github.com/gianlucam76/k8s-cleaner/commits?author=gianlucam76" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://keybase.io/oliverbaehler"><img src="https://avatars.githubusercontent.com/u/26610571?v=4?s=100" width="100px;" alt="Oliver Bähler"/><br /><sub><b>Oliver Bähler</b></sub></a><br /><a href="https://github.com/gianlucam76/k8s-cleaner/commits?author=oliverbaehler" title="Code">💻</a></td>
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->


