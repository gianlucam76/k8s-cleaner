// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { useSignal } from '@preact/signals';
import { summaryData, cleanersData, configData } from '../../app';
import { StatusPanel } from '../dashboard/StatusPanel';
import { CleanerCard } from '../dashboard/CleanerCard';
import { CleanerDetail } from '../dashboard/CleanerDetail';
import { SlidePanel } from '../dashboard/SlidePanel';
import { TriggerAllButton } from '../dashboard/TriggerAllButton';

export function DashboardPage() {
  const summary = summaryData.value;
  const cleaners = cleanersData.value;
  const readOnly = configData.value?.readOnly ?? true;
  const selected = useSignal(null);

  if (!summary) {
    return (
      <div class="space-y-4">
        <div class="skeleton h-24 rounded-lg" />
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} class="skeleton h-32 rounded-lg" />
          ))}
        </div>
      </div>
    );
  }

  const sorted = cleaners
    ? [...cleaners].sort((a, b) => {
        if (b.flaggedCount !== a.flaggedCount) return b.flaggedCount - a.flaggedCount;
        return a.name.localeCompare(b.name);
      })
    : [];

  const selectedCleaner = sorted.find((c) => c.name === selected.value);

  function onSelect(name) {
    selected.value = selected.value === name ? null : name;
  }

  return (
    <div class="space-y-6">
      <StatusPanel summary={summary} />

      <div class="flex items-center justify-between">
        <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-300">
          Cleaners ({sorted.length})
        </h2>
        <TriggerAllButton cleaners={sorted} readOnly={readOnly} />
      </div>

      {sorted.length > 0 ? (
        <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4">
          {sorted.map((c) => (
            <CleanerCard
              key={c.name}
              cleaner={c}
              isSelected={c.name === selected.value}
              onSelect={onSelect}
            />
          ))}
        </div>
      ) : (
        <div class="panel text-center py-12">
          <svg class="w-12 h-12 mx-auto text-gray-300 dark:text-gray-600 mb-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5m6 4.125l2.25 2.25m0 0l2.25 2.25M12 13.875l2.25-2.25M12 13.875l-2.25 2.25M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z" />
          </svg>
          <p class="text-sm text-gray-500 dark:text-gray-400">No cleaners configured</p>
          <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">
            Create Cleaner CRDs to start scanning for orphaned resources
          </p>
        </div>
      )}

      {/* Slide-over detail panel */}
      <SlidePanel
        isOpen={!!selectedCleaner}
        onClose={() => { selected.value = null; }}
        cleaner={selectedCleaner}
      >
        {selectedCleaner && <CleanerDetail cleaner={selectedCleaner} />}
      </SlidePanel>
    </div>
  );
}
