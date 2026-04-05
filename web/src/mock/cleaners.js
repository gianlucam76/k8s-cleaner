// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

const now = Date.now();

export const mockCleaners = [
  {
    name: 'unused-configmaps',
    schedule: '0 6-22 * * *',
    action: 'Scan',
    lastRunTime: new Date(now - 3600000).toISOString(),
    nextScheduleTime: new Date(now + 3600000).toISOString(),
    flaggedCount: 23,
    selectors: [
      { group: '', version: 'v1', kind: 'ConfigMap' },
      { group: '', version: 'v1', kind: 'Pod' },
      { group: 'apps', version: 'v1', kind: 'Deployment' },
    ],
    luaScript: '-- Finds ConfigMaps not referenced by any workload\nfunction evaluate()\n  ...\nend',
  },
  {
    name: 'unused-secrets',
    schedule: '5 6-22 * * *',
    action: 'Scan',
    lastRunTime: new Date(now - 3300000).toISOString(),
    nextScheduleTime: new Date(now + 3900000).toISOString(),
    flaggedCount: 12,
    selectors: [
      { group: '', version: 'v1', kind: 'Secret' },
      { group: '', version: 'v1', kind: 'Pod' },
    ],
  },
  {
    name: 'pvc-scan',
    schedule: '15 6-22 * * *',
    action: 'Scan',
    lastRunTime: new Date(now - 2700000).toISOString(),
    nextScheduleTime: new Date(now + 4500000).toISOString(),
    flaggedCount: 8,
    selectors: [
      { group: '', version: 'v1', kind: 'Pod' },
      { group: '', version: 'v1', kind: 'PersistentVolumeClaim' },
    ],
  },
  {
    name: 'deployment-with-zero-replicas',
    schedule: '20 6-22 * * *',
    action: 'Scan',
    lastRunTime: new Date(now - 2400000).toISOString(),
    nextScheduleTime: new Date(now + 5100000).toISOString(),
    flaggedCount: 4,
    selectors: [{ group: 'apps', version: 'v1', kind: 'Deployment' }],
  },
  {
    name: 'deployments-not-gitops',
    schedule: '25 6-22 * * *',
    action: 'Scan',
    lastRunTime: new Date(now - 2100000).toISOString(),
    nextScheduleTime: new Date(now + 5400000).toISOString(),
    flaggedCount: 0,
    selectors: [{ group: 'apps', version: 'v1', kind: 'Deployment' }],
  },
  {
    name: 'cnpg-orphan-resources',
    schedule: '30 6-22 * * *',
    action: 'Scan',
    lastRunTime: new Date(now - 1800000).toISOString(),
    nextScheduleTime: new Date(now + 5700000).toISOString(),
    flaggedCount: 0,
    selectors: [
      { group: 'postgresql.cnpg.io', version: 'v1', kind: 'ScheduledBackup' },
      { group: 'postgresql.cnpg.io', version: 'v1', kind: 'Cluster' },
    ],
  },
  {
    name: 'cnpg-orphan-prometheusrules',
    schedule: '35 6-22 * * *',
    action: 'Scan',
    lastRunTime: new Date(now - 1500000).toISOString(),
    nextScheduleTime: new Date(now + 6000000).toISOString(),
    flaggedCount: 0,
    selectors: [
      { group: 'monitoring.coreos.com', version: 'v1', kind: 'PrometheusRule' },
      { group: 'postgresql.cnpg.io', version: 'v1', kind: 'Cluster' },
    ],
  },
  {
    name: 'helm-not-gitops',
    schedule: '40 6-22 * * *',
    action: 'Scan',
    lastRunTime: new Date(now - 1200000).toISOString(),
    nextScheduleTime: new Date(now + 6300000).toISOString(),
    flaggedCount: 0,
    selectors: [{ group: 'apps', version: 'v1', kind: 'Deployment' }],
  },
  {
    name: 'secrets-non-infisical',
    schedule: '10 6-22 * * *',
    action: 'Scan',
    lastRunTime: new Date(now - 900000).toISOString(),
    nextScheduleTime: new Date(now + 6600000).toISOString(),
    flaggedCount: 0,
    selectors: [{ group: '', version: 'v1', kind: 'Secret' }],
  },
  {
    name: 's3bkp-missing-secrets',
    schedule: '45 6-22 * * *',
    action: 'Scan',
    lastRunTime: new Date(now - 600000).toISOString(),
    nextScheduleTime: new Date(now + 6900000).toISOString(),
    flaggedCount: 0,
    selectors: [
      { group: 'apps', version: 'v1', kind: 'Deployment' },
      { group: 'apps', version: 'v1', kind: 'StatefulSet' },
    ],
  },
  {
    name: 'krelay-delete',
    schedule: '30 5 * * *',
    action: 'Delete',
    lastRunTime: new Date(now - 86400000).toISOString(),
    nextScheduleTime: new Date(now + 43200000).toISOString(),
    flaggedCount: 0,
    selectors: [
      { group: '', version: 'v1', kind: 'ConfigMap' },
      { group: '', version: 'v1', kind: 'Service' },
    ],
  },
];
