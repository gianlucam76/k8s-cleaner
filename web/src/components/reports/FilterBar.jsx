// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

export function FilterBar({ filters, onFilterChange, cleaners }) {
  // Kinds come from cleaner selectors (what they're configured to scan)
  const allKinds = cleaners
    ? [...new Set(cleaners.flatMap((c) => (c.selectors || []).map((s) => s.kind)))].sort()
    : [];

  return (
    <div class="flex flex-wrap items-center gap-2">
      <select
        class="text-xs rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-2 py-1.5 text-gray-700 dark:text-gray-300"
        value={filters.kind}
        onChange={(e) => onFilterChange({ ...filters, kind: e.target.value })}
      >
        <option value="">All resource kinds</option>
        {allKinds.map((k) => (
          <option key={k} value={k}>{k}</option>
        ))}
      </select>

      {filters.kind && (
        <button
          class="text-xs text-blue-600 hover:underline"
          onClick={() => onFilterChange({ kind: '' })}
        >
          Clear
        </button>
      )}
    </div>
  );
}
