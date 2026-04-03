// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

export function FilterBar({ filters, onFilterChange, reports }) {
  // Extract unique values for dropdowns
  const cleanerNames = reports ? [...new Set(reports.map((r) => r.name))].sort() : [];
  const allKinds = reports
    ? [...new Set(reports.flatMap((r) => r.resources.map((res) => res.kind)))].sort()
    : [];
  const allNamespaces = reports
    ? [...new Set(reports.flatMap((r) => r.resources.map((res) => res.namespace).filter(Boolean)))].sort()
    : [];

  return (
    <div class="flex flex-wrap items-center gap-2">
      <select
        class="text-xs rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-2 py-1.5 text-gray-700 dark:text-gray-300"
        value={filters.cleaner}
        onChange={(e) => onFilterChange({ ...filters, cleaner: e.target.value })}
      >
        <option value="">All cleaners</option>
        {cleanerNames.map((n) => (
          <option key={n} value={n}>{n}</option>
        ))}
      </select>

      <select
        class="text-xs rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-2 py-1.5 text-gray-700 dark:text-gray-300"
        value={filters.kind}
        onChange={(e) => onFilterChange({ ...filters, kind: e.target.value })}
      >
        <option value="">All kinds</option>
        {allKinds.map((k) => (
          <option key={k} value={k}>{k}</option>
        ))}
      </select>

      <select
        class="text-xs rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-2 py-1.5 text-gray-700 dark:text-gray-300"
        value={filters.namespace}
        onChange={(e) => onFilterChange({ ...filters, namespace: e.target.value })}
      >
        <option value="">All namespaces</option>
        {allNamespaces.map((ns) => (
          <option key={ns} value={ns}>{ns}</option>
        ))}
      </select>

      {(filters.cleaner || filters.kind || filters.namespace) && (
        <button
          class="text-xs text-blue-600 hover:underline"
          onClick={() => onFilterChange({ cleaner: '', kind: '', namespace: '' })}
        >
          Clear filters
        </button>
      )}
    </div>
  );
}
