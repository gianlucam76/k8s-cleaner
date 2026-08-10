// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

export const mockLibrary = [
  {
    id: 'unused-clusterroles',
    category: 'unused-resources',
    group: 'RBAC',
    title: 'Unused ClusterRoles',
    description: 'Finds ClusterRole instances not referenced by any ClusterRoleBinding or RoleBinding.',
    schedule: '* 0 * * *',
    selectors: [
      { group: 'rbac.authorization.k8s.io', version: 'v1', kind: 'ClusterRole' },
      { group: 'rbac.authorization.k8s.io', version: 'v1', kind: 'ClusterRoleBinding' },
      { group: 'rbac.authorization.k8s.io', version: 'v1', kind: 'RoleBinding' },
    ],
  },
  {
    id: 'unused-roles',
    category: 'unused-resources',
    group: 'RBAC',
    title: 'Unused Roles',
    description: 'Finds Role instances, across all namespaces, not referenced by any RoleBinding.',
    schedule: '* 0 * * *',
    selectors: [
      { group: 'rbac.authorization.k8s.io', version: 'v1', kind: 'Role' },
      { group: 'rbac.authorization.k8s.io', version: 'v1', kind: 'RoleBinding' },
    ],
  },
  {
    id: 'unused-configmaps',
    category: 'unused-resources',
    group: 'Config & Secrets',
    title: 'Orphaned ConfigMaps',
    description: 'Finds ConfigMaps not used by any Pod through volumes, environment variables, or envFrom.',
    schedule: '* 0 * * *',
    selectors: [
      { group: '', version: 'v1', kind: 'ConfigMap' },
      { group: '', version: 'v1', kind: 'Pod' },
    ],
  },
  {
    id: 'completed-jobs',
    category: 'unused-resources',
    group: 'Workloads',
    title: 'Completed Jobs',
    description: 'Finds Jobs that completed successfully and have no running or pending Pods left.',
    schedule: '* 0 * * *',
    selectors: [{ group: 'batch', version: 'v1', kind: 'Job' }],
  },
  {
    id: 'ttl-based-cleaner',
    category: 'unused-resources',
    group: 'Time-Based',
    title: 'Resources Past Their TTL',
    description: 'Finds Deployments, StatefulSets, and Services carrying a cleaner/ttl annotation whose ' +
      'time-to-live has elapsed.',
    schedule: '0 * * * *',
    selectors: [
      { group: 'apps', version: 'v1', kind: 'Deployment' },
      { group: 'apps', version: 'v1', kind: 'StatefulSet' },
      { group: '', version: 'v1', kind: 'Service' },
    ],
  },
  {
    id: 'unhealthy-ingresses',
    category: 'unhealthy-resources',
    group: 'Broken References',
    title: 'Ingresses Referencing Missing Services',
    description: 'Finds Ingresses whose default backend, or a Service referenced via spec.rules, no longer exists.',
    schedule: '* 0 * * *',
    selectors: [
      { group: 'networking.k8s.io', version: 'v1', kind: 'Ingress' },
      { group: '', version: 'v1', kind: 'Service' },
    ],
  },
  {
    id: 'terminating-pods',
    category: 'unhealthy-resources',
    group: 'Pods',
    title: 'Pods Stuck Terminating',
    description: 'Finds Pods that have a deletionTimestamp set, i.e. stuck in a terminating state.',
    schedule: '*/5 * * * *',
    selectors: [{ group: '', version: 'v1', kind: 'Pod' }],
  },
];

export const mockLibraryDetail = {
  'unused-clusterroles': {
    id: 'unused-clusterroles',
    category: 'unused-resources',
    group: 'RBAC',
    title: 'Unused ClusterRoles',
    description: 'Finds ClusterRole instances not referenced by any ClusterRoleBinding or RoleBinding.',
    schedule: '* 0 * * *',
    selectors: [
      { group: 'rbac.authorization.k8s.io', version: 'v1', kind: 'ClusterRole' },
      { group: 'rbac.authorization.k8s.io', version: 'v1', kind: 'ClusterRoleBinding' },
      { group: 'rbac.authorization.k8s.io', version: 'v1', kind: 'RoleBinding' },
    ],
    luaScript: 'function evaluate()\n  local hs = {}\n  -- find ClusterRoles with no ClusterRoleBinding/RoleBinding\n  return hs\nend',
  },
};
