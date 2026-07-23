#!/bin/bash
#
# Build a Probo device posture agent macOS installer (.pkg) from a
# pre-built fat `probo-agent` binary (arm64 + x86_64).
#
# Required arguments:
#   --binary  PATH    Path to a compiled probo-agent fat binary.
#   --version VER     Agent version, e.g. 0.1.0. Defaults to the
#                     content of cmd/probo-agent/VERSION.
#   --output  PATH    Output .pkg path. Defaults to
#                     dist/probo-agent_${VER}_darwin.pkg.
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
# probo:// URL handler + privileged helper) from enroll-ui/.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"
ENROLL_UI_DIR="${SCRIPT_DIR}/enroll-ui"

BINARY=""
VERSION=""
OUTPUT=""
IDENTIFIER="com.probo.agent"
APP_NAME="Probo Agent.app"
URL_HANDLER_NAME="probo-agent-url-handler"
HELPER_LABEL="com.probo.agent.helper"
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

# Distribution.xml advertises hostArchitectures=arm64,x86_64. Refuse a
# binary that lacks either slice so Installer cannot install on a CPU
# the agent cannot run on.
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
if [ "${has_arm64}" != true ] || [ "${has_x86_64}" != true ]; then
    echo "error: --binary must be a fat binary with arm64 and x86_64 slices (got: ${BINARY_ARCHS}); use lipo -create" >&2
    exit 2
fi

if [ -z "${VERSION}" ]; then
    VERSION="$(cat "${REPO_ROOT}/cmd/probo-agent/VERSION")"
fi
if [ -z "${OUTPUT}" ]; then
    mkdir -p "${REPO_ROOT}/dist"
    OUTPUT="${REPO_ROOT}/dist/probo-agent_${VERSION}_darwin.pkg"
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

codesign_runtime() {
    local path="$1"
    codesign \
        --force \
        --options runtime \
        --timestamp \
        --sign "${CODESIGN_IDENTITY}" \
        "${path}"
    codesign --verify --verbose=2 "${path}"
}

client_requirement() {
    printf 'anchor apple generic and identifier "com.probo.agent.url-handler" and certificate leaf[subject.OU] = "%s"' "${APPLE_TEAM_ID}"
}

team_id_option() {
    printf '"%s"' "${APPLE_TEAM_ID}"
}

# Build Probo Agent.app (URL handler + embedded privileged helper) into
# parent_dir. Signs nested Mach-Os then the .app bundle (bottom-up).
build_probo_agent_app() {
    local parent_dir="$1"
    local build_dir render_dir
    local helper_info_plist helper_launchd_plist
    local helper_binary url_handler_binary bin_dir
    local app_root contents macos launch_services launch_daemons
    local plist embedded_helper embedded_launchd
    local helper_requirement
    local -a helper_linker_flags swift_arch_args

    build_dir="${STAGE}/enroll-ui-build"
    render_dir="${build_dir}/rendered"
    mkdir -p "${render_dir}"

    swift_arch_args=(--arch arm64 --arch x86_64)

    sed \
        -e "s|@@VERSION@@|${VERSION}|g" \
        "${ENROLL_UI_DIR}/Shared/HelperVersion.generated.swift.tmpl" \
        > "${ENROLL_UI_DIR}/Shared/HelperVersion.generated.swift"

    sed \
        -e "s|@@TEAM_ID_OPTION@@|$(team_id_option)|g" \
        "${ENROLL_UI_DIR}/Shared/SigningConstants.generated.swift.tmpl" \
        > "${ENROLL_UI_DIR}/Shared/SigningConstants.generated.swift"

    helper_info_plist="${render_dir}/helper-info.plist"
    helper_launchd_plist="${render_dir}/helper-launchd.plist"

    sed \
        -e "s|@@VERSION@@|${VERSION}|g" \
        -e "s|@@CLIENT_DESIGNATED_REQUIREMENT@@|$(client_requirement)|g" \
        "${ENROLL_UI_DIR}/HelperTool/Info.plist.tmpl" > "${helper_info_plist}"

    cp "${ENROLL_UI_DIR}/HelperTool/Launchd.plist.tmpl" "${helper_launchd_plist}"

    helper_linker_flags=(
        -Xlinker -sectcreate -Xlinker __TEXT -Xlinker __info_plist
        -Xlinker "${helper_info_plist}"
        -Xlinker -sectcreate -Xlinker __TEXT -Xlinker __launchd_plist
        -Xlinker "${helper_launchd_plist}"
    )

    pushd "${ENROLL_UI_DIR}" >/dev/null
    swift build -c release "${swift_arch_args[@]}" \
        --scratch-path "${build_dir}/swift" \
        --product "${HELPER_LABEL}" \
        "${helper_linker_flags[@]}"

    swift build -c release "${swift_arch_args[@]}" \
        --scratch-path "${build_dir}/swift" \
        --product "${URL_HANDLER_NAME}"

    bin_dir="$(swift build -c release "${swift_arch_args[@]}" \
        --scratch-path "${build_dir}/swift" --show-bin-path)"
    helper_binary="${bin_dir}/${HELPER_LABEL}"
    url_handler_binary="${bin_dir}/${URL_HANDLER_NAME}"
    popd >/dev/null

    if [ ! -x "${helper_binary}" ] || [ ! -x "${url_handler_binary}" ]; then
        echo "error: expected release binaries were not produced" >&2
        exit 1
    fi

    app_root="${parent_dir}/${APP_NAME}"
    contents="${app_root}/Contents"
    macos="${contents}/MacOS"
    launch_services="${contents}/Library/LaunchServices"
    launch_daemons="${contents}/Library/LaunchDaemons"
    plist="${contents}/Info.plist"
    embedded_helper="${launch_services}/${HELPER_LABEL}"
    embedded_launchd="${launch_daemons}/${HELPER_LABEL}.plist"

    rm -rf "${app_root}"
    mkdir -p "${macos}" "${launch_services}" "${launch_daemons}"

    install -m 0755 "${url_handler_binary}" "${macos}/${URL_HANDLER_NAME}"
    install -m 0755 "${helper_binary}" "${embedded_helper}"
    install -m 0644 "${helper_launchd_plist}" "${embedded_launchd}"

    # Sign helper before writing Info.plist so SMPrivilegedExecutables
    # can embed the helper's designated requirement.
    codesign_runtime "${embedded_helper}"
    # codesign prints "Executable=…" on stderr and either
    # "# designated => …" (modern) or "designated => …" (older) on stdout.
    helper_requirement="$(
        codesign -d -r- "${embedded_helper}" 2>&1 \
            | sed -n -e 's/^# designated => //p' -e 's/^designated => //p'
    )"
    if [ -z "${helper_requirement}" ]; then
        echo "error: cannot extract designated requirement from signed helper" >&2
        codesign -d -r- "${embedded_helper}" 2>&1 >&2 || true
        exit 1
    fi
    echo "Helper designated requirement: ${helper_requirement}"

    sed \
        -e "s|@@VERSION@@|${VERSION}|g" \
        -e "s|@@HELPER_DESIGNATED_REQUIREMENT@@|${helper_requirement}|g" \
        "${ENROLL_UI_DIR}/Info.plist.tmpl" > "${plist}"

    if ! plutil -lint "${plist}" >/dev/null; then
        echo "error: rendered Info.plist failed plutil -lint" >&2
        exit 1
    fi
    if ! grep -q '<string>probo</string>' "${plist}"; then
        echo "error: Info.plist is missing probo URL scheme" >&2
        exit 1
    fi

    codesign_runtime "${macos}/${URL_HANDLER_NAME}"
    codesign_runtime "${app_root}"

    echo "Built ${app_root}"
}

