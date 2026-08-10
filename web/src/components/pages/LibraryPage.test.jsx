// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { render, screen, fireEvent, waitFor } from '@testing-library/preact';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { LibraryPage } from './LibraryPage';
import { api } from '../../utils/fetch';

vi.mock('../../utils/fetch', () => ({ api: vi.fn() }));

const entries = [
  {
    id: 'unused-clusterroles', category: 'unused-resources', group: 'RBAC',
    title: 'Unused ClusterRoles', description: 'Finds ClusterRole instances not referenced.',
    schedule: '* 0 * * *', selectors: [{ group: '', version: 'v1', kind: 'ClusterRole' }],
  },
  {
    id: 'unused-roles', category: 'unused-resources', group: 'RBAC',
    title: 'Unused Roles', description: 'Finds Role instances not referenced.',
    schedule: '* 0 * * *', selectors: [{ group: '', version: 'v1', kind: 'Role' }],
  },
  {
    id: 'unused-configmaps', category: 'unused-resources', group: 'Config & Secrets',
    title: 'Orphaned ConfigMaps', description: 'Finds ConfigMaps unused by any Pod.',
    schedule: '* 0 * * *', selectors: [{ group: '', version: 'v1', kind: 'ConfigMap' }],
  },
  {
    id: 'terminating-pods', category: 'unhealthy-resources', group: 'Pods',
    title: 'Pods Stuck Terminating', description: 'Finds terminating pods.',
    schedule: '*/5 * * * *', selectors: [{ group: '', version: 'v1', kind: 'Pod' }],
  },
];

describe('LibraryPage', () => {
  beforeEach(() => {
    api.mockReset();
    api.mockImplementation((url) => (url === '/api/v1/library' ? Promise.resolve(entries) : Promise.resolve(null)));
  });

  it('defaults to the Unused Resources tab, grouped into sections', async () => {
    render(<LibraryPage />);
    await waitFor(() => expect(screen.getByText('Unused ClusterRoles')).toBeTruthy());

    expect(screen.getByText('Unused Roles')).toBeTruthy();
    expect(screen.getByText('Orphaned ConfigMaps')).toBeTruthy();
    expect(screen.queryByText('Pods Stuck Terminating')).toBeNull();

    expect(screen.getByText('RBAC')).toBeTruthy();
    expect(screen.getByText('Config & Secrets')).toBeTruthy();
  });

  it('colors group headings instead of leaving them gray', async () => {
    render(<LibraryPage />);
    await waitFor(() => expect(screen.getByText('Unused ClusterRoles')).toBeTruthy());

    expect(screen.getByText('RBAC').className).toMatch(/text-indigo-700/);
    expect(screen.getByText('Config & Secrets').className).toMatch(/text-indigo-700/);
  });

  it('switches to the Unhealthy Resources tab', async () => {
    render(<LibraryPage />);
    await waitFor(() => expect(screen.getByText('Unused ClusterRoles')).toBeTruthy());

    fireEvent.click(screen.getByRole('tab', { name: /Unhealthy Resources/ }));

    expect(screen.getByText('Pods Stuck Terminating')).toBeTruthy();
    expect(screen.queryByText('Unused ClusterRoles')).toBeNull();
  });

  it('filters by search term across title and description', async () => {
    render(<LibraryPage />);
    await waitFor(() => expect(screen.getByText('Unused ClusterRoles')).toBeTruthy());

    fireEvent.input(screen.getByPlaceholderText('Search recipes...'), { target: { value: 'configmap' } });

    expect(screen.getByText('Orphaned ConfigMaps')).toBeTruthy();
    expect(screen.queryByText('Unused ClusterRoles')).toBeNull();
    expect(screen.queryByText('Unused Roles')).toBeNull();
  });

  it('filters by resource-kind chip', async () => {
    render(<LibraryPage />);
    await waitFor(() => expect(screen.getByText('Unused ClusterRoles')).toBeTruthy());

    fireEvent.click(screen.getByRole('button', { name: 'Role' }));

    expect(screen.getByText('Unused Roles')).toBeTruthy();
    expect(screen.queryByText('Unused ClusterRoles')).toBeNull();
    expect(screen.queryByText('Orphaned ConfigMaps')).toBeNull();
  });

  it('clears the kind filter and shows an empty state when a search matches nothing', async () => {
    render(<LibraryPage />);
    await waitFor(() => expect(screen.getByText('Unused ClusterRoles')).toBeTruthy());

    fireEvent.input(screen.getByPlaceholderText('Search recipes...'), { target: { value: 'nothing-matches-this' } });

    expect(screen.getByText('No recipes match')).toBeTruthy();
  });

  it('resets the kind filter when switching category tabs', async () => {
    render(<LibraryPage />);
    await waitFor(() => expect(screen.getByText('Unused ClusterRoles')).toBeTruthy());

    fireEvent.click(screen.getByRole('button', { name: 'Role' }));
    fireEvent.click(screen.getByRole('tab', { name: /Unhealthy Resources/ }));
    fireEvent.click(screen.getByRole('tab', { name: /Unused Resources/ }));

    // Kind filter should no longer be applied after switching away and back
    expect(screen.getByText('Unused ClusterRoles')).toBeTruthy();
    expect(screen.getByText('Unused Roles')).toBeTruthy();
  });
});
