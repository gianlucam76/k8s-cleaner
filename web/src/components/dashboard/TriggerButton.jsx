// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { useSignal } from '@preact/signals';
import { api } from '../../utils/fetch';
import { enableFastRefresh } from '../../app';
import { ENDPOINTS } from '../../utils/constants';

// 'idle' | 'loading' | 'success' | 'error'
export function TriggerButton({ cleanerName, readOnly }) {
  const state = useSignal('idle');
  const errorMsg = useSignal('');

  if (readOnly) return null;

  async function handleTrigger(e) {
    e.stopPropagation();
    if (state.value === 'loading') return;

    state.value = 'loading';
    try {
      await api(ENDPOINTS.trigger(cleanerName), { method: 'POST' });
      state.value = 'success';
      enableFastRefresh();
      setTimeout(() => { state.value = 'idle'; }, 2000);
    } catch (err) {
      state.value = 'error';
      errorMsg.value = err.message;
      setTimeout(() => { state.value = 'idle'; }, 5000);
    }
  }

  const s = state.value;

  if (s === 'success') {
    return (
      <span class="inline-flex items-center gap-1 text-xs text-green-600 dark:text-green-400 font-medium">
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
        </svg>
        Triggered
      </span>
    );
  }

  if (s === 'error') {
    return (
      <span class="inline-flex items-center gap-1 text-xs text-red-600 dark:text-red-400 max-w-full">
        <svg class="w-3.5 h-3.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
        <span class="truncate">{errorMsg.value}</span>
      </span>
    );
  }

  return (
    <button onClick={handleTrigger} class="action-primary" disabled={s === 'loading'}>
      {s === 'loading' ? (
        <>
          <svg class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          Running...
        </>
      ) : (
        <>
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182" />
          </svg>
          Trigger Scan
        </>
      )}
    </button>
  );
}
