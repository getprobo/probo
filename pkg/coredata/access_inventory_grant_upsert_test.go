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

package coredata_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type accessInventoryFixture struct {
	scope          *coredata.Scope
	organizationID gid.GID
	sourceID       gid.GID
}

func seedAccessInventoryFixture(t *testing.T, ctx context.Context, client *pg.Client) accessInventoryFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	sourceID := gid.New(tenantID, coredata.AccessReviewSourceEntityType)
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "Access Inventory Test Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		source := &coredata.AccessReviewSource{
			ID:             sourceID,
			OrganizationID: organizationID,
			Name:           "Access Inventory Test Source",
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if _, err := source.Insert(ctx, tx, scope); err != nil {
			return err
		}

		return nil
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			if _, err := tx.Exec(ctx, `DELETE FROM access_inventory_grants WHERE access_review_source_id = $1`, sourceID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM access_inventory_principals WHERE access_review_source_id = $1`, sourceID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM access_inventory_resources WHERE access_review_source_id = $1`, sourceID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM access_review_sources WHERE id = $1`, sourceID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID); err != nil {
				return err
			}

			return nil
		})
	})

	return accessInventoryFixture{
		scope:          scope,
		organizationID: organizationID,
		sourceID:       sourceID,
	}
}

func TestAccessInventoryGrantUpsertPreservesFirstSeenAt(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedAccessInventoryFixture(t, ctx, client)
	tenantID := fx.scope.GetTenantID()
	now := time.Now().UTC().Truncate(time.Millisecond)
	firstSeen := now.Add(-time.Hour)

	principalID := gid.New(tenantID, coredata.AccessInventoryPrincipalEntityType)
	resourceID := gid.New(tenantID, coredata.AccessInventoryResourceEntityType)
	grantID := gid.New(tenantID, coredata.AccessInventoryGrantEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		principal := &coredata.AccessInventoryPrincipal{
			ID:                   principalID,
			OrganizationID:       fx.organizationID,
			AccessReviewSourceID: fx.sourceID,
			PrincipalKey:         "jane@example.com|ext-1",
			Email:                "jane@example.com",
			FullName:             "Jane Doe",
			ExternalID:           "ext-1",
			JobTitle:             "Engineer",
			MFAStatus:            coredata.MFAStatusUnknown,
			AuthMethod:           coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType:          coredata.AccessReviewEntryAccountTypeUser,
			FirstSeenAt:          firstSeen,
			LastSeenAt:           firstSeen,
			CreatedAt:            firstSeen,
			UpdatedAt:            firstSeen,
		}
		inserted, err := principal.Upsert(ctx, tx, fx.scope)
		if err != nil {
			return err
		}
		require.True(t, inserted)
		principalID = principal.ID

		resource := &coredata.AccessInventoryResource{
			ID:                   resourceID,
			OrganizationID:       fx.organizationID,
			AccessReviewSourceID: fx.sourceID,
			ResourceKey:          "system",
			DisplayName:          "Access Inventory Test Source",
			Kind:                 "system",
			FirstSeenAt:          firstSeen,
			LastSeenAt:           firstSeen,
			CreatedAt:            firstSeen,
			UpdatedAt:            firstSeen,
		}
		inserted, err = resource.Upsert(ctx, tx, fx.scope)
		if err != nil {
			return err
		}
		require.True(t, inserted)
		resourceID = resource.ID

		grant := &coredata.AccessInventoryGrant{
			ID:                      grantID,
			OrganizationID:          fx.organizationID,
			AccessReviewSourceID:    fx.sourceID,
			PrincipalID:             principalID,
			ResourceID:              resourceID,
			PrincipalKey:            "jane@example.com|ext-1",
			ResourceKey:             "system",
			EntitlementKey:          "admin",
			EntitlementDisplayName:  "Admin",
			Presence:                coredata.AccessInventoryGrantPresencePresent,
			ConsecutiveMissingSyncs: 0,
			FirstSeenAt:             firstSeen,
			LastSeenAt:              firstSeen,
			CreatedAt:               firstSeen,
			UpdatedAt:               firstSeen,
		}
		inserted, err = grant.Upsert(ctx, tx, fx.scope)
		if err != nil {
			return err
		}
		require.True(t, inserted)
		grantID = grant.ID

		return nil
	}))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		later := now
		grant := &coredata.AccessInventoryGrant{
			ID:                      gid.New(tenantID, coredata.AccessInventoryGrantEntityType),
			OrganizationID:          fx.organizationID,
			AccessReviewSourceID:    fx.sourceID,
			PrincipalID:             principalID,
			ResourceID:              resourceID,
			PrincipalKey:            "jane@example.com|ext-1",
			ResourceKey:             "system",
			EntitlementKey:          "admin",
			EntitlementDisplayName:  "Administrator",
			Presence:                coredata.AccessInventoryGrantPresencePresent,
			ConsecutiveMissingSyncs: 0,
			FirstSeenAt:             later,
			LastSeenAt:              later,
			CreatedAt:               later,
			UpdatedAt:               later,
		}
		inserted, err := grant.Upsert(ctx, tx, fx.scope)
		if err != nil {
			return err
		}
		require.False(t, inserted)
		assert.Equal(t, grantID, grant.ID)
		assert.Equal(t, "Administrator", grant.EntitlementDisplayName)
		assert.WithinDuration(t, firstSeen, grant.FirstSeenAt, time.Millisecond)
		assert.WithinDuration(t, later, grant.LastSeenAt, time.Millisecond)

		loaded := &coredata.AccessInventoryGrant{}
		if err := loaded.LoadByNaturalKey(
			ctx,
			tx,
			fx.scope,
			fx.sourceID,
			"jane@example.com|ext-1",
			"system",
			"admin",
		); err != nil {
			return err
		}

		assert.Equal(t, grantID, loaded.ID)
		assert.Equal(t, coredata.AccessInventoryGrantPresencePresent, loaded.Presence)
		assert.Equal(t, 0, loaded.ConsecutiveMissingSyncs)
		assert.WithinDuration(t, firstSeen, loaded.FirstSeenAt, time.Millisecond)

		return nil
	}))
}

func TestAccessInventoryGrantRejectsCrossSourcePrincipal(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	fxA := seedAccessInventoryFixture(t, ctx, client)
	fxB := seedAccessInventoryFixture(t, ctx, client)
	tenantID := fxA.scope.GetTenantID()
	now := time.Now().UTC()

	var principalFromA gid.GID
	var resourceFromB gid.GID

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		principal := &coredata.AccessInventoryPrincipal{
			ID:                   gid.New(tenantID, coredata.AccessInventoryPrincipalEntityType),
			OrganizationID:       fxA.organizationID,
			AccessReviewSourceID: fxA.sourceID,
			PrincipalKey:         "cross-source@example.com",
			Email:                "cross-source@example.com",
			FullName:             "Cross Source",
			ExternalID:           "",
			JobTitle:             "",
			MFAStatus:            coredata.MFAStatusUnknown,
			AuthMethod:           coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType:          coredata.AccessReviewEntryAccountTypeUser,
			FirstSeenAt:          now,
			LastSeenAt:           now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if _, err := principal.Upsert(ctx, tx, fxA.scope); err != nil {
			return err
		}
		principalFromA = principal.ID

		resource := &coredata.AccessInventoryResource{
			ID:                   gid.New(tenantID, coredata.AccessInventoryResourceEntityType),
			OrganizationID:       fxB.organizationID,
			AccessReviewSourceID: fxB.sourceID,
			ResourceKey:          "system",
			DisplayName:          "Source B",
			Kind:                 "system",
			FirstSeenAt:          now,
			LastSeenAt:           now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if _, err := resource.Upsert(ctx, tx, fxB.scope); err != nil {
			return err
		}
		resourceFromB = resource.ID

		return nil
	}))

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		grant := &coredata.AccessInventoryGrant{
			ID:                      gid.New(tenantID, coredata.AccessInventoryGrantEntityType),
			OrganizationID:          fxB.organizationID,
			AccessReviewSourceID:    fxB.sourceID,
			PrincipalID:             principalFromA,
			ResourceID:              resourceFromB,
			PrincipalKey:            "cross-source@example.com",
			ResourceKey:             "system",
			EntitlementKey:          "member",
			EntitlementDisplayName:  "Member",
			Presence:                coredata.AccessInventoryGrantPresencePresent,
			ConsecutiveMissingSyncs: 0,
			FirstSeenAt:             now,
			LastSeenAt:              now,
			CreatedAt:               now,
			UpdatedAt:               now,
		}
		_, err := grant.Upsert(ctx, tx, fxB.scope)

		return err
	})
	require.Error(t, err)

	var pgErr *pgconn.PgError
	require.True(t, errors.As(err, &pgErr))
	assert.Equal(t, "23503", pgErr.Code)
}
