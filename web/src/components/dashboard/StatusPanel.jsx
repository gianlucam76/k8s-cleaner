// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

export function StatusPanel({ summary }) {
  return (
    <div class="panel">
      <div class="flex items-center gap-6 flex-wrap">
        <div>
          <div class="text-2xl font-bold">{summary.totalCleaners}</div>
          <div class="text-xs text-gray-500 dark:text-gray-400">Cleaners</div>
        </div>
        <div>
          <div class={`text-2xl font-bold ${summary.totalFlaggedResources > 0 ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400'}`}>
            {summary.totalFlaggedResources}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400">Flagged Resources</div>
        </div>
        <div>
          <div class={`text-2xl font-bold ${summary.cleanersWithFindings > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-green-600 dark:text-green-400'}`}>
            {summary.cleanersWithFindings}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400">With Findings</div>
        </div>
        {summary.lastScanTime && (
          <div class="ml-auto text-xs text-gray-400">
            Last scan: {new Date(summary.lastScanTime).toLocaleTimeString()}
          </div>
        )}
      </div>
    </div>
  );
}
