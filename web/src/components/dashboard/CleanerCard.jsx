// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

export function CleanerCard({ cleaner, isSelected, onSelect }) {
  const hasFlagged = cleaner.flaggedCount > 0;
  const borderColor = hasFlagged ? 'border-l-red-500' : 'border-l-green-500';

  function handleClick(e) {
    if (e.target.closest('button') || e.target.closest('a')) return;
    onSelect(cleaner.name);
  }

  function handleKeyDown(e) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onSelect(cleaner.name);
    }
  }

  return (
    <div
      role="button"
      tabIndex={0}
      aria-expanded={isSelected}
      class={`panel-interactive border-l-4 ${borderColor} cursor-pointer select-none ${
        isSelected ? 'ring-2 ring-blue-500 ring-offset-1 dark:ring-offset-gray-900' : ''
      }`}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
    >
      <div class="flex items-start justify-between mb-2">
        <div class="flex items-center gap-2 min-w-0">
          <svg
            class={`w-3.5 h-3.5 text-gray-400 dark:text-gray-500 transition-transform flex-shrink-0 ${isSelected ? 'rotate-90' : ''}`}
            fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
          </svg>
          <h3 class="text-sm font-semibold truncate">{cleaner.name}</h3>
        </div>
        {hasFlagged ? (
          <span class="badge-alert flex-shrink-0">{cleaner.flaggedCount}</span>
        ) : (
          <span class="badge-ok flex-shrink-0">Clean</span>
        )}
      </div>

      <div class="text-xs text-gray-500 dark:text-gray-400 space-y-0.5 ml-[1.375rem]">
        <div class="flex items-center gap-3">
          <span title="Schedule">{cleaner.schedule}</span>
          <span class="text-gray-300 dark:text-gray-600">|</span>
          <span>{cleaner.action}</span>
        </div>
        <div>
          {cleaner.selectors?.map((s) => s.kind).join(', ')}
        </div>
        {cleaner.lastRunTime && (
          <div class="text-gray-400 dark:text-gray-500">
            Last: {new Date(cleaner.lastRunTime).toLocaleTimeString()}
            {cleaner.nextScheduleTime && (
              <> / Next: {new Date(cleaner.nextScheduleTime).toLocaleTimeString()}</>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
