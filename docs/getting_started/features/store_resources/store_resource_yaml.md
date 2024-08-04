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

This is a small section describing how to store resources before the k8s-cleaner deletes or modifies them. The k8s-cleaner has an optional field called `StoreResourcePath`.

When this option is set, the k8s-cleaner will dump all the maching resources before any modification (update and/or deletion) is performed.

The maching resource will be stored in the below directory.

```bash
/<__StoreResourcePath__ value>/<Cleaner name>/<resourceNamespace>/<resource Kind>/<resource Name>.yaml
```
## Example - Unsused ConfigMap

### Step 1 - Create PersistentVolumeClaim
!!! example "PersistentVolumeClaim"

	```yaml
	apiVersion: v1
	kind: PersistentVolumeClaim
	metadata:
	  name: cleaner-pvc
	  namespace: projectsveltos
	  labels:
	    app: k8s-cleaner
	spec:
	  storageClassName: standard
	  accessModes:
	    - ReadWriteOnce
	resources:
	  requests:
	    storage: 2Gi
	```

The above YAML definition will create a `PersistentVolumeClaim` of 2Gi. In case more storage is required, simply update the YAML definition.

```bash
$ kubectl apply -f "pvc.yaml"
```

### Step 2 - Update k8s-cleaner-controller Deployment

The next is to update the `k8s-cleaner-controller` deployment located in the `projectsveltos` namespace. Then, we will define the `PersistentVolumeClaim` and the actual storage location.

```bash
$ kubectl get deploy -n projectsveltos                        
NAME                     READY   UP-TO-DATE   AVAILABLE   AGE
k8s-cleaner-controller   1/1     1            1           10m
$ kubectl edit deploy k8s-cleaner-controller -n projectsveltos
```

!!! example "k8s-cleaner-controller"

	```yaml
    volumes:
    - name: volume
      persistentVolumeClaim:
        claimName: cleaner-pvc

      volumeMounts:
      - mountPath: /pvc/
        name: volume
	```

The YAML defition files will be stored in `/pvc/`.

### Step 3 - Cleaner Resource Creation

In step 3, we will create a Cleaner Resource and define the deletion of any unused `configMap` resources based on a cron job. To store the resources before performing any deletions, we will add the argument ` storeResourcePath: "/pvc/"` and store the resources inside the `/pvc/` directory.

!!! example "Cleaner Resource"

	```yaml
	apiVersion: apps.projectsveltos.io/v1alpha1
	kind: Cleaner
	metadata:
	  name: unused-configmaps
	spec:
	  storeResourcePath: "/pvc/"
	  schedule: "* 0 * * *"
	  action: Delete
	```

When cleaner find the ununsed `ConfigMap`, it will first store the resource definition and then delete the actual resource.

### Validation

```bash
docker exec -i cleaner-management-worker ls /var/local-path-provisioner/pvc-8314c600-dc54-4e23-a796-06b73080f589_projectsveltos_cleaner-pvc
unused-configmaps

/var/local-path-provisioner/pvc-8314c600-dc54-4e23-a796-06b73080f589_projectsveltos_cleaner-pvc/unused-configmaps/test/ConfigMap:
kube-root-ca.crt.yaml
my-configmap.yaml
```