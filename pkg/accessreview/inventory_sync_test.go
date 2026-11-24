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
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type inventorySyncFixture struct {
	scope          *coredata.Scope
	organizationID gid.GID
	sourceID       gid.GID
}

func seedInventorySyncFixture(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	csvData string,
) inventorySyncFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	fx := inventorySyncFixture{
		scope:          coredata.NewScope(tenantID, gid.Nil),
		organizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		sourceID:       gid.New(tenantID, coredata.AccessReviewSourceEntityType),
	}
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        fx.organizationID,
			TenantID:  tenantID,
			Name:      "Inventory Sync",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		source := &coredata.AccessReviewSource{
			ID:             fx.sourceID,
			OrganizationID: fx.organizationID,
			Name:           "CSV Roster",
			CsvData:        &csvData,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		_, err := source.Insert(ctx, tx, fx.scope)

		return err
	}))

	return fx
}

func insertQueuedInventorySyncRun(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx inventorySyncFixture,
) *coredata.AccessSourceSyncRun {
	t.Helper()

	now := time.Now().UTC()
	run := &coredata.AccessSourceSyncRun{
		ID:                   gid.New(fx.scope.GetTenantID(), coredata.AccessSourceSyncRunEntityType),
		OrganizationID:       fx.organizationID,
		AccessReviewSourceID: fx.sourceID,
		Status:               coredata.AccessSourceSyncRunStatusQueued,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return run.Insert(ctx, tx, fx.scope)
	}))

	return run
}

func loadInventoryGrants(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx inventorySyncFixture,
) []*coredata.AccessInventoryGrant {
	t.Helper()

	var grants coredata.AccessInventoryGrants

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		cursor := page.NewCursor(
			100,
			nil,
			page.Head,
			page.OrderBy[coredata.AccessInventoryGrantOrderField]{
				Field:     coredata.AccessInventoryGrantOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		)

		return grants.LoadBySourceID(ctx, conn, fx.scope, fx.sourceID, cursor)
	}))

	return grants
}

func loadInventoryChangeEvents(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx inventorySyncFixture,
) []*coredata.AccessInventoryChangeEvent {
	t.Helper()

	var events coredata.AccessInventoryChangeEvents

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		cursor := page.NewCursor(
			100,
			nil,
			page.Head,
			page.OrderBy[coredata.AccessInventoryChangeEventOrderField]{
				Field:     coredata.AccessInventoryChangeEventOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		)

		return events.LoadBySourceID(ctx, conn, fx.scope, fx.sourceID, cursor)
	}))

	return events
}

func loadSyncRun(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx inventorySyncFixture,
	runID gid.GID,
) *coredata.AccessSourceSyncRun {
	t.Helper()

	run := &coredata.AccessSourceSyncRun{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return run.LoadByID(ctx, conn, fx.scope, runID)
	}))

	return run
}

func markSyncRunRunning(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx inventorySyncFixture,
	run *coredata.AccessSourceSyncRun,
) {
	t.Helper()

	now := time.Now().UTC()
	run.Status = coredata.AccessSourceSyncRunStatusRunning
	run.StartedAt = &now
	run.UpdatedAt = now

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return run.Update(ctx, tx, fx.scope)
	}))
}

func TestSyncInventorySource_AppearedOnNewEntitlements(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	csvV1 := "email,full_name,role\nalice@example.com,Alice,Admin\n"
	fx := seedInventorySyncFixture(t, ctx, client, csvV1)
	run1 := insertQueuedInventorySyncRun(t, ctx, client, fx)
	markSyncRunRunning(t, ctx, client, fx, run1)

	require.NoError(t, svc.SyncInventorySource(ctx, fx.scope, run1))

	loadedRun1 := loadSyncRun(t, ctx, client, fx, run1.ID)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusSuccess, loadedRun1.Status)
	assert.Equal(t, 1, loadedRun1.GrantsUpsertedCount)

	grants := loadInventoryGrants(t, ctx, client, fx)
	require.Len(t, grants, 1)
	assert.Equal(t, "alice@example.com", grants[0].PrincipalKey)
	assert.Equal(t, DefaultResourceKey, grants[0].ResourceKey)
	assert.Equal(t, "Admin", grants[0].EntitlementKey)
	assert.Equal(t, "Admin", grants[0].EntitlementDisplayName)
	assert.Equal(t, coredata.AccessInventoryGrantPresencePresent, grants[0].Presence)

	events := loadInventoryChangeEvents(t, ctx, client, fx)
	require.Len(t, events, 1)
	assert.Equal(t, coredata.AccessInventoryChangeTypeAppeared, events[0].ChangeType)
	assert.Equal(t, run1.ID, *events[0].SyncRunID)

	csvV2 := "email,full_name,role\nalice@example.com,Alice,Owner\n"

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		source := &coredata.AccessReviewSource{}
		if err := source.LoadByID(ctx, tx, fx.scope, fx.sourceID); err != nil {
			return err
		}

		source.CsvData = &csvV2
		source.UpdatedAt = time.Now().UTC()

		return source.Update(ctx, tx, fx.scope)
	}))

	run2 := insertQueuedInventorySyncRun(t, ctx, client, fx)
	markSyncRunRunning(t, ctx, client, fx, run2)
	require.NoError(t, svc.SyncInventorySource(ctx, fx.scope, run2))

	grants = loadInventoryGrants(t, ctx, client, fx)
	require.Len(t, grants, 2, "Owner is a new entitlement key; Admin remains until soft-disappear")

	events = loadInventoryChangeEvents(t, ctx, client, fx)
	require.Len(t, events, 2)
	assert.Equal(t, coredata.AccessInventoryChangeTypeAppeared, events[1].ChangeType)
	assert.Equal(t, "Owner", events[1].EntitlementKey)
	require.NotNil(t, events[1].SyncRunID)
	assert.Equal(t, run2.ID, *events[1].SyncRunID)
}

