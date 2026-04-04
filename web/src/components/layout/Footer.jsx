// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { configData } from '../../app';

export function Footer() {
  const ver = configData.value?.version;

  return (
    <footer class="border-t border-gray-200 dark:border-gray-700 mt-auto">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div class="flex items-center justify-between text-xs text-gray-400 dark:text-gray-500">
          <span>
            k8s-cleaner{ver ? ` ${ver}` : ''}
          </span>
          <a
            href="https://github.com/gianlucam76/k8s-cleaner"
            target="_blank"
            rel="noopener noreferrer"
            class="hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>
  );
}
