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

package mcp_v1

import (
	"go.probo.inc/probo/pkg/cloudaccount"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

// Cloud-account MCP type adapters. The tool resolvers live in
// schema.resolvers.go (the file mcpgen owns) and call into these
// helpers; keeping them in this file isolates the typed conversions
// from the auto-generated stubs and matches the access_review.go /
// asset.go layout used by other entity packages.

// newMCPCloudAccount maps a coredata.CloudAccount to the MCP types.
// Surfaces external_id unconditionally — the MCP role gate is
// enforced at MustAuthorize time, the JSON shape is uniform.
func newMCPCloudAccount(account *coredata.CloudAccount) *types.CloudAccount {
	return &types.CloudAccount{
		ID:                  account.ID,
		OrganizationID:      account.OrganizationID,
		Label:               account.Label,
		Provider:            account.Provider,
		CredentialKind:      account.CredentialKind,
		ScopeKind:           account.ScopeKind,
		ScopeIdentifier:     account.ScopeIdentifier,
		Status:              account.Status,
		EnabledAuditModules: account.EnabledAuditModules,
		ExternalID:          account.ExternalID,
		LastProbeAt:         account.LastProbeAt,
		LastProbeError:      account.LastProbeError,
		LastVerifiedAt:      account.LastVerifiedAt,
		CreatedAt:           account.CreatedAt,
		UpdatedAt:           account.UpdatedAt,
	}
}

func newMCPAWSInstallAssets(a *cloudaccount.AWSInstallAssets) *types.AWSInstallAssets {
	return &types.AWSInstallAssets{
		QuickCreateURL:  a.QuickCreateURL,
		ExternalID:      a.ExternalID,
		PrincipalArn:    a.PrincipalARN,
		RequiredActions: a.RequiredActions,
	}
}

func newMCPGCPInstallAssets(a *cloudaccount.GCPInstallAssets) *types.GCPInstallAssets {
	return &types.GCPInstallAssets{
		SetupScript:   a.SetupScript,
		RequiredRoles: a.RequiredRoles,
		RequiredApis:  a.RequiredAPIs,
	}
}

func newMCPAzureInstallAssets(a *cloudaccount.AzureInstallGuide) *types.AzureInstallAssets {
	steps := make([]*types.AzureInstallStep, 0, len(a.Steps))
	for i := range a.Steps {
		s := a.Steps[i]
		step := &types.AzureInstallStep{
			Title: s.Title,
			Body:  s.Body,
		}
		if s.Code != "" {
			code := s.Code
			step.Code = &code
		}
		steps = append(steps, step)
	}
	return &types.AzureInstallAssets{
		Steps:                    steps,
		RequiredRbacRoles:        a.RequiredRBACRoles,
		RequiredGraphPermissions: a.RequiredGraphPermissions,
	}
}
