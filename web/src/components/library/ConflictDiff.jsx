// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { useSignal } from '@preact/signals';
import { api } from '../../utils/fetch';
import { ENDPOINTS } from '../../utils/constants';
import { LuaViewer } from '../dashboard/LuaViewer';

function sameJSON(a, b) {
  return JSON.stringify(a) === JSON.stringify(b);
}

function NotificationList({ notifications }) {
  if (!notifications?.length) {
    return <span class="text-gray-400 dark:text-gray-500 italic">none</span>;
  }
  return (
    <div class="flex flex-wrap gap-1">
      {notifications.map((n, i) => (
        <span key={i} class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300">
          {n.name} <span class="text-gray-400 dark:text-gray-500 ml-1">{n.type}</span>
        </span>
      ))}
    </div>
  );
}

function SelectorList({ selectors }) {
  if (!selectors?.length) {
    return <span class="text-gray-400 dark:text-gray-500 italic">none</span>;
  }
  return (
    <div class="flex flex-wrap gap-1">
      {selectors.map((s, i) => (
        <span key={i} class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300">
          {s.kind}
        </span>
      ))}
    </div>
  );
}

function DiffRow({ label, children }) {
  return (
    <div class="grid grid-cols-[100px_1fr] gap-x-2 items-start">
      <span class="text-gray-500 dark:text-gray-400 pt-0.5">{label}</span>
      {children}
    </div>
  );
}

// Shown when posting a Cleaner hits a name conflict (409): a structured
// before/after comparison - not a line-level text diff - so the user can see
// what "Update" would actually change before it overwrites the existing
// Cleaner. Updating always resyncs everything (selectors/Lua/action included)
// from the library recipe; that's the whole point of posting from here.
export function ConflictDiff({ name, sourceLibraryId, existing, proposed, onUpdated, onCancel }) {
  const updateState = useSignal('idle'); // idle | loading | success | error
  const updateMessage = useSignal('');
  const luaTab = useSignal('current'); // current | proposed

  const scheduleChanged = existing.schedule !== proposed.schedule;
  const actionChanged = existing.action !== proposed.action;
  const notificationsChanged = !sameJSON(existing.notifications, proposed.notifications);
  const selectionChanged = !sameJSON(existing.selectors, proposed.selectors) || existing.luaScript !== proposed.luaScript;

  async function handleUpdate(e) {
    e.stopPropagation();
    if (updateState.value === 'loading') return;

    updateState.value = 'loading';
    try {
      const updated = await api(ENDPOINTS.cleaner(name), {
        method: 'PUT',
        body: JSON.stringify({
          sourceLibraryId,
          schedule: proposed.schedule,
          action: proposed.action,
          notifications: proposed.notifications,
        }),
      });
      updateState.value = 'success';
      updateMessage.value = `Updated "${updated.name}"`;
      onUpdated?.(updated);
    } catch (err) {
      updateState.value = 'error';
      updateMessage.value = err.message;
    }
  }

  return (
    <div class="rounded-lg border border-amber-300 dark:border-amber-700 bg-amber-50 dark:bg-amber-900/10 p-4 space-y-4">
      <div class="flex items-start gap-2">
        <svg class="w-4 h-4 text-amber-500 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
        <p class="text-xs text-amber-800 dark:text-amber-300">
          A Cleaner named <span class="font-mono">{name}</span> already exists. Review what Update would change.
        </p>
      </div>

      <div class="space-y-2 text-xs">
        <DiffRow label="Schedule">
          {scheduleChanged ? (
            <span>
              <span class="font-mono line-through text-gray-400">{existing.schedule}</span>
              {' → '}
              <span class="font-mono text-gray-900 dark:text-white">{proposed.schedule}</span>
            </span>
          ) : (
            <span class="font-mono text-gray-500 dark:text-gray-400">{proposed.schedule} (unchanged)</span>
          )}
        </DiffRow>

        <DiffRow label="Action">
          {actionChanged ? (
            <span>
              <span class="font-mono line-through text-gray-400">{existing.action}</span>
              {' → '}
              <span class="font-mono text-gray-900 dark:text-white">{proposed.action}</span>
            </span>
          ) : (
            <span class="font-mono text-gray-500 dark:text-gray-400">{proposed.action} (unchanged)</span>
          )}
        </DiffRow>

        <DiffRow label="Notifications">
          {notificationsChanged ? (
            <div class="space-y-1">
              <div>
                <div class="text-[10px] text-gray-400 mb-0.5">Current</div>
                <NotificationList notifications={existing.notifications} />
              </div>
              <div>
                <div class="text-[10px] text-gray-400 mb-0.5">New</div>
                <NotificationList notifications={proposed.notifications} />
              </div>
            </div>
          ) : (
            <NotificationList notifications={proposed.notifications} />
          )}
        </DiffRow>

        <DiffRow label="Selectors">
          {selectionChanged ? (
            <div class="space-y-1">
              <div>
                <div class="text-[10px] text-gray-400 mb-0.5">Current</div>
                <SelectorList selectors={existing.selectors} />
              </div>
              <div>
                <div class="text-[10px] text-gray-400 mb-0.5">New (from library)</div>
                <SelectorList selectors={proposed.selectors} />
              </div>
            </div>
          ) : (
            <SelectorList selectors={proposed.selectors} />
          )}
        </DiffRow>
      </div>

      {selectionChanged && (
        <div>
          <div class="flex items-center gap-1 mb-1">
            <button
              type="button"
              onClick={() => { luaTab.value = 'current'; }}
              class={`px-2 py-0.5 rounded text-[10px] font-semibold ${
                luaTab.value === 'current'
                  ? 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white'
                  : 'text-gray-500 dark:text-gray-400'
              }`}
            >
              Current
            </button>
            <button
              type="button"
              onClick={() => { luaTab.value = 'proposed'; }}
              class={`px-2 py-0.5 rounded text-[10px] font-semibold ${
                luaTab.value === 'proposed'
                  ? 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white'
                  : 'text-gray-500 dark:text-gray-400'
              }`}
            >
              New (from library)
            </button>
          </div>
          <LuaViewer code={luaTab.value === 'current' ? existing.luaScript : proposed.luaScript} />
        </div>
      )}

      <div class="flex items-center gap-3 pt-1">
        <button
          onClick={handleUpdate}
          class="action-primary"
          disabled={updateState.value === 'loading' || updateState.value === 'success'}
        >
          {updateState.value === 'loading' ? 'Updating...' : 'Update Cleaner'}
        </button>
        <button
          type="button"
          onClick={onCancel}
          class="action border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700"
          disabled={updateState.value === 'loading'}
        >
          Cancel
        </button>
        {updateState.value === 'success' && (
          <span class="text-xs text-green-600 dark:text-green-400 font-medium">{updateMessage.value}</span>
        )}
        {updateState.value === 'error' && (
          <span class="text-xs text-red-600 dark:text-red-400">{updateMessage.value}</span>
        )}
      </div>
    </div>
  );
}
