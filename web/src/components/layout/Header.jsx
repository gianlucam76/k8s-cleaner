// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { ThemeToggle } from './ThemeToggle';

function NavItem({ href, children }) {
  const active = typeof window !== 'undefined' && window.location.pathname === href;
  return (
    <a
      href={href}
      class={`px-3 py-1.5 text-[11px] font-semibold rounded-lg transition-colors ${
        active
          ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white'
          : 'text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
      }`}
    >
      {children}
    </a>
  );
}

export function Header() {
  return (
    <header class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-14 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <a href="/" class="flex items-center gap-2 text-gray-900 dark:text-white hover:opacity-80">
            <svg class="w-5 h-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 3.104v5.714a2.25 2.25 0 01-.659 1.591L5 14.5M9.75 3.104c-.251.023-.501.05-.75.082m.75-.082a24.301 24.301 0 014.5 0m0 0v5.714c0 .597.237 1.17.659 1.591L19.8 15.3M14.25 3.104c.251.023.501.05.75.082M19.8 15.3l-1.57.393A9.065 9.065 0 0112 15a9.065 9.065 0 00-6.23.693L5 14.5m14.8.8l1.402 1.402c1.232 1.232.65 3.318-1.067 3.611A48.309 48.309 0 0112 21c-2.773 0-5.491-.235-8.135-.687-1.718-.293-2.3-2.379-1.067-3.61L5 14.5" />
            </svg>
            <span class="font-semibold text-sm">k8s-cleaner</span>
          </a>
          <nav class="flex items-center gap-1 ml-5">
            <NavItem href="/">Dashboard</NavItem>
            <NavItem href="/reports">Reports</NavItem>
          </nav>
        </div>
        <ThemeToggle />
      </div>
    </header>
  );
}
