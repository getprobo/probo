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
