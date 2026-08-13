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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestResetStaleSlackDeliveryOperations_RequeuesWhenAttemptsRemain(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC()
	operation := insertStaleSlackDeliveryOperation(t, pgClient, now, 1)

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return coredata.ResetStaleSlackDeliveryOperations(
					ctx,
					conn,
					now,
					10*time.Minute,
				)
			},
		),
	)

	persisted := loadSlackDeliveryOperation(t, pgClient, operation)
	assert.Nil(t, persisted.ProcessingStartedAt)
	require.NotNil(t, persisted.NextAttemptAt)
	assert.WithinDuration(t, now, *persisted.NextAttemptAt, time.Second)
	assert.Nil(t, persisted.DeadLetteredAt)
	require.NotNil(t, persisted.LastError)
	assert.Equal(t, "Slack delivery operation lease expired", *persisted.LastError)
}

func TestResetStaleSlackDeliveryOperations_DeadLettersWhenAttemptsExhausted(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC()
	operation := insertStaleSlackDeliveryOperation(
		t,
		pgClient,
		now,
		coredata.SlackDeliveryOperationDefaultMaxAttempts,
	)

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return coredata.ResetStaleSlackDeliveryOperations(
					ctx,
					conn,
					now,
					10*time.Minute,
				)
			},
		),
	)

	persisted := loadSlackDeliveryOperation(t, pgClient, operation)
	assert.Nil(t, persisted.ProcessingStartedAt)
	assert.Nil(t, persisted.NextAttemptAt)
	require.NotNil(t, persisted.DeadLetteredAt)
	assert.WithinDuration(t, now, *persisted.DeadLetteredAt, time.Second)
	require.NotNil(t, persisted.LastError)
	assert.Equal(t, "Slack delivery operation lease expired", *persisted.LastError)
}

func insertStaleSlackDeliveryOperation(
	t *testing.T,
	pgClient *pg.Client,
	now time.Time,
	attemptCount int,
) *coredata.SlackDeliveryOperation {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organization := coredata.Organization{
		ID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TenantID:  tenantID,
		Name:      "Slack delivery operation stale reset test",
		CreatedAt: now,
		UpdatedAt: now,
	}
	startedAt := now.Add(-time.Hour)
	operation := coredata.NewSlackDeliveryOperation(
		scope,
		organization.ID,
		"stale-"+organization.ID.String(),
		coredata.SlackDeliveryOperationKindAddReaction,
		map[string]any{"channel": "C123"},
	)
	operation.ProcessingStartedAt = &startedAt
	operation.AttemptCount = attemptCount

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				if err := organization.Insert(ctx, tx); err != nil {
					return err
				}

				_, err := operation.Upsert(ctx, tx, scope)

				return err
			},
		),
	)
	t.Cleanup(
		func() {
			_ = pgClient.WithTx(
				context.Background(),
				func(ctx context.Context, tx pg.Tx) error {
					return organization.Delete(ctx, tx, organization.ID)
				},
			)
		},
	)

	return operation
}

func loadSlackDeliveryOperation(
	t *testing.T,
	pgClient *pg.Client,
	operation *coredata.SlackDeliveryOperation,
) *coredata.SlackDeliveryOperation {
	t.Helper()

	scope := coredata.NewScope(operation.ID.TenantID())
	var persisted coredata.SlackDeliveryOperation

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return persisted.LoadByID(ctx, conn, scope, operation.ID)
			},
		),
	)

	return &persisted
}
