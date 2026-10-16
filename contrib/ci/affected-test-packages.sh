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

# Print test packages affected by the diff between two commits, one import
# path per line. Prints "all" when the diff cannot be scoped (missing
# commits, dependency files, or workflow changes) so the caller runs the
# full suite.
set -euo pipefail

base_sha="${1:?base SHA is required}"
head_sha="${2:?head SHA is required}"

repository_root="$(git rev-parse --show-toplevel)"

if ! git cat-file -e "${base_sha}^{commit}" 2>/dev/null \
  || ! git cat-file -e "${head_sha}^{commit}" 2>/dev/null; then
  echo all
  exit 0
fi

changed_files="$(git diff --name-only --no-renames "$base_sha" "$head_sha")"
if [ -z "$changed_files" ]; then
  echo all
  exit 0
fi

while IFS= read -r changed_file; do
  [ -n "$changed_file" ] || continue
  case "$changed_file" in
    go.mod | go.sum | GNUmakefile | .github/actions/* | .github/workflows/* | contrib/ci/*)
      echo all
      exit 0
      ;;
  esac
done <<<"$changed_files"

list_output="$(
  go list -f '{{.ImportPath}}	{{.Dir}}	{{if or .TestGoFiles .XTestGoFiles}}test{{end}}	{{join .Imports ","}}	{{join .TestImports ","}}	{{join .XTestImports ","}}' ./...
)" || {
  echo all
  exit 0
}

changed_list="$(mktemp)"
trap 'rm -f "$changed_list"' EXIT
printf '%s\n' "$changed_files" >"$changed_list"

awk -v root="$repository_root" -v changed_list="$changed_list" '
  function own(file,   absolute, path, dir, best, best_len) {
    absolute = root "/" file
    best = ""
    best_len = -1
    for (path in dirs) {
      dir = dirs[path]
      if (absolute != dir && index(absolute, dir "/") != 1) {
        continue
      }
      if (length(dir) > best_len) {
        best = path
        best_len = length(dir)
      }
    }
    return best
  }

  BEGIN {
    FS = "\t"
    while ((getline file < changed_list) > 0) {
      if (file != "") {
        files[++nfiles] = file
      }
    }
    close(changed_list)
  }
  {
    path = $1
    if (path == "") {
      next
    }
    dirs[path] = $2
    if ($3 == "test") {
      tests[path] = 1
    }
    raw[path] = $4 "," $5 "," $6
    packages[++npackages] = path
  }
  END {
    for (i = 1; i <= npackages; i++) {
      path = packages[i]
      count = split(raw[path], deps, ",")
      for (j = 1; j <= count; j++) {
        dependency = deps[j]
        if (dependency != "" && dependency in dirs) {
          reverse[dependency] = reverse[dependency] " " path
        }
      }
    }

    for (i = 1; i <= nfiles; i++) {
      file = files[i]
      owner = own(file)
      if (owner == "") {
        continue
      }
      affected[owner] = 1
      if (file !~ /_test\.go$/) {
        expand[owner] = 1
        queue[++nqueue] = owner
      }
    }

    for (head = 1; head <= nqueue; head++) {
      current = queue[head]
      count = split(reverse[current], dependents, " ")
      for (j = 1; j <= count; j++) {
        dependent = dependents[j]
        # A test-only edit is affected but not queued. Skip only packages
        # already expanded, so this dependency edit still walks their importers.
        if (dependent == "" || dependent in expand) {
          continue
        }
        affected[dependent] = 1
        expand[dependent] = 1
        queue[++nqueue] = dependent
      }
    }

    for (path in affected) {
      if (path in tests) {
        print path
      }
    }
  }
' <<<"$list_output" | sort
