// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { useSignal } from '@preact/signals';
import { reportsData } from '../../app';
import { ReportCard } from '../reports/ReportCard';
import { FilterBar } from '../reports/FilterBar';

export function ReportsPage() {
  const reports = reportsData.value;
  const filters = useSignal({ cleaner: '', kind: '', namespace: '' });

  if (!reports) {
    return (
      <div class="space-y-4">
        <div class="skeleton h-10 w-64 rounded" />
        {[1, 2, 3].map((i) => (
          <div key={i} class="skeleton h-20 rounded-lg" />
        ))}
      </div>
    );
  }

  // Apply filters
  const filtered = reports.filter((r) => {
    if (filters.value.cleaner && r.name !== filters.value.cleaner) return false;
    if (filters.value.kind || filters.value.namespace) {
      // Check if report has any resources matching the kind/namespace filter
      const hasMatch = r.resources.some((res) => {
        if (filters.value.kind && res.kind !== filters.value.kind) return false;
        if (filters.value.namespace && res.namespace !== filters.value.namespace) return false;
        return true;
      });
      // Show reports with matching resources, or reports with no resources if no kind/ns filter would exclude them
      if (!hasMatch && r.resources.length > 0) return false;
      if (r.resources.length === 0 && (filters.value.kind || filters.value.namespace)) return false;
    }
    return true;
  });

  // Sort: flagged first
  const sorted = [...filtered].sort((a, b) => b.resources.length - a.resources.length);

  const totalFlagged = sorted.reduce((sum, r) => sum + r.resources.length, 0);

  return (
    <div class="space-y-4">
      {/* Header + filters */}
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-300">
          Reports ({sorted.length})
          {totalFlagged > 0 && (
            <span class="ml-2 text-xs font-normal text-red-500 dark:text-red-400">
              {totalFlagged} flagged resources
            </span>
          )}
        </h2>
        <FilterBar
          filters={filters.value}
          onFilterChange={(f) => { filters.value = f; }}
          reports={reports}
        />
      </div>

      {/* Status bar */}
      {sorted.length > 0 && (
        <div class="flex h-2 rounded-full overflow-hidden bg-gray-200 dark:bg-gray-700">
          {(() => {
            const withFindings = sorted.filter((r) => r.resources.length > 0).length;
            const clean = sorted.length - withFindings;
            const flaggedPct = (withFindings / sorted.length) * 100;
            return (
              <>
                {withFindings > 0 && (
                  <div
                    class="bg-red-500 transition-all"
                    style={{ width: `${flaggedPct}%` }}
                    title={`${withFindings} with findings`}
                  />
                )}
                {clean > 0 && (
                  <div
                    class="bg-green-500 transition-all"
                    style={{ width: `${100 - flaggedPct}%` }}
                    title={`${clean} clean`}
                  />
                )}
              </>
            );
          })()}
        </div>
      )}

      {/* Report cards */}
      {sorted.length > 0 ? (
        <div class="space-y-3">
          {sorted.map((r) => (
            <ReportCard key={r.name} report={r} />
          ))}
        </div>
      ) : (
        <div class="panel text-center py-12">
          <svg class="w-12 h-12 mx-auto text-gray-300 dark:text-gray-600 mb-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
          </svg>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {filters.value.cleaner || filters.value.kind || filters.value.namespace
              ? 'No reports match the current filters'
              : 'No reports found'}
          </p>
          {(filters.value.cleaner || filters.value.kind || filters.value.namespace) && (
            <button
              class="text-xs text-blue-600 hover:underline mt-2"
              onClick={() => { filters.value = { cleaner: '', kind: '', namespace: '' }; }}
            >
              Clear filters
            </button>
          )}
        </div>
      )}
    </div>
  );
}
