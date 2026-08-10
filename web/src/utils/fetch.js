// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

const MOCK_MODE = import.meta.env.VITE_USE_MOCK_DATA === 'true';

const mockLookup = {
  '/api/v1/summary': () => import('../mock/summary.js').then((m) => m.mockSummary),
  '/api/v1/cleaners': () => import('../mock/cleaners.js').then((m) => m.mockCleaners),
  '/api/v1/reports': () => import('../mock/reports.js').then((m) => m.mockReports),
  '/api/v1/config': () => import('../mock/config.js').then((m) => m.mockConfig),
  '/api/v1/health': () => Promise.resolve({ status: 'ok' }),
  '/api/v1/library': () => import('../mock/library.js').then((m) => m.mockLibrary),
};

export async function api(url, opts = {}) {
  if (MOCK_MODE) {
    await new Promise((r) => setTimeout(r, 150));
    const fn = mockLookup[url];
    if (fn) return fn();

    if (url === '/api/v1/cleaners' && opts.method === 'POST') {
      await new Promise((r) => setTimeout(r, 400));
      const body = JSON.parse(opts.body);
      return { name: body.name, schedule: body.schedule, action: 'Scan' };
    }
    if (url.startsWith('/api/v1/cleaners/') && opts.method === 'PUT') {
      await new Promise((r) => setTimeout(r, 400));
      const name = url.split('/').pop();
      const body = JSON.parse(opts.body);
      return { name, schedule: body.schedule, action: 'Scan' };
    }
    if (url.startsWith('/api/v1/cleaners/') && !url.includes('/trigger')) {
      const name = url.split('/').pop();
      const { mockCleaners } = await import('../mock/cleaners.js');
      return mockCleaners.find((c) => c.name === name) || null;
    }
    if (url.startsWith('/api/v1/library/')) {
      const id = url.split('/').pop();
      const { mockLibrary, mockLibraryDetail } = await import('../mock/library.js');
      const summary = mockLibrary.find((e) => e.id === id);
      return mockLibraryDetail[id] || (summary ? { ...summary, luaScript: '' } : null);
    }
    if (url.startsWith('/api/v1/reports/')) {
      const name = url.split('/').pop();
      const { mockReports } = await import('../mock/reports.js');
      return mockReports.find((r) => r.name === name) || null;
    }
    if (url.includes('/rollback')) {
      await new Promise((r) => setTimeout(r, 400));
      const name = url.split('/')[4];
      const { mockReports } = await import('../mock/reports.js');
      const report = mockReports.find((r) => r.name === name);
      return (report?.resources || []).map((res) => ({
        kind: res.kind,
        namespace: res.namespace,
        name: res.name,
        success: true,
      }));
    }
    if (url.includes('/trigger')) {
      await new Promise((r) => setTimeout(r, 400));
      return { message: 'scan triggered', cleaner: url.split('/')[4] };
    }
    return null;
  }

  const res = await fetch(url, {
    headers: { 'Content-Type': 'application/json', ...opts.headers },
    ...opts,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    const err = new Error(body.error || `HTTP ${res.status}`);
    err.status = res.status;
    throw err;
  }
  return res.json();
}
