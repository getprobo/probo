#!/bin/sh
#
# probo-agent installer for Darwin, Linux, and FreeBSD.
#
# Downloads the matching GitHub Release binary, verifies its sha256
# checksum (embedded in this script at release time), installs to
# /usr/local/bin, then enrolls the device.
#
# Usage:
#
#   # Interactive — curl install.sh from the target probo-agent/v* release
#   curl -fsSL "https://github.com/getprobo/probo/releases/download/probo-agent/vX.Y.Z/install.sh" | sudo sh
#
#   # Unattended / MDM
#   curl -fsSL "…/install.sh" | sudo \
#     PROBO_SERVER_URL=https://us.probo.com \
#     PROBO_ENROLLMENT_TOKEN='…' sh
#
#   # Mirror the release assets (must match the embedded release tag)
#   PROBO_AGENT_RELEASE_BASE="https://release-base/probo-agent/vX.Y.Z" \
#     curl -fsSL "…/install.sh" | sudo sh
#
#   # Explicit flags
#   curl -fsSL "…/install.sh" | sudo sh -s -- \
#     --server https://us.probo.com \
#     --enrollment-token '…'
#
# Environment variables:
#   PROBO_AGENT_RELEASE_BASE       Release download base URL (default: embedded tag)
#   PROBO_AGENT_RELEASE_TAG        Override embedded release tag (local dev)
#   PROBO_AGENT_SKIP_CHECKSUM_VERIFY  Set to true to skip SHA-256 verification (local dev)
#   PROBO_AGENT_STATE_DIR            Agent state directory passed as --dir (default: /var/lib/probo-agent)
#   PROBO_SERVER_URL               Probo server base URL
#   PROBO_ENROLLMENT_TOKEN         One-shot enrollment token
#   PROBO_NO_AUTO_UPDATE           Set to true to pass --no-auto-update
#
# Never pass the enrollment token in the curl URL.

set -eu
# pipefail is a bash/ksh extension; enable when available.
# shellcheck disable=SC3040
(set -o pipefail 2>/dev/null) && set -o pipefail

BINARY_PATH="/usr/local/bin/probo-agent"
GITHUB_RELEASES_URL="https://github.com/getprobo/probo/releases/download"

# Injected at release time by .github/workflows/release-probo-agent.yaml
RELEASE_TAG="__PROBO_AGENT_RELEASE_TAG__"
if [ -n "${PROBO_AGENT_RELEASE_TAG:-}" ]; then
	RELEASE_TAG="$PROBO_AGENT_RELEASE_TAG"
fi

RELEASE_BASE="${PROBO_AGENT_RELEASE_BASE:-}"
SERVER_URL="${PROBO_SERVER_URL:-}"
ENROLLMENT_TOKEN="${PROBO_ENROLLMENT_TOKEN:-}"
STATE_DIR="${PROBO_AGENT_STATE_DIR:-}"
NO_AUTO_UPDATE="${PROBO_NO_AUTO_UPDATE:-}"
SKIP_SERVICE=false

die() {
	echo "error: $*" >&2
	exit 1
}

can_prompt() {
	[ -t 0 ] && return 0
	[ -r /dev/tty ] && [ -w /dev/tty ]
}

read_user() {
	if [ -t 0 ]; then
		IFS= read -r "$1"
	else
		IFS= read -r "$1" < /dev/tty
	fi
}

usage() {
	cat <<'EOF'
probo-agent installer for Darwin, Linux, and FreeBSD.

Usage:
  curl -fsSL "…/install.sh" | sudo sh
  curl -fsSL "…/install.sh" | sudo sh -s -- --server URL --enrollment-token TOKEN

Environment variables:
  PROBO_AGENT_RELEASE_BASE          Release download base URL (default: embedded tag)
  PROBO_AGENT_RELEASE_TAG             Override embedded release tag (local dev)
  PROBO_AGENT_SKIP_CHECKSUM_VERIFY    Set to true to skip SHA-256 verification (local dev)
  PROBO_AGENT_STATE_DIR               Agent state directory (--dir; default /var/lib/probo-agent)
  PROBO_SERVER_URL                    Probo server base URL
  PROBO_ENROLLMENT_TOKEN              One-shot enrollment token
  PROBO_NO_AUTO_UPDATE                Set to true to disable auto-update
EOF
}

