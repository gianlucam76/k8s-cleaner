// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { render, screen, waitFor } from '@testing-library/preact';
import { describe, it, expect } from 'vitest';
import { LuaViewer } from './LuaViewer';

describe('LuaViewer', () => {
  it('renders the given code', async () => {
    render(<LuaViewer code={'function evaluate() -- first\nend'} />);
    await waitFor(() => expect(screen.getByText('-- first')).toBeTruthy());
  });

  it('shows a placeholder when there is no code', () => {
    render(<LuaViewer code="" />);
    expect(screen.getByText('No Lua script available')).toBeTruthy();
  });

  it('re-highlights when the code prop changes on an already-mounted instance', async () => {
    // Regression test: highlight.js marks a DOM node data-highlighted="yes"
    // and silently refuses to re-highlight it, so swapping `code` under a
    // live LuaViewer (e.g. a Current/New toggle) must reset that marker.
    const { rerender } = render(<LuaViewer code={'function evaluate() -- first\nend'} />);
    await waitFor(() => expect(screen.getByText('-- first')).toBeTruthy());

    rerender(<LuaViewer code={'function evaluate() -- second\nend'} />);

    await waitFor(() => expect(screen.getByText('-- second')).toBeTruthy());
    expect(screen.queryByText('-- first')).toBeNull();
  });
});
