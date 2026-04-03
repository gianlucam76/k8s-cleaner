// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/preact';
import { describe, it, expect } from 'vitest';
import { Header } from './Header';

describe('Header', () => {
  it('renders the app name', () => {
    render(<Header />);
    expect(screen.getByText('k8s-cleaner')).toBeTruthy();
  });

  it('renders navigation links', () => {
    render(<Header />);
    expect(screen.getByText('Dashboard')).toBeTruthy();
    expect(screen.getByText('Reports')).toBeTruthy();
  });
});
