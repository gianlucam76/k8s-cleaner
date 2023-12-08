The Kubernetes controller __Pruner__ efficiently removes or updates stale resources in your cluster. It's designed to handle different types of resources and supports label filtering and Lua-based selection criteria.

It provides a flexible and customizable approach to identifying and removing or updating outdated resources, helping to maintain a clean and efficient Kubernetes environment. The ability to select resources based on labels and utilize Lua-based selection criteria further enhances its applicability to various scenarios.

## Removing All Secrets

To remove all Secrets from the test namespace every day at 1 AM, use the following YAML configuration:

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Pruner
metadata:
  name: pruner-sample
spec:
  schedule: "* 1 * * *"
  staleResources:
  - namespace: test
    kind: Secret
    group: ""
    version: v1
    action: Delete
```

## Selecting Resources with Label Filters

__Pruner__ can select resources based on their labels. For example, the following configuration removes all Deployment instances in the __test__ namespace that have both __serving=api__ and __environment!=production__ labels:

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Pruner
metadata:
  name: pruner-sample1
spec:
  schedule: "* 0 * * *"
  staleResources:
  - namespace: test
    kind: Deployment
    group: "apps"
    version: v1
    labelFilters:
    - key: serving
      operation: Equal
      value: api
    - key: environment
      operation: Different
      value: prouction
    action: Delete
```

## Using Lua for Advanced Selection

__Pruner__ allows you to define __Lua__ functions named ``evaluate`` for customized selection criteria. This function receives the resource object as obj.

For instance, the following configuration selects all Service instances in the foo namespace that expose port ``443`` or ``8443``:

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Pruner
metadata:
  name: pruner-sample2
spec:
  schedule: "* 0 * * *"
  staleResources:
  - namespace: foo
    kind: Service
    group: ""
    version: v1
    evaluate: |
      function evaluate()
        hs = {}
        hs.matching = false
        if obj.spec.ports ~= nil then
          for _,p in pairs(obj.spec.ports) do
            if p.port == 443 or p.port == 8443 then
              hs.matching = true
            end
          end
        end
        return hs
      end
    action: Delete
```

## Updating Resources

Besides removing stale resources, __Pruner__ also enables you to update existing resources. This feature empowers you to dynamically modify resource configurations based on specific criteria. For instance, you can replace outdated labels with updated ones, or alter resource settings to align with changing requirements.

Consider the scenario where you want to update Service objects in the foo namespace to use __version2__ apps. 
The __evaluate__ function allows you to select resources, Services in the __foo__ namespace pointing to ``version1``  apps. 
The __trasnform__ function will change any such a resources, by updating ``obj.spec.selector["app"]`` to ``version2``.

```yaml
apiVersion: apps.projectsveltos.io/v1alpha1
kind: Pruner
metadata:
  name: pruner-sample3
spec:
  schedule: "* 0 * * *"
  staleResources:
  - namespace: foo
    kind: Service
    group: ""
    version: v1
    evaluate: |
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
    transform: |
      function transform()
        hs = {}
        obj.spec.selector["app"] = "version2"
        hs.resource = obj
        return hs
        end
    action: Transform
```