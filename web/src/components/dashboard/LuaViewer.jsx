// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { useRef, useEffect } from 'preact/hooks';
import { useSignal } from '@preact/signals';
import hljs from 'highlight.js/lib/core';
import lua from 'highlight.js/lib/languages/lua';

hljs.registerLanguage('lua', lua);

export function LuaViewer({ code }) {
  const codeRef = useRef(null);
  const copied = useSignal(false);
  const search = useSignal('');
  const matchCount = useSignal(0);

  useEffect(() => {
    if (!codeRef.current || !code) return;

    if (search.value) {
      // Highlight search matches by wrapping them in <mark> tags
      const escaped = search.value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      const regex = new RegExp(`(${escaped})`, 'gi');
      const highlighted = code
        .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
        .replace(regex, '<mark class="bg-yellow-300 dark:bg-yellow-600 text-inherit rounded-sm px-0.5">$1</mark>');
      const matches = code.match(regex);
      matchCount.value = matches ? matches.length : 0;
      codeRef.current.innerHTML = highlighted;
    } else {
      matchCount.value = 0;
      codeRef.current.textContent = code;
      hljs.highlightElement(codeRef.current);
    }
  }, [code, search.value]);

  function copyToClipboard(e) {
    e.stopPropagation();
    navigator.clipboard.writeText(code).then(() => {
      copied.value = true;
      setTimeout(() => { copied.value = false; }, 2000);
    });
  }

  if (!code) {
    return (
      <div class="text-xs text-gray-400 dark:text-gray-500 italic">
        No Lua script available
      </div>
    );
  }

  return (
    <div class="rounded-md overflow-hidden border border-gray-200 dark:border-gray-600">
      <div class="flex items-center justify-between px-3 py-1.5 bg-gray-100 dark:bg-gray-700 border-b border-gray-200 dark:border-gray-600 gap-2">
        <span class="text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider flex-shrink-0">Lua</span>
        <div class="flex items-center gap-1.5 flex-1 justify-end">
          <div class="relative flex items-center">
            <svg class="w-3 h-3 text-gray-400 absolute left-1.5 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
            </svg>
            <input
              type="text"
              placeholder="Search..."
              class="w-28 text-[10px] pl-5 pr-1.5 py-0.5 rounded border border-gray-300 dark:border-gray-500 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-blue-500"
              value={search.value}
              onInput={(e) => { search.value = e.target.value; }}
              onClick={(e) => e.stopPropagation()}
            />
            {search.value && (
              <span class="text-[9px] text-gray-400 dark:text-gray-500 ml-1 flex-shrink-0">
                {matchCount.value}
              </span>
            )}
          </div>
          <span class="text-[10px] text-gray-400 dark:text-gray-500 flex-shrink-0">{code.length} chars</span>
          <button
            onClick={copyToClipboard}
            class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors flex-shrink-0"
            title="Copy to clipboard"
          >
            {copied.value ? (
              <svg class="w-3.5 h-3.5 text-green-500" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
              </svg>
            ) : (
              <svg class="w-3.5 h-3.5 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" />
              </svg>
            )}
          </button>
        </div>
      </div>
      <pre class="p-3 overflow-x-auto bg-gray-50 dark:bg-gray-900 text-xs leading-relaxed max-h-64 overflow-y-auto">
        <code ref={codeRef} class="language-lua">{code}</code>
      </pre>
    </div>
  );
}
