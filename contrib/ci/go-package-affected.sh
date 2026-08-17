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
package="${3:?Go package is required}"

repository_root="$(git rev-parse --show-toplevel)"

# A shallow checkout can omit the comparison commit after a multi-commit push.
# Build in that case instead of risking a false negative.
if ! git cat-file -e "${base_sha}^{commit}" 2>/dev/null \
  || ! git cat-file -e "${head_sha}^{commit}" 2>/dev/null; then
  echo "true"
  exit 0
fi

mapfile -t changed_files < <(
  git diff --name-only --no-renames "$base_sha" "$head_sha"
)

for changed_file in "${changed_files[@]}"; do
  case "$changed_file" in
    go.mod | go.sum | .github/workflows/make.yaml | contrib/ci/go-package-affected.sh)
      echo "true"
      exit 0
      ;;
  esac
done

if ! dependency_output="$(
  CGO_ENABLED=0 go list -deps \
    -f '{{if .Module}}{{if .Module.Main}}{{.Dir}}{{end}}{{end}}' \
    "$package"
)"; then
  echo "true"
  exit 0
fi

mapfile -t dependency_dirs < <(
  printf '%s\n' "$dependency_output" | awk 'NF' | sort -u
)

for changed_file in "${changed_files[@]}"; do
  if [[ "$changed_file" == *_test.go ]]; then
    continue
  fi

  absolute_path="${repository_root}/${changed_file}"
  for dependency_dir in "${dependency_dirs[@]}"; do
    if [[ "$absolute_path" == "$dependency_dir" ||
      "$absolute_path" == "$dependency_dir/"* ]]; then
      echo "true"
      exit 0
    fi
  done
done

echo "false"