func TestSyncInventorySource_EmptyRosterUpsertsDefaultResource(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedInventorySyncFixture(t, ctx, client, "email,full_name,role\n")
	run := insertQueuedInventorySyncRun(t, ctx, client, fx)
	markSyncRunRunning(t, ctx, client, fx, run)
	require.NoError(t, svc.SyncInventorySource(ctx, fx.scope, run))

	resource := &coredata.AccessInventoryResource{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return resource.LoadBySourceIDAndResourceKey(ctx, conn, fx.scope, fx.sourceID, DefaultResourceKey)
	}))
	assert.Equal(t, DefaultResourceKind, resource.Kind)
	assert.Equal(t, "CSV Roster", resource.DisplayName)
	assert.Empty(t, loadInventoryGrants(t, ctx, client, fx))
}

func TestInventoryChangeIdempotencyKey_Unambiguous(t *testing.T) {
	t.Parallel()

	a := inventoryChangeIdempotencyKey("run", "a|b", "c", "d", "APPEARED", "appeared")
	b := inventoryChangeIdempotencyKey("run", "a", "b|c", "d", "APPEARED", "appeared")
	assert.NotEqual(t, a, b)
}

func TestSyncInventorySource_IdempotentOnUnchangedRoster(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	csvData := "email,full_name,role\nbob@example.com,Bob,Member\n"
	fx := seedInventorySyncFixture(t, ctx, client, csvData)

	run1 := insertQueuedInventorySyncRun(t, ctx, client, fx)
	markSyncRunRunning(t, ctx, client, fx, run1)
	require.NoError(t, svc.SyncInventorySource(ctx, fx.scope, run1))

	firstSeen := loadInventoryGrants(t, ctx, client, fx)[0].FirstSeenAt
	require.Len(t, loadInventoryChangeEvents(t, ctx, client, fx), 1)

	time.Sleep(2 * time.Millisecond)

	run2 := insertQueuedInventorySyncRun(t, ctx, client, fx)
	markSyncRunRunning(t, ctx, client, fx, run2)
	require.NoError(t, svc.SyncInventorySource(ctx, fx.scope, run2))

	grants := loadInventoryGrants(t, ctx, client, fx)
	require.Len(t, grants, 1)
	assert.Equal(t, firstSeen.UTC(), grants[0].FirstSeenAt.UTC())
	assert.True(t, grants[0].LastSeenAt.After(firstSeen))

	events := loadInventoryChangeEvents(t, ctx, client, fx)
	assert.Len(t, events, 1, "unchanged roster must not emit another change event")
}

func TestSyncInventorySource_AttributeChangedDisplayName(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	// Seed an existing grant whose entitlement key matches but display name differs
	// from what the next roster sync will write — ATTRIBUTE_CHANGED, not APPEARED.
	csvData := "email,full_name,role\ncarol@example.com,Carol,admin\n"
	fx := seedInventorySyncFixture(t, ctx, client, csvData)
	tenantID := fx.scope.GetTenantID()
	now := time.Now().UTC().Truncate(time.Millisecond)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		principal := &coredata.AccessInventoryPrincipal{
			ID:                   gid.New(tenantID, coredata.AccessInventoryPrincipalEntityType),
			OrganizationID:       fx.organizationID,
			AccessReviewSourceID: fx.sourceID,
			PrincipalKey:         "carol@example.com",
			Email:                "carol@example.com",
			FullName:             "Carol",
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
			ResourceKey:          DefaultResourceKey,
			DisplayName:          "CSV Roster",
			Kind:                 DefaultResourceKind,
			FirstSeenAt:          now,
			LastSeenAt:           now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if _, err := resource.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		grant := &coredata.AccessInventoryGrant{
			ID:                      gid.New(tenantID, coredata.AccessInventoryGrantEntityType),
			OrganizationID:          fx.organizationID,
			AccessReviewSourceID:    fx.sourceID,
			PrincipalID:             principal.ID,
			ResourceID:              resource.ID,
			PrincipalKey:            "carol@example.com",
			ResourceKey:             DefaultResourceKey,
			EntitlementKey:          "admin",
			EntitlementDisplayName:  "Admin (legacy)",
			Presence:                coredata.AccessInventoryGrantPresencePresent,
			ConsecutiveMissingSyncs: 0,
			FirstSeenAt:             now,
			LastSeenAt:              now,
			CreatedAt:               now,
			UpdatedAt:               now,
		}
		_, err := grant.Upsert(ctx, tx, fx.scope)

		return err
	}))

	run := insertQueuedInventorySyncRun(t, ctx, client, fx)
	markSyncRunRunning(t, ctx, client, fx, run)
	require.NoError(t, svc.SyncInventorySource(ctx, fx.scope, run))

	grants := loadInventoryGrants(t, ctx, client, fx)
	require.Len(t, grants, 1)
	assert.Equal(t, "admin", grants[0].EntitlementDisplayName)

	events := loadInventoryChangeEvents(t, ctx, client, fx)
	require.Len(t, events, 1)
	assert.Equal(t, coredata.AccessInventoryChangeTypeAttributeChanged, events[0].ChangeType)
	require.NotNil(t, events[0].Before)
	require.NotNil(t, events[0].After)
}
