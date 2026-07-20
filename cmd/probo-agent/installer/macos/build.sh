#!/bin/bash
#
# Build a Probo device posture agent macOS installer (.pkg) from a
# pre-built `probo-agent` binary.
#
# Required arguments:
#   --binary  PATH    Path to a compiled probo-agent binary.
#   --arch    ARCH    Target architecture: amd64, arm64, or universal.
#   --version VER     Agent version, e.g. 0.1.0. Defaults to the
#                     content of cmd/probo-agent/VERSION.
#   --output  PATH    Output .pkg path. Defaults to
#                     dist/probo-agent_${VER}_darwin_${ARCH}.pkg.
#
# Required environment variables:
#   CODESIGN_IDENTITY    Developer ID Application identity. Signs the
#                        agent binary, Probo Agent.app, and embedded helper.
#   APPLE_TEAM_ID        Apple Developer Team ID (helper client requirement).
#
# Optional (auditor-mode compatible):
#   INSTALLER_IDENTITY            Developer ID Installer identity. When
#                                 set, passes --sign to productbuild.
#   APPLE_ID                      Apple ID for notarytool store-credentials.
#   APPLE_ID_PASSWORD             App-specific password; used only to
#                                 populate a keychain profile (not passed
#                                 to long-lived notarytool submit).
#   NOTARYTOOL_KEYCHAIN_PROFILE   Keychain profile name for store/submit.
#                                 Defaults to probo-agent-notary.
#
# Notarization is enabled when APPLE_ID and APPLE_ID_PASSWORD are both
# set. INSTALLER_IDENTITY is then required. The script stores credentials
# into the keychain profile, then notarizes and staples the .app before
# packaging and the signed .pkg via --keychain-profile so the password
# is not on submit argv for the long --wait.
#
# Must run on macOS: pkgbuild, productbuild, and swift build are
# Apple-only tools. The build also compiles Probo Agent.app (the
# probo:// URL handler) from enroll-ui/.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"

BINARY=""
ARCH=""
VERSION=""
OUTPUT=""
IDENTIFIER="com.probo.agent"
CODESIGN_IDENTITY="${CODESIGN_IDENTITY:-}"
INSTALLER_IDENTITY="${INSTALLER_IDENTITY:-}"
APPLE_ID="${APPLE_ID:-}"
APPLE_ID_PASSWORD="${APPLE_ID_PASSWORD:-}"
APPLE_TEAM_ID="${APPLE_TEAM_ID:-}"
NOTARYTOOL_KEYCHAIN_PROFILE="${NOTARYTOOL_KEYCHAIN_PROFILE:-probo-agent-notary}"

usage() {
    sed -ne '/^#/!q; s/^# \{0,1\}//; 2,$ p' < "$0"
}

while [ $# -gt 0 ]; do
    case "$1" in
        --binary)     BINARY="$2";     shift 2 ;;
        --arch)       ARCH="$2";       shift 2 ;;
        --version)    VERSION="$2";    shift 2 ;;
        --output)     OUTPUT="$2";     shift 2 ;;
        --identifier) IDENTIFIER="$2"; shift 2 ;;
        -h|--help)    usage; exit 0 ;;
        *)            echo "unknown flag: $1" >&2; usage >&2; exit 2 ;;
    esac
done

if [ -z "${BINARY}" ] || [ ! -x "${BINARY}" ]; then
    echo "error: --binary <path-to-probo-agent> is required and must be executable" >&2
    exit 2
fi
case "${ARCH}" in
    amd64)
        PKG_ARCH="x86_64"
        HOST_ARCHS="x86_64"
        OUTPUT_ARCH="x86_64"
        ;;
    arm64)
        PKG_ARCH="arm64"
        HOST_ARCHS="arm64"
        OUTPUT_ARCH="arm64"
        ;;
    universal)
        PKG_ARCH="arm64"
        HOST_ARCHS="arm64,x86_64"
        OUTPUT_ARCH="universal"
        ;;
    "")
        echo "error: --arch (amd64|arm64|universal) is required" >&2
        exit 2
        ;;
    *)
        echo "error: unsupported --arch '${ARCH}' (want amd64, arm64, or universal)" >&2
        exit 2
        ;;
esac

# --arch sets hostArchitectures in Distribution.xml. Refuse a binary that
# lacks the advertised slice(s) so Installer cannot install on a CPU the
# agent cannot run on.
if ! command -v lipo >/dev/null 2>&1; then
    echo "error: lipo is required to validate --binary architecture (run on macOS)" >&2
    exit 1
fi
BINARY_ARCHS="$(lipo -archs "${BINARY}")"
has_arm64=false
has_x86_64=false
for arch_slice in ${BINARY_ARCHS}; do
    case "${arch_slice}" in
        arm64)  has_arm64=true ;;
        x86_64) has_x86_64=true ;;
    esac
