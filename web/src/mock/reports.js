// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

export const mockReports = [
  {
    name: 'unused-configmaps',
    action: 'Scan',
    resources: [
      { kind: 'ConfigMap', namespace: 'default', name: 'old-app-config', apiVersion: 'v1', message: 'Not referenced by any workload' },
      { kind: 'ConfigMap', namespace: 'monitoring', name: 'stale-dashboard', apiVersion: 'v1', message: 'Not referenced by any workload' },
      { kind: 'ConfigMap', namespace: 'kube-system', name: 'legacy-coredns-config', apiVersion: 'v1', message: 'Not referenced by any workload' },
    ],
  },
  {
    name: 'unused-secrets',
    action: 'Scan',
    resources: [
      { kind: 'Secret', namespace: 'default', name: 'old-tls-cert', apiVersion: 'v1', message: 'Not referenced by any workload or Ingress TLS' },
      { kind: 'Secret', namespace: 'staging', name: 'orphan-pull-secret', apiVersion: 'v1', message: 'Not referenced by any workload' },
    ],
  },
  {
    name: 'pvc-scan',
    action: 'Scan',
    resources: [
      { kind: 'PersistentVolumeClaim', namespace: 'default', name: 'data-old-postgres-0', apiVersion: 'v1', message: 'Not mounted by any Pod' },
      { kind: 'PersistentVolumeClaim', namespace: 'staging', name: 'cache-redis', apiVersion: 'v1', message: 'Not mounted by any Pod' },
    ],
  },
  {
    name: 'deployment-with-zero-replicas',
    action: 'Scan',
    resources: [
      { kind: 'Deployment', namespace: 'staging', name: 'old-api-server', apiVersion: 'apps/v1', message: 'Scaled to zero replicas' },
      { kind: 'Deployment', namespace: 'default', name: 'debug-tool', apiVersion: 'apps/v1', message: 'Scaled to zero replicas' },
    ],
  },
  { name: 'deployments-not-gitops', action: 'Scan', resources: [] },
  { name: 'cnpg-orphan-resources', action: 'Scan', resources: [] },
  { name: 'cnpg-orphan-prometheusrules', action: 'Scan', resources: [] },
  { name: 'helm-not-gitops', action: 'Scan', resources: [] },
  { name: 'secrets-non-infisical', action: 'Scan', resources: [] },
  { name: 's3bkp-missing-secrets', action: 'Scan', resources: [] },
  { name: 'krelay-delete', action: 'Delete', resources: [] },
];
