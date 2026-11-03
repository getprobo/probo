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

# Print the "Verifying this release" section of a GitHub release note.
#
# Usage: release-verify-notes.sh <example-archive> [<image:tag>]
#
# Must run inside the release workflow: the signing identity is taken from
# GITHUB_WORKFLOW_REF, so the printed commands pin the exact workflow file
# and tag that signed the checksums, archives and (optionally) the image.

set -euo pipefail

archive="${1:?example archive name is required}"
image="${2:-}"

: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"
: "${GITHUB_WORKFLOW_REF:?GITHUB_WORKFLOW_REF is required}"

issuer="https://token.actions.githubusercontent.com"
identity="https://github.com/${GITHUB_WORKFLOW_REF}"
workflow="${GITHUB_WORKFLOW_REF%@*}"

cat <<EOF

## Verifying this release

\`\`\`bash
# 1. cosign signature over the checksum file (identity = the release workflow).
cosign verify-blob \\
  --bundle checksums.txt.bundle \\
  --certificate-oidc-issuer ${issuer} \\
  --certificate-identity '${identity}' \\
  checksums.txt

# 2. hash every downloaded artifact against the now-trusted checksums.
shasum -a 256 --ignore-missing -c checksums.txt

# 3. GitHub build provenance (repeat for whichever archive you pulled).
gh attestation verify ${archive} \\
  --repo ${GITHUB_REPOSITORY} \\
  --signer-workflow ${workflow}

# 4. SBOM attestation binding sbom.json to the archive.
gh attestation verify ${archive} \\
  --repo ${GITHUB_REPOSITORY} \\
  --predicate-type https://cyclonedx.org/bom
\`\`\`
EOF

if [ -n "${image}" ]; then
  cat <<EOF

The Docker image is signed by the same workflow:

\`\`\`bash
cosign verify \\
  --certificate-oidc-issuer ${issuer} \\
  --certificate-identity '${identity}' \\
  ${image}

gh attestation verify oci://${image} \\
  --repo ${GITHUB_REPOSITORY} \\
  --predicate-type https://cyclonedx.org/bom
\`\`\`
EOF
fi

cat <<EOF

The release ships one SBOM for the whole build (\`sbom.json\`, CycloneDX JSON).
EOF
