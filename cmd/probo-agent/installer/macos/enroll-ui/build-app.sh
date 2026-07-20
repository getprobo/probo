#!/bin/bash
#
# Build Probo Agent.app — headless macOS app bundle with probo:// handler
# and embedded privileged helper (installed by PKG postinstall).
#
# Required arguments:
#   --arch     amd64, arm64, or universal
#   --version  Agent version, e.g. 0.1.0
#   --output   Parent directory; creates "Probo Agent.app" inside it
#
# Required environment variables:
#   CODESIGN_IDENTITY  Developer ID Application identity. Signs the
#                      embedded helper (for SMPrivilegedExecutables DR),
#                      URL handler, and app bundle.
#   APPLE_TEAM_ID      Apple Developer Team ID for SMAuthorizedClients.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

ARCH=""
VERSION=""
OUTPUT=""
APP_NAME="Probo Agent.app"
URL_HANDLER_NAME="probo-agent-url-handler"
HELPER_LABEL="com.probo.agent.helper"
CODESIGN_IDENTITY="${CODESIGN_IDENTITY:-}"
APPLE_TEAM_ID="${APPLE_TEAM_ID:-}"

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
if [ -z "${CODESIGN_IDENTITY}" ]; then
    echo "error: CODESIGN_IDENTITY is required (privileged helper must be signed)" >&2
    exit 2
fi
if [ -z "${APPLE_TEAM_ID}" ]; then
    echo "error: APPLE_TEAM_ID is required (SMAuthorizedClients team requirement)" >&2
    exit 2
fi

BUILD_DIR="$(mktemp -d -t probo-agent-enroll-ui-build)"
RENDER_DIR="${BUILD_DIR}/rendered"
trap 'rm -rf "${BUILD_DIR}"' EXIT
mkdir -p "${RENDER_DIR}"

client_requirement() {
    if [ -z "${APPLE_TEAM_ID}" ]; then
        echo "error: APPLE_TEAM_ID is required (client designated requirement)" >&2
        exit 2
    fi
    printf 'anchor apple generic and identifier "com.probo.agent.url-handler" and certificate leaf[subject.OU] = "%s"' "${APPLE_TEAM_ID}"
}

team_id_option() {
    if [ -n "${APPLE_TEAM_ID}" ]; then
        printf '"%s"' "${APPLE_TEAM_ID}"
        return
    fi
    printf 'nil'
}

sed \
    -e "s|@@VERSION@@|${VERSION}|g" \
    "${SCRIPT_DIR}/Shared/HelperVersion.generated.swift.tmpl" \
    > "${SCRIPT_DIR}/Shared/HelperVersion.generated.swift"

sed \
    -e "s|@@TEAM_ID_OPTION@@|$(team_id_option)|g" \
    "${SCRIPT_DIR}/Shared/SigningConstants.generated.swift.tmpl" \
    > "${SCRIPT_DIR}/Shared/SigningConstants.generated.swift"

HELPER_INFO_PLIST="${RENDER_DIR}/helper-info.plist"
HELPER_LAUNCHD_PLIST="${RENDER_DIR}/helper-launchd.plist"

sed \
    -e "s|@@VERSION@@|${VERSION}|g" \
    -e "s|@@CLIENT_DESIGNATED_REQUIREMENT@@|$(client_requirement)|g" \
    "${SCRIPT_DIR}/HelperTool/Info.plist.tmpl" > "${HELPER_INFO_PLIST}"

cp "${SCRIPT_DIR}/HelperTool/Launchd.plist.tmpl" "${HELPER_LAUNCHD_PLIST}"

HELPER_LINKER_FLAGS=(
    -Xlinker -sectcreate -Xlinker __TEXT -Xlinker __info_plist
    -Xlinker "${HELPER_INFO_PLIST}"
    -Xlinker -sectcreate -Xlinker __TEXT -Xlinker __launchd_plist
    -Xlinker "${HELPER_LAUNCHD_PLIST}"
)

