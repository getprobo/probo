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

func TestWebhookEvent_ClaimAndCompleteDelivery(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC()
	event := insertWebhookDeliveryFixture(t, pgClient, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))

	var claimed coredata.WebhookEvent
	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return claimed.ClaimNextForUpdateSkipLocked(ctx, tx, now)
			},
		),
	)

	assert.Equal(t, event.ID, claimed.ID)
	assert.Equal(t, 1, claimed.AttemptCount)
	require.NotNil(t, claimed.ProcessingOwnerToken)
	require.NotNil(t, claimed.ProcessingStartedAt)

	completedAt := now.Add(time.Second)
	claimed.Status = coredata.WebhookEventStatusSucceeded
	claimed.ProcessingStartedAt = nil
	claimed.CompletedAt = &completedAt
	claimed.UpdatedAt = completedAt
	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return claimed.UpdateDeliveryState(
					ctx,
					conn,
					coredata.NewScope(claimed.ID.TenantID()),
				)
			},
		),
	)

	persisted := loadWebhookEvent(t, pgClient, event.ID)
	assert.Equal(t, coredata.WebhookEventStatusSucceeded, persisted.Status)
	assert.Nil(t, persisted.ProcessingOwnerToken)
	assert.Nil(t, persisted.ProcessingStartedAt)
	require.NotNil(t, persisted.CompletedAt)
	assert.WithinDuration(t, completedAt, *persisted.CompletedAt, time.Second)
}

func TestResetStaleWebhookEvents_RequeuesOrDeadLetters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		attemptCount int
		wantStatus   coredata.WebhookEventStatus
		wantDead     bool
	}{
		{
			name:         "requeues while attempts remain",
			attemptCount: 1,
			wantStatus:   coredata.WebhookEventStatusPending,
		},
		{
			name:         "dead letters after attempts are exhausted",
			attemptCount: coredata.WebhookEventDefaultMaxAttempts,
			wantStatus:   coredata.WebhookEventStatusFailed,
			wantDead:     true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				pgClient := test.PGClient(t)
				now := time.Now().UTC()
				event := insertWebhookDeliveryFixture(t, pgClient, now)
				startedAt := now.Add(-time.Hour)
				event.ProcessingOwnerToken = new("owner")
				event.ProcessingStartedAt = &startedAt
				event.AttemptCount = tt.attemptCount
				updateWebhookEventLease(t, pgClient, event)

				require.NoError(
					t,
					pgClient.WithConn(
						t.Context(),
						func(ctx context.Context, conn pg.Querier) error {
							return coredata.ResetStaleWebhookEvents(
								ctx,
								conn,
								now,
								10*time.Minute,
							)
						},
					),
				)

				persisted := loadWebhookEvent(t, pgClient, event.ID)
				assert.Equal(t, tt.wantStatus, persisted.Status)
				assert.Nil(t, persisted.ProcessingOwnerToken)
				assert.Nil(t, persisted.ProcessingStartedAt)
				assert.Equal(t, tt.wantDead, persisted.DeadLetteredAt != nil)
				assert.Equal(t, !tt.wantDead, persisted.NextAttemptAt != nil)
			},
		)
	}
}

func insertWebhookDeliveryFixture(
	t *testing.T,
	pgClient *pg.Client,
	now time.Time,
) *coredata.WebhookEvent {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organization := coredata.Organization{
		ID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TenantID:  tenantID,
		Name:      "Webhook event test",
		CreatedAt: now,
		UpdatedAt: now,
	}
	subscription := coredata.WebhookSubscription{
		ID:                     gid.New(tenantID, coredata.WebhookSubscriptionEntityType),
		OrganizationID:         organization.ID,
		EndpointURL:            "https://example.test/webhook",
		SelectedEvents:         coredata.WebhookEventTypes{coredata.WebhookEventTypeUserCreated},
		EncryptedSigningSecret: []byte("encrypted"),
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	processedAt := now
	data := coredata.WebhookData{
		ID:             gid.New(tenantID, coredata.WebhookDataEntityType),
		OrganizationID: organization.ID,
		EventType:      coredata.WebhookEventTypeUserCreated,
		Data:           []byte(`{"id":"user"}`),
		CreatedAt:      now,
		ProcessedAt:    &processedAt,
	}
	event := coredata.WebhookEvent{
		ID:                    gid.New(tenantID, coredata.WebhookEventEntityType),
		WebhookDataID:         data.ID,
		WebhookSubscriptionID: subscription.ID,
		Status:                coredata.WebhookEventStatusPending,
		MaxAttempts:           coredata.WebhookEventDefaultMaxAttempts,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				if err := organization.Insert(ctx, tx); err != nil {
					return err
				}
				if err := subscription.Insert(ctx, tx, scope); err != nil {
					return err
				}
				if err := data.Insert(ctx, tx, scope); err != nil {
					return err
				}

				return event.Insert(ctx, tx, scope)
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

	return &event
}

func updateWebhookEventLease(
	t *testing.T,
	pgClient *pg.Client,
	event *coredata.WebhookEvent,
) {
	t.Helper()

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(
					ctx,
					`UPDATE webhook_events
					SET processing_owner_token = $1,
						processing_started_at = $2,
						attempt_count = $3
					WHERE id = $4`,
					event.ProcessingOwnerToken,
					event.ProcessingStartedAt,
					event.AttemptCount,
					event.ID,
				)

				return err
			},
		),
	)
}

func loadWebhookEvent(
	t *testing.T,
	pgClient *pg.Client,
	id gid.GID,
) *coredata.WebhookEvent {
	t.Helper()

	var event coredata.WebhookEvent
	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return event.LoadByID(ctx, conn, coredata.NewScope(id.TenantID()), id)
			},
		),
	)

	return &event
}
