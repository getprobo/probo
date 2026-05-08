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

package console_v1

import (
	"context"
	"errors"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/cloudaccount"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

// resolveCloudAccountOrgID maps a cloud-account GID to the
// organization GID that owns it, so resolvers that take only
// cloud_account_id (Verify, Rotate, Delete) can authorize against
// the right organization scope. Returns NotFound through gqlutils
// so the user-facing error matches every other "object not found"
// path, and Internal otherwise (logged before mapping).
func (r *Resolver) resolveCloudAccountOrgID(
	ctx context.Context,
	cloudAccountID gid.GID,
) (gid.GID, error) {
	prb := r.ProboService(ctx, cloudAccountID.TenantID())
	account, err := prb.CloudAccounts.GetMetadata(ctx, cloudAccountID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return gid.GID{}, gqlutils.NotFound(ctx, err)
		}
		r.logger.ErrorCtx(ctx, "cannot resolve cloud account org id", log.Error(err))
		return gid.GID{}, gqlutils.Internal(ctx)
	}
	return account.OrganizationID, nil
}

// newCloudAccount maps a coredata.CloudAccount entity to the
// GraphQL types.CloudAccount value. revealCredentialAdjacent
// controls whether scope.identifier and last_probe_error are
// populated (true for OWNER/ADMIN, false for AUDITOR/VIEWER).
func newCloudAccount(account *coredata.CloudAccount, revealCredentialAdjacent bool) *types.CloudAccount {
	scope := &types.CloudAccountScope{Kind: account.ScopeKind}
	if revealCredentialAdjacent {
		identifier := account.ScopeIdentifier
		scope.Identifier = &identifier
	}
	out := &types.CloudAccount{
		ID:                  account.ID,
		Provider:            account.Provider,
		Label:               account.Label,
		Status:              account.Status,
		Scope:               scope,
		CredentialKind:      account.CredentialKind,
		EnabledAuditModules: append([]coredata.CloudAccountAuditModule{}, account.EnabledAuditModules...),
		LastProbeAt:         account.LastProbeAt,
		LastVerifiedAt:      account.LastVerifiedAt,
		CreatedAt:           account.CreatedAt,
		UpdatedAt:           account.UpdatedAt,
	}
	if revealCredentialAdjacent {
		out.LastProbeError = account.LastProbeError
	}
	return out
}

// newCloudAccountForList is the list-time variant. The split
// exists so a future divergence in list-vs-get fields (e.g. trim
// some heavy fields in the list view) does not require touching
// every call site.
func newCloudAccountForList(account *coredata.CloudAccount, revealCredentialAdjacent bool) *types.CloudAccount {
	return newCloudAccount(account, revealCredentialAdjacent)
}

func newAWSInstallAssets(a *cloudaccount.AWSInstallAssets) *types.AWSInstallAssets {
	return &types.AWSInstallAssets{
		QuickCreateURL:  a.QuickCreateURL,
		ExternalID:      a.ExternalID,
		PrincipalArn:    a.PrincipalARN,
		RequiredActions: append([]string{}, a.RequiredActions...),
	}
}

func newGCPInstallAssets(a *cloudaccount.GCPInstallAssets) *types.GCPInstallAssets {
	return &types.GCPInstallAssets{
		SetupScript:   a.SetupScript,
		RequiredRoles: append([]string{}, a.RequiredRoles...),
		RequiredApis:  append([]string{}, a.RequiredAPIs...),
	}
}

func newAzureInstallAssets(a *cloudaccount.AzureInstallGuide) *types.AzureInstallAssets {
	steps := make([]*types.AzureInstallStep, len(a.Steps))
	for i, s := range a.Steps {
		step := &types.AzureInstallStep{
			Title: s.Title,
			Body:  s.Body,
		}
		if s.Code != "" {
			c := s.Code
			step.Code = &c
		}
		steps[i] = step
	}
	return &types.AzureInstallAssets{
		Steps:                    steps,
		RequiredRbacRoles:        append([]string{}, a.RequiredRBACRoles...),
		RequiredGraphPermissions: append([]string{}, a.RequiredGraphPermissions...),
	}
}
