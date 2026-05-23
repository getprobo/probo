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

package iam_test

import (
	"fmt"
	"testing"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
)

func TestAuthorizeRelatedErrors_Error(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	identityID := gid.New(gid.NilTenant, coredata.IdentityEntityType)
	resourceID := gid.New(tenantID, coredata.FrameworkEntityType)
	membershipID := gid.New(tenantID, coredata.MembershipEntityType)
	sessionID := gid.New(gid.NilTenant, coredata.SessionEntityType)
	action := iam.Action("core:framework:get")
	mixedOrgIDs := []string{
		gid.New(tenantID, coredata.OrganizationEntityType).String(),
		gid.New(tenantID, coredata.OrganizationEntityType).String(),
	}
	mixedEntityTypes := []uint16{coredata.OrganizationEntityType, coredata.FrameworkEntityType}

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "insufficient permissions",
			err:  iam.NewInsufficientPermissionsError(identityID, resourceID, action),
			want: fmt.Sprintf(
				"identity %q does not have sufficient permissions to perform action %s on entity %q",
				identityID,
				action,
				resourceID,
			),
		},
		{
			name: "unsupported principal type",
			err:  iam.NewUnsupportedPrincipalTypeError(coredata.OrganizationEntityType),
			want: fmt.Sprintf("unsupported principal type: %d", coredata.OrganizationEntityType),
		},
		{
			name: "empty resource batch",
			err:  iam.NewEmptyResourceBatchError(action),
			want: fmt.Sprintf("cannot authorize batch action %s with an empty resource set", action),
		},
		{
			name: "mixed entity type batch",
			err:  iam.NewMixedEntityTypeBatchError(action, mixedEntityTypes),
			want: fmt.Sprintf("cannot authorize batch action %s across entity types %v", action, mixedEntityTypes),
		},
		{
			name: "mixed organization batch",
			err:  iam.NewMixedOrganizationBatchError(action, mixedOrgIDs),
			want: fmt.Sprintf("cannot authorize batch action %s across organization ids %q", action, mixedOrgIDs),
		},
		{
			name: "batch unsupported resource type",
			err:  iam.NewBatchAuthorizationUnsupportedResourceTypeError(coredata.OAuth2AccessTokenEntityType),
			want: fmt.Sprintf(
				"resource type %d does not support batch authorization attributes",
				coredata.OAuth2AccessTokenEntityType,
			),
		},
		{
			name: "assumption required",
			err:  iam.NewAssumptionRequiredError(identityID, membershipID),
			want: fmt.Sprintf("assumption for identity %q required for membership %q", identityID, membershipID),
		},
		{
			name: "session not found with nil id",
			err:  iam.NewSessionNotFoundError(gid.Nil),
			want: "session not found",
		},
		{
			name: "session not found with specific id",
			err:  iam.NewSessionNotFoundError(sessionID),
			want: fmt.Sprintf("session %q not found", sessionID),
		},
		{
			name: "session expired",
			err:  iam.NewSessionExpiredError(sessionID),
			want: fmt.Sprintf("session %q expired", sessionID),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.err.Error() != tt.want {
				t.Errorf("Error() = %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}
