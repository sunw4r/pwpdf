#!/usr/bin/env bash
set -Eeuo pipefail

# Release build matrix helper.
#
# Execution model:
# - native-arch builds run through `wails build` on the current host
# - foreign-arch Linux builds can fall back to Docker automatically
# - macOS desktop artifacts still require a macOS host or runner
#
# Practical prerequisites:
# - `go`, `npm`, and `wails` available on the host
# - for native Linux builds, host packages that satisfy the selected WebKit tag
# - for Docker-backed Linux builds, a working Docker daemon with binfmt/qemu support
#
# Useful overrides when a machine is missing something:
# - `DOCKER_CROSS_LINUX=0` to disable the Docker fallback entirely
# - `DOCKER_NETWORK_MODE=bridge` if `host` networking is not allowed in your environment
# - `LINUX_PLATFORMS=...`, `WINDOWS_PLATFORMS=...`, `MACOS_PLATFORMS=...` to narrow the matrix
# - `--selector ...` is required and decides which part of the matrix actually runs
#
# If a foreign-arch Linux build fails before the Go compile starts, first check:
# - Docker is installed and the daemon is running
# - `docker run --rm --platform linux/amd64 alpine uname -m` works for the target architecture
# - the builder image in `build/docker/wails-linux-builder.Dockerfile` still installs successfully
#
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

go run ./scripts/syncicons

WAILS_BIN="${WAILS_BIN:-}"
DIST_DIR="${DIST_DIR:-$ROOT_DIR/dist}"
FAIL_ON_SKIP="${FAIL_ON_SKIP:-0}"
FAIL_ON_FAILURE="${FAIL_ON_FAILURE:-1}"
DOCKER_BIN="${DOCKER_BIN:-docker}"
DOCKER_CROSS_LINUX="${DOCKER_CROSS_LINUX:-1}"
DOCKER_LINUX_BUILDER_FILE="${DOCKER_LINUX_BUILDER_FILE:-$ROOT_DIR/build/docker/wails-linux-builder.Dockerfile}"
DOCKER_LINUX_BUILDER_IMAGE_PREFIX="${DOCKER_LINUX_BUILDER_IMAGE_PREFIX:-pwpdf-wails-linux-builder}"
DOCKER_LINUX_TAGS="${DOCKER_LINUX_TAGS:-webkit2_41}"
DOCKER_NETWORK_MODE="${DOCKER_NETWORK_MODE:-host}"

HOST_OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
HOST_ARCH_RAW="$(uname -m)"
BUILD_SELECTOR=""
LINUX_PLATFORMS_RAW="${LINUX_PLATFORMS:-${LINUX_PLATFORM:-linux/amd64 linux/arm64}}"
WINDOWS_PLATFORMS_RAW="${WINDOWS_PLATFORMS:-${WINDOWS_PLATFORM:-windows/amd64 windows/arm64}}"
MACOS_PLATFORMS_RAW="${MACOS_PLATFORMS:-${MACOS_PLATFORM:-darwin/universal}}"

declare -a BUILT_PLATFORMS=()
declare -a SKIPPED_PLATFORMS=()
declare -a FAILED_PLATFORMS=()
declare -a LINUX_PLATFORMS=()
declare -a WINDOWS_PLATFORMS=()
declare -a MACOS_PLATFORMS=()
FRONTEND_BUILD_COMPLETED=0

print_usage() {
  cat <<'EOF'
Usage:
  scripts/build-all.sh --selector <value>

Required selector values:
  all
  linux
  windows
  macos
  linux/<arch>
  windows/<arch>
  darwin/<arch>

Examples:
  scripts/build-all.sh --selector all
  scripts/build-all.sh --selector linux
  scripts/build-all.sh --selector linux/amd64
  scripts/build-all.sh --selector windows/arm64

Notes:
  - Group selectors use the platform lists configured by LINUX_PLATFORMS, WINDOWS_PLATFORMS, and MACOS_PLATFORMS.
  - Exact platform selectors ignore the other platform groups and build only that target.
EOF
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --selector | -s)
        if [[ $# -lt 2 ]] || [[ -z "$2" ]]; then
          echo "--selector requires a value." >&2
          print_usage >&2
          exit 1
        fi
        BUILD_SELECTOR="$2"
        shift 2
        ;;
      --help | -h)
        print_usage
        exit 0
        ;;
      *)
        echo "Unknown argument: $1" >&2
        print_usage >&2
        exit 1
        ;;
    esac
  done

  if [[ -z "$BUILD_SELECTOR" ]]; then
    echo "--selector is required." >&2
    print_usage >&2
    exit 1
  fi
}