embedded_checksums() {
	cat <<'EOF'
# __PROBO_AGENT_CHECKSUMS_BEGIN__
# __PROBO_AGENT_CHECKSUMS_END__
EOF
}

resolve_embedded_release() {
	if [ "$RELEASE_TAG" = "__PROBO_AGENT_RELEASE_TAG__" ]; then
		die "this install.sh was not published by a probo-agent release; curl install.sh from the target release"
	fi

	if [ -n "$RELEASE_BASE" ]; then
		RELEASE_BASE="${RELEASE_BASE%/}"
		case "$RELEASE_BASE" in
		*/"$RELEASE_TAG") ;;
		*) die "PROBO_AGENT_RELEASE_BASE must end with release tag ${RELEASE_TAG}" ;;
		esac
		return 0
	fi

	RELEASE_BASE="${GITHUB_RELEASES_URL}/${RELEASE_TAG}"
	printf 'Using release %s\n' "$RELEASE_TAG"
}

require_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		die "required command not found: $1"
	fi
}

detect_platform() {
	os="$(uname -s)"
	arch="$(uname -m)"

	case "$os" in
	Darwin) os_label="Darwin" ;;
	Linux) os_label="Linux" ;;
	FreeBSD) os_label="Freebsd" ;;
	*) die "unsupported operating system: $os (Darwin, Linux, and FreeBSD only)" ;;
	esac

	case "$arch" in
	x86_64 | amd64) arch_label="x86_64" ;;
	arm64 | aarch64) arch_label="arm64" ;;
	*) die "unsupported CPU architecture: $arch" ;;
	esac

	archive_dir="probo-agent_${os_label}_${arch_label}"
	archive_name="${archive_dir}.tar.gz"
}

sha256_file() {
	file="$1"
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$file" | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$file" | awk '{print $1}'
	elif command -v sha256 >/dev/null 2>&1; then
		sha256 -q "$file"
	else
		die "no sha256 tool found (need sha256sum, shasum, or sha256)"
	fi
}

verify_embedded_checksum() {
	case "${PROBO_AGENT_SKIP_CHECKSUM_VERIFY:-}" in
	1 | true | TRUE | yes | YES) return 0 ;;
	esac

	archive_file="$1"
	archive_basename="$(basename "$archive_file")"

	expected="$(
		embedded_checksums | awk -v name="$archive_basename" '
			$0 ~ /^#/ { next }
			$2 == name { print $1; exit }
		'
	)"
	if [ -z "$expected" ]; then
		die "archive ${archive_basename} not found in embedded release checksums"
	fi

	actual="$(sha256_file "$archive_file")"
	if [ "$expected" != "$actual" ]; then
		die "checksum mismatch for ${archive_file}"
	fi
}

read_secret() {
	prompt_text="$1"
	printf '%s' "$prompt_text"
	if command -v stty >/dev/null 2>&1; then
		if [ -t 0 ]; then
			old_stty="$(stty -g 2>/dev/null || true)"
			stty -echo 2>/dev/null || true
			IFS= read -r REPLY || REPLY=""
			printf '\n'
			if [ -n "${old_stty:-}" ]; then
				stty "$old_stty" 2>/dev/null || stty echo 2>/dev/null || true
			fi
		else
			old_stty="$(stty -g < /dev/tty 2>/dev/null || true)"
			stty -echo < /dev/tty 2>/dev/null || true
			IFS= read -r REPLY < /dev/tty || REPLY=""
			printf '\n' >/dev/tty
			if [ -n "${old_stty:-}" ]; then
				stty "$old_stty" < /dev/tty 2>/dev/null || stty echo < /dev/tty 2>/dev/null || true
			fi
		fi
	else
		read_user REPLY
	fi
	ENROLLMENT_TOKEN="$REPLY"
}

