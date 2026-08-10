// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

export function LibraryCard({ entry, isSelected, onSelect }) {
  function handleClick(e) {
    if (e.target.closest('button') || e.target.closest('a')) return;
    onSelect(entry.id);
  }

  function handleKeyDown(e) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onSelect(entry.id);
    }
  }

  return (
    <div
      role="button"
      tabIndex={0}
      aria-expanded={isSelected}
      class={`panel-interactive cursor-pointer select-none ${
        isSelected ? 'ring-2 ring-blue-500 ring-offset-1 dark:ring-offset-gray-900' : ''
      }`}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
    >
      <h3 class="text-sm font-semibold mb-1">{entry.title}</h3>
      <p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mb-2">{entry.description}</p>

      {entry.selectors?.length > 0 && (
        <div class="flex flex-wrap gap-1 mb-2">
          {entry.selectors.map((s, i) => (
            <span key={i} class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300">
              {s.kind}
            </span>
          ))}
        </div>
      )}

      <div class="flex items-center gap-1 text-[11px] text-gray-400 dark:text-gray-500">
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span class="font-mono">{entry.schedule}</span>
      </div>
    </div>
  );
}
