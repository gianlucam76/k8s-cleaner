// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { useSignal } from '@preact/signals';
import { useEffect } from 'preact/hooks';
import { api } from '../../utils/fetch';
import { ENDPOINTS } from '../../utils/constants';
import { LibraryCard } from '../library/LibraryCard';
import { LibraryDetailPanel } from '../library/LibraryDetailPanel';

const CATEGORIES = [
  { key: 'unused-resources', label: 'Unused Resources' },
  { key: 'unhealthy-resources', label: 'Unhealthy Resources' },
];

// Same color for every group heading - just needs to not be gray. Reuses the
// bg-100/text-700 (light) and bg-900/30/text-400 (dark) formula the existing
// .badge-ok/.badge-alert/.badge-warn classes in index.css already use.
const GROUP_COLOR = 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-400';

// Splits an already-ordered list of entries into consecutive-same-group runs.
// The backend guarantees entries sharing a group are contiguous within a
// category (see library_test.go), so this never needs to re-sort.
function groupConsecutive(entries) {
  const groups = [];
  for (const entry of entries) {
    const last = groups[groups.length - 1];
    if (last && last.name === entry.group) {
      last.items.push(entry);
    } else {
      groups.push({ name: entry.group, items: [entry] });
    }
  }
  return groups;
}

export function LibraryPage() {
  const entries = useSignal(null);
  const selected = useSignal(null);
  const activeCategory = useSignal(CATEGORIES[0].key);
  const search = useSignal('');
  const activeKind = useSignal(null);

  useEffect(() => {
    let cancelled = false;
    api(ENDPOINTS.library)
      .then((data) => { if (!cancelled) entries.value = data; })
      .catch((err) => console.error('Failed to load library:', err));
    return () => { cancelled = true; };
  }, []);

  if (!entries.value) {
    return (
      <div class="space-y-4">
        <div class="skeleton h-6 w-40 rounded" />
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} class="skeleton h-24 rounded-lg" />
          ))}
        </div>
      </div>
    );
  }

  function switchCategory(key) {
    activeCategory.value = key;
    activeKind.value = null; // kinds differ per category, so the filter no longer applies
  }

  function onSelect(id) {
    selected.value = selected.value === id ? null : id;
  }

  const inCategory = entries.value.filter((e) => e.category === activeCategory.value);
  const kinds = [...new Set(inCategory.flatMap((e) => (e.selectors || []).map((s) => s.kind)))].sort();

  const term = search.value.trim().toLowerCase();
  const filtered = inCategory.filter((e) => {
    if (activeKind.value && !(e.selectors || []).some((s) => s.kind === activeKind.value)) return false;
    if (!term) return true;
    return e.title.toLowerCase().includes(term) || e.description.toLowerCase().includes(term);
  });
  const groups = groupConsecutive(filtered);

  return (
    <div class="space-y-6">
      <div>
        <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Library</h2>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
          Curated Cleaner recipes. Pick one, set its schedule and notifications, and post it to the cluster.
        </p>
      </div>

      <div class="flex items-center gap-1 border-b border-gray-200 dark:border-gray-700" role="tablist">
        {CATEGORIES.map(({ key, label }) => {
          const count = entries.value.filter((e) => e.category === key).length;
          const active = activeCategory.value === key;
          return (
            <button
              key={key}
              role="tab"
              aria-selected={active}
              onClick={() => switchCategory(key)}
              class={`px-4 py-2 text-xs font-semibold border-b-2 -mb-px transition-colors ${
                active
                  ? 'border-blue-600 text-blue-600 dark:border-blue-400 dark:text-blue-400'
                  : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
              }`}
            >
              {label} <span class="text-gray-400 dark:text-gray-500">({count})</span>
            </button>
          );
        })}
      </div>

      <div class="flex flex-wrap items-center gap-3">
        <div class="relative w-full sm:w-64">
          <svg class="w-4 h-4 absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
          </svg>
          <input
            type="text"
            value={search.value}
            onInput={(e) => { search.value = e.currentTarget.value; }}
            placeholder="Search recipes..."
            aria-label="Search recipes"
            class="w-full pl-8 pr-2 py-1.5 text-xs rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-gray-900 dark:text-white"
          />
        </div>

        {kinds.length > 0 && (
          <div class="flex flex-wrap items-center gap-1">
            {kinds.map((kind) => {
              const active = activeKind.value === kind;
              return (
                <button
                  key={kind}
                  onClick={() => { activeKind.value = active ? null : kind; }}
                  aria-pressed={active}
                  class={`px-2 py-0.5 rounded-full text-[10px] font-mono border transition-colors ${
                    active
                      ? 'bg-blue-600 border-blue-600 text-white'
                      : 'border-gray-300 dark:border-gray-600 text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
                  }`}
                >
                  {kind}
                </button>
              );
            })}
          </div>
        )}
      </div>

      {filtered.length === 0 ? (
        <div class="panel text-center py-12">
          <p class="text-sm text-gray-500 dark:text-gray-400">No recipes match</p>
        </div>
      ) : (
        <div class="space-y-8">
          {groups.map((group) => (
            <div key={group.name}>
              <div class="flex items-center gap-2 mb-3">
                <span class={`badge ${GROUP_COLOR}`}>{group.name}</span>
                <span class="text-xs text-gray-400 dark:text-gray-500">{group.items.length}</span>
              </div>
              <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4">
                {group.items.map((entry) => (
                  <LibraryCard
                    key={entry.id}
                    entry={entry}
                    isSelected={entry.id === selected.value}
                    onSelect={onSelect}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      <LibraryDetailPanel
        entryId={selected.value}
        isOpen={!!selected.value}
        onClose={() => { selected.value = null; }}
      />
    </div>
  );
}
