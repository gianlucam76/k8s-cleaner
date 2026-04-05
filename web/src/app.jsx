// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { signal } from '@preact/signals';
import { useEffect } from 'preact/hooks';
import Router from 'preact-router';

import { Header } from './components/layout/Header';
import { Footer } from './components/layout/Footer';
import { ConnectionStatus } from './components/layout/ConnectionStatus';
import { DashboardPage } from './components/pages/DashboardPage';
import { ReportsPage } from './components/pages/ReportsPage';
import { api } from './utils/fetch';
import { ENDPOINTS, POLL_INTERVAL_MS, FAST_POLL_INTERVAL_MS, FAST_POLL_DURATION_MS } from './utils/constants';

// Activate theme side-effect
import './utils/theme';

// Global state
export const connState = signal('loading');
export const summaryData = signal(null);
export const cleanersData = signal(null);
export const reportsData = signal(null);
export const configData = signal(null);

let interval = POLL_INTERVAL_MS;
let fastTimer = null;

async function refresh() {
  try {
    const [s, c, r, cfg] = await Promise.all([
      api(ENDPOINTS.summary),
      api(ENDPOINTS.cleaners),
      api(ENDPOINTS.reports),
      api(ENDPOINTS.config),
    ]);
    summaryData.value = s;
    cleanersData.value = c;
    reportsData.value = r;
    configData.value = cfg;
    connState.value = 'connected';
  } catch (err) {
    console.error('Data fetch failed:', err);
    connState.value = summaryData.value ? 'disconnected' : 'loading';
  }
}

export function enableFastRefresh() {
  interval = FAST_POLL_INTERVAL_MS;
  if (fastTimer) clearTimeout(fastTimer);
  fastTimer = setTimeout(() => { interval = POLL_INTERVAL_MS; fastTimer = null; }, FAST_POLL_DURATION_MS);
}

function usePolling() {
  useEffect(() => {
    let active = true;
    refresh();
    const tick = () => {
      setTimeout(async () => {
        if (!active) return;
        await refresh();
        tick();
      }, interval);
    };
    tick();
    return () => { active = false; };
  }, []);
}

export function App() {
  usePolling();

  return (
    <div class="min-h-screen flex flex-col">
      <ConnectionStatus />
      <Header />
      <main class="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <Router>
          <DashboardPage path="/" />
          <ReportsPage path="/reports" />
          <NotFoundPage default />
        </Router>
      </main>
      <Footer />
    </div>
  );
}

function NotFoundPage() {
  return (
    <div class="text-center py-20">
      <h1 class="text-4xl font-bold text-gray-300 dark:text-gray-600">404</h1>
      <p class="text-sm text-gray-500 dark:text-gray-400 mt-2">Page not found</p>
      <a href="/" class="text-sm text-blue-600 hover:underline mt-4 inline-block">Back to Dashboard</a>
    </div>
  );
}
