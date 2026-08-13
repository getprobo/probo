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

package bot

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestEnqueueMessagePersistsClonedAttributesInCallerTransaction(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := botTestDatabase(t)
	attributes := map[string]any{"resource_id": "original"}

	var message *coredata.BotMessage

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				var err error

				message, err = NewService(ServiceConfig{}).EnqueueMessage(
					ctx,
					tx,
					scope,
					MessageParams{
						OrganizationID:   organizationID,
						Capability:       "compliance_access",
						MessageType:      "ACCESS_REQUEST",
						Attributes:       attributes,
						SubjectNamespace: "compliance_portal_access",
						SubjectKey:       "access-1",
						EventKey:         "created",
						Purpose:          coredata.BotMessagePurposePost,
					},
				)
				attributes["resource_id"] = "mutated"

				return err
			},
		),
	)

	var loaded coredata.BotMessage

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return loaded.LoadByID(ctx, conn, scope, message.ID)
			},
		),
	)

	var persisted map[string]any
	require.NoError(t, json.Unmarshal(loaded.Attributes, &persisted))
	assert.Equal(t, "original", persisted["resource_id"])
	assert.Equal(t, coredata.BotMessageEntityType, message.ID.EntityType())
	assert.Equal(t, coredata.BotMessagePurposePost, loaded.Purpose)
}

func TestEnqueueMessageRollsBackWithCallerTransaction(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := botTestDatabase(t)

	var message *coredata.BotMessage

	rollbackErr := errors.New("roll back")
	err := pgClient.WithTx(
		t.Context(),
		func(ctx context.Context, tx pg.Tx) error {
			var err error

			message, err = NewService(ServiceConfig{}).EnqueueMessage(
				ctx,
				tx,
				scope,
				MessageParams{
					OrganizationID:   organizationID,
					Capability:       "test",
					MessageType:      "test",
					SubjectNamespace: "test",
					SubjectKey:       "test",
					EventKey:         "created",
					Purpose:          coredata.BotMessagePurposePost,
				},
			)
			require.NoError(t, err)

			return rollbackErr
		},
	)
	require.ErrorIs(t, err, rollbackErr)

	var loaded coredata.BotMessage

	err = pgClient.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			return loaded.LoadByID(ctx, conn, scope, message.ID)
		},
	)
	assert.ErrorIs(t, err, coredata.ErrResourceNotFound)
}

func TestEnqueueMessageQueuesDistinctIdempotentEvents(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := botTestDatabase(t)
	service := NewService(ServiceConfig{})
	params := MessageParams{
		OrganizationID:   organizationID,
		Capability:       "compliance_access",
		MessageType:      "ACCESS_REQUEST",
		Attributes:       map[string]any{"resource_id": "access-1"},
		SubjectNamespace: "compliance_portal_access",
		SubjectKey:       "access-1",
		EventKey:         "created",
		Purpose:          coredata.BotMessagePurposePost,
	}

	var first, initialReplay, refresh, refreshReplay, secondRefresh *coredata.BotMessage

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				var err error

				first, err = service.EnqueueMessage(ctx, tx, scope, params)
				if err != nil {
					return err
				}

				initialReplay, err = service.EnqueueMessage(ctx, tx, scope, params)
				if err != nil {
					return err
				}

				params.EventKey = "grant:operation-1"
				params.Purpose = coredata.BotMessagePurposeUpdate

				refresh, err = service.EnqueueMessage(ctx, tx, scope, params)
				if err != nil {
					return err
				}

				refreshReplay, err = service.EnqueueMessage(ctx, tx, scope, params)
				if err != nil {
					return err
				}

				params.EventKey = "reject:operation-2"
				secondRefresh, err = service.EnqueueMessage(ctx, tx, scope, params)

				return err
			},
		),
	)
	assert.Equal(t, first.ID, initialReplay.ID)
	assert.Equal(t, refresh.ID, refreshReplay.ID)
	assert.NotEqual(t, first.ID, refresh.ID)
	assert.NotEqual(t, refresh.ID, secondRefresh.ID)
	assert.Equal(t, coredata.BotMessagePurposePost, first.Purpose)
	assert.Equal(t, coredata.BotMessagePurposeUpdate, refresh.Purpose)

	var count int

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return conn.QueryRow(
					ctx,
					`SELECT count(*) FROM bot_messages WHERE organization_id = $1`,
					organizationID,
				).Scan(&count)
			},
		),
	)
	assert.Equal(t, 3, count)
}

func TestEnqueueMessageDoesNotQueueWhenDisabled(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := botTestDatabase(t)
	service := NewService(ServiceConfig{Disabled: true})

	var message *coredata.BotMessage

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				var err error

				message, err = service.EnqueueMessage(
					ctx,
					tx,
					scope,
					MessageParams{
						OrganizationID:   organizationID,
						Capability:       "compliance_access",
						MessageType:      "ACCESS_REQUEST",
						SubjectNamespace: "compliance_portal_access",
						SubjectKey:       "access-1",
						EventKey:         "created",
						Purpose:          coredata.BotMessagePurposePost,
					},
				)

				return err
			},
		),
	)
	assert.Nil(t, message)

	var count int

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return conn.QueryRow(
					ctx,
					`SELECT count(*) FROM bot_messages WHERE organization_id = $1`,
					organizationID,
				).Scan(&count)
			},
		),
	)
	assert.Zero(t, count)
}

func TestStableEventKeyNormalizesComponentOrder(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		StableEventKey("update", "document:2:GRANTED", "document:1:REJECTED"),
		StableEventKey("update", "document:1:REJECTED", "document:2:GRANTED"),
	)
	assert.NotEqual(
		t,
		StableEventKey("grant", "document:1"),
		StableEventKey("reject", "document:1"),
	)
}

func botTestDatabase(
	t *testing.T,
) (*pg.Client, coredata.Scoper, gid.GID) {
	t.Helper()

	pgClient := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(
					ctx,
					`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
					organizationID,
					tenantID,
					"bot-test-"+organizationID.String(),
					now,
					now,
				)

				return err
			},
		),
	)
	t.Cleanup(
		func() {
			_ = pgClient.WithConn(
				context.Background(),
				func(ctx context.Context, conn pg.Querier) error {
					_, err := conn.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID)
					return err
				},
			)
		},
	)

	return pgClient, scope, organizationID
}