pushd "${SCRIPT_DIR}" >/dev/null
swift build -c release "${SWIFT_ARCH_ARGS[@]}" \
    --scratch-path "${BUILD_DIR}/swift" \
    --product "${HELPER_LABEL}" \
    "${HELPER_LINKER_FLAGS[@]}"

swift build -c release "${SWIFT_ARCH_ARGS[@]}" \
    --scratch-path "${BUILD_DIR}/swift" \
    --product "${URL_HANDLER_NAME}"

BIN_DIR="$(swift build -c release "${SWIFT_ARCH_ARGS[@]}" \
    --scratch-path "${BUILD_DIR}/swift" --show-bin-path)"
HELPER_BINARY="${BIN_DIR}/${HELPER_LABEL}"
URL_HANDLER_BINARY="${BIN_DIR}/${URL_HANDLER_NAME}"
popd >/dev/null

if [ ! -x "${HELPER_BINARY}" ] || [ ! -x "${URL_HANDLER_BINARY}" ]; then
    echo "error: expected release binaries were not produced" >&2
    exit 1
fi

APP_ROOT="${OUTPUT}/${APP_NAME}"
CONTENTS="${APP_ROOT}/Contents"
MACOS="${CONTENTS}/MacOS"
LAUNCH_SERVICES="${CONTENTS}/Library/LaunchServices"
LAUNCH_DAEMONS="${CONTENTS}/Library/LaunchDaemons"
PLIST="${CONTENTS}/Info.plist"
EMBEDDED_HELPER="${LAUNCH_SERVICES}/${HELPER_LABEL}"
EMBEDDED_LAUNCHD="${LAUNCH_DAEMONS}/${HELPER_LABEL}.plist"

rm -rf "${APP_ROOT}"
mkdir -p "${MACOS}" "${LAUNCH_SERVICES}" "${LAUNCH_DAEMONS}"

install -m 0755 "${URL_HANDLER_BINARY}" "${MACOS}/${URL_HANDLER_NAME}"
install -m 0755 "${HELPER_BINARY}" "${EMBEDDED_HELPER}"
install -m 0644 "${HELPER_LAUNCHD_PLIST}" "${EMBEDDED_LAUNCHD}"

codesign \
    --force \
    --options runtime \
    --timestamp \
    --sign "${CODESIGN_IDENTITY}" \
    "${EMBEDDED_HELPER}"
codesign --verify --verbose=2 "${EMBEDDED_HELPER}"
# codesign prints "Executable=…" on stderr and either
# "# designated => …" (modern) or "designated => …" (older) on stdout.
HELPER_REQUIREMENT="$(
    codesign -d -r- "${EMBEDDED_HELPER}" 2>&1 \
        | sed -n -e 's/^# designated => //p' -e 's/^designated => //p'
)"
if [ -z "${HELPER_REQUIREMENT}" ]; then
    echo "error: cannot extract designated requirement from signed helper" >&2
    codesign -d -r- "${EMBEDDED_HELPER}" 2>&1 >&2 || true
    exit 1
fi
echo "Helper designated requirement: ${HELPER_REQUIREMENT}"

sed \
    -e "s|@@VERSION@@|${VERSION}|g" \
    -e "s|@@HELPER_DESIGNATED_REQUIREMENT@@|${HELPER_REQUIREMENT}|g" \
    "${SCRIPT_DIR}/Info.plist.tmpl" > "${PLIST}"

if ! plutil -lint "${PLIST}" >/dev/null; then
    echo "error: rendered Info.plist failed plutil -lint" >&2
    exit 1
fi
if ! grep -q '<string>probo</string>' "${PLIST}"; then
    echo "error: Info.plist is missing probo URL scheme" >&2
    exit 1
fi

codesign \
    --force \
    --options runtime \
    --timestamp \
    --sign "${CODESIGN_IDENTITY}" \
    "${MACOS}/${URL_HANDLER_NAME}"
codesign \
    --force \
    --options runtime \
    --timestamp \
    --sign "${CODESIGN_IDENTITY}" \
    "${APP_ROOT}"
codesign --verify --verbose=2 "${APP_ROOT}"

echo "Built ${APP_ROOT}"
