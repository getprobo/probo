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
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func seedAccessInventoryGrant(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx accessInventoryFixture,
) gid.GID {
	t.Helper()

	tenantID := fx.scope.GetTenantID()
	now := time.Now().UTC()

	var grantID gid.GID

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		principal := &coredata.AccessInventoryPrincipal{
			ID:                   gid.New(tenantID, coredata.AccessInventoryPrincipalEntityType),
			OrganizationID:       fx.organizationID,
			AccessReviewSourceID: fx.sourceID,
			PrincipalKey:         "change-event@example.com",
			Email:                "change-event@example.com",
			FullName:             "Change Event",
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

		inserted, err := principal.Upsert(ctx, tx, fx.scope)
		if err != nil {
			return err
		}

		require.True(t, inserted)

		resource := &coredata.AccessInventoryResource{
			ID:                   gid.New(tenantID, coredata.AccessInventoryResourceEntityType),
			OrganizationID:       fx.organizationID,
			AccessReviewSourceID: fx.sourceID,
			ResourceKey:          "system",
			DisplayName:          "Change Event Source",
			Kind:                 "system",
			FirstSeenAt:          now,
			LastSeenAt:           now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		inserted, err = resource.Upsert(ctx, tx, fx.scope)
		if err != nil {
			return err
		}

		require.True(t, inserted)

		grant := &coredata.AccessInventoryGrant{
			ID:                      gid.New(tenantID, coredata.AccessInventoryGrantEntityType),
			OrganizationID:          fx.organizationID,
			AccessReviewSourceID:    fx.sourceID,
			PrincipalID:             principal.ID,
			ResourceID:              resource.ID,
			PrincipalKey:            "change-event@example.com",
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

		inserted, err = grant.Upsert(ctx, tx, fx.scope)
		if err != nil {
			return err
		}

		require.True(t, inserted)

		grantID = grant.ID

		return nil
	}))

	return grantID
}

func TestAccessInventoryChangeEvent_Insert_Idempotent(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedAccessInventoryFixture(t, ctx, client)
	grantID := seedAccessInventoryGrant(t, ctx, client, fx)
	tenantID := fx.scope.GetTenantID()
	now := time.Now().UTC().Truncate(time.Millisecond)
	afterPayload := json.RawMessage(`{"presence":"PRESENT"}`)
	idempotencyKey := "sync-1|change-event@example.com|system|member|APPEARED|appeared"

	eventID := gid.New(tenantID, coredata.AccessInventoryChangeEventEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		event := &coredata.AccessInventoryChangeEvent{
			ID:                   eventID,
			OrganizationID:       fx.organizationID,
			AccessReviewSourceID: fx.sourceID,
			GrantID:              &grantID,
			PrincipalKey:         "change-event@example.com",
			ResourceKey:          "system",
			EntitlementKey:       "member",
			ChangeType:           coredata.AccessInventoryChangeTypeAppeared,
			ObservedAt:           now,
			After:                &afterPayload,
			IdempotencyKey:       idempotencyKey,
			CreatedAt:            now,
		}

		inserted, err := event.Insert(ctx, tx, fx.scope)
		if err != nil {
			return err
		}

		require.True(t, inserted)

		duplicate := &coredata.AccessInventoryChangeEvent{
			ID:                   gid.New(tenantID, coredata.AccessInventoryChangeEventEntityType),
			OrganizationID:       fx.organizationID,
			AccessReviewSourceID: fx.sourceID,
			GrantID:              &grantID,
			PrincipalKey:         "change-event@example.com",
			ResourceKey:          "system",
			EntitlementKey:       "member",
			ChangeType:           coredata.AccessInventoryChangeTypeAppeared,
			ObservedAt:           now.Add(time.Minute),
			After:                &afterPayload,
			IdempotencyKey:       idempotencyKey,
			CreatedAt:            now.Add(time.Minute),
		}

		inserted, err = duplicate.Insert(ctx, tx, fx.scope)
		if err != nil {
			return err
		}

		require.False(t, inserted)

		loaded := &coredata.AccessInventoryChangeEvent{}
		if err := loaded.LoadByID(ctx, tx, fx.scope, eventID); err != nil {
			return err
		}

		assert.Equal(t, coredata.AccessInventoryChangeTypeAppeared, loaded.ChangeType)
		assert.Equal(t, idempotencyKey, loaded.IdempotencyKey)
		assert.Equal(t, grantID, *loaded.GrantID)
		require.NotNil(t, loaded.After)
		assert.JSONEq(t, string(afterPayload), string(*loaded.After))
		assert.Nil(t, loaded.Before)
		assert.Nil(t, loaded.SyncRunID)
		assert.WithinDuration(t, now, loaded.ObservedAt, time.Millisecond)

		return nil
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		cursor := page.NewCursor(
			10,
			nil,
			page.Head,
			page.OrderBy[coredata.AccessInventoryChangeEventOrderField]{
				Field:     coredata.AccessInventoryChangeEventOrderFieldObservedAt,
				Direction: page.OrderDirectionAsc,
			},
		)

		var byGrant coredata.AccessInventoryChangeEvents
		if err := byGrant.LoadByGrantID(ctx, conn, fx.scope, grantID, cursor); err != nil {
			return err
		}

		require.Len(t, byGrant, 1)
		assert.Equal(t, eventID, byGrant[0].ID)

		var bySource coredata.AccessInventoryChangeEvents
		if err := bySource.LoadBySourceID(ctx, conn, fx.scope, fx.sourceID, cursor); err != nil {
			return err
		}

		require.Len(t, bySource, 1)
		assert.Equal(t, eventID, bySource[0].ID)

		return nil
	}))
}
