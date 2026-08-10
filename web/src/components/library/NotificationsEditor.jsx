// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { NOTIFICATION_TYPES } from '../../utils/constants';

// Editable list of Notification{name, type} entries. Starts pre-populated with a
// CleanerReport (see LibraryDetailPanel), but the user is free to remove or
// change any row and add others (Slack, Event, etc.).
export function NotificationsEditor({ notifications, onChange }) {
  function updateRow(index, field, value) {
    const next = notifications.map((n, i) => (i === index ? { ...n, [field]: value } : n));
    onChange(next);
  }

  function removeRow(index) {
    onChange(notifications.filter((_, i) => i !== index));
  }

  function addRow() {
    onChange([...notifications, { name: '', type: 'CleanerReport' }]);
  }

  return (
    <div class="space-y-2">
      {notifications.map((n, i) => (
        <div key={i} class="flex items-center gap-2">
          <input
            type="text"
            value={n.name}
            onInput={(e) => updateRow(i, 'name', e.currentTarget.value)}
            placeholder="name"
            class="flex-1 min-w-0 px-2 py-1 text-xs rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-gray-900 dark:text-white"
          />
          <select
            value={n.type}
            onChange={(e) => updateRow(i, 'type', e.currentTarget.value)}
            class="px-2 py-1 text-xs rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-gray-900 dark:text-white"
          >
            {NOTIFICATION_TYPES.map((t) => (
              <option key={t} value={t}>{t}</option>
            ))}
          </select>
          <button
            type="button"
            onClick={() => removeRow(i)}
            aria-label="Remove notification"
            class="p-1 rounded-md text-gray-400 hover:text-rose-500 hover:bg-rose-50 dark:hover:bg-rose-900/20"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      ))}
      <button type="button" onClick={addRow} class="text-xs text-blue-600 dark:text-blue-400 hover:underline">
        + Add notification
      </button>
    </div>
  );
}
