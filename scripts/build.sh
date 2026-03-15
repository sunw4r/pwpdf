#!/usr/bin/env bash
set -euo pipefail

if ! command -v wails >/dev/null 2>&1; then
  echo "Wails CLI is required. Install it with:" >&2
  echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest" >&2
  exit 1
fi

build_tags=()

if [[ "$(uname -s)" == "Linux" ]] && command -v pkg-config >/dev/null 2>&1; then
  if pkg-config --exists webkit2gtk-4.1; then
    build_tags=(-tags webkit2_41)
  fi
fi

export_darwin_cli_binary() {
  local output_dir="$1"
  local app_bundle
  local binary_name
  local embedded_binary

  app_bundle="$(find "$output_dir" -maxdepth 1 -type d -name '*.app' | head -n 1)"
  [[ -n "$app_bundle" ]] || return 0

  binary_name="$(basename "$app_bundle" .app)"
  embedded_binary="$app_bundle/Contents/MacOS/$binary_name"

  if [[ ! -f "$embedded_binary" ]]; then
    echo "Missing macOS CLI binary inside app bundle: $embedded_binary" >&2
    return 1
  fi

  cp "$embedded_binary" "$output_dir/$binary_name"
  chmod +x "$output_dir/$binary_name"
}

go run ./scripts/syncicons
go test ./...
wails build -clean "${build_tags[@]}"

if [[ "$(uname -s)" == "Darwin" ]]; then
  export_darwin_cli_binary "build/bin"
fi