done
case "${ARCH}" in
    amd64)
        if [ "${has_x86_64}" != true ]; then
            echo "error: --arch amd64 requires a binary with an x86_64 slice (got: ${BINARY_ARCHS})" >&2
            exit 2
        fi
        ;;
    arm64)
        if [ "${has_arm64}" != true ]; then
            echo "error: --arch arm64 requires a binary with an arm64 slice (got: ${BINARY_ARCHS})" >&2
            exit 2
        fi
        ;;
    universal)
        if [ "${has_arm64}" != true ] || [ "${has_x86_64}" != true ]; then
            echo "error: --arch universal requires a fat binary with arm64 and x86_64 slices (got: ${BINARY_ARCHS}); use lipo -create" >&2
            exit 2
        fi
        ;;
esac

if [ -z "${VERSION}" ]; then
    VERSION="$(cat "${REPO_ROOT}/cmd/probo-agent/VERSION")"
fi
if [ -z "${OUTPUT}" ]; then
    mkdir -p "${REPO_ROOT}/dist"
    OUTPUT="${REPO_ROOT}/dist/probo-agent_${VERSION}_darwin_${OUTPUT_ARCH}.pkg"
fi

if ! command -v pkgbuild >/dev/null 2>&1 || ! command -v productbuild >/dev/null 2>&1; then
    echo "error: pkgbuild and productbuild are required (run on macOS)" >&2
    exit 1
fi
if ! command -v swift >/dev/null 2>&1; then
    echo "error: swift is required to build Probo Agent.app (run on macOS)" >&2
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

notarize_enabled=false
if [ -n "${APPLE_ID}" ] && [ -n "${APPLE_ID_PASSWORD}" ]; then
    notarize_enabled=true
fi
if [ "${notarize_enabled}" = true ] && [ -z "${INSTALLER_IDENTITY}" ]; then
    echo "error: notarization requires INSTALLER_IDENTITY" >&2
    exit 2
fi

sign_macho() {
    local path="$1"
    codesign \
        --force \
        --options runtime \
        --timestamp \
        --sign "${CODESIGN_IDENTITY}" \
        "${path}"
    codesign --verify --verbose=2 "${path}"
}

sign_app_bundle() {
    local app_path="$1"
    local helper_path="${app_path}/Contents/Library/LaunchServices/com.probo.agent.helper"
    if [ -x "${helper_path}" ]; then
        codesign \
            --force \
            --options runtime \
            --timestamp \
            --sign "${CODESIGN_IDENTITY}" \
            "${helper_path}"
        codesign --verify --verbose=2 "${helper_path}"
    fi
    codesign \
        --force \
        --options runtime \
        --timestamp \
        --sign "${CODESIGN_IDENTITY}" \
        "${app_path}/Contents/MacOS/probo-agent-url-handler"
    codesign \
        --force \
        --options runtime \
        --timestamp \
        --sign "${CODESIGN_IDENTITY}" \
        "${app_path}"
    codesign --verify --verbose=2 "${app_path}"
}

ensure_notarytool_credentials() {
    if [ -z "${APPLE_ID}" ] || [ -z "${APPLE_ID_PASSWORD}" ]; then
        echo "error: APPLE_ID and APPLE_ID_PASSWORD are required to store notarytool credentials" >&2
        exit 2
    fi
    # Password appears on argv only for this short-lived store. Submits
    # use --keychain-profile so concurrent processes cannot read it.
    xcrun notarytool store-credentials "${NOTARYTOOL_KEYCHAIN_PROFILE}" \
        --apple-id "${APPLE_ID}" \
        --password "${APPLE_ID_PASSWORD}" \
        --team-id "${APPLE_TEAM_ID}"
}

notarytool_submit() {
    local path="$1"
    xcrun notarytool submit "${path}" \
        --keychain-profile "${NOTARYTOOL_KEYCHAIN_PROFILE}" \
        --wait
}

notarize_and_staple_app() {
    local app_path="$1"
    local zip_path

    zip_path="${STAGE}/probo-agent-app.zip"
    ditto -c -k --keepParent "${app_path}" "${zip_path}"
    notarytool_submit "${zip_path}"
    rm -f "${zip_path}"
    xcrun stapler staple "${app_path}"
}

notarize_and_staple_pkg() {
    local pkg_path="$1"
    notarytool_submit "${pkg_path}"
    xcrun stapler staple "${pkg_path}"
}

STAGE="$(mktemp -d -t probo-agent-pkg)"
trap 'rm -rf "${STAGE}"' EXIT

PAYLOAD="${STAGE}/payload"
SCRIPTS="${STAGE}/scripts"
RESOURCES="${STAGE}/Resources"
mkdir -p "${PAYLOAD}/usr/local/bin" "${SCRIPTS}" "${RESOURCES}"

install -m 0755 "${BINARY}" "${PAYLOAD}/usr/local/bin/probo-agent"
sign_macho "${PAYLOAD}/usr/local/bin/probo-agent"

mkdir -p "${PAYLOAD}/Applications"
"${SCRIPT_DIR}/enroll-ui/build-app.sh" \
    --arch "${ARCH}" \
    --version "${VERSION}" \
    --output "${PAYLOAD}/Applications"

