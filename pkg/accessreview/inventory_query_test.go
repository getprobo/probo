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

package accessreview

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func TestListInventoryGrants_HidesPastThreshold(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)
	fx := seedInventorySyncFixture(t, ctx, client, "")

	now := time.Now().UTC()
	tenantID := fx.scope.GetTenantID()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		principal := &coredata.AccessInventoryPrincipal{
			ID:                   gid.New(tenantID, coredata.AccessInventoryPrincipalEntityType),
			OrganizationID:       fx.organizationID,
			AccessReviewSourceID: fx.sourceID,
			PrincipalKey:         "user@example.com",
			Email:                "user@example.com",
			FullName:             "User",
			ExternalID:           "ext-1",
			MFAStatus:            coredata.MFAStatusUnknown,
			AuthMethod:           coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType:          coredata.AccessReviewEntryAccountTypeUser,
			FirstSeenAt:          now,
			LastSeenAt:           now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if _, err := principal.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		resource := &coredata.AccessInventoryResource{
			ID:                   gid.New(tenantID, coredata.AccessInventoryResourceEntityType),
			OrganizationID:       fx.organizationID,
			AccessReviewSourceID: fx.sourceID,
			ResourceKey:          "system",
			DisplayName:          "System",
			Kind:                 "connected_system",
			FirstSeenAt:          now,
			LastSeenAt:           now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if _, err := resource.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		visible := &coredata.AccessInventoryGrant{
			ID:                      gid.New(tenantID, coredata.AccessInventoryGrantEntityType),
			OrganizationID:          fx.organizationID,
			AccessReviewSourceID:    fx.sourceID,
			PrincipalID:             principal.ID,
			ResourceID:              resource.ID,
			PrincipalKey:            principal.PrincipalKey,
			ResourceKey:             resource.ResourceKey,
			EntitlementKey:          "member",
			EntitlementDisplayName:  "Member",
			Presence:                coredata.AccessInventoryGrantPresenceDisappeared,
			ConsecutiveMissingSyncs: 1,
			FirstSeenAt:             now,
			LastSeenAt:              now,
			CreatedAt:               now,
			UpdatedAt:               now,
		}
		if _, err := visible.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		hidden := &coredata.AccessInventoryGrant{
			ID:                      gid.New(tenantID, coredata.AccessInventoryGrantEntityType),
			OrganizationID:          fx.organizationID,
			AccessReviewSourceID:    fx.sourceID,
			PrincipalID:             principal.ID,
			ResourceID:              resource.ID,
			PrincipalKey:            principal.PrincipalKey,
			ResourceKey:             resource.ResourceKey,
			EntitlementKey:          "admin",
			EntitlementDisplayName:  "Admin",
			Presence:                coredata.AccessInventoryGrantPresenceDisappeared,
			ConsecutiveMissingSyncs: coredata.DefaultHideAfterMissingSyncs,
			FirstSeenAt:             now,
			LastSeenAt:              now,
			CreatedAt:               now,
			UpdatedAt:               now,
		}
		if _, err := hidden.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		return nil
	}))

	cursor := page.NewCursor(
		10,
		nil,
		page.Head,
		page.OrderBy[coredata.AccessInventoryGrantOrderField]{
			Field:     coredata.AccessInventoryGrantOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
	)

	defaultPage, err := svc.ListInventoryGrantsForOrganizationID(
		ctx,
		fx.scope,
		fx.organizationID,
		cursor,
		nil,
	)
	require.NoError(t, err)
	require.Len(t, defaultPage.Data, 1)
	assert.Equal(t, "member", defaultPage.Data[0].EntitlementKey)

	include := true
	allPage, err := svc.ListInventoryGrantsForOrganizationID(
		ctx,
		fx.scope,
		fx.organizationID,
		cursor,
		&coredata.AccessInventoryGrantFilter{IncludeDisappeared: &include},
	)
	require.NoError(t, err)
	require.Len(t, allPage.Data, 2)
}
