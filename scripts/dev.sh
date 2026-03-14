#!/usr/bin/env bash
set -euo pipefail

if ! command -v wails >/dev/null 2>&1; then
  echo "Wails CLI is required. Install it with:" >&2
  echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest" >&2
  exit 1
fi

wails dev -tags webkit2_41