APP_PATH="${PAYLOAD}/Applications/Probo Agent.app"
sign_app_bundle "${APP_PATH}"

if [ "${notarize_enabled}" = true ]; then
    ensure_notarytool_credentials
    echo "Notarizing Probo Agent.app before packaging..."
    notarize_and_staple_app "${APP_PATH}"
fi

# Avoid AppleDouble (._*) and resource-fork noise in the package.
export COPYFILE_DISABLE=1

# ditto --norsrc/--noextattr copies without resource forks / xattrs.
ditto --norsrc --noextattr "${SCRIPT_DIR}/scripts/preinstall" "${SCRIPTS}/preinstall"
ditto --norsrc --noextattr "${SCRIPT_DIR}/scripts/postinstall" "${SCRIPTS}/postinstall"
ditto --norsrc --noextattr \
    "${REPO_ROOT}/pkg/deviceagent/tray/launchagent.plist.tmpl" \
    "${SCRIPTS}/launchagent.plist.tmpl"
chmod 0755 "${SCRIPTS}/preinstall" "${SCRIPTS}/postinstall"
chmod 0644 "${SCRIPTS}/launchagent.plist.tmpl"

ditto --norsrc --noextattr "${SCRIPT_DIR}/Resources/welcome.html"    "${RESOURCES}/welcome.html"
ditto --norsrc --noextattr "${SCRIPT_DIR}/Resources/conclusion.html" "${RESOURCES}/conclusion.html"
ditto --norsrc --noextattr "${REPO_ROOT}/LICENSE"                    "${RESOURCES}/license.txt"

# Strip any xattrs that tools may have reattached (codesign, etc.).
xattr -cr "${PAYLOAD}" "${SCRIPTS}" "${RESOURCES}" 2>/dev/null || true
find "${PAYLOAD}" "${SCRIPTS}" "${RESOURCES}" -name '._*' -delete 2>/dev/null || true

# Component package: payload + scripts only.
COMPONENT_PKG="${STAGE}/probo-agent-component.pkg"
pkgbuild \
    --root "${PAYLOAD}" \
    --scripts "${SCRIPTS}" \
    --identifier "${IDENTIFIER}" \
    --version "${VERSION}" \
    --install-location "/" \
    "${COMPONENT_PKG}"

# pkgbuild records protected com.apple.provenance xattrs as empty
# AppleDouble (._*) Bom entries. Rewrite the Bom with mkbom so the
# installer does not lay down those stubs next to real files.
rewrite_component_bom() {
    local pkg="$1"
    local expand_dir root_dir flat_pkg

    expand_dir="${STAGE}/component-expand"
    root_dir="${STAGE}/component-root"
    flat_pkg="${STAGE}/probo-agent-component-clean.pkg"
    rm -rf "${expand_dir}" "${root_dir}" "${flat_pkg}"
    # pkgutil --expand creates the destination directory itself.
    pkgutil --expand "${pkg}" "${expand_dir}"
    find "${expand_dir}/Scripts" -name '._*' -delete 2>/dev/null || true

    mkdir -p "${root_dir}"
    (
        cd "${root_dir}"
        gzip -dc "${expand_dir}/Payload" | cpio -idmu 2>/dev/null
    )
    find "${root_dir}" -name '._*' -delete 2>/dev/null || true
    mkbom "${root_dir}" "${expand_dir}/Bom"
    if lsbom "${expand_dir}/Bom" | grep -q '/\._'; then
        echo "error: rewritten Bom still contains AppleDouble entries" >&2
        return 1
    fi
    pkgutil --flatten "${expand_dir}" "${flat_pkg}"
    mv "${flat_pkg}" "${pkg}"
}

rewrite_component_bom "${COMPONENT_PKG}"

# Render Distribution.xml from its template.
DISTRIBUTION="${STAGE}/Distribution.xml"
sed \
    -e "s|@@VERSION@@|${VERSION}|g" \
    -e "s|@@PKG_ARCH@@|${PKG_ARCH}|g" \
    -e "s|@@HOST_ARCHS@@|${HOST_ARCHS}|g" \
    -e "s|@@IDENTIFIER@@|${IDENTIFIER}|g" \
    "${SCRIPT_DIR}/Distribution.xml.tmpl" > "${DISTRIBUTION}"

mkdir -p "$(dirname "${OUTPUT}")"

PRODUCTBUILD_ARGS=(
    --distribution "${DISTRIBUTION}"
    --package-path "${STAGE}"
    --resources "${RESOURCES}"
)
if [ -n "${INSTALLER_IDENTITY}" ]; then
    PRODUCTBUILD_ARGS+=(--sign "${INSTALLER_IDENTITY}")
fi
PRODUCTBUILD_ARGS+=("${OUTPUT}")

productbuild "${PRODUCTBUILD_ARGS[@]}"

if [ "${notarize_enabled}" = true ]; then
    echo "Notarizing ${OUTPUT}..."
    notarize_and_staple_pkg "${OUTPUT}"
fi

echo "Built ${OUTPUT}"
