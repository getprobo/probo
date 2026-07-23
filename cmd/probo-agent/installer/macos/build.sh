#!/bin/bash
#
# Build a Probo device posture agent macOS installer (.pkg) from a
# pre-built `probo-agent` binary.
#
# Required arguments:
#   --binary  PATH    Path to a compiled probo-agent binary.
#   --arch    ARCH    Target architecture: amd64 or arm64.
#   --version VER     Agent version, e.g. 0.1.0. Defaults to the
#                     content of cmd/probo-agent/VERSION.
#   --output  PATH    Output .pkg path. Defaults to
#                     dist/probo-agent_${VER}_${OS}.pkg.
#
# The resulting flat distribution package is unsigned. Apple
# Developer ID signing + notarization are out of scope for this
# script; consumers can chain `productsign` and `xcrun notarytool`
# afterwards.
#
# Must run on macOS: pkgbuild and productbuild are Apple-only tools.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"

BINARY=""
ARCH=""
VERSION=""
OUTPUT=""
IDENTIFIER="com.getprobo.agent"

usage() {
  sed -ne '/^#/!q; s/^# \{0,1\}//; 2,$ p' <"$0"
}

while [ $# -gt 0 ]; do
  case "$1" in
    --binary)
      BINARY="$2"
      shift 2
      ;;
    --arch)
      ARCH="$2"
      shift 2
      ;;
    --version)
      VERSION="$2"
      shift 2
      ;;
    --output)
      OUTPUT="$2"
      shift 2
      ;;
    --identifier)
      IDENTIFIER="$2"
      shift 2
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      echo "unknown flag: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [ -z "${BINARY}" ] || [ ! -x "${BINARY}" ]; then
  echo "error: --binary <path-to-probo-agent> is required and must be executable" >&2
  exit 2
fi
case "${ARCH}" in
  amd64) PKG_ARCH="x86_64" ;;
  arm64) PKG_ARCH="arm64" ;;
  "")
    echo "error: --arch (amd64|arm64) is required" >&2
    exit 2
    ;;
  *)
    echo "error: unsupported --arch '${ARCH}' (want amd64 or arm64)" >&2
    exit 2
    ;;
esac
if [ -z "${VERSION}" ]; then
  VERSION="$(cat "${REPO_ROOT}/cmd/probo-agent/VERSION")"
fi
if [ -z "${OUTPUT}" ]; then
  mkdir -p "${REPO_ROOT}/dist"
  OUTPUT="${REPO_ROOT}/dist/probo-agent_${VERSION}_darwin_${PKG_ARCH}.pkg"
fi

if ! command -v pkgbuild >/dev/null 2>&1 || ! command -v productbuild >/dev/null 2>&1; then
  echo "error: pkgbuild and productbuild are required (run on macOS)" >&2
  exit 1
fi

STAGE="$(mktemp -d -t probo-agent-pkg)"
trap 'rm -rf "${STAGE}"' EXIT

PAYLOAD="${STAGE}/payload"
SCRIPTS="${STAGE}/scripts"
RESOURCES="${STAGE}/Resources"
mkdir -p "${PAYLOAD}/usr/local/bin" "${SCRIPTS}" "${RESOURCES}"

install -m 0755 "${BINARY}" "${PAYLOAD}/usr/local/bin/probo-agent"

install -m 0755 "${SCRIPT_DIR}/scripts/postinstall" "${SCRIPTS}/postinstall"

cp "${SCRIPT_DIR}/Resources/welcome.html" "${RESOURCES}/welcome.html"
cp "${SCRIPT_DIR}/Resources/conclusion.html" "${RESOURCES}/conclusion.html"
cp "${REPO_ROOT}/LICENSE" "${RESOURCES}/license.txt"

# Component package: payload + scripts only.
COMPONENT_PKG="${STAGE}/probo-agent-component.pkg"
pkgbuild \
  --root "${PAYLOAD}" \
  --scripts "${SCRIPTS}" \
  --identifier "${IDENTIFIER}" \
  --version "${VERSION}" \
  --install-location "/" \
  "${COMPONENT_PKG}"

# Render Distribution.xml from its template.
DISTRIBUTION="${STAGE}/Distribution.xml"
sed \
  -e "s|@@VERSION@@|${VERSION}|g" \
  -e "s|@@PKG_ARCH@@|${PKG_ARCH}|g" \
  -e "s|@@HOST_ARCHS@@|${PKG_ARCH}|g" \
  "${SCRIPT_DIR}/Distribution.xml.tmpl" >"${DISTRIBUTION}"

mkdir -p "$(dirname "${OUTPUT}")"
productbuild \
  --distribution "${DISTRIBUTION}" \
  --package-path "${STAGE}" \
  --resources "${RESOURCES}" \
  "${OUTPUT}"

echo "Built ${OUTPUT}"
