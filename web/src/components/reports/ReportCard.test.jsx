// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { render, screen, fireEvent } from '@testing-library/preact';
import { describe, it, expect } from 'vitest';
import { ReportCard } from './ReportCard';

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
});
