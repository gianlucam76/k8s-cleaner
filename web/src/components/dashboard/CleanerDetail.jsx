// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { useSignal } from '@preact/signals';
import { useEffect } from 'preact/hooks';
import { api } from '../../utils/fetch';
import { configData } from '../../app';
import { ENDPOINTS } from '../../utils/constants';
import { LuaViewer } from './LuaViewer';
import { TriggerButton } from './TriggerButton';

export function CleanerDetail({ cleaner }) {
  const detail = useSignal(null);
  const report = useSignal(null);
  const loading = useSignal(true);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      loading.value = true;
      try {
        const [d, r] = await Promise.all([
          api(ENDPOINTS.cleaner(cleaner.name)),
          api(ENDPOINTS.report(cleaner.name)).catch(() => null),
        ]);
        if (!cancelled) {
          detail.value = d;
          report.value = r;
        }
      } catch (err) {
        console.error('Failed to load cleaner detail:', err);
      } finally {
        if (!cancelled) loading.value = false;
      }
    }
    load();
    return () => { cancelled = true; };
  }, [cleaner.name]);

  if (loading.value) {
    return (
      <div class="mt-3 pt-3 border-t border-gray-200 dark:border-gray-600 space-y-2">
        <div class="skeleton h-4 w-32 rounded" />
        <div class="skeleton h-20 rounded" />
      </div>
    );
  }

  const readOnly = configData.value?.readOnly ?? true;
  const resources = report.value?.resources || [];

  return (
    <div class="mt-3 pt-3 border-t border-gray-200 dark:border-gray-600 space-y-3">
      {/* Selectors */}
      <div>
        <div class="text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">
          Resource Selectors
        </div>
        <div class="flex flex-wrap gap-1">
          {cleaner.selectors?.map((s, i) => (
            <span key={i} class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300">
              {s.kind}
              {s.group && <span class="text-gray-400 dark:text-gray-500 ml-0.5">({s.group})</span>}
            </span>
          ))}
        </div>
      </div>

      {/* Lua Script */}
      <div>
        <div class="text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">
          Lua Script
        </div>
        <LuaViewer code={detail.value?.luaScript} />
      </div>

      {/* Flagged Resources */}
      {resources.length > 0 && (
        <div>
          <div class="text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">
            Flagged Resources ({resources.length})
          </div>
          <div class="space-y-1 max-h-48 overflow-y-auto">
            {resources.map((res, i) => (
              <div key={i} class="flex items-center gap-2 text-xs text-gray-600 dark:text-gray-400">
                <span class="font-mono bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400 px-1 rounded text-[10px]">
                  {res.kind}
                </span>
                <span class="truncate">
                  {res.namespace && <span class="text-gray-400 dark:text-gray-500">{res.namespace}/</span>}
                  {res.name}
                </span>
                {res.message && (
                  <span class="text-gray-400 dark:text-gray-500 truncate ml-auto text-[10px]">{res.message}</span>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {resources.length === 0 && cleaner.flaggedCount === 0 && (
        <div class="text-xs text-gray-400 dark:text-gray-500 italic">
          No flagged resources in latest scan
        </div>
      )}

      {/* Actions */}
      <div class="flex items-center justify-between pt-1">
        <TriggerButton cleanerName={cleaner.name} readOnly={readOnly} />
        {resources.length > 0 && (
          <a href={`/reports#${cleaner.name}`} class="text-xs text-blue-600 hover:underline">
            View full report
          </a>
        )}
      </div>
    </div>
  );
}
