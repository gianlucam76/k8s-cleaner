// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { render, screen, fireEvent } from '@testing-library/preact';
import { describe, it, expect, vi } from 'vitest';
import { LibraryCard } from './LibraryCard';

describe('LibraryCard', () => {
  const entry = {
    id: 'unused-clusterroles',
    category: 'unused-resources',
    title: 'Unused ClusterRoles',
    description: 'Finds ClusterRole instances not referenced by any ClusterRoleBinding or RoleBinding.',
    schedule: '* 0 * * *',
  };

  it('renders title, description and schedule', () => {
    render(<LibraryCard entry={entry} isSelected={false} onSelect={() => {}} />);
    expect(screen.getByText('Unused ClusterRoles')).toBeTruthy();
    expect(screen.getByText(entry.description)).toBeTruthy();
    expect(screen.getByText('* 0 * * *')).toBeTruthy();
  });

  it('calls onSelect with the entry id when clicked', () => {
    const onSelect = vi.fn();
    render(<LibraryCard entry={entry} isSelected={false} onSelect={onSelect} />);
    fireEvent.click(screen.getByText('Unused ClusterRoles'));
    expect(onSelect).toHaveBeenCalledWith('unused-clusterroles');
  });

  it('calls onSelect on Enter key', () => {
    const onSelect = vi.fn();
    render(<LibraryCard entry={entry} isSelected={false} onSelect={onSelect} />);
    fireEvent.keyDown(screen.getByRole('button'), { key: 'Enter' });
    expect(onSelect).toHaveBeenCalledWith('unused-clusterroles');
  });

  it('renders a chip for each resource kind the recipe selects', () => {
    const withSelectors = {
      ...entry,
      selectors: [
        { group: 'rbac.authorization.k8s.io', version: 'v1', kind: 'ClusterRole' },
        { group: 'rbac.authorization.k8s.io', version: 'v1', kind: 'RoleBinding' },
      ],
    };
    render(<LibraryCard entry={withSelectors} isSelected={false} onSelect={() => {}} />);
    expect(screen.getAllByText('ClusterRole').length).toBeGreaterThan(0);
    expect(screen.getByText('RoleBinding')).toBeTruthy();
  });
});
