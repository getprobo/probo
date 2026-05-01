// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package cloudaccount

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/rand"
)

const (
	// gcpScannerProjectSuffixByteLen is the entropy bytes consumed
	// by the random project-id suffix. 4 bytes -> 8 hex chars,
	// keeping the full project id under GCP's 30-char limit even
	// when the prefix grows.
	gcpScannerProjectSuffixByteLen = 4

	gcpScriptTmpl = `#!/usr/bin/env bash
# Probo cloud-account install script
# scope kind:       {{.ScopeKind}}
# scope identifier: {{.ScopeIdentifier}}
#
# Run this script with the gcloud CLI authenticated as a user with
# Project Creator + Organization Admin (org scope) or Project Owner
# (project scope) permissions.
set -euo pipefail

SCANNER_PROJECT_ID="{{.ScannerProjectID}}"
SA_NAME="probo-cloud-scanner"
ROLE_ID="ProboCloudScanner"
SCOPE_KIND="{{.ScopeKind}}"
SCOPE_IDENTIFIER="{{.ScopeIdentifier}}"

# 1. Create the dedicated scanner project owned by Probo's SA.
gcloud projects create "${SCANNER_PROJECT_ID}" --name="Probo Cloud Scanner"
gcloud config set project "${SCANNER_PROJECT_ID}"

# 2. Enable the APIs required by the configured audit modules.
{{range .APIs}}gcloud services enable {{.}} --project="${SCANNER_PROJECT_ID}"
{{end}}

# 3. Create the dedicated service account.
gcloud iam service-accounts create "${SA_NAME}" \
    --project="${SCANNER_PROJECT_ID}" \
    --display-name="Probo Cloud Scanner"

SA_EMAIL="${SA_NAME}@${SCANNER_PROJECT_ID}.iam.gserviceaccount.com"

# 4. Grant the predefined roles at the customer's chosen scope.
{{if eq .ScopeKind "GCP_PROJECT"}}{{range .Roles}}gcloud projects add-iam-policy-binding "${SCOPE_IDENTIFIER}" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="{{.}}"
{{end}}{{else if eq .ScopeKind "GCP_ORGANIZATION"}}{{range .Roles}}gcloud organizations add-iam-policy-binding "${SCOPE_IDENTIFIER}" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="{{.}}"
{{end}}{{end}}

# 5. Mint a JSON key the customer pastes back into Probo.
KEY_FILE="probo-cloud-account-${SCANNER_PROJECT_ID}.json"
gcloud iam service-accounts keys create "${KEY_FILE}" \
    --iam-account="${SA_EMAIL}" \
    --project="${SCANNER_PROJECT_ID}"

echo
echo "Service account email: ${SA_EMAIL}"
echo "JSON key file:         ${KEY_FILE}"
echo
echo "Paste the JSON key contents into Probo's Cloud Accounts UI to complete the install."
`
)

type (
	// GCPInstallAssets is the structured payload returned to the
	// frontend when the customer starts a GCP install.
	GCPInstallAssets struct {
		SetupScript      string                             `json:"setup_script"`
		ScannerProjectID string                             `json:"scanner_project_id"`
		RequiredRoles    []string                           `json:"required_roles"`
		RequiredAPIs     []string                           `json:"required_apis"`
		Modules          []coredata.CloudAccountAuditModule `json:"modules"`
	}

	gcpScriptParams struct {
		ScannerProjectID string
		ScopeKind        string
		ScopeIdentifier  string
		Roles            []string
		APIs             []string
	}
)

// BuildGCPInstallScript renders a self-contained gcloud bash script
// that creates a dedicated probo-scanner project, enables the
// required APIs, mints a service account and JSON key, and grants
// the scope-appropriate roles. Roles and APIs are scope-dependent
// (see permissions.go for the rationale).
func BuildGCPInstallScript(
	modules []coredata.CloudAccountAuditModule,
	scopeKind coredata.CloudAccountScopeKind,
	scopeIdentifier string,
) (GCPInstallAssets, error) {
	if scopeKind != coredata.CloudAccountScopeKindGCPProject &&
		scopeKind != coredata.CloudAccountScopeKindGCPOrganization {
		return GCPInstallAssets{}, fmt.Errorf("cannot build gcp install script: unsupported scope kind %q", scopeKind)
	}

	roles := GCPRolesForModules(scopeKind, modules)
	apis := GCPAPIsForModules(scopeKind, modules)
	if len(roles) == 0 || len(apis) == 0 {
		return GCPInstallAssets{}, fmt.Errorf("cannot build gcp install script: no roles or apis for scope %q and modules %v", scopeKind, modules)
	}

	suffix, err := rand.HexString(gcpScannerProjectSuffixByteLen)
	if err != nil {
		return GCPInstallAssets{}, fmt.Errorf("cannot generate gcp scanner project suffix: %w", err)
	}

	scannerProjectID := fmt.Sprintf("probo-scanner-%s", suffix)

	tmpl, err := template.New("gcp-install").Parse(gcpScriptTmpl)
	if err != nil {
		return GCPInstallAssets{}, fmt.Errorf("cannot parse gcp install template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, gcpScriptParams{
		ScannerProjectID: scannerProjectID,
		ScopeKind:        scopeKind.String(),
		ScopeIdentifier:  scopeIdentifier,
		Roles:            roles,
		APIs:             apis,
	}); err != nil {
		return GCPInstallAssets{}, fmt.Errorf("cannot render gcp install template: %w", err)
	}

	return GCPInstallAssets{
		SetupScript:      strings.TrimSpace(buf.String()) + "\n",
		ScannerProjectID: scannerProjectID,
		RequiredRoles:    roles,
		RequiredAPIs:     apis,
		Modules:          modules,
	}, nil
}
