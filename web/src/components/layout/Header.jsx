// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { useSignal } from '@preact/signals';
import { useEffect } from 'preact/hooks';
import { ThemeToggle } from './ThemeToggle';

export function Header() {
  const currentPath = useSignal(typeof window !== 'undefined' ? window.location.pathname : '/');

  // Listen for route changes
  useEffect(() => {
    function onRouteChange() {
      currentPath.value = window.location.pathname;
    }
    window.addEventListener('popstate', onRouteChange);
    // Also intercept clicks on links for pushState navigation
    const observer = new MutationObserver(() => {
      if (currentPath.value !== window.location.pathname) {
        currentPath.value = window.location.pathname;
      }
    });
    observer.observe(document.body, { childList: true, subtree: true });
    return () => {
      window.removeEventListener('popstate', onRouteChange);
      observer.disconnect();
    };
  }, []);

  function navClass(href) {
    const active = currentPath.value === href;
    return `px-3 py-1.5 text-[11px] font-semibold rounded-lg transition-colors ${
      active
        ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white'
        : 'text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
    }`;
  }

  return (
    <header class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-14 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <a href="/" class="flex items-center gap-2 text-gray-900 dark:text-white hover:opacity-80">
            <img src="/logo.png" alt="k8s-cleaner" class="h-8 w-auto" />
            <span class="font-semibold text-sm">k8s-cleaner</span>
          </a>
          <nav class="flex items-center gap-1 ml-5">
            <a href="/" class={navClass('/')}>Dashboard</a>
            <a href="/reports" class={navClass('/reports')}>Reports</a>
          </nav>
        </div>
        <ThemeToggle />
      </div>
    </header>
  );
}
