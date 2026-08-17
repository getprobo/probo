#!/bin/bash
#
# Fully remove a macOS PKG install of probo-agent.
#
# Idempotent: missing pieces are skipped. Must run as root.
#
# Removes:
#   - agent service, tray LaunchAgent, privileged helper
#   - binary, Probo Agent.app (and .localized variants)
#   - state/run dirs, logs
#   - PKG receipt (com.probo.agent)
#   - stale Launch Services registration for the URL handler

set -u

BINARY="/usr/local/bin/probo-agent"
STATE_DIR="/var/lib/probo-agent"
RUN_DIR="/var/run/probo-agent"
DAEMON_PLIST="/Library/LaunchDaemons/com.probo.agent.plist"
HELPER_LABEL="com.probo.agent.helper"
HELPER_PLIST="/Library/LaunchDaemons/${HELPER_LABEL}.plist"
HELPER_BINARY="/Library/PrivilegedHelperTools/${HELPER_LABEL}"
PKG_ID="com.probo.agent"
LSREGISTER="/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"

log() {
  printf '%s\n' "$*"
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

require_root() {
  if [ "$(id -u)" -ne 0 ]; then
    die "must run as root (try: sudo make -C cmd/probo-agent uninstall)"
  fi
}

bootout_system_plist() {
  local plist="$1"
  if [ -f "${plist}" ]; then
    launchctl bootout system "${plist}" 2>/dev/null || true
    log "Booted out ${plist}"
  fi
}

unregister_apps() {
  local path
  for path in \
    "/Applications/Probo Agent.app" \
    "/Applications/Probo Agent.localized/Probo Agent.app"; do
    if [ -d "${path}" ] && [ -x "${LSREGISTER}" ]; then
      "${LSREGISTER}" -u "${path}" 2>/dev/null || true
      log "Unregistered Launch Services entry for ${path}"
    fi
  done
}

kill_leftovers() {
  # Best-effort; deleted-but-running binaries otherwise keep claiming probo://.
  pkill -x probo-agent-url-handler 2>/dev/null || true
  pkill -f '/usr/local/bin/probo-agent tray' 2>/dev/null || true
  pkill -f '/Library/PrivilegedHelperTools/com.probo.agent.helper' 2>/dev/null || true
  # Agent daemon may still be running after plist bootout races.
  pkill -x probo-agent 2>/dev/null || true
}

require_root

if [ "$(uname -s)" != "Darwin" ]; then
  die "this uninstall script is macOS-only"
fi

log "=== probo-agent macOS uninstall $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

# Prefer the agent's own uninstall for service/tray/state when present.
if [ -x "${BINARY}" ]; then
  if "${BINARY}" uninstall; then
    log "Ran: ${BINARY} uninstall"
  else
    log "warning: ${BINARY} uninstall failed; continuing with manual cleanup"
  fi
else
  log "Binary not found at ${BINARY}; skipping probo-agent uninstall"
fi

bootout_system_plist "${DAEMON_PLIST}"
bootout_system_plist "${HELPER_PLIST}"

kill_leftovers

rm -f "${DAEMON_PLIST}" "${HELPER_PLIST}" "${HELPER_BINARY}"
log "Removed LaunchDaemon / helper files (if present)"

unregister_apps
rm -rf \
  "/Applications/Probo Agent.app" \
  "/Applications/Probo Agent.localized"
log "Removed Probo Agent.app (if present)"

rm -f "${BINARY}"
rm -rf "${STATE_DIR}" "${RUN_DIR}"
rm -f \
  /var/log/probo-agent.log \
  /var/log/probo-agent-install.log \
  /tmp/probo-agent.conf
log "Removed binary, state, run dir, logs, and staged conf (if present)"

if pkgutil --pkg-info "${PKG_ID}" >/dev/null 2>&1; then
  if ! pkgutil --forget "${PKG_ID}" >/dev/null; then
    die "failed to forget PKG receipt ${PKG_ID}"
  fi
  log "Forgot PKG receipt ${PKG_ID}"
fi

log "=== uninstall done ==="
