#!/bin/bash
#
# Build Probo Agent.app — a headless macOS app bundle that registers
# the probo:// URL scheme and forwards enrollment links to probo-agent.
#
# Required arguments:
#   --arch     amd64, arm64, or universal
#   --version  Agent version, e.g. 0.1.0
#   --output   Parent directory; creates "Probo Agent.app" inside it
#
# Must run on macOS with the Swift toolchain (swift build).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

ARCH=""
VERSION=""
OUTPUT=""
APP_NAME="Probo Agent.app"
EXECUTABLE_NAME="probo-agent-url-handler"

usage() {
    sed -ne '/^#/!q; s/^# \{0,1\}//; 2,$ p' < "$0"
}

while [ $# -gt 0 ]; do
    case "$1" in
        --arch)    ARCH="$2";    shift 2 ;;
        --version) VERSION="$2"; shift 2 ;;
        --output)  OUTPUT="$2";  shift 2 ;;
        -h|--help) usage; exit 0 ;;
        *)         echo "unknown flag: $1" >&2; usage >&2; exit 2 ;;
    esac
done

if [ -z "${ARCH}" ]; then
    echo "error: --arch (amd64|arm64|universal) is required" >&2
    exit 2
fi
case "${ARCH}" in
    amd64)     SWIFT_ARCH_ARGS=(--arch x86_64) ;;
    arm64)     SWIFT_ARCH_ARGS=(--arch arm64) ;;
    universal) SWIFT_ARCH_ARGS=(--arch arm64 --arch x86_64) ;;
    *)
        echo "error: unsupported --arch '${ARCH}' (want amd64, arm64, or universal)" >&2
        exit 2
        ;;
esac
if [ -z "${VERSION}" ]; then
    echo "error: --version is required" >&2
    exit 2
fi
if [ -z "${OUTPUT}" ]; then
    echo "error: --output is required" >&2
    exit 2
fi

if ! command -v swift >/dev/null 2>&1; then
    echo "error: swift is required (run on macOS with Xcode or Swift toolchain)" >&2
    exit 1
fi

BUILD_DIR="$(mktemp -d -t probo-agent-url-handler-build)"
trap 'rm -rf "${BUILD_DIR}"' EXIT

pushd "${SCRIPT_DIR}" >/dev/null
swift build -c release "${SWIFT_ARCH_ARGS[@]}" --scratch-path "${BUILD_DIR}"
BIN_DIR="$(swift build -c release "${SWIFT_ARCH_ARGS[@]}" --scratch-path "${BUILD_DIR}" --show-bin-path)"
BINARY="${BIN_DIR}/${EXECUTABLE_NAME}"
popd >/dev/null

if [ ! -x "${BINARY}" ]; then
    echo "error: release binary not found at ${BINARY}" >&2
    exit 1
fi

APP_ROOT="${OUTPUT}/${APP_NAME}"
CONTENTS="${APP_ROOT}/Contents"
MACOS="${CONTENTS}/MacOS"
PLIST="${CONTENTS}/Info.plist"

rm -rf "${APP_ROOT}"
mkdir -p "${MACOS}"

install -m 0755 "${BINARY}" "${MACOS}/${EXECUTABLE_NAME}"

sed \
    -e "s|@@VERSION@@|${VERSION}|g" \
    "${SCRIPT_DIR}/Info.plist.tmpl" > "${PLIST}"

if ! plutil -lint "${PLIST}" >/dev/null; then
    echo "error: rendered Info.plist failed plutil -lint" >&2
    exit 1
fi
if ! grep -q '<string>probo</string>' "${PLIST}"; then
    echo "error: Info.plist is missing probo URL scheme" >&2
    exit 1
fi

echo "Built ${APP_ROOT}"
