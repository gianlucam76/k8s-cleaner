// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { signal, effect } from '@preact/signals';

function readStorage(key) {
  try { return localStorage.getItem(key); } catch { return null; }
}

function writeStorage(key, val) {
  try { localStorage.setItem(key, val); } catch { /* noop */ }
}

function prefersDark() {
  try { return window.matchMedia('(prefers-color-scheme: dark)').matches; } catch { return false; }
}

// Possible values: 'light', 'dark', 'auto'
export const theme = signal(readStorage('k8s-cleaner-theme') || 'auto');

effect(() => {
  const current = theme.value;
  const dark = current === 'dark' || (current === 'auto' && prefersDark());
  if (typeof document !== 'undefined') {
    document.documentElement.classList.toggle('dark', dark);
  }
  writeStorage('k8s-cleaner-theme', current);
});

export function nextTheme() {
  const order = ['light', 'dark', 'auto'];
  theme.value = order[(order.indexOf(theme.value) + 1) % order.length];
}
