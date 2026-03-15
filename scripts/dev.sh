#!/usr/bin/env bash
set -euo pipefail

if ! command -v wails >/dev/null 2>&1; then
  echo "Wails CLI is required. Install it with:" >&2
  echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest" >&2
  exit 1
fi

dev_tags=()

if [[ "$(uname -s)" == "Linux" ]] && command -v pkg-config >/dev/null 2>&1; then
  if pkg-config --exists webkit2gtk-4.1; then
    dev_tags=(-tags webkit2_41)
  fi
fi

wails dev "${dev_tags[@]}"
