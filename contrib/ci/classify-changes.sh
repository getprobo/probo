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

base_sha="${1:?base SHA is required}"
head_sha="${2:?head SHA is required}"

emit_all() {
  local value="$1"

  echo "agent=$value"
  echo "e2e=$value"
  echo "go=$value"
  echo "javascript=$value"
  echo "shell=$value"
  echo "snapshot=$value"
  echo "swift=$value"
}

if ! git cat-file -e "${base_sha}^{commit}" 2>/dev/null \
  || ! git cat-file -e "${head_sha}^{commit}" 2>/dev/null; then
  emit_all "true"
  exit 0
fi

mapfile -t changed_files < <(
  git diff --name-only --no-renames "$base_sha" "$head_sha"
)

for changed_file in "${changed_files[@]}"; do
  case "$changed_file" in
    .github/actions/* | .github/workflows/* | \
      contrib/ci/* | GNUmakefile | go.mod | go.sum)
      emit_all "true"
      exit 0
      ;;
  esac
done

agent="$(
  contrib/ci/go-package-affected.sh \
    "$base_sha" "$head_sha" ./cmd/probo-agent
)"
probod="$(
  contrib/ci/go-package-affected.sh \
    "$base_sha" "$head_sha" ./cmd/probod
)"

go_changed="false"
javascript="false"
shell="false"
swift="false"
e2e="$probod"

for changed_file in "${changed_files[@]}"; do
  case "$changed_file" in
    *.go | *.graphql | *.tmpl | .golangci.yml)
      go_changed="true"
      ;;
  esac

  case "$changed_file" in
    *.cjs | *.css | *.js | *.jsx | *.mjs | *.ts | *.tsx | \
      package.json | package-lock.json | turbo.json | \
      apps/* | packages/*)
      javascript="true"
      ;;
  esac

  case "$changed_file" in
    *.sh | GNUmakefile)
      shell="true"
      ;;
  esac

  case "$changed_file" in
    .swift-format | .swiftlint.yml | \
      cmd/probo-agent/installer/macos/enroll-ui/*)
      swift="true"
      ;;
  esac

  case "$changed_file" in
    Dockerfile | cfg/* | compose.yaml | compose.github-action.yaml | \
      compose/* | e2e/*)
      e2e="true"
      ;;
  esac
done

snapshot="$probod"
if [[ "$javascript" == "true" ]]; then
  snapshot="true"
fi

echo "agent=$agent"
echo "e2e=$e2e"
echo "go=$go_changed"
echo "javascript=$javascript"
echo "shell=$shell"
echo "snapshot=$snapshot"
echo "swift=$swift"
