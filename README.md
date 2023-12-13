[![CI](https://github.com/gianlucam76/k8s-cleaner/actions/workflows/main.yaml/badge.svg)](https://github.com/gianlucam76/k8s-cleaner/actions)
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-1-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->
[![Go Report Card](https://goreportcard.com/badge/github.com/gianlucam76/k8s-cleaner)](https://goreportcard.com/report/github.com/gianlucam76/k8s-cleaner)
[![Slack](https://img.shields.io/badge/join%20slack-%23projectsveltos-brighteen)](https://join.slack.com/t/projectsveltos/shared_invite/zt-1hraownbr-W8NTs6LTimxLPB8Erj8Q6Q)
[![License](https://img.shields.io/badge/license-Apache-blue.svg)](LICENSE)
[![Twitter Follow](https://img.shields.io/twitter/follow/projectsveltos?style=social)](https://twitter.com/projectsveltos)

The Kubernetes controller __Cleaner__ proactively identifies, removes, or updates stale resources to maintain a clean and efficient Kubernetes environment. It's designed to handle any Kubernetes resource types (including your own custom resources) and provides sophisticated filtering capabilities, including label-based selection and custom Lua-based criteria.

- 👉 For feature requests and bugs, file an [issue](https://github.com/gianlucam76/k8s-cleaner/issues).
- 👉 To get updates [⭐️ star](https://github.com/gianlucam76/k8s-cleaner/stargazers) this repository.

# Contribute to Cleaner Examples

We encourage you to contribute to the _example_ directory by adding your own Cleaner configurations 💡. This will help the community benefit from your expertise and build a stronger knowledge base of Cleaner use cases.

To add an example, simply create a new file in the example directory with a descriptive name and put your Cleaner configuration within the file. Once you've added your example, feel free to submit a pull request to share it with the community.

🤝 together we can expand the range of Cleaner applications and make it an even more valuable tool for managing Kubernetes resources efficiently.

## Flexibility and Customization:

1️⃣ **Schedule**: Specify the frequency at which the Cleaner should scan the cluster and identify stale resources. Utilize the Cron syntax to define recurring schedules.

2️⃣ **DryRun**: Enable safe testing of the Cleaner's filtering logic without affecting actual resource configurations. Resources matching the criteria will be identified, but no changes will be applied.

3️⃣ **Label Filtering**: Select resources based on user-defined labels, filtering out unwanted or outdated components. Refine the selection based on label key, operation (equal, different, etc.), and value.

4️⃣ **Lua-based Selection Criteria**: Leverage Lua scripting to create complex and dynamic selection criteria, catering to specific resource management needs. Define custom logic to identify and handle stale resources.

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
  dryRun: false
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
  resourcePolicySet:
    resourceSelectors:
    - namespace: foo
      kind: Deployment
      group: ""
      version: v1
    - namespace: foo
      kind: HorizontalPodAutoscaler
      group: "autoscaling"
      version: v2beta1
    action: Delete # Delete matching resources
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
                table.insert(deploymentWithNoAutoscaler, deployment)
                break
            end
        end

        hs.resources = deploymentWithNoAutoscaler
        return hs
      end
```

## DryRun 

To preview which resources match the __Cleaner__'s criteria, set the __DryRun__ flag to true. The Cleaner will still execute its logic but will not actually delete or update any resources. To identify matching resources, search the controller logs for the message "resource is a match for cleaner".

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: cleaner-sample1
spec:
  schedule: "* 0 * * *" # Runs every day at midnight
  dryRun: true  # Set to true to preview matching resources
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
    action: Delete
```

By setting DryRun to true, you can safely test the Cleaner's filtering logic without affecting your actual deployment configurations. Once you're confident in the filtering criteria, you can set DryRun back to false to enable automatic resource deletion.

## Schedule

The __schedule__ field specifies when the __Cleaner__ should run its logic to identify and potentially delete or update matching resources. It adheres to the Cron syntax, which is a widely adopted scheduling language for tasks and events.

The Cron syntax consists of five fields, separated by spaces, each representing a specific part of the scheduling period: minute, hour, day of month, month and day of week, in that order. 

It also accepts

- Standard crontab specs, e.g. "* * * * ?"
- Descriptors, e.g. "@midnight", "@every 1h30m"

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


