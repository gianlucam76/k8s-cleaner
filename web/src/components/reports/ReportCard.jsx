// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { useSignal } from '@preact/signals';

export function ReportCard({ report }) {
  const expanded = useSignal(false);
  const hasFlagged = report.resources.length > 0;

  function toggle(e) {
    if (e.target.closest('a')) return;
    expanded.value = !expanded.value;
  }

  function onKeyDown(e) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      expanded.value = !expanded.value;
    }
  }

  return (
    <div
      role="button"
      tabIndex={0}
      aria-expanded={expanded.value}
      class={`panel-interactive border-l-4 cursor-pointer select-none ${
        hasFlagged ? 'border-l-red-500' : 'border-l-green-500'
      }`}
      onClick={toggle}
      onKeyDown={onKeyDown}
    >
      {/* Header */}
      <div class="flex items-center justify-between mb-1">
        <div class="flex items-center gap-2 min-w-0">
          <svg
            class={`w-3.5 h-3.5 text-gray-400 dark:text-gray-500 transition-transform flex-shrink-0 ${expanded.value ? 'rotate-90' : ''}`}
            fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
          </svg>
          <h3 class="text-sm font-semibold truncate">{report.name}</h3>
        </div>
        {hasFlagged ? (
          <span class="badge-alert flex-shrink-0">{report.resources.length} flagged</span>
        ) : (
          <span class="badge-ok flex-shrink-0">Clean</span>
        )}
      </div>

      <div class="text-xs text-gray-500 dark:text-gray-400 ml-[1.375rem]">
        Action: {report.action}
        {hasFlagged && (
          <span class="ml-2">
            Kinds: {[...new Set(report.resources.map((r) => r.kind))].join(', ')}
          </span>
        )}
      </div>

      {/* Expanded resource list */}
      {expanded.value && hasFlagged && (
        <div class="mt-3 pt-3 border-t border-gray-200 dark:border-gray-600">
          <div class="text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
            Flagged Resources ({report.resources.length})
          </div>
          <div class="space-y-1.5 max-h-72 overflow-y-auto">
            {report.resources.map((res, i) => (
              <div key={i} class="flex items-start gap-2 text-xs">
                <span class="font-mono bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400 px-1.5 py-0.5 rounded text-[10px] flex-shrink-0">
                  {res.kind}
                </span>
                <div class="min-w-0">
                  <div class="text-gray-700 dark:text-gray-300 truncate">
                    {res.namespace && (
                      <span class="text-gray-400 dark:text-gray-500">{res.namespace}/</span>
                    )}
                    {res.name}
                  </div>
                  {res.message && (
                    <div class="text-[10px] text-gray-400 dark:text-gray-500 truncate">
                      {res.message}
                    </div>
                  )}
                </div>
                {res.apiVersion && (
                  <span class="text-[10px] text-gray-400 dark:text-gray-500 ml-auto flex-shrink-0">
                    {res.apiVersion}
                  </span>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {expanded.value && !hasFlagged && (
        <div class="mt-3 pt-3 border-t border-gray-200 dark:border-gray-600">
          <div class="text-xs text-gray-400 dark:text-gray-500 italic">
            No flagged resources in latest scan
          </div>
        </div>
      )}
    </div>
  );
}
