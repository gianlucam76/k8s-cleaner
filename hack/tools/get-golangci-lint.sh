#!/bin/bash

set -euo pipefail

# Define the URL for downloading the golangci-lint archive
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(pwd)/bin "$1"