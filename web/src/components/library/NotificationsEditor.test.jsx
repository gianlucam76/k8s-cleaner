// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { render, screen, fireEvent } from '@testing-library/preact';
import { describe, it, expect, vi } from 'vitest';
import { NotificationsEditor } from './NotificationsEditor';

describe('NotificationsEditor', () => {
  it('renders one row per notification', () => {
    const notifications = [
      { name: 'report', type: 'CleanerReport' },
      { name: 'alerts', type: 'Slack' },
    ];
    render(<NotificationsEditor notifications={notifications} onChange={() => {}} />);
    expect(screen.getByDisplayValue('report')).toBeTruthy();
    expect(screen.getByDisplayValue('alerts')).toBeTruthy();
  });

  it('adds a new CleanerReport row', () => {
    const onChange = vi.fn();
    render(<NotificationsEditor notifications={[]} onChange={onChange} />);
    fireEvent.click(screen.getByText('+ Add notification'));
    expect(onChange).toHaveBeenCalledWith([{ name: '', type: 'CleanerReport' }]);
  });

  it('removes a row', () => {
    const onChange = vi.fn();
    const notifications = [
      { name: 'report', type: 'CleanerReport' },
      { name: 'alerts', type: 'Slack' },
    ];
    render(<NotificationsEditor notifications={notifications} onChange={onChange} />);
    fireEvent.click(screen.getAllByLabelText('Remove notification')[0]);
    expect(onChange).toHaveBeenCalledWith([{ name: 'alerts', type: 'Slack' }]);
  });

  it('updates a row name', () => {
    const onChange = vi.fn();
    const notifications = [{ name: 'report', type: 'CleanerReport' }];
    render(<NotificationsEditor notifications={notifications} onChange={onChange} />);
    fireEvent.input(screen.getByDisplayValue('report'), { target: { value: 'my-report' } });
    expect(onChange).toHaveBeenCalledWith([{ name: 'my-report', type: 'CleanerReport' }]);
  });

  it('updates a row type', () => {
    const onChange = vi.fn();
    const notifications = [{ name: 'report', type: 'CleanerReport' }];
    render(<NotificationsEditor notifications={notifications} onChange={onChange} />);
    fireEvent.change(screen.getByDisplayValue('CleanerReport'), { target: { value: 'Event' } });
    expect(onChange).toHaveBeenCalledWith([{ name: 'report', type: 'Event' }]);
  });
});
