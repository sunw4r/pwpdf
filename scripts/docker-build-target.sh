#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET_PLATFORM="${PWPDF_TARGET_PLATFORM:-}"
BUILD_TAGS="${PWPDF_BUILD_TAGS:-}"
TARGET_OS="${TARGET_PLATFORM%%/*}"
OUTPUT_PATH="${PWPDF_OUTPUT_PATH:-}"

if [[ -z "$TARGET_PLATFORM" ]]; then
  echo "PWPDF_TARGET_PLATFORM is required." >&2
  exit 1
fi

if [[ "$TARGET_OS" != "linux" ]]; then
  echo "Docker build helper only supports Linux targets. Received: $TARGET_PLATFORM" >&2
  exit 1
fi

if [[ -z "$OUTPUT_PATH" ]]; then
  OUTPUT_PATH="build/bin/pwpdf"
fi

mkdir -p \
  "${HOME:-/tmp/pwpdf-home}" \
  "${GOCACHE:-/tmp/pwpdf-go-cache}" \
  "${GOMODCACHE:-/tmp/pwpdf-go-mod-cache}" \
  "${npm_config_cache:-/tmp/pwpdf-npm-cache}"

cd "$ROOT_DIR"

rm -rf build/bin
mkdir -p "$(dirname "$OUTPUT_PATH")"

build_args=(build -o "$OUTPUT_PATH")

if [[ -n "$BUILD_TAGS" ]]; then
  build_args+=(-tags "$BUILD_TAGS")
fi

build_args+=(.)

go "${build_args[@]}"
