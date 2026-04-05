// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { connState } from '../../app';

export function ConnectionStatus() {
  const s = connState.value;
  if (s === 'connected') return null;

  if (s === 'loading') {
    return <div class="h-1 w-full bg-gray-200 dark:bg-gray-700 overflow-hidden"><div class="h-full w-1/3 bg-blue-600 skeleton" /></div>;
  }

  return (
    <div class="h-7 w-full bg-rose-600 text-white text-[11px] font-semibold flex items-center justify-center gap-1.5">
      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
      </svg>
      API unreachable
    </div>
  );
}
