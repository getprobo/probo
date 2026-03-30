#!/usr/bin/env bash
# Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
#
# Permission to use, copy, modify, and/or distribute this software for any
# purpose with or without fee is hereby granted, provided that the above
# copyright notice and this permission notice appear in all copies.
#
# THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
# REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
# AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
# INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
# LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
# OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
# PERFORMANCE OF THIS SOFTWARE.

set -euo pipefail

YEAR="$(date +%Y)"

# The license text without comment markers. Each line is prefixed at runtime.
HEADER_LINES=(
	"Copyright (c) %YEAR% Probo Inc <hello@getprobo.com>."
	""
	"Permission to use, copy, modify, and/or distribute this software for any"
	"purpose with or without fee is hereby granted, provided that the above"
	"copyright notice and this permission notice appear in all copies."
	""
	"THE SOFTWARE IS PROVIDED \"AS IS\" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH"
	"REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY"
	"AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,"
	"INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM"
	"LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR"
	"OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR"
	"PERFORMANCE OF THIS SOFTWARE."
)

# Number of lines the header occupies (used to strip an existing header).
HEADER_LINE_COUNT=${#HEADER_LINES[@]}

# Files containing "--license-header-ignore" in the first 3 lines are skipped.
is_excluded() {
	head -3 "$1" | grep -q -- "--license-header-ignore"
}

# ---------------------------------------------------------------------------
# Comment prefix by file extension
# ---------------------------------------------------------------------------

comment_prefix() {
	case "$1" in
		*.go|*.ts|*.tsx) echo "//" ;;
		*.sql)           echo "--" ;;
	esac
}

# ---------------------------------------------------------------------------
# Build a header string for a given file
# ---------------------------------------------------------------------------

build_header() {
	local prefix="$1" year="$2"
	local line
	for line in "${HEADER_LINES[@]}"; do
		line="${line//%YEAR%/$year}"
		if [ -z "$line" ]; then
			echo "$prefix"
		else
			echo "$prefix $line"
		fi
	done
}

# ---------------------------------------------------------------------------
# Year helpers
# ---------------------------------------------------------------------------

git_year() {
	local f="$1"
	local created modified

	created=$(git log --diff-filter=A --format='%ad' --date=format:'%Y' -- "$f" 2>/dev/null | tail -1)
	modified=$(git log -1 --format='%ad' --date=format:'%Y' -- "$f" 2>/dev/null)

	: "${created:=$YEAR}"
	: "${modified:=$YEAR}"

	if [ "$created" = "$modified" ]; then
		echo "$created"
	else
		echo "${created}-${modified}"
	fi
}

# ---------------------------------------------------------------------------
# Detect whether a file already has the header and extract its year
# ---------------------------------------------------------------------------

# Returns 0 if any of the first 5 lines contains a Probo copyright notice
# using the expected prefix. Prints the existing year string.
has_header() {
	local f="$1" prefix="$2"
	local line year
	for line in $(seq 1 5); do
		local text
		text=$(sed -n "${line}p" "$f")
		if [[ "$text" =~ ^"$prefix Copyright (c) "([0-9]{4}(-[0-9]{4})?)" Probo Inc" ]]; then
			echo "${BASH_REMATCH[1]}"
			return 0
		fi
	done
	return 1
}

# ---------------------------------------------------------------------------
# Strip an existing header (HEADER_LINE_COUNT lines + optional blank line)
# ---------------------------------------------------------------------------

strip_header() {
	local f="$1"
	local skip=$HEADER_LINE_COUNT

	# If the line right after the header is blank, skip it too.
	local next_line
	next_line=$(sed -n "$((skip + 1))p" "$f")
	if [ -z "$next_line" ]; then
		skip=$((skip + 1))
	fi

	tail -n +"$((skip + 1))" "$f"
}

# ---------------------------------------------------------------------------
# Collect files
# ---------------------------------------------------------------------------

find_source_files() {
	find . -type f \( \
		-name "*.go" -o \
		-name "*.ts" -o -name "*.tsx" -o \
		-name "*.sql" \
	\) \
		-not -path "*/node_modules/*" \
		-not -path "*/.context/*" \
		-not -path "*/dist/*" \
		-not -path "*/__generated__/*" \
		-not -path "*/vendor/*"
}

# ---------------------------------------------------------------------------
# Modes
# ---------------------------------------------------------------------------

check_files() {
	local missing=0
	while IFS= read -r f; do
		is_excluded "$f" && continue

		local prefix
		prefix=$(comment_prefix "$f")
		if ! has_header "$f" "$prefix" > /dev/null; then
			echo "missing header: $f"
			missing=$((missing + 1))
		fi
	done < <(find_source_files)

	if [ "$missing" -gt 0 ]; then
		echo ""
		echo "error: $missing file(s) missing ISC license header"
		echo "run 'contrib/license-header.sh fix' to add them"
		exit 1
	fi
}

fix_files() {
	local fixed=0
	while IFS= read -r f; do
		is_excluded "$f" && continue

		local prefix
		prefix=$(comment_prefix "$f")
		local year
		year=$(git_year "$f")

		local existing_year
		if existing_year=$(has_header "$f" "$prefix"); then
			# Header exists — update only if the year changed.
			if [ "$existing_year" = "$year" ]; then
				continue
			fi
			# Strip old header, will re-add below.
			local body
			body=$(strip_header "$f")
			local tmpfile
			tmpfile=$(mktemp)
			{ build_header "$prefix" "$year"; echo; echo "$body"; } > "$tmpfile" && mv "$tmpfile" "$f"
		else
			# No header — prepend.
			local tmpfile
			tmpfile=$(mktemp)
			{ build_header "$prefix" "$year"; echo; cat "$f"; } > "$tmpfile" && mv "$tmpfile" "$f"
		fi

		fixed=$((fixed + 1))
		echo "fixed: $f ($year)"
	done < <(find_source_files)

	echo ""
	echo "$fixed file(s) fixed"
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

usage() {
	echo "Usage: $0 {check|fix}" >&2
	exit 1
}

case "${1:-}" in
	check) check_files ;;
	fix)   fix_files ;;
	*)     usage ;;
esac
