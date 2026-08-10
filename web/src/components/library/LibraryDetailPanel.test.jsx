// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { render, screen, fireEvent, waitFor } from '@testing-library/preact';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { LibraryDetailPanel } from './LibraryDetailPanel';
import { configData } from '../../app';
import { api } from '../../utils/fetch';

vi.mock('../../utils/fetch', () => ({ api: vi.fn() }));

const detail = {
  id: 'unused-clusterroles',
  title: 'Unused ClusterRoles',
  description: 'Finds ClusterRole instances not referenced by any ClusterRoleBinding or RoleBinding.',
  schedule: '* 0 * * *',
  selectors: [{ group: 'rbac.authorization.k8s.io', version: 'v1', kind: 'ClusterRole' }],
  luaScript: 'function evaluate() end',
};

describe('LibraryDetailPanel', () => {
  beforeEach(() => {
    configData.value = { readOnly: false };
    api.mockReset();
  });

  it('renders nothing when closed', () => {
    render(<LibraryDetailPanel entryId="unused-clusterroles" isOpen={false} onClose={() => {}} />);
    expect(screen.queryByText('Unused ClusterRoles')).toBeNull();
  });

  it('loads and displays the recipe, prefilling name and schedule', async () => {
    api.mockResolvedValue(detail);
    render(<LibraryDetailPanel entryId="unused-clusterroles" isOpen onClose={() => {}} />);

    await waitFor(() => expect(screen.getByText('Unused ClusterRoles')).toBeTruthy());
    expect(screen.getByDisplayValue('unused-clusterroles')).toBeTruthy();
    expect(screen.getByDisplayValue('* 0 * * *')).toBeTruthy();
    expect(screen.getByText('ClusterRole')).toBeTruthy();
  });

  it('starts with one CleanerReport notification', async () => {
    api.mockResolvedValue(detail);
    render(<LibraryDetailPanel entryId="unused-clusterroles" isOpen onClose={() => {}} />);

    await waitFor(() => expect(screen.getByDisplayValue('report')).toBeTruthy());
    expect(screen.getByDisplayValue('CleanerReport')).toBeTruthy();
  });

  it('defaults Action to Scan and warns when switched to Delete', async () => {
    api.mockResolvedValue(detail);
    render(<LibraryDetailPanel entryId="unused-clusterroles" isOpen onClose={() => {}} />);

    await waitFor(() => expect(screen.getByDisplayValue('Scan (report only)')).toBeTruthy());
    expect(screen.queryByText(/Consider testing as Scan first/)).toBeNull();

    fireEvent.change(screen.getByDisplayValue('Scan (report only)'), { target: { value: 'Delete' } });

    expect(screen.getByText(/Consider testing as Scan first/)).toBeTruthy();
  });

  it('includes the selected action in the post body', async () => {
    api.mockResolvedValueOnce(detail).mockResolvedValueOnce({ name: 'unused-clusterroles', schedule: '* 0 * * *' });
    render(<LibraryDetailPanel entryId="unused-clusterroles" isOpen onClose={() => {}} />);

    await waitFor(() => expect(screen.getByDisplayValue('Scan (report only)')).toBeTruthy());
    fireEvent.change(screen.getByDisplayValue('Scan (report only)'), { target: { value: 'Delete' } });
    fireEvent.click(screen.getByText('Post Cleaner'));

    await waitFor(() => expect(api).toHaveBeenCalledTimes(2));
    const body = JSON.parse(api.mock.calls[1][1].body);
    expect(body.action).toEqual('Delete');
  });

  it('posts the edited cleaner and reports success', async () => {
    api.mockResolvedValueOnce(detail).mockResolvedValueOnce({ name: 'my-clusterroles', schedule: '0 3 * * *' });
    render(<LibraryDetailPanel entryId="unused-clusterroles" isOpen onClose={() => {}} />);

    await waitFor(() => expect(screen.getByDisplayValue('unused-clusterroles')).toBeTruthy());
    fireEvent.input(screen.getByDisplayValue('unused-clusterroles'), { target: { value: 'my-clusterroles' } });
    fireEvent.click(screen.getByText('Post Cleaner'));

    await waitFor(() => expect(screen.getByText('Created "my-clusterroles"')).toBeTruthy());
    expect(api).toHaveBeenLastCalledWith('/api/v1/cleaners', expect.objectContaining({ method: 'POST' }));

    const body = JSON.parse(api.mock.calls[1][1].body);
    expect(body).toEqual({
      sourceLibraryId: 'unused-clusterroles',
      name: 'my-clusterroles',
      schedule: '* 0 * * *',
      action: 'Scan',
      notifications: [{ name: 'report', type: 'CleanerReport' }],
    });
  });

  it('reports an error message when posting fails for a non-conflict reason', async () => {
    api.mockResolvedValueOnce(detail).mockRejectedValueOnce(new Error('failed to create cleaner'));
    render(<LibraryDetailPanel entryId="unused-clusterroles" isOpen onClose={() => {}} />);

    await waitFor(() => expect(screen.getByDisplayValue('unused-clusterroles')).toBeTruthy());
    fireEvent.click(screen.getByText('Post Cleaner'));

    await waitFor(() => expect(screen.getByText('failed to create cleaner')).toBeTruthy());
  });

  it('shows a conflict comparison on a 409 and can update from it', async () => {
    const conflictError = new Error('a cleaner with this name already exists');
    conflictError.status = 409;
    const existingCleaner = {
      name: 'unused-clusterroles', schedule: '0 * * * *', action: 'Delete',
      notifications: [{ name: 'old', type: 'Slack' }],
      selectors: [{ group: '', version: 'v1', kind: 'ConfigMap' }],
      luaScript: 'function evaluate() -- old\nend',
    };
    api
      .mockResolvedValueOnce(detail) // GET library entry
      .mockRejectedValueOnce(conflictError) // POST -> 409
      .mockResolvedValueOnce(existingCleaner) // GET existing cleaner for the diff
      .mockResolvedValueOnce({ name: 'unused-clusterroles', schedule: '* 0 * * *', action: 'Scan' }); // PUT update

    render(<LibraryDetailPanel entryId="unused-clusterroles" isOpen onClose={() => {}} />);
    await waitFor(() => expect(screen.getByDisplayValue('unused-clusterroles')).toBeTruthy());

    fireEvent.click(screen.getByText('Post Cleaner'));

    await waitFor(() => expect(screen.getByText(/already exists/)).toBeTruthy());
    expect(screen.getByText('Update Cleaner')).toBeTruthy();

    fireEvent.click(screen.getByText('Update Cleaner'));

    await waitFor(() => expect(screen.getByText('Updated "unused-clusterroles"')).toBeTruthy());
    expect(api).toHaveBeenLastCalledWith('/api/v1/cleaners/unused-clusterroles',
      expect.objectContaining({ method: 'PUT' }));
  });

  it('returns to the edit form when the conflict comparison is cancelled', async () => {
    const conflictError = new Error('a cleaner with this name already exists');
    conflictError.status = 409;
    const existingCleaner = {
      name: 'unused-clusterroles', schedule: '0 * * * *', action: 'Delete',
      notifications: [], selectors: [], luaScript: '',
    };
    api
      .mockResolvedValueOnce(detail)
      .mockRejectedValueOnce(conflictError)
      .mockResolvedValueOnce(existingCleaner);

    render(<LibraryDetailPanel entryId="unused-clusterroles" isOpen onClose={() => {}} />);
    await waitFor(() => expect(screen.getByDisplayValue('unused-clusterroles')).toBeTruthy());

    fireEvent.click(screen.getByText('Post Cleaner'));
    await waitFor(() => expect(screen.getByText('Update Cleaner')).toBeTruthy());

    fireEvent.click(screen.getByText('Cancel'));

    expect(screen.getByText('Post Cleaner')).toBeTruthy();
    expect(screen.queryByText('Update Cleaner')).toBeNull();
  });

  it('shows the conflict comparison on re-post after closing and reopening the same entry', async () => {
    // Regression test for the real usage pattern: post once (success), close the
    // panel, reselect the same library entry, post again - the second post must
    // still route through the conflict view, not the plain error message.
    api
      .mockResolvedValueOnce(detail) // 1st open: GET library entry
      .mockResolvedValueOnce({ name: 'unused-clusterroles', schedule: '* 0 * * *', action: 'Scan' }); // 1st POST -> success

    const { rerender } = render(<LibraryDetailPanel entryId="unused-clusterroles" isOpen onClose={() => {}} />);
    await waitFor(() => expect(screen.getByDisplayValue('unused-clusterroles')).toBeTruthy());
    fireEvent.click(screen.getByText('Post Cleaner'));
    await waitFor(() => expect(screen.getByText('Created "unused-clusterroles"')).toBeTruthy());

    rerender(<LibraryDetailPanel entryId="unused-clusterroles" isOpen={false} onClose={() => {}} />); // close

    const conflictError = new Error('a cleaner with this name already exists');
    conflictError.status = 409;
    api
      .mockResolvedValueOnce(detail) // 2nd open: GET library entry again
      .mockRejectedValueOnce(conflictError) // 2nd POST -> 409
      .mockResolvedValueOnce({ // GET existing cleaner for the diff
        name: 'unused-clusterroles', schedule: '* 0 * * *', action: 'Scan',
        notifications: [], selectors: [], luaScript: '',
      });

    rerender(<LibraryDetailPanel entryId="unused-clusterroles" isOpen onClose={() => {}} />); // reopen, same entry

    await waitFor(() => expect(screen.getByDisplayValue('unused-clusterroles')).toBeTruthy());
    fireEvent.click(screen.getByText('Post Cleaner'));

    await waitFor(() => expect(screen.getByText(/already exists/)).toBeTruthy());
    expect(screen.getByText('Update Cleaner')).toBeTruthy();
  });

  it('hides the Post button and shows a notice in read-only mode', async () => {
    configData.value = { readOnly: true };
    api.mockResolvedValue(detail);
    render(<LibraryDetailPanel entryId="unused-clusterroles" isOpen onClose={() => {}} />);

    await waitFor(() => expect(screen.getByText('Unused ClusterRoles')).toBeTruthy());
    expect(screen.queryByText('Post Cleaner')).toBeNull();
    expect(screen.getByText('Dashboard is in read-only mode.')).toBeTruthy();
  });

  it('calls onClose when the close button is clicked', async () => {
    api.mockResolvedValue(detail);
    const onClose = vi.fn();
    render(<LibraryDetailPanel entryId="unused-clusterroles" isOpen onClose={onClose} />);

    await waitFor(() => expect(screen.getByText('Unused ClusterRoles')).toBeTruthy());
    fireEvent.click(screen.getByLabelText('Close panel'));
    expect(onClose).toHaveBeenCalled();
  });
});
