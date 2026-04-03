// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/preact';
import { describe, it, expect } from 'vitest';
import { StatusPanel } from './StatusPanel';

describe('StatusPanel', () => {
  it('renders cleaner count', () => {
    const summary = {
      totalCleaners: 11,
      totalFlaggedResources: 5,
      cleanersWithFindings: 2,
      lastScanTime: new Date().toISOString(),
    };
    render(<StatusPanel summary={summary} />);
    expect(screen.getByText('11')).toBeTruthy();
    expect(screen.getByText('5')).toBeTruthy();
    expect(screen.getByText('2')).toBeTruthy();
  });

  it('renders zero state', () => {
    const summary = {
      totalCleaners: 0,
      totalFlaggedResources: 0,
      cleanersWithFindings: 0,
      lastScanTime: null,
    };
    render(<StatusPanel summary={summary} />);
    expect(screen.getAllByText('0').length).toBeGreaterThanOrEqual(3);
  });
});
