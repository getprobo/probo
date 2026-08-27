#!/usr/bin/env bash
#
# Copyright (c) 2026 Probo Inc <hello@probo.com>.
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

set -euo pipefail

# Pin the real upstream battery. Bump this deliberately after reviewing upstream
# workload/schema changes; never benchmark a moving main branch in CI.
readonly DEFAULT_REF="8feb8be8be203e3b878bf2cb5919601094f3c816"
readonly REF="${AUTOMERGE_BATTERY_REF:-$DEFAULT_REF}"
readonly TOOLCHAIN="${AUTOMERGE_BATTERY_RUST_TOOLCHAIN:-1.90.0}"
readonly CACHE_ROOT="${XDG_CACHE_HOME:-$HOME/.cache}/probo/automerge-battery"
readonly CHECKOUT="$CACHE_ROOT/$REF"
SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIRECTORY
readonly LOCKFILE="$SCRIPT_DIRECTORY/automerge-battery.Cargo.lock"

if [[ ! -d "$CHECKOUT/.git" ]]; then
  mkdir -p "$CHECKOUT"
  git -C "$CHECKOUT" init --quiet
  git -C "$CHECKOUT" remote add origin https://github.com/automerge/automerge.git
fi

git -C "$CHECKOUT" fetch --quiet --depth 1 origin "$REF"
COMMIT="$(git -C "$CHECKOUT" rev-parse --verify "FETCH_HEAD^{commit}")"
readonly COMMIT

git -C "$CHECKOUT" checkout --quiet --detach "$COMMIT"
cp "$LOCKFILE" "$CHECKOUT/rust/Cargo.lock"

if ! rustup run "$TOOLCHAIN" rustc --version >/dev/null 2>&1; then
  rustup toolchain install "$TOOLCHAIN" --profile minimal
fi

exec cargo +"$TOOLCHAIN" run \
  --release \
  --locked \
  --manifest-path "$CHECKOUT/rust/Cargo.toml" \
  -p benchmark-battery \
  -- "$@"