ensure_command() {
  local command_name="$1"
  if ! command -v "$command_name" >/dev/null 2>&1; then
    # Keep the failure explicit here so missing tooling is obvious in CI logs.
    echo "Missing required command: $command_name" >&2
    exit 1
  fi
}

resolve_wails_bin() {
  if [[ -n "$WAILS_BIN" ]]; then
    echo "$WAILS_BIN"
    return
  fi

  if command -v wails >/dev/null 2>&1; then
    echo "wails"
    return
  fi

  local gopath_bin
  gopath_bin="$(go env GOPATH 2>/dev/null)/bin/wails"
  if [[ -x "$gopath_bin" ]]; then
    echo "$gopath_bin"
    return
  fi

  echo "wails"
}

linux_tags() {
  # The host-side Wails Linux build uses the WebKit tag only when pkg-config can
  # actually resolve the matching development package. If this returns empty on a
  # Linux machine that should build natively, install the appropriate WebKitGTK
  # development package first.
  if [[ "$HOST_OS" == "linux" ]] && command -v pkg-config >/dev/null 2>&1; then
    if pkg-config --exists webkit2gtk-4.1; then
      printf '%s' "webkit2_41"
      return
    fi
  fi

  printf '%s' ""
}

normalize_arch() {
  case "$1" in
    x86_64 | amd64)
      printf '%s' "amd64"
      ;;
    aarch64 | arm64)
      printf '%s' "arm64"
      ;;
    *)
      printf '%s' "$1"
      ;;
  esac
}

platform_os() {
  printf '%s' "${1%%/*}"
}

platform_arch() {
  printf '%s' "${1##*/}"
}

parse_platform_list() {
  local raw_list="${1//,/ }"
  local platform

  for platform in $raw_list; do
    [[ -n "$platform" ]] && printf '%s\n' "$platform"
  done
}

array_has_values() {
  local array_name="$1"
  eval "[[ \${${array_name}[0]+x} ]]"
}

run_platform_group() {
  local array_name="$1"
  local tags="$2"
  local platform

  # Bash 3.2 with `set -u` treats empty arrays as unbound during direct
  # expansion, so expand through `eval` only when the array actually has items.
  eval "set -- \${${array_name}[@]+\"\${${array_name}[@]}\"}"

  for platform in "$@"; do
    run_wails_build "$platform" "$tags"
  done
}

copy_build_output() {
  local platform_slug="$1"
  local target_dir="$DIST_DIR/$platform_slug"

  if [[ ! -d build/bin ]]; then
    echo "Build output directory is missing: $ROOT_DIR/build/bin" >&2
    return 1
  fi

  mkdir -p "$target_dir"
  find "$target_dir" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
  cp -R build/bin/. "$target_dir/"
}

