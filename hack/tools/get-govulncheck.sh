#!/bin/bash

set -euo pipefail

GOBIN=$(pwd)/bin go install golang.org/x/vuln/cmd/govulncheck@$1
