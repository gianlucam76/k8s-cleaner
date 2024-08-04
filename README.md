
[![CI](https://github.com/gianlucam76/k8s-cleaner/actions/workflows/main.yaml/badge.svg)](https://github.com/gianlucam76/k8s-cleaner/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/gianlucam76/k8s-cleaner)](https://goreportcard.com/report/github.com/gianlucam76/k8s-cleaner)
[![Slack](https://img.shields.io/badge/join%20slack-%23projectsveltos-brighteen)](https://join.slack.com/t/projectsveltos/shared_invite/zt-1hraownbr-W8NTs6LTimxLPB8Erj8Q6Q)
[![License](https://img.shields.io/badge/license-Apache-blue.svg)](LICENSE)
[![Twitter Follow](https://img.shields.io/twitter/follow/projectsveltos?style=social)](https://twitter.com/projectsveltos)
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-2-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

<img src="https://raw.githubusercontent.com/gianlucam76/k8s-cleaner/main/assets/logo.png" width="200">

The Kubernetes controller __Cleaner__ identifies, removes, or updates stale/orphaned or unhealthy resources to maintain a clean and efficient Kubernetes cluster. It is designed to handle any Kubernetes resource types (including your own custom resources) and provides sophisticated filtering capabilities, including label-based selection and custom Lua-based criteria.
__Cleaner__ can also be used to identify unhealthy resources.

k8s-cleaner keeps you in the loop with handy notifications through:

1. <img src="assets/slack_logo.png" alt="Slack" width="30" />  [__Slack__](#slack-notification)
2. <img src="assets/webex_logo.png" alt="Webex" width="30" />  [__Webex__](#webex-notifications)
3. <img src="assets/discord_logo.png" alt="Discord" width="30" />  [__Discord__](#discord-notifications)
3. <img src="assets/teams_logo.svg" alt="Teams" width="30" />  [__Teams__](#teams-notifications)
4.  [__reports__](#cleaner-report)
  
Each notification contains list of all resources successfully deleted (or modified) by k8s-cleaner. Choose what works best for you!

Currently k8s-cleaner has rich set of working examples to **identify** and **list** **unused**:

- [ConfigMaps](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/configmaps)/[Secrets](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/secrets)
- [ClusterRoles](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/clusterroles)/[Roles](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/roles)
- [ServiceAccounts](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/service-accounts)
- [PersistentVolumes](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/persistent-volumes)/[PersistentVolumeClaims](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/persistent-volume-claims)
- [Deployments](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/deployments)/[StatefulSets](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/stateful-sets)
- [Identify resources based on annotation indicating the maximum lifespan or the expiration date](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources/time_based_delete)
- More

There are also examples to identify unhealthy resources:

  - [Pods Mounting Secrets with Old Content](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unhealthy-resources/pod-with-outdated-secrets): Detect pods that are not utilizing the most recent Secret data.
  - [Pods Using Expired Certificates](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unhealthy-resources/pod-with-expired-certificates): Pinpoint pods that are operating with expired security certificates.

## Features and Capabilities

1️⃣ **Schedule**: Specify the frequency at which the k8s-cleaner should scan the cluster and identify stale resources. Utilise the Cron syntax to define recurring schedules.

2️⃣ **DryRun**: Enable safe testing of the k8s-cleaner filtering logic without affecting actual resource configurations. Resources matching the criteria will get identified, but no changes will get applied.

3️⃣ **Label Filtering**: Select resources based on user-defined labels, filtering out unwanted or outdated components. Refine the selection based on label key, operation (equal, different, etc.), and value.

4️⃣ **Lua-based Selection Criteria**: Leverage the [Lua](https://lua.org/) scripting language to create complex and dynamic selection criteria, catering to specific resource management needs. Define custom logic to identify and handle stale resources. To validate the Cleaner configuration, have a look [here](#validate-the-cleaner-configuration).

5️⃣ **Notifications**: Stay informed! The k8s-cleaner keeps users in the loop about every cleaned-up resource, whether removed or optimized. Get detailed notification lists and pick your preferred channel: Slack, Webex, Discord, Teams or reports.

For a complete list of **features** with **examples**, have a look at the [link](https://gianlucam76.github.io/k8s-cleaner/getting_started/features/dryrun/dryrun/).

## Benefits

💪 **Resource Removal**: Efficiently remove stale resources from your cluster, reclaiming unused resources and improving resource utilisation.

💪 **Resource Updates**: Update outdated resources to ensure they align with the latest configurations and maintain consistent functionality.

💪 **Reduced Resource Bloat**: Minimize resource bloat and maintain a clean and organized cluster, improving overall performance and stability.

By combining the **flexibility** of **scheduling**, the **accuracy** of **label filtering**, the **power** of **Lua-based criteria**, and the ability to **remove** or **update** stale resources, the k8s-cleaner empowers users to effectively manage Kubernetes environments and optimise resource usage.

## How to work with us

- 👉 For feature requests and bugs, file an [issue](https://github.com/gianlucam76/k8s-cleaner/issues).
- 👉 To get updates [⭐️ star](https://github.com/gianlucam76/k8s-cleaner/stargazers) this repository.
- 👉 Working examples can be found in the [examples](https://github.com/gianlucam76/k8s-cleaner/tree/main/examples-unused-resources) section.

## Getting Started Guide
- [Complete Documentation](http://k8scleaner.projectsveltos.io/)

## Validate Cleaner Configuration 

To verify the correctness of the __Cleaner__ configuration, follow the comprehensive instructions provided in the documentation: [here](https://github.com/gianlucam76/k8s-cleaner/blob/main/internal/controller/executor/validate_transform/README.md) and [here](https://github.com/gianlucam76/k8s-cleaner/blob/main/internal/controller/executor/validate_transform/README.md).

In essence, the Cleaner YAML file alongside the YAML files representing matching and non-matching resources need to get provided, and then by executing the simple ```make ut``` command the resutls will appear. This will validate whether your configuration correctly identifies and manages the desired resources.

## Contribute

We encourage everyone to contribute to the example directory by adding new Cleaner configurations 💡. This will help the community benefit from different expertise and build a stronger knowledge base of the Cleaner use cases.

To add an example, simply create a new file in the example directory with a descriptive name and put your Cleaner configuration within the file. Once you've added your example, feel free to submit a pull request to share it with the community.

🤝 Together we can expand the range of Cleaner applications and make it an even more valuable tool for managing Kubernetes resources efficiently.

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