prompt_server_url() {
	if [ -n "$SERVER_URL" ]; then
		return 0
	fi
	if ! can_prompt; then
		die "PROBO_SERVER_URL is required in non-interactive mode"
	fi

	printf '\nProbo server URL:\n'
	printf '  1) https://us.probo.com (United States)\n'
	printf '  2) https://eu.probo.com (European Union)\n'
	printf '  3) Enter a custom URL\n'
	printf 'Choice [1]: '
	read_user choice

	case "${choice:-1}" in
	1 | "") SERVER_URL="https://us.probo.com" ;;
	2) SERVER_URL="https://eu.probo.com" ;;
	3)
		printf 'Server URL: '
		read_user SERVER_URL
		;;
	*) SERVER_URL="$choice" ;;
	esac

	if [ -z "$SERVER_URL" ]; then
		die "server URL is required"
	fi
}

prompt_enrollment_token() {
	if [ -n "$ENROLLMENT_TOKEN" ]; then
		return 0
	fi
	if ! can_prompt; then
		die "PROBO_ENROLLMENT_TOKEN is required in non-interactive mode"
	fi

	read_secret "Enrollment token: "
	if [ -z "$ENROLLMENT_TOKEN" ]; then
		die "enrollment token is required"
	fi
}

parse_args() {
	while [ $# -gt 0 ]; do
		case "$1" in
		--server)
			[ $# -ge 2 ] || die "--server requires a value"
			SERVER_URL="$2"
			shift 2
			;;
		--enrollment-token)
			[ $# -ge 2 ] || die "--enrollment-token requires a value"
			ENROLLMENT_TOKEN="$2"
			shift 2
			;;
		--no-auto-update)
			NO_AUTO_UPDATE=true
			shift
			;;
		--skip-service)
			SKIP_SERVICE=true
			shift
			;;
		--dir)
			[ $# -ge 2 ] || die "--dir requires a value"
			STATE_DIR="$2"
			shift 2
			;;
		-h | --help)
			usage
			exit 0
			;;
		*)
			die "unknown option: $1 (try --help)"
			;;
		esac
	done
}

run_agent_install() {
	set -- --server "$SERVER_URL" --enrollment-token "$ENROLLMENT_TOKEN"
	if [ -n "$STATE_DIR" ]; then
		set -- "$@" --dir "$STATE_DIR"
	fi
	case "$NO_AUTO_UPDATE" in
	1 | true | TRUE | yes | YES) set -- "$@" --no-auto-update ;;
	esac
	case "$SKIP_SERVICE" in
	1 | true | TRUE | yes | YES) set -- "$@" --skip-service ;;
	esac
	"$BINARY_PATH" install "$@"
}

main() {
	parse_args "$@"

	if [ "$(id -u)" -ne 0 ]; then
		die "this installer must run as root; re-run with: curl -fsSL \"…/install.sh\" | sudo sh"
	fi

	require_cmd curl
	require_cmd tar
	require_cmd install

	detect_platform
	resolve_embedded_release

	workdir="$(mktemp -d "${TMPDIR:-/tmp}/probo-agent-install.XXXXXX")"
	trap 'rm -rf "$workdir"' EXIT INT HUP TERM

	printf 'Downloading probo-agent %s …\n' "$archive_name"

	curl -fsSL "${RELEASE_BASE}/${archive_name}" -o "${workdir}/${archive_name}"

	verify_embedded_checksum "${workdir}/${archive_name}"

	tar -xzf "${workdir}/${archive_name}" -C "$workdir"
	if [ ! -f "${workdir}/${archive_dir}/probo-agent" ]; then
		die "archive did not contain probo-agent binary"
	fi

	install -m 0755 "${workdir}/${archive_dir}/probo-agent" "$BINARY_PATH"
	printf 'Installed %s\n' "$BINARY_PATH"

	prompt_server_url
	prompt_enrollment_token

	printf 'Enrolling device …\n'
	if run_agent_install; then
		enroll_ok=true
	else
		enroll_ok=false
	fi

	if [ "$enroll_ok" = true ]; then
		case "$SKIP_SERVICE" in
		1 | true | TRUE | yes | YES)
			printf 'Device enrolled (service installation skipped).\n'
			;;
		*)
			printf 'Device enrolled and service installed.\n'
			;;
		esac
	else
		printf 'warning: probo-agent install failed; binary is at %s\n' "$BINARY_PATH" >&2
		printf 'Re-run: %s install --server … --enrollment-token …\n' "$BINARY_PATH" >&2
		exit 1
	fi
}

main "$@"
