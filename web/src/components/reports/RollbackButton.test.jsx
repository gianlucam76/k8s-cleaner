// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { render, screen, fireEvent, waitFor } from '@testing-library/preact';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { RollbackButton } from './RollbackButton';
import { configData } from '../../app';
import { api } from '../../utils/fetch';

vi.mock('../../utils/fetch', () => ({ api: vi.fn() }));

describe('RollbackButton', () => {
  beforeEach(() => {
    configData.value = { readOnly: false };
    api.mockReset();
  });

  it('renders nothing in read-only mode', () => {
    configData.value = { readOnly: true };
    render(<RollbackButton cleanerName="unused-configmaps" resourceCount={2} />);
    expect(screen.queryByText('Rollback')).toBeNull();
  });

  it('shows a confirmation dialog before rolling back', () => {
    render(<RollbackButton cleanerName="unused-configmaps" resourceCount={2} />);
    fireEvent.click(screen.getByText('Rollback'));
    expect(screen.getByText('Roll Back Last Execution?')).toBeTruthy();
  });

  it('cancels without calling the API', () => {
    render(<RollbackButton cleanerName="unused-configmaps" resourceCount={2} />);
    fireEvent.click(screen.getByText('Rollback'));
    fireEvent.click(screen.getByText('Cancel'));
    expect(screen.queryByText('Roll Back Last Execution?')).toBeNull();
    expect(api).not.toHaveBeenCalled();
  });

  it('rolls back and reports success', async () => {
    api.mockResolvedValue([
      { kind: 'ConfigMap', namespace: 'default', name: 'old-cm', success: true },
      { kind: 'ConfigMap', namespace: 'default', name: 'stale', success: true },
    ]);
    render(<RollbackButton cleanerName="unused-configmaps" resourceCount={2} />);
    fireEvent.click(screen.getByText('Rollback'));
    fireEvent.click(screen.getByText('Roll Back'));

    await waitFor(() => expect(screen.getByText('Rolled back 2/2')).toBeTruthy());
    expect(api).toHaveBeenCalledWith('/api/v1/reports/unused-configmaps/rollback', { method: 'POST' });
  });

  it('reports partial failures', async () => {
    api.mockResolvedValue([
      { kind: 'ConfigMap', namespace: 'default', name: 'old-cm', success: true },
      { kind: 'ConfigMap', namespace: 'default', name: 'stale', success: false, message: 'no rollback data' },
    ]);
    render(<RollbackButton cleanerName="unused-configmaps" resourceCount={2} />);
    fireEvent.click(screen.getByText('Rollback'));
    fireEvent.click(screen.getByText('Roll Back'));

    await waitFor(() => expect(screen.getByText('Rolled back 1/2, 1 failed')).toBeTruthy());
  });

  it('reports a structural failure, e.g. no report found', async () => {
    api.mockRejectedValue(new Error('report not found'));
    render(<RollbackButton cleanerName="unused-configmaps" resourceCount={2} />);
    fireEvent.click(screen.getByText('Rollback'));
    fireEvent.click(screen.getByText('Roll Back'));

    await waitFor(() => expect(screen.getByText('report not found')).toBeTruthy());
  });
});
