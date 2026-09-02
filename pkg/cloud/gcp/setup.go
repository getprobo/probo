// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package gcp

import (
	"fmt"
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

const (
	// DefaultTerraformModuleSource is the public registry address of the GCP
	// audit-role module. Phase 3 owns the module; this package only formats
	// the snippet.
	DefaultTerraformModuleSource = "getprobo/audit-role/gcp"

	// DefaultServiceAccountName is the service account the customer setup
	// template creates.
	DefaultServiceAccountName = "probo-audit"

	// AudienceTemplate is the JWT audience the customer will configure after
	// they create the pool. The exact audience is unknown at setup time.
	AudienceTemplate = "https://iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/POOL_ID/providers/PROVIDER_ID"
)

type (
	// ConnectorSetup is the server-derived values the connect dialog shows so
	// the customer never retypes an issuer, subject or audience.
	ConnectorSetup struct {
		Issuer                      string
		Audience                    string
		Subject                     string
		SuggestedServiceAccountName string
		TerraformSnippet            string
	}

	// ConnectorInstallConfig is the deployment-supplied install artifact used
	// to fill the Terraform snippet. An empty source omits the snippet.
	ConnectorInstallConfig struct {
		TerraformModuleSource string
	}

	// ConnectorSetupInput is what BuildConnectorSetup needs. Every Probo-derived
	// string is supplied already resolved so this function never talks to config
	// or an issuer.
	ConnectorSetupInput struct {
		IssuerURL             string
		Subject               string
		TerraformModuleSource string
	}
)

// BuildConnectorSetup fills issuer, subject and snippets from already-resolved
// values. It does not mint a token and does not read global config.
func BuildConnectorSetup(in ConnectorSetupInput) (ConnectorSetup, error) {
	if in.IssuerURL == "" {
		return ConnectorSetup{}, fmt.Errorf("cannot build gcp connector setup: issuer is required")
	}

	if in.Subject == "" {
		return ConnectorSetup{}, fmt.Errorf("cannot build gcp connector setup: subject is required")
	}

	return ConnectorSetup{
		Issuer:                      in.IssuerURL,
		Audience:                    AudienceTemplate,
		Subject:                     in.Subject,
		SuggestedServiceAccountName: DefaultServiceAccountName,
		TerraformSnippet:            terraformSnippet(in.TerraformModuleSource, in.IssuerURL, in.Subject),
	}, nil
}

// ConnectorSetupFor resolves issuer and subject from the live issuer, then
// fills snippets. It is what GraphQL and MCP call so those surfaces cannot
// drift from token minting.
func ConnectorSetupFor(
	issuer *identityfederation.Issuer,
	organizationID gid.GID,
	install ConnectorInstallConfig,
) (ConnectorSetup, error) {
	if issuer == nil {
		return ConnectorSetup{}, fmt.Errorf("cannot build gcp connector setup: identity federation is not configured")
	}

	issuerURL, err := issuer.IssuerURL(organizationID)
	if err != nil {
		return ConnectorSetup{}, fmt.Errorf("cannot build gcp connector setup: %w", err)
	}

	return BuildConnectorSetup(
		ConnectorSetupInput{
			IssuerURL:             issuerURL,
			Subject:               organizationID.String(),
			TerraformModuleSource: install.TerraformModuleSource,
		},
	)
}

func terraformSnippet(moduleSource, issuerURL, subject string) string {
	if moduleSource == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString("module \"probo_audit\" {\n")
	b.WriteString("  source = ")
	b.WriteString(strconv.Quote(moduleSource))
	b.WriteString("\n\n")
	b.WriteString("  probo_issuer_url     = ")
	b.WriteString(strconv.Quote(issuerURL))
	b.WriteString("\n")
	b.WriteString("  probo_subject        = ")
	b.WriteString(strconv.Quote(subject))
	b.WriteString("\n")
	b.WriteString("  service_account_name = ")
	b.WriteString(strconv.Quote(DefaultServiceAccountName))
	b.WriteString("\n")
	b.WriteString("}\n")

	return b.String()
}
