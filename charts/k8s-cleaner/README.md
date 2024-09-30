# K8s-cleaner

![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

Cleaner identifies, removes, or updates stale/orphaned or unhealthy resources to maintain a clean and efficient Kubernetes cluster

The chart is under active development and may contain bugs/unfinished documentation. Any testing/contributions are welcome! :)

**Homepage:** <https://github.com/gianlucam76/k8s-cleaner>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| gianlucam76 |  |  |
| oliverbaehler |  |  |

# Major Changes

Major Changes to functions are documented with the version affected. **Before upgrading the dependency version, check this section out!**

| **Change** | **Chart Version** | **Description** | **Commits/PRs** |
| :--------- | :---------------- | :-------------- | :-------------- |
|||||

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity |
| controller.args | object | `{}` | Controller ARguments |
| controller.image.pullPolicy | string | `"IfNotPresent"` | Controller Image pull policy |
| controller.image.registry | string | `"docker.io"` | Controller Image Registry |
| controller.image.repository | string | `"projectsveltos/k8s-cleaner"` | Controller Image Repository |
| controller.image.tag | string | `"v0.9.0"` | ControllerImage Tag |
| controller.livenessProbe | object | `{"enabled":true,"httpGet":{"path":"/healthz","port":"healthz","scheme":"HTTP"},"initialDelaySeconds":15,"periodSeconds":20}` | Controller LivenessProbe   |
| controller.ports[0].containerPort | int | `8443` |  |
| controller.ports[0].name | string | `"metrics"` |  |
| controller.ports[0].protocol | string | `"TCP"` |  |
| controller.ports[1].containerPort | int | `9440` |  |
| controller.ports[1].name | string | `"healthz"` |  |
| controller.ports[1].protocol | string | `"TCP"` |  |
| controller.readinessProbe | object | `{"enabled":true,"httpGet":{"path":"/readyz","port":"healthz","scheme":"HTTP"},"initialDelaySeconds":5,"periodSeconds":10}` | Controller ReadinessProbe |
| controller.resources | object | `{}` | Resource limits and requests for the controller |
| controller.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"enabled":true,"runAsNonRoot":true}` | Controller SecurityCOntext |
| controller.volumeMounts | list | `[]` | Controller VolumeMounts |
| crds.install | bool | `true` | Install the CustomResourceDefinitions (This also manages the lifecycle of the CRDs for update operations) |
| crds.keep | bool | `true` | Keep the CustomResourceDefinitions (when the chart is deleted) |
| fullnameOverride | string | `""` | Full name overwrite |
| historyLimit | int | `3` | The number of old ReplicaSets to retain for a Deployment (default=10) |
| imagePullSecrets | list | `[]` | ImagePullSecrets |
| nameOverride | string | `""` | Partial name overwrite |
| nodeSelector | object | `{}` | NodeSelector |
| podAnnotations | object | `{}` | Pod Annotations |
| podLabels | object | `{}` | Pod Labels |
| podSecurityContext | object | `{"enabled":true,"runAsNonRoot":true,"seccompProfile":{"type":"RuntimeDefault"}}` | Pod Security Context |
| rbac.create | bool | `true` | Create RBAC resources |
| replicaCount | int | `1` | Amount of replicas |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.automount | bool | `true` | Automatically mount a ServiceAccount's API credentials? |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| serviceMonitor.annotations | object | `{}` | Assign additional Annotations |
| serviceMonitor.enabled | bool | `false` | Enable ServiceMonitor |
| serviceMonitor.endpoint.interval | string | `"15s"` | Set the scrape interval for the endpoint of the serviceMonitor |
| serviceMonitor.endpoint.metricRelabelings | list | `[]` | Set metricRelabelings for the endpoint of the serviceMonitor |
| serviceMonitor.endpoint.relabelings | list | `[]` | Set relabelings for the endpoint of the serviceMonitor |
| serviceMonitor.endpoint.scrapeTimeout | string | `""` | Set the scrape timeout for the endpoint of the serviceMonitor |
| serviceMonitor.jobLabel | string | `"app.kubernetes.io/name"` | Set JobLabel for the serviceMonitor |
| serviceMonitor.labels | object | `{}` | Assign additional labels according to Prometheus' serviceMonitorSelector matching labels |
| serviceMonitor.matchLabels | object | `{}` | Change matching labels |
| serviceMonitor.namespace | string | `""` | Install the ServiceMonitor into a different Namespace, as the monitoring stack one (default: the release one) |
| serviceMonitor.targetLabels | list | `[]` | Set targetLabels for the serviceMonitor |
| tolerations | list | `[]` | Tolerations |
| volumes | list | `[]` | Additional volumes on the output Deployment definition. |
