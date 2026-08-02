// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { useSignal } from '@preact/signals';
import { api } from '../../utils/fetch';
import { configData, enableFastRefresh } from '../../app';
import { ENDPOINTS } from '../../utils/constants';

// 'idle' | 'loading' | 'success' | 'error'
export function RollbackButton({ cleanerName, resourceCount }) {
  const state = useSignal('idle');
  const message = useSignal('');
  const showConfirm = useSignal(false);
  const readOnly = configData.value?.readOnly ?? true;

  if (readOnly) return null;

  async function handleRollback(e) {
    e.stopPropagation();
    showConfirm.value = false;
    if (state.value === 'loading') return;

    state.value = 'loading';
    try {
      const results = await api(ENDPOINTS.rollback(cleanerName), { method: 'POST' });
      const succeeded = results.filter((r) => r.success).length;
      const failed = results.length - succeeded;
      state.value = failed === 0 ? 'success' : 'error';
      message.value = failed === 0
        ? `Rolled back ${succeeded}/${results.length}`
        : `Rolled back ${succeeded}/${results.length}, ${failed} failed`;
      enableFastRefresh();
      setTimeout(() => { state.value = 'idle'; }, failed === 0 ? 3000 : 6000);
    } catch (err) {
      state.value = 'error';
      message.value = err.message;
      setTimeout(() => { state.value = 'idle'; }, 6000);
    }
  }

  const s = state.value;

  if (s === 'success') {
    return (
      <span class="inline-flex items-center gap-1 text-xs text-green-600 dark:text-green-400 font-medium">
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
        </svg>
        {message.value}
      </span>
    );
  }

  if (s === 'error') {
    return (
      <span class="inline-flex items-center gap-1 text-xs text-red-600 dark:text-red-400 max-w-full">
        <svg class="w-3.5 h-3.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
        <span class="truncate">{message.value}</span>
      </span>
    );
  }

  return (
    <>
      <button
        onClick={(e) => { e.stopPropagation(); showConfirm.value = true; }}
        class="action"
        disabled={s === 'loading'}
      >
        {s === 'loading' ? (
          <>
            <svg class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            Rolling back...
          </>
        ) : (
          <>
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 15L3 9m0 0l6-6M3 9h12a6 6 0 010 12h-3" />
            </svg>
            Rollback
          </>
        )}
      </button>

      {showConfirm.value && (
        <div
          class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
          role="dialog"
          aria-modal="true"
          aria-label="Rollback confirmation"
          onClick={(e) => { e.stopPropagation(); showConfirm.value = false; }}
        >
          <div class="panel max-w-sm mx-4" onClick={(e) => e.stopPropagation()}>
            <h3 class="text-sm font-semibold mb-2">Roll Back Last Execution?</h3>
            <p class="text-xs text-gray-500 dark:text-gray-400 mb-4">
              This recreates the {resourceCount} resource{resourceCount === 1 ? '' : 's'} deleted, or restores{' '}
              {resourceCount === 1 ? 'it' : 'them'} to its previous state if transformed, by the last run of{' '}
              <span class="font-mono">{cleanerName}</span>. Only the most recent execution can be rolled back.
            </p>
            <div class="flex items-center justify-end gap-2">
              <button
                onClick={(e) => { e.stopPropagation(); showConfirm.value = false; }}
                class="action border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700"
              >
                Cancel
              </button>
              <button onClick={handleRollback} class="action bg-blue-600 text-white border-blue-600 hover:bg-blue-700">
                Roll Back
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
