// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { render, screen, fireEvent, waitFor } from '@testing-library/preact';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ConflictDiff } from './ConflictDiff';
import { api } from '../../utils/fetch';

vi.mock('../../utils/fetch', () => ({ api: vi.fn() }));

const existing = {
  schedule: '0 * * * *',
  action: 'Delete',
  notifications: [{ name: 'old', type: 'Slack' }],
  selectors: [{ group: '', version: 'v1', kind: 'ConfigMap' }],
  luaScript: 'function evaluate() -- old\nend',
};

const proposed = {
  schedule: '0 5 * * *',
  action: 'Scan',
  notifications: [{ name: 'report', type: 'CleanerReport' }],
  selectors: [{ group: 'rbac.authorization.k8s.io', version: 'v1', kind: 'ClusterRole' }],
  luaScript: 'function evaluate() -- new\nend',
};

describe('ConflictDiff', () => {
  beforeEach(() => {
    api.mockReset();
  });

  it('shows the conflicting name and a schedule/action diff', () => {
    render(
      <ConflictDiff
        name="my-cleaner"
        sourceLibraryId="unused-clusterroles"
        existing={existing}
        proposed={proposed}
        onUpdated={() => {}}
        onCancel={() => {}}
      />,
    );
    expect(screen.getByText('my-cleaner')).toBeTruthy();
    expect(screen.getByText('0 * * * *')).toBeTruthy();
    expect(screen.getByText('0 5 * * *')).toBeTruthy();
    expect(screen.getByText('Delete')).toBeTruthy();
    expect(screen.getByText('Scan')).toBeTruthy();
  });

  it('shows a Current/New Lua toggle when selection logic differs', async () => {
    render(
      <ConflictDiff
        name="my-cleaner"
        sourceLibraryId="unused-clusterroles"
        existing={existing}
        proposed={proposed}
        onUpdated={() => {}}
        onCancel={() => {}}
      />,
    );
    await waitFor(() => expect(screen.getByText('-- old')).toBeTruthy()); // defaults to the "Current" tab
    fireEvent.click(screen.getByRole('button', { name: 'New (from library)' }));
    await waitFor(() => expect(screen.getByText('-- new')).toBeTruthy());
  });

  it('does not show a Lua toggle when selection logic is identical', () => {
    const sameSelection = { ...proposed, selectors: existing.selectors, luaScript: existing.luaScript };
    render(
      <ConflictDiff
        name="my-cleaner"
        sourceLibraryId="unused-clusterroles"
        existing={existing}
        proposed={sameSelection}
        onUpdated={() => {}}
        onCancel={() => {}}
      />,
    );
    expect(screen.queryByText('New (from library)')).toBeNull();
  });

  it('updates the cleaner and reports success', async () => {
    api.mockResolvedValue({ name: 'my-cleaner', schedule: '0 5 * * *', action: 'Scan' });
    const onUpdated = vi.fn();
    render(
      <ConflictDiff
        name="my-cleaner"
        sourceLibraryId="unused-clusterroles"
        existing={existing}
        proposed={proposed}
        onUpdated={onUpdated}
        onCancel={() => {}}
      />,
    );

    fireEvent.click(screen.getByText('Update Cleaner'));

    await waitFor(() => expect(screen.getByText('Updated "my-cleaner"')).toBeTruthy());
    expect(api).toHaveBeenCalledWith('/api/v1/cleaners/my-cleaner', {
      method: 'PUT',
      body: JSON.stringify({
        sourceLibraryId: 'unused-clusterroles',
        schedule: '0 5 * * *',
        action: proposed.action,
        notifications: proposed.notifications,
      }),
    });
    expect(onUpdated).toHaveBeenCalledWith({ name: 'my-cleaner', schedule: '0 5 * * *', action: 'Scan' });
  });

  it('reports an error message when the update fails', async () => {
    api.mockRejectedValue(new Error('cleaner was modified concurrently, please retry'));
    render(
      <ConflictDiff
        name="my-cleaner"
        sourceLibraryId="unused-clusterroles"
        existing={existing}
        proposed={proposed}
        onUpdated={() => {}}
        onCancel={() => {}}
      />,
    );

    fireEvent.click(screen.getByText('Update Cleaner'));

    await waitFor(() => expect(screen.getByText('cleaner was modified concurrently, please retry')).toBeTruthy());
  });

  it('calls onCancel when Cancel is clicked', () => {
    const onCancel = vi.fn();
    render(
      <ConflictDiff
        name="my-cleaner"
        sourceLibraryId="unused-clusterroles"
        existing={existing}
        proposed={proposed}
        onUpdated={() => {}}
        onCancel={onCancel}
      />,
    );
    fireEvent.click(screen.getByText('Cancel'));
    expect(onCancel).toHaveBeenCalled();
  });
});
