The Kubernetes controller __Cleaner__ proactively identifies, removes, or updates stale resources to maintain a clean and efficient Kubernetes environment. It's designed to handle any Kubernetes resource types (including your own custom resources) and provides sophisticated filtering capabilities, including label-based selection and custom Lua-based criteria.

## Flexibility and Customization:

- **Schedule**: Specify the frequency at which the Cleaner should scan the cluster and identify stale resources. Utilize the Cron syntax to define recurring schedules.

- **DryRun**: Enable safe testing of the Cleaner's filtering logic without affecting actual resource configurations. Resources matching the criteria will be identified, but no changes will be applied.

- **Label Filtering**: Select resources based on user-defined labels, filtering out unwanted or outdated components. Refine the selection based on label key, operation (equal, different, etc.), and value.

- **Lua-based Selection Criteria**: Leverage Lua scripting to create complex and dynamic selection criteria, catering to specific resource management needs. Define custom logic to identify and handle stale resources.

## Maintaining a Clean and Efficient Cluster:

- **Resource Removal**: Efficiently remove stale resources from your cluster, reclaiming unused resources and improving resource utilization.

- **Resource Updates**: Update outdated resources to ensure they align with the latest configurations and maintain consistent functionality.

- **Reduced Resource Bloat**: Minimize resource bloat and maintain a clean and organized cluster, improving overall performance and stability.

By combining the flexibility of scheduling, the accuracy of label filtering, the power of Lua-based criteria, and the ability to remove or update stale resources, Cleaner empowers users to effectively manage their Kubernetes environments and optimize resource usage.

## Removing Unwanted Secrets

To remove all Secrets from the test namespace every day at 1 AM, use the following YAML configuration:

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: cleaner-sample
spec:
  schedule: "* 1 * * *" # Runs every day at 1 AM
  matchingResources:
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
  matchingResources:
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
  matchingResources:
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
  matchingResources:
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
  matchingResources:
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