selector_target_os() {
  case "$1" in
    linux | linux/*)
      printf '%s' "linux"
      ;;
    windows | windows/*)
      printf '%s' "windows"
      ;;
    macos | darwin/*)
      printf '%s' "darwin"
      ;;
    *)
      printf '%s' ""
      ;;
  esac
}

validate_selector() {
  case "$1" in
    all | linux | windows | macos | linux/* | windows/* | darwin/*)
      return 0
      ;;
    *)
      echo "Unsupported selector: $1" >&2
      print_usage >&2
      exit 1
      ;;
  esac
}

filter_platforms_for_selector() {
  local selector="$1"
  local selector_os

  validate_selector "$selector"

  case "$selector" in
    all)
      return 0
      ;;
    linux)
      WINDOWS_PLATFORMS=()
      MACOS_PLATFORMS=()
      return 0
      ;;
    windows)
      LINUX_PLATFORMS=()
      MACOS_PLATFORMS=()
      return 0
      ;;
    macos)
      LINUX_PLATFORMS=()
      WINDOWS_PLATFORMS=()
      return 0
      ;;
  esac

  selector_os="$(selector_target_os "$selector")"

  case "$selector_os" in
    linux)
      LINUX_PLATFORMS=("$selector")
      WINDOWS_PLATFORMS=()
      MACOS_PLATFORMS=()
      ;;
    windows)
      LINUX_PLATFORMS=()
      WINDOWS_PLATFORMS=("$selector")
      MACOS_PLATFORMS=()
      ;;
    darwin)
      LINUX_PLATFORMS=()
      WINDOWS_PLATFORMS=()
      MACOS_PLATFORMS=("$selector")
      ;;
  esac
}

ensure_frontend_build() {
  if [[ "$FRONTEND_BUILD_COMPLETED" == "1" ]]; then
    return 0
  fi

  # Build the frontend on the host once and reuse it across targets.
  # For foreign-arch Linux builds this avoids running Node/npm inside qemu,
  # which is less reliable than using Docker only for the final Go compile.
  echo
  echo "==> Building frontend assets on host"

  if ! npm run build --prefix frontend; then
    echo "Failed to build frontend assets" >&2
    return 1
  fi

  FRONTEND_BUILD_COMPLETED=1
}

linux_platform_needs_docker() {
  local platform="$1"
  local target_os
  local target_arch

  target_os="$(platform_os "$platform")"
  target_arch="$(platform_arch "$platform")"

  [[ "$DOCKER_CROSS_LINUX" == "1" ]] || return 1
  [[ "$HOST_OS" == "linux" ]] || return 1
  [[ "$target_os" == "linux" ]] || return 1
  # Same-arch Linux builds stay native. Docker is only for foreign-arch Linux.
  [[ "$target_arch" != "$HOST_ARCH" ]] || return 1

  return 0
}

run_wails_build_in_docker() {
  local platform="$1"
  local tags="$2"
  local target_arch
  local container_platform
  local builder_image
  local builder_context

  target_arch="$(platform_arch "$platform")"
  container_platform="linux/$target_arch"
  builder_image="$DOCKER_LINUX_BUILDER_IMAGE_PREFIX:$target_arch"
  builder_context="$(dirname "$DOCKER_LINUX_BUILDER_FILE")"

  if ! command -v "$DOCKER_BIN" >/dev/null 2>&1; then
    echo "Skipping $platform: Docker is required for foreign-architecture Linux builds on this host." >&2
    echo "Install Docker or rerun with DOCKER_CROSS_LINUX=0 and a host-native toolchain for $platform." >&2
    FAILED_PLATFORMS+=("$platform")
    return 0
  fi

  if [[ -z "$tags" ]]; then
    tags="$DOCKER_LINUX_TAGS"
  fi

  echo
  echo "==> Building $platform in Docker ($container_platform)"

  if ! ensure_frontend_build; then
    FAILED_PLATFORMS+=("$platform")
    return 0
  fi

  if ! "$DOCKER_BIN" build \
    --network "$DOCKER_NETWORK_MODE" \
    --platform "$container_platform" \
    -f "$DOCKER_LINUX_BUILDER_FILE" \
    -t "$builder_image" \
    "$builder_context"; then
    # Common causes here:
    # - Docker cannot emulate the requested architecture yet
    # - the builder Dockerfile needs extra system packages
    # - module/package downloads need a different Docker network mode
    echo "Failed to build Docker builder image for $platform" >&2
    FAILED_PLATFORMS+=("$platform")
    return 0
  fi

  if ! "$DOCKER_BIN" run \
    --rm \
    --network "$DOCKER_NETWORK_MODE" \
    --platform "$container_platform" \
    --user "$(id -u):$(id -g)" \
    -e HOME=/tmp/pwpdf-home \
    -e GOCACHE=/tmp/pwpdf-go-cache \
    -e GOMODCACHE=/tmp/pwpdf-go-mod-cache \
    -e npm_config_cache=/tmp/pwpdf-npm-cache \
    -e PWPDF_TARGET_PLATFORM="$platform" \
    -e PWPDF_BUILD_TAGS="$tags" \
    -v "$ROOT_DIR:/workspace" \
    -w /workspace \
    "$builder_image" \
    /workspace/scripts/docker-build-target.sh; then
    # If this step fails after the image was created, inspect:
    # - qemu/binfmt stability for the target architecture
    # - missing Go/system dependencies in the builder image
    # - frontend assets under `frontend/dist`, which are produced on the host
    echo "Failed to build $platform in Docker" >&2
    FAILED_PLATFORMS+=("$platform")
    return 0
  fi

  if ! copy_build_output "${platform//\//-}"; then
    echo "Failed to collect build artifacts for $platform" >&2
    FAILED_PLATFORMS+=("$platform")
    return 0
  fi

  BUILT_PLATFORMS+=("$platform")
}

run_wails_build() {
  local platform="$1"
  local tags="$2"
  local platform_slug="${platform//\//-}"
  local -a build_args=(build -clean -platform "$platform")

  echo
  echo "==> Building $platform"

  if [[ "$platform" == darwin/* ]] && [[ "$HOST_OS" != "darwin" ]]; then
    # Wails desktop packaging for macOS remains host-specific.
    echo "Skipping $platform: Wails v2 does not support crosscompiling macOS from $HOST_OS."
    SKIPPED_PLATFORMS+=("$platform")
    return 0
  fi

  if linux_platform_needs_docker "$platform"; then
    run_wails_build_in_docker "$platform" "$tags"
    return 0
  fi

  if [[ -n "$tags" ]]; then
    build_args+=(-tags "$tags")
  fi

  # Native host path: prefer Wails directly whenever the target matches the host
  # architecture, because it packages the desktop app layout for that platform.
  if ! "$WAILS_BIN" "${build_args[@]}"; then
    echo "Failed to build $platform" >&2
    FAILED_PLATFORMS+=("$platform")
    return 0
  fi

  if ! copy_build_output "$platform_slug"; then
    echo "Failed to collect build artifacts for $platform" >&2
    FAILED_PLATFORMS+=("$platform")
    return 0
  fi

  BUILT_PLATFORMS+=("$platform")
}

print_summary() {
  echo
  echo "Build summary"
  echo "Artifacts directory: $DIST_DIR"

  if array_has_values BUILT_PLATFORMS; then
    printf 'Built: %s\n' "${BUILT_PLATFORMS[*]}"
  fi

  if array_has_values SKIPPED_PLATFORMS; then
    printf 'Skipped: %s\n' "${SKIPPED_PLATFORMS[*]}"
    echo "macOS artifacts require running this script on macOS or using CI with a macOS runner."
  fi

  if array_has_values FAILED_PLATFORMS; then
    printf 'Failed: %s\n' "${FAILED_PLATFORMS[*]}"
    echo "Some targets may require host-specific SDKs, cross toolchains, or Wails support on the current machine."
  fi
}

HOST_ARCH="$(normalize_arch "$HOST_ARCH_RAW")"

parse_args "$@"

# Host-side minimum requirements. Docker is checked lazily because it is only
# needed when the script actually decides to route a Linux target through it.
ensure_command go
ensure_command npm

WAILS_BIN="$(resolve_wails_bin)"
ensure_command "$WAILS_BIN"

LINUX_PLATFORMS=($(parse_platform_list "$LINUX_PLATFORMS_RAW"))
WINDOWS_PLATFORMS=($(parse_platform_list "$WINDOWS_PLATFORMS_RAW"))
MACOS_PLATFORMS=($(parse_platform_list "$MACOS_PLATFORMS_RAW"))

# The selector is required so CI callers have to be explicit about whether they
# want the full matrix, one OS family, or a single exact target.
filter_platforms_for_selector "$BUILD_SELECTOR"

mkdir -p "$DIST_DIR"

LINUX_TAGS="$(linux_tags)"

run_platform_group LINUX_PLATFORMS "$LINUX_TAGS"
run_platform_group WINDOWS_PLATFORMS ""
run_platform_group MACOS_PLATFORMS ""

print_summary

if [[ "$FAIL_ON_SKIP" == "1" ]] && array_has_values SKIPPED_PLATFORMS; then
  exit 1
fi

if [[ "$FAIL_ON_FAILURE" == "1" ]] && array_has_values FAILED_PLATFORMS; then
  exit 1
fi
