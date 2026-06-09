#!/bin/bash
#
# Build the Windows enrollment UI helper (probo-agent-enroll-ui.exe).
#
# Required arguments:
#   --arch ARCH   Target architecture: amd64 or arm64.
#   --output PATH Output directory. Defaults to dist/.
#
# Requires the .NET 8 SDK. Cross-compiles from macOS or Linux when
# --arch is set; on Windows, omit --arch to build for the host.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"
PROJECT="${SCRIPT_DIR}/enroll-ui/probo-agent-enroll-ui.csproj"

ARCH=""
OUTPUT="${REPO_ROOT}/dist"

usage() {
    sed -ne '/^#/!q; s/^# \{0,1\}//; 2,$ p' < "$0"
}

while [ $# -gt 0 ]; do
    case "$1" in
        --arch)   ARCH="$2";   shift 2 ;;
        --output) OUTPUT="$2"; shift 2 ;;
        -h|--help) usage; exit 0 ;;
        *) echo "unknown flag: $1" >&2; usage >&2; exit 2 ;;
    esac
done

if ! command -v dotnet >/dev/null 2>&1; then
    echo "error: dotnet SDK is required" >&2
    exit 1
fi

RUNTIME=""
case "${ARCH}" in
    amd64) RUNTIME="win-x64" ;;
    arm64) RUNTIME="win-arm64" ;;
    "")
        if [ "$(uname -s)" = "MINGW"* ] || [ "$(uname -s)" = "MSYS"* ] || [ "$(uname -s)" = "CYGWIN"* ]; then
            RUNTIME="win-x64"
        else
            echo "error: --arch (amd64|arm64) is required when not building on Windows" >&2
            exit 2
        fi
        ;;
    *)
        echo "error: unsupported --arch '${ARCH}' (want amd64 or arm64)" >&2
        exit 2
        ;;
esac

mkdir -p "${OUTPUT}"

echo "Building probo-agent-enroll-ui (${RUNTIME})..."
dotnet publish "${PROJECT}" \
    -c Release \
    -r "${RUNTIME}" \
    --self-contained true \
    -p:PublishSingleFile=true \
    -p:IncludeNativeLibrariesForSelfExtract=true \
    -o "${OUTPUT}"

if [ ! -f "${OUTPUT}/probo-agent-enroll-ui.exe" ]; then
    echo "error: build did not produce ${OUTPUT}/probo-agent-enroll-ui.exe" >&2
    exit 1
fi

if [ ! -f "${OUTPUT}/regions.json" ]; then
    echo "error: build did not produce ${OUTPUT}/regions.json" >&2
    exit 1
fi

echo "Built ${OUTPUT}/probo-agent-enroll-ui.exe"
