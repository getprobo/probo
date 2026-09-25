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
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func insertLinearWebhookEventFixture(
	t *testing.T,
	pgClient *pg.Client,
	now time.Time,
	ownerToken *string,
	startedAt *time.Time,
) *coredata.LinearWebhookEvent {
	t.Helper()

	event := coredata.NewLinearWebhookEvent(
		"lease-"+gid.New(gid.NewTenantID(), coredata.TaskSyncJobEntityType).String(),
		[]byte(`{}`),
	)
	event.CreatedAt = now
	event.UpdatedAt = now
	event.ProcessingOwnerToken = ownerToken
	event.ProcessingStartedAt = startedAt

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				_, err := event.Insert(ctx, conn)

				return err
			},
		),
	)
	t.Cleanup(
		func() {
			_ = pgClient.WithConn(
				context.Background(),
				func(ctx context.Context, conn pg.Querier) error {
					_, err := conn.Exec(
						ctx,
						`DELETE FROM linear_webhook_events WHERE delivery_id = @delivery_id`,
						pgx.StrictNamedArgs{"delivery_id": event.DeliveryID},
					)

					return err
				},
			)
		},
	)

	return event
}

func loadLinearWebhookEvent(
	t *testing.T,
	pgClient *pg.Client,
	deliveryID string,
) *coredata.LinearWebhookEvent {
	t.Helper()

	var event coredata.LinearWebhookEvent

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return event.LoadByDeliveryID(ctx, conn, deliveryID)
			},
		),
	)

	return &event
}

func TestLinearWebhookEvent_TouchLease_RenewsStartedAt(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC()
	startedAt := now.Add(-time.Minute)
	event := insertLinearWebhookEventFixture(
		t,
		pgClient,
		now,
		new("owner"),
		&startedAt,
	)

	touchedAt := now.Add(time.Minute)

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return event.TouchLease(ctx, conn, touchedAt)
			},
		),
	)

	persisted := loadLinearWebhookEvent(t, pgClient, event.DeliveryID)
	require.NotNil(t, persisted.ProcessingOwnerToken)
	assert.Equal(t, "owner", *persisted.ProcessingOwnerToken)
	require.NotNil(t, persisted.ProcessingStartedAt)
	assert.WithinDuration(t, touchedAt, *persisted.ProcessingStartedAt, time.Second)
}

func TestLinearWebhookEvent_TouchLease_WrongToken(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC()
	startedAt := now.Add(-time.Minute)
	event := insertLinearWebhookEventFixture(
		t,
		pgClient,
		now,
		new("owner-a"),
		&startedAt,
	)

	event.ProcessingOwnerToken = new("owner-b")

	err := pgClient.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			return event.TouchLease(ctx, conn, now)
		},
	)
	require.ErrorIs(t, err, coredata.ErrProcessingLeaseLost)

	persisted := loadLinearWebhookEvent(t, pgClient, event.DeliveryID)
	require.NotNil(t, persisted.ProcessingOwnerToken)
	assert.Equal(t, "owner-a", *persisted.ProcessingOwnerToken)
	require.NotNil(t, persisted.ProcessingStartedAt)
	assert.WithinDuration(t, startedAt, *persisted.ProcessingStartedAt, time.Microsecond)
}
