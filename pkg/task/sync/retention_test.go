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

package tasksync

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestRetentionHandlerDeletesExpiredLinearWebhookEvents(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC().Round(time.Microsecond)
	expiredProcessedAt := now.Add(-2 * time.Hour)
	keptProcessedAt := now.Add(-30 * time.Minute)
	expiredDeadLetteredAt := now.Add(-4 * time.Hour)

	expiredProcessed := coredata.NewLinearWebhookEvent("retention-processed-expired", []byte(`{}`))
	expiredProcessed.ProcessedAt = &expiredProcessedAt
	expiredProcessed.UpdatedAt = expiredProcessedAt

	keptProcessed := coredata.NewLinearWebhookEvent("retention-processed-kept", []byte(`{}`))
	keptProcessed.ProcessedAt = &keptProcessedAt
	keptProcessed.UpdatedAt = keptProcessedAt

	expiredDeadLetter := coredata.NewLinearWebhookEvent("retention-dead-letter-expired", []byte(`{}`))
	expiredDeadLetter.DeadLetteredAt = &expiredDeadLetteredAt
	expiredDeadLetter.UpdatedAt = expiredDeadLetteredAt

	pending := coredata.NewLinearWebhookEvent("retention-pending", []byte(`{}`))

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				for _, event := range []*coredata.LinearWebhookEvent{
					expiredProcessed,
					keptProcessed,
					expiredDeadLetter,
					pending,
				} {
					if _, err := event.Insert(ctx, tx); err != nil {
						return err
					}
				}

				return nil
			},
		),
	)

	h := &retentionHandler{
		pg:                  pgClient,
		logger:              log.NewLogger(log.WithOutput(io.Discard)),
		retention:           time.Hour,
		deadLetterRetention: 3 * time.Hour,
		batchSize:           10,
		now:                 func() time.Time { return now },
	}
	require.NoError(t, h.Run(t.Context()))

	var remaining []string

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				rows, err := conn.Query(
					ctx,
					`SELECT delivery_id
					FROM linear_webhook_events
					WHERE delivery_id IN (
						@expired_processed,
						@kept_processed,
						@expired_dead_letter,
						@pending
					)
					ORDER BY delivery_id`,
					pgx.StrictNamedArgs{
						"expired_processed":   expiredProcessed.DeliveryID,
						"kept_processed":      keptProcessed.DeliveryID,
						"expired_dead_letter": expiredDeadLetter.DeliveryID,
						"pending":             pending.DeliveryID,
					},
				)
				if err != nil {
					return err
				}

				remaining, err = pgx.CollectRows(rows, pgx.RowTo[string])

				return err
			},
		),
	)
	require.Equal(t, []string{pending.DeliveryID, keptProcessed.DeliveryID}, remaining)
}

func TestRetentionHandlerDeletesExpiredTaskSyncJobs(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC().Round(time.Microsecond)
	organization := insertRetentionOrganization(t, pgClient, now)

	expiredSucceededAt := now.Add(-2 * time.Hour)
	keptSucceededAt := now.Add(-30 * time.Minute)
	expiredFailedAt := now.Add(-4 * time.Hour)
	keptFailedAt := now.Add(-2 * time.Hour)

	expiredSucceeded := insertRetentionTaskSyncJob(
		t,
		pgClient,
		organization,
		now,
		coredata.TaskSyncJobStatusSucceeded,
		&expiredSucceededAt,
	)
	keptSucceeded := insertRetentionTaskSyncJob(
		t,
		pgClient,
		organization,
		now,
		coredata.TaskSyncJobStatusSucceeded,
		&keptSucceededAt,
	)
	expiredFailed := insertRetentionTaskSyncJob(
		t,
		pgClient,
		organization,
		now,
		coredata.TaskSyncJobStatusFailed,
		&expiredFailedAt,
	)
	keptFailed := insertRetentionTaskSyncJob(
		t,
		pgClient,
		organization,
		now,
		coredata.TaskSyncJobStatusFailed,
		&keptFailedAt,
	)
	pending := insertRetentionTaskSyncJob(
		t,
		pgClient,
		organization,
		now,
		coredata.TaskSyncJobStatusPending,
		nil,
	)
	processing := insertRetentionTaskSyncJob(
		t,
		pgClient,
		organization,
		now,
		coredata.TaskSyncJobStatusProcessing,
		nil,
	)

	h := &retentionHandler{
		pg:                  pgClient,
		logger:              log.NewLogger(log.WithOutput(io.Discard)),
		retention:           time.Hour,
		deadLetterRetention: 3 * time.Hour,
		batchSize:           10,
		now:                 func() time.Time { return now },
	}
	require.NoError(t, h.Run(t.Context()))

	var remaining []string

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				rows, err := conn.Query(
					ctx,
					`SELECT id
					FROM task_sync_jobs
					WHERE id IN (
						@expired_succeeded,
						@kept_succeeded,
						@expired_failed,
						@kept_failed,
						@pending,
						@processing
					)
					ORDER BY id`,
					pgx.StrictNamedArgs{
						"expired_succeeded": expiredSucceeded.ID,
						"kept_succeeded":    keptSucceeded.ID,
						"expired_failed":    expiredFailed.ID,
						"kept_failed":       keptFailed.ID,
						"pending":           pending.ID,
						"processing":        processing.ID,
					},
				)
				if err != nil {
					return err
				}

				remaining, err = pgx.CollectRows(rows, pgx.RowTo[string])

				return err
			},
		),
	)

	require.ElementsMatch(
		t,
		[]string{
			keptSucceeded.ID.String(),
			keptFailed.ID.String(),
			pending.ID.String(),
			processing.ID.String(),
		},
		remaining,
	)
}

func insertRetentionOrganization(
	t *testing.T,
	pgClient *pg.Client,
	now time.Time,
) *coredata.Organization {
	t.Helper()

	tenantID := gid.NewTenantID()
	organization := coredata.Organization{
		ID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TenantID:  tenantID,
		Name:      "Task sync retention test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return organization.Insert(ctx, tx)
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

	return &organization
}

func insertRetentionTaskSyncJob(
	t *testing.T,
	pgClient *pg.Client,
	organization *coredata.Organization,
	now time.Time,
	status coredata.TaskSyncJobStatus,
	completedAt *time.Time,
) *coredata.TaskSyncJob {
	t.Helper()

	job := coredata.TaskSyncJob{
		ID:             gid.New(organization.TenantID, coredata.TaskSyncJobEntityType),
		OrganizationID: organization.ID,
		Direction:      coredata.TaskSyncJobDirectionOutbound,
		Status:         status,
		Payload:        []byte(`{}`),
		CreatedAt:      now,
		UpdatedAt:      now,
		CompletedAt:    completedAt,
	}

	if completedAt != nil {
		job.UpdatedAt = *completedAt
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return job.Insert(ctx, tx, coredata.NewScope(organization.TenantID))
			},
		),
	)

	return &job
}
