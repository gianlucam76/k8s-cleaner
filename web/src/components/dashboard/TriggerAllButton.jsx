// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { signal } from '@preact/signals';
import { api } from '../../utils/fetch';
import { enableFastRefresh } from '../../app';
import { ENDPOINTS } from '../../utils/constants';

const showConfirm = signal(false);
const triggerState = signal('idle'); // 'idle' | 'loading' | 'success'
const progress = signal({ current: 0, total: 0 });

export function TriggerAllButton({ cleaners, readOnly }) {
  if (readOnly || !cleaners || cleaners.length === 0) return null;

  async function handleTriggerAll() {
    showConfirm.value = false;
    triggerState.value = 'loading';
    progress.value = { current: 0, total: cleaners.length };

    for (let i = 0; i < cleaners.length; i++) {
      try {
        await api(ENDPOINTS.trigger(cleaners[i].name), { method: 'POST' });
      } catch (err) {
        // Continue triggering others even if one fails
      }
      progress.value = { current: i + 1, total: cleaners.length };
    }

    triggerState.value = 'success';
    enableFastRefresh();
    setTimeout(() => { triggerState.value = 'idle'; }, 3000);
  }

  const s = triggerState.value;

  if (s === 'success') {
    return (
      <span class="inline-flex items-center gap-1.5 text-xs text-green-600 dark:text-green-400 font-medium">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
        </svg>
        All {progress.value.total} scans triggered
      </span>
    );
  }

  if (s === 'loading') {
    return (
      <span class="inline-flex items-center gap-1.5 text-xs text-blue-600 font-medium">
        <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        {progress.value.current}/{progress.value.total} triggered
      </span>
    );
  }

  return (
    <>
      <button onClick={() => { showConfirm.value = true; }} class="action-primary">
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182" />
        </svg>
        Trigger All
      </button>

      {showConfirm.value && (
        <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" role="dialog" aria-modal="true" aria-label="Trigger all scans confirmation" onClick={() => { showConfirm.value = false; }}>
          <div class="panel max-w-sm mx-4" onClick={(e) => e.stopPropagation()}>
            <h3 class="text-sm font-semibold mb-2">Trigger All Scans?</h3>
            <p class="text-xs text-gray-500 dark:text-gray-400 mb-4">
              This will trigger {cleaners.length} cleaners sequentially. Scans will run in the background.
            </p>
            <div class="flex items-center justify-end gap-2">
              <button onClick={() => { showConfirm.value = false; }} class="action border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700">
                Cancel
              </button>
              <button onClick={handleTriggerAll} class="action bg-blue-600 text-white border-blue-600 hover:bg-blue-700">
                Trigger {cleaners.length} Cleaners
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
