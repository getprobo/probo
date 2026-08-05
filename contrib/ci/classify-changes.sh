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
prb="$(
  contrib/ci/go-package-affected.sh \
    "$base_sha" "$head_sha" ./cmd/prb
)"
probod="$(
  contrib/ci/go-package-affected.sh \
    "$base_sha" "$head_sha" ./cmd/probod
)"
probod_bootstrap="$(
  contrib/ci/go-package-affected.sh \
    "$base_sha" "$head_sha" ./cmd/probod-bootstrap
)"

go_changed="false"
javascript="false"
shell="false"
swift="false"
e2e="$probod"
frontend="false"
snapshot="false"

for changed_file in "${changed_files[@]}"; do
  # go gates lint-go and unit tests. Match first-party package dirs so
  # testdata and //go:embed inputs are not skipped; do not rely on *.go.
  # enroll-ui is Swift. Mixed JS/Go trees stay on the extension list.
  case "$changed_file" in
    cmd/probo-agent/installer/macos/enroll-ui/*)
      ;;
    cmd/* | pkg/* | e2e/* | internal/* | \
      *.go | *.graphql | *.tmpl | *.sql | .golangci.yml | \
      */gqlgen.yaml | */mcpgen.yaml | */specification.yaml)
      go_changed="true"
      ;;
  esac

  # javascript gates lint-js, which is the only PR job that runs
  # make relay. Schema files, relay.config.json, and the merge
  # script must match here or a schema-only PR skips Relay.
  case "$changed_file" in
    *.cjs | *.css | *.js | *.jsx | *.mjs | *.ts | *.tsx | \
      *.graphql | relay.config.json | contrib/merge-graphql-schema.sh | \
      package.json | package-lock.json | turbo.json | \
      apps/* | packages/*)
      javascript="true"
      ;;
  esac

  case "$changed_file" in
    .nvmrc | package.json | package-lock.json | turbo.json | \
      apps/compliance-portal/* | apps/console/* | \
      packages/coredata/* | packages/emails/* | packages/helpers/* | \
      packages/hooks/* | packages/i18n/* | packages/prosemirror/* | \
      packages/react-lazy/* | packages/relay/* | packages/routes/* | \
      packages/tsconfig/* | packages/ui/*)
      frontend="true"
      ;;
  esac

  case "$changed_file" in
    *.sh | .shellcheckrc | GNUmakefile)
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
    cfg/* | compose.yaml | compose.github-action.yaml | \
      compose/* | e2e/*)
      e2e="true"
      ;;
  esac

  # Image + scanner config. Snapshot builds the Docker image (Trivy)
  # and runs snapshot-scan (Grype). E2E uses Compose + host bin/probod.
  case "$changed_file" in
    Dockerfile | entrypoint.sh | .trivyignore.yaml | .grype.yaml)
      snapshot="true"
      ;;
  esac
done

if [[ "$probod" == "true" ||
  "$prb" == "true" ||
  "$probod_bootstrap" == "true" ||
  "$frontend" == "true" ]]; then
  snapshot="true"
fi

echo "agent=$agent"
echo "e2e=$e2e"
echo "go=$go_changed"
echo "javascript=$javascript"
echo "shell=$shell"
echo "snapshot=$snapshot"
echo "swift=$swift"
