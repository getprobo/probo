#!/bin/bash
#
# Tear down any previous macOS install, then install a signed .pkg.
# Must run as root. Used by: make -C cmd/probo-agent install
#
# Usage: reinstall.sh /path/to/probo-agent_*.pkg

set -eu

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PKG="${1:-}"

if [ "$(id -u)" -ne 0 ]; then
    echo "error: must run as root (try: sudo make -C cmd/probo-agent install)" >&2
    exit 1
fi

if [ -z "${PKG}" ] || [ ! -f "${PKG}" ]; then
    echo "error: usage: $0 /path/to/probo-agent_*.pkg" >&2
    exit 2
fi

"${SCRIPT_DIR}/uninstall.sh"
echo "Installing ${PKG}…"
installer -pkg "${PKG}" -target /
echo "PKG install complete."
