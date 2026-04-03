// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useRef } from 'preact/hooks';

export function SlidePanel({ isOpen, onClose, cleaner, children }) {
  const panelRef = useRef(null);

  // Close on Escape
  useEffect(() => {
    if (!isOpen) return;
    function handleKey(e) {
      if (e.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [isOpen, onClose]);

  // Close on click outside panel
  useEffect(() => {
    if (!isOpen) return;
    function handleClick(e) {
      if (panelRef.current && !panelRef.current.contains(e.target)) {
        onClose();
      }
    }
    // Delay to avoid closing on the same click that opened it
    const timer = setTimeout(() => {
      window.addEventListener('click', handleClick);
    }, 10);
    return () => {
      clearTimeout(timer);
      window.removeEventListener('click', handleClick);
    };
  }, [isOpen, onClose]);

  if (!isOpen || !cleaner) return null;

  const hasFlagged = cleaner.flaggedCount > 0;

  return (
    <div
      ref={panelRef}
      class="fixed top-0 right-0 bottom-0 w-full sm:w-[58%] lg:w-[52%] xl:w-[48%] z-30 bg-white dark:bg-gray-800 shadow-[-20px_0_60px_rgba(0,0,0,0.15)] dark:shadow-[-20px_0_60px_rgba(0,0,0,0.4)] flex flex-col transition-transform duration-300 ease-[cubic-bezier(0.16,1,0.3,1)]"
      style={{ transform: 'translateX(0)' }}
    >
      {/* Header */}
      <div class="px-6 pt-5 pb-4 border-b border-gray-100 dark:border-gray-700/50 flex-shrink-0 bg-white dark:bg-gray-800">
        <div class="flex items-start justify-between">
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2.5 mb-2">
              <h2 class="text-base font-bold truncate text-gray-900 dark:text-white">
                {cleaner.name}
              </h2>
              {hasFlagged ? (
                <span class="badge-alert">{cleaner.flaggedCount}</span>
              ) : (
                <span class="badge-ok">Clean</span>
              )}
            </div>
            <div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
              <span class="flex items-center gap-1">
                <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                {cleaner.schedule}
              </span>
              <span>{cleaner.action}</span>
              {cleaner.lastRunTime && (
                <span>Last: {new Date(cleaner.lastRunTime).toLocaleTimeString()}</span>
              )}
            </div>
          </div>
          <button
            onClick={onClose}
            class="p-1.5 -mt-1 -mr-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
            aria-label="Close panel"
          >
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      {/* Scrollable content */}
      <div class="flex-1 overflow-y-auto px-6 py-5">
        {children}
      </div>
    </div>
  );
}
