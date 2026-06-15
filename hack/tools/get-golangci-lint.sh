#!/bin/bash

set -euo pipefail

VERSION="$1"
BINARY="$(pwd)/bin/golangci-lint"

# Skip download if the binary already exists at the requested version.
if [[ -x "$BINARY" ]] && "$BINARY" --version 2>/dev/null | grep -qF "${VERSION#v}"; then
    echo "golangci-lint ${VERSION} already installed, skipping download"
    exit 0
fi

# Retry up to 3 times — the GitHub CDN occasionally returns 504.
for attempt in 1 2 3; do
    if curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b "$(pwd)/bin" "$VERSION"; then
        exit 0
    fi
    echo "golangci-lint download attempt ${attempt} failed, retrying..." >&2
    sleep $((attempt * 5))
done

echo "golangci-lint download failed after 3 attempts" >&2
exit 1
