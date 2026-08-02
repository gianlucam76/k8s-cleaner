// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { render, screen, fireEvent } from '@testing-library/preact';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ReportCard } from './ReportCard';
import { configData } from '../../app';

vi.mock('../../utils/fetch', () => ({ api: vi.fn() }));

describe('ReportCard', () => {
  const cleanReport = { name: 'pvc-scan', action: 'Scan', resources: [] };
  const flaggedReport = {
    name: 'unused-configmaps',
    action: 'Scan',
    resources: [
      { kind: 'ConfigMap', namespace: 'default', name: 'old-cm', apiVersion: 'v1', message: 'orphaned' },
      { kind: 'ConfigMap', namespace: 'staging', name: 'stale', apiVersion: 'v1', message: 'orphaned' },
    ],
  };
  const deletedReport = {
    name: 'unused-secrets',
    action: 'Delete',
    resources: [
      { kind: 'Secret', namespace: 'default', name: 'old-secret', apiVersion: 'v1', message: 'orphaned' },
    ],
  };

  beforeEach(() => {
    configData.value = { readOnly: false };
  });

  it('renders clean report with Clean badge', () => {
    render(<ReportCard report={cleanReport} />);
    expect(screen.getByText('pvc-scan')).toBeTruthy();
    expect(screen.getByText('Clean')).toBeTruthy();
  });

  it('renders flagged report with count badge', () => {
    render(<ReportCard report={flaggedReport} />);
    expect(screen.getByText('unused-configmaps')).toBeTruthy();
    expect(screen.getByText('2 flagged')).toBeTruthy();
  });

  it('expands to show resources on click', () => {
    render(<ReportCard report={flaggedReport} />);
    // Resources not visible initially
    expect(screen.queryByText('old-cm')).toBeNull();

    // Click to expand
    fireEvent.click(screen.getByText('unused-configmaps'));
    expect(screen.getByText('old-cm')).toBeTruthy();
    expect(screen.getByText('stale')).toBeTruthy();
  });

  it('shows empty message for clean report when expanded', () => {
    render(<ReportCard report={cleanReport} />);
    fireEvent.click(screen.getByText('pvc-scan'));
    expect(screen.getByText('No flagged resources in latest scan')).toBeTruthy();
  });

  it('shows a Rollback button for a flagged Delete report', () => {
    render(<ReportCard report={deletedReport} />);
    fireEvent.click(screen.getByText('unused-secrets'));
    expect(screen.getByText('Rollback')).toBeTruthy();
  });

  it('does not show a Rollback button for a flagged Scan report', () => {
    render(<ReportCard report={flaggedReport} />);
    fireEvent.click(screen.getByText('unused-configmaps'));
    expect(screen.queryByText('Rollback')).toBeNull();
  });
});