notarytool_submit() {
    local path="$1"
    xcrun notarytool submit "${path}" \
        --keychain-profile "${NOTARYTOOL_KEYCHAIN_PROFILE}" \
        --wait
}

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

STAGE="$(mktemp -d -t probo-agent-pkg)"
trap 'rm -rf "${STAGE}"' EXIT

PAYLOAD="${STAGE}/payload"
SCRIPTS="${STAGE}/scripts"
RESOURCES="${STAGE}/Resources"
mkdir -p "${PAYLOAD}/usr/local/bin" "${SCRIPTS}" "${RESOURCES}"

install -m 0755 "${BINARY}" "${PAYLOAD}/usr/local/bin/probo-agent"
codesign_runtime "${PAYLOAD}/usr/local/bin/probo-agent"

mkdir -p "${PAYLOAD}/Applications"
build_probo_agent_app "${PAYLOAD}/Applications"
APP_PATH="${PAYLOAD}/Applications/${APP_NAME}"

if [ "${notarize_enabled}" = true ]; then
    # Password appears on argv only for this short-lived store. Submits
    # use --keychain-profile so concurrent processes cannot read it.
    xcrun notarytool store-credentials "${NOTARYTOOL_KEYCHAIN_PROFILE}" \
        --apple-id "${APPLE_ID}" \
        --password "${APPLE_ID_PASSWORD}" \
        --team-id "${APPLE_TEAM_ID}"
    echo "Notarizing Probo Agent.app before packaging..."
    zip_path="${STAGE}/probo-agent-app.zip"
    ditto -c -k --keepParent "${APP_PATH}" "${zip_path}"
    notarytool_submit "${zip_path}"
    rm -f "${zip_path}"
    xcrun stapler staple "${APP_PATH}"
fi

# AppleDouble / xattr hygiene: COPYFILE_DISABLE + ditto --norsrc/--noextattr
# avoid forks on copy; xattr -cr / find '._*' strip anything codesign
# reattached; rewrite_component_bom drops provenance stubs from the Bom.
export COPYFILE_DISABLE=1

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

xattr -cr "${PAYLOAD}" "${SCRIPTS}" "${RESOURCES}" 2>/dev/null || true
find "${PAYLOAD}" "${SCRIPTS}" "${RESOURCES}" -name '._*' -delete 2>/dev/null || true

COMPONENT_PKG="${STAGE}/probo-agent-component.pkg"
pkgbuild \
    --root "${PAYLOAD}" \
    --scripts "${SCRIPTS}" \
    --identifier "${IDENTIFIER}" \
    --version "${VERSION}" \
    --install-location "/" \
    "${COMPONENT_PKG}"

rewrite_component_bom "${COMPONENT_PKG}"

DISTRIBUTION="${STAGE}/Distribution.xml"
sed \
    -e "s|@@VERSION@@|${VERSION}|g" \
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
    notarytool_submit "${OUTPUT}"
    xcrun stapler staple "${OUTPUT}"
fi

echo "Built ${OUTPUT}"
