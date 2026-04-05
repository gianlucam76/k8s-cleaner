// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

export const API_BASE = '/api/v1';

export const POLL_INTERVAL_MS = 30000;       // 30s normal polling
export const FAST_POLL_INTERVAL_MS = 5000;   // 5s after user action
export const FAST_POLL_DURATION_MS = 120000; // 2min fast polling window

export const ENDPOINTS = {
  summary: `${API_BASE}/summary`,
  cleaners: `${API_BASE}/cleaners`,
  reports: `${API_BASE}/reports`,
  config: `${API_BASE}/config`,
  health: `${API_BASE}/health`,
  triggerAll: `${API_BASE}/trigger-all`,
  trigger: (name) => `${API_BASE}/cleaners/${encodeURIComponent(name)}/trigger`,
  cleaner: (name) => `${API_BASE}/cleaners/${encodeURIComponent(name)}`,
  report: (name) => `${API_BASE}/reports/${encodeURIComponent(name)}`,
};
