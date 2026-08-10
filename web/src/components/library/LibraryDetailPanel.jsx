// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { useSignal } from '@preact/signals';
import { useEffect, useRef } from 'preact/hooks';
import { api } from '../../utils/fetch';
import { configData, enableFastRefresh } from '../../app';
import { ENDPOINTS } from '../../utils/constants';
import { LuaViewer } from '../dashboard/LuaViewer';
import { NotificationsEditor } from './NotificationsEditor';
import { ConflictDiff } from './ConflictDiff';

// Slide-over panel: read-only preview of a library recipe (selectors + Lua),
// plus a form limited to name/schedule/notifications - the recipe's selection
// logic is never editable here, it always comes from the library as-is.
export function LibraryDetailPanel({ entryId, isOpen, onClose, onCreated }) {
  const panelRef = useRef(null);
  const detail = useSignal(null);
  const loading = useSignal(true);

  const name = useSignal('');
  const schedule = useSignal('');
  const action = useSignal('Scan');
  const notifications = useSignal([{ name: 'report', type: 'CleanerReport' }]);

  const postState = useSignal('idle'); // idle | loading | success | error | conflict
  const postMessage = useSignal('');
  const conflictExisting = useSignal(null);

  useEffect(() => {
    if (!isOpen || !entryId) return;
    let cancelled = false;

    async function load() {
      loading.value = true;
      postState.value = 'idle';
      conflictExisting.value = null;
      try {
        const d = await api(ENDPOINTS.libraryEntry(entryId));
        if (cancelled) return;
        detail.value = d;
        name.value = d?.id || '';
        schedule.value = d?.schedule || '';
        action.value = 'Scan'; // every library recipe is authored as Scan
        notifications.value = [{ name: 'report', type: 'CleanerReport' }];
      } catch (err) {
        console.error('Failed to load library entry:', err);
      } finally {
        if (!cancelled) loading.value = false;
      }
    }
    load();
    return () => { cancelled = true; };
  }, [entryId, isOpen]);

  useEffect(() => {
    if (!isOpen) return;
    function handleKey(e) {
      if (e.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [isOpen, onClose]);

  useEffect(() => {
    if (!isOpen) return;
    function handleClick(e) {
      if (panelRef.current && !panelRef.current.contains(e.target)) onClose();
    }
    const timer = setTimeout(() => window.addEventListener('click', handleClick), 10);
    return () => {
      clearTimeout(timer);
      window.removeEventListener('click', handleClick);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const readOnly = configData.value?.readOnly ?? true;

  async function handlePost(e) {
    e.stopPropagation();
    if (postState.value === 'loading') return;
    if (!name.value.trim() || !schedule.value.trim()) {
      postState.value = 'error';
      postMessage.value = 'Name and schedule are required';
      return;
    }

    postState.value = 'loading';
    try {
      const created = await api(ENDPOINTS.cleaners, {
        method: 'POST',
        body: JSON.stringify({
          sourceLibraryId: entryId,
          name: name.value.trim(),
          schedule: schedule.value.trim(),
          action: action.value,
          notifications: notifications.value,
        }),
      });
      postState.value = 'success';
      postMessage.value = `Created "${created.name}"`;
      enableFastRefresh();
      onCreated?.(created);
    } catch (err) {
      if (err.status === 409) {
        try {
          conflictExisting.value = await api(ENDPOINTS.cleaner(name.value.trim()));
          postState.value = 'conflict';
          return;
        } catch {
          // fall through to the generic error below if the existing cleaner can't be fetched
        }
      }
      postState.value = 'error';
      postMessage.value = err.message;
    }
  }

  function handleUpdated(updated) {
    enableFastRefresh();
    onCreated?.(updated);
  }

  return (
    <div
      ref={panelRef}
      class="fixed top-3 right-3 bottom-3 w-full sm:w-[58%] lg:w-[52%] xl:w-[48%] z-30 bg-white dark:bg-gray-800 shadow-[-20px_0_60px_rgba(0,0,0,0.15)] dark:shadow-[-20px_0_60px_rgba(0,0,0,0.4)] flex flex-col transition-transform duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] rounded-2xl"
    >
      <div class="px-6 pt-5 pb-4 border-b border-gray-100 dark:border-gray-700/50 flex-shrink-0 rounded-t-2xl">
        <div class="flex items-start justify-between">
          <h2 class="text-base font-bold text-gray-900 dark:text-white">{detail.value?.title || entryId}</h2>
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
        {detail.value?.description && (
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">{detail.value.description}</p>
        )}
      </div>

      <div class="flex-1 overflow-y-auto px-6 py-5">
        {loading.value ? (
          <div class="space-y-2">
            <div class="skeleton h-4 w-32 rounded" />
            <div class="skeleton h-20 rounded" />
          </div>
        ) : (
          <div class="space-y-5">
            <div>
              <div class="text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">
                Resource Selectors
              </div>
              <div class="flex flex-wrap gap-1">
                {detail.value?.selectors?.map((s, i) => (
                  <span key={i} class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300">
                    {s.kind}
                    {s.group && <span class="text-gray-400 dark:text-gray-500 ml-0.5">({s.group})</span>}
                  </span>
                ))}
              </div>
            </div>

            <div>
              <div class="text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">
                Lua Script
              </div>
              <LuaViewer code={detail.value?.luaScript} />
            </div>

            <div class="border-t border-gray-100 dark:border-gray-700/50 pt-4 space-y-4">
              <div>
                <label class="block text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">
                  Cleaner Name
                </label>
                <input
                  type="text"
                  value={name.value}
                  onInput={(e) => { name.value = e.currentTarget.value; }}
                  class="w-full px-2 py-1.5 text-sm rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label class="block text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">
                  Schedule (cron)
                </label>
                <input
                  type="text"
                  value={schedule.value}
                  onInput={(e) => { schedule.value = e.currentTarget.value; }}
                  class="w-full px-2 py-1.5 text-sm font-mono rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label class="block text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">
                  Action
                </label>
                <select
                  value={action.value}
                  onChange={(e) => { action.value = e.currentTarget.value; }}
                  class="w-full px-2 py-1.5 text-sm rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-gray-900 dark:text-white"
                >
                  <option value="Scan">Scan (report only)</option>
                  <option value="Delete">Delete (remove matching resources)</option>
                </select>
                {action.value === 'Delete' && (
                  <p class="text-[11px] text-amber-600 dark:text-amber-400 mt-1">
                    Delete removes matching resources on every scheduled run. Consider testing as Scan first.
                  </p>
                )}
              </div>

              <div>
                <label class="block text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">
                  Notifications
                </label>
                <NotificationsEditor
                  notifications={notifications.value}
                  onChange={(next) => { notifications.value = next; }}
                />
              </div>

              {readOnly ? (
                <p class="text-xs text-gray-400 dark:text-gray-500 italic">
                  Dashboard is in read-only mode.
                </p>
              ) : postState.value === 'conflict' ? (
                <ConflictDiff
                  name={name.value.trim()}
                  sourceLibraryId={entryId}
                  existing={conflictExisting.value}
                  proposed={{
                    schedule: schedule.value.trim(),
                    action: action.value,
                    notifications: notifications.value,
                    selectors: detail.value?.selectors || [],
                    luaScript: detail.value?.luaScript || '',
                  }}
                  onUpdated={handleUpdated}
                  onCancel={() => { postState.value = 'idle'; conflictExisting.value = null; }}
                />
              ) : (
                <div class="flex items-center gap-3">
                  <button
                    onClick={handlePost}
                    class="action-primary"
                    disabled={postState.value === 'loading' || postState.value === 'success'}
                  >
                    {postState.value === 'loading' ? 'Posting...' : 'Post Cleaner'}
                  </button>
                  {postState.value === 'success' && (
                    <span class="text-xs text-green-600 dark:text-green-400 font-medium">{postMessage.value}</span>
                  )}
                  {postState.value === 'error' && (
                    <span class="text-xs text-red-600 dark:text-red-400">{postMessage.value}</span>
                  )}
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
