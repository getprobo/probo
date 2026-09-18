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

func insertTaskSyncJobFixture(
	t *testing.T,
	pgClient *pg.Client,
	now time.Time,
	status coredata.TaskSyncJobStatus,
	startedAt *time.Time,
	ownerToken *string,
	attemptCount int,
) *coredata.TaskSyncJob {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organization := coredata.Organization{
		ID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TenantID:  tenantID,
		Name:      "Task sync job test",
		CreatedAt: now,
		UpdatedAt: now,
	}
	job := coredata.TaskSyncJob{
		ID:                   gid.New(tenantID, coredata.TaskSyncJobEntityType),
		OrganizationID:       organization.ID,
		Direction:            coredata.TaskSyncJobDirectionOutbound,
		Status:               status,
		Payload:              []byte(`{}`),
		CreatedAt:            now,
		UpdatedAt:            now,
		StartedAt:            startedAt,
		AttemptCount:         attemptCount,
		ProcessingOwnerToken: ownerToken,
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				if err := organization.Insert(ctx, tx); err != nil {
					return err
				}

				return job.Insert(ctx, tx, scope)
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

	return &job
}

func loadTaskSyncJob(
	t *testing.T,
	pgClient *pg.Client,
	id gid.GID,
) *coredata.TaskSyncJob {
	t.Helper()

	var job coredata.TaskSyncJob

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return job.LoadByID(ctx, conn, coredata.NewScope(id.TenantID()), id)
			},
		),
	)

	return &job
}

func TestTaskSyncJob_ClaimAndCompleteProcessing(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC()
	job := insertTaskSyncJobFixture(
		t,
		pgClient,
		time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		coredata.TaskSyncJobStatusPending,
		nil,
		nil,
		0,
	)

	var claimed coredata.TaskSyncJob

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return claimed.ClaimNextForUpdateSkipLocked(ctx, tx, now)
			},
		),
	)

	assert.Equal(t, job.ID, claimed.ID)
	assert.Equal(t, coredata.TaskSyncJobStatusProcessing, claimed.Status)
	require.NotNil(t, claimed.ProcessingOwnerToken)
	require.NotNil(t, claimed.StartedAt)

	completedAt := now.Add(time.Second)
	claimed.Status = coredata.TaskSyncJobStatusSucceeded
	claimed.CompletedAt = &completedAt
	claimed.UpdatedAt = completedAt
	claimed.Error = nil

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return claimed.UpdateProcessingState(
					ctx,
					conn,
					coredata.NewScope(claimed.ID.TenantID()),
				)
			},
		),
	)

	persisted := loadTaskSyncJob(t, pgClient, claimed.ID)
	assert.Equal(t, coredata.TaskSyncJobStatusSucceeded, persisted.Status)
	assert.Nil(t, persisted.ProcessingOwnerToken)
	require.NotNil(t, persisted.CompletedAt)
	assert.WithinDuration(t, completedAt, *persisted.CompletedAt, time.Second)
}

func TestTaskSyncJob_UpdateProcessingState_WrongToken(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC()
	startedAt := now.Add(-time.Minute)
	job := insertTaskSyncJobFixture(
		t,
		pgClient,
		now,
		coredata.TaskSyncJobStatusProcessing,
		&startedAt,
		new("owner-a"),
		0,
	)

	job.ProcessingOwnerToken = new("owner-b")
	job.Status = coredata.TaskSyncJobStatusSucceeded
	job.CompletedAt = &now
	job.UpdatedAt = now

	err := pgClient.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			return job.UpdateProcessingState(
				ctx,
				conn,
				coredata.NewScope(job.ID.TenantID()),
			)
		},
	)
	require.ErrorIs(t, err, coredata.ErrProcessingLeaseLost)

	persisted := loadTaskSyncJob(t, pgClient, job.ID)
	assert.Equal(t, coredata.TaskSyncJobStatusProcessing, persisted.Status)
	require.NotNil(t, persisted.ProcessingOwnerToken)
	assert.Equal(t, "owner-a", *persisted.ProcessingOwnerToken)
	assert.Nil(t, persisted.CompletedAt)
}

func TestTaskSyncJob_TouchLease_RenewsStartedAt(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC()
	startedAt := now.Add(-time.Minute)
	job := insertTaskSyncJobFixture(
		t,
		pgClient,
		now,
		coredata.TaskSyncJobStatusProcessing,
		&startedAt,
		new("owner"),
		0,
	)

	touchedAt := now.Add(time.Minute)

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return job.TouchLease(
					ctx,
					conn,
					coredata.NewScope(job.ID.TenantID()),
					touchedAt,
				)
			},
		),
	)

	persisted := loadTaskSyncJob(t, pgClient, job.ID)
	assert.Equal(t, coredata.TaskSyncJobStatusProcessing, persisted.Status)
	require.NotNil(t, persisted.ProcessingOwnerToken)
	assert.Equal(t, "owner", *persisted.ProcessingOwnerToken)
	require.NotNil(t, persisted.StartedAt)
	assert.WithinDuration(t, touchedAt, *persisted.StartedAt, time.Second)
}

func TestTaskSyncJob_TouchLease_WrongToken(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC()
	startedAt := now.Add(-time.Minute)
	job := insertTaskSyncJobFixture(
		t,
		pgClient,
		now,
		coredata.TaskSyncJobStatusProcessing,
		&startedAt,
		new("owner-a"),
		0,
	)

	job.ProcessingOwnerToken = new("owner-b")

	err := pgClient.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			return job.TouchLease(
				ctx,
				conn,
				coredata.NewScope(job.ID.TenantID()),
				now,
			)
		},
	)
	require.ErrorIs(t, err, coredata.ErrProcessingLeaseLost)
}

func TestResetStaleTaskSyncJobs_RequeuesOrFails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		attemptCount int
		wantStatus   coredata.TaskSyncJobStatus
	}{
		{
			name:         "requeues while attempts remain",
			attemptCount: 1,
			wantStatus:   coredata.TaskSyncJobStatusPending,
		},
		{
			name:         "fails after attempts are exhausted",
			attemptCount: coredata.TaskSyncJobDefaultMaxAttempts - 1,
			wantStatus:   coredata.TaskSyncJobStatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				pgClient := test.PGClient(t)
				now := time.Now().UTC()
				startedAt := now.Add(-time.Hour)
				job := insertTaskSyncJobFixture(
					t,
					pgClient,
					now,
					coredata.TaskSyncJobStatusProcessing,
					&startedAt,
					new("owner"),
					tt.attemptCount,
				)

				require.NoError(
					t,
					pgClient.WithConn(
						t.Context(),
						func(ctx context.Context, conn pg.Querier) error {
							return coredata.ResetStaleTaskSyncJobs(ctx, conn, 10*time.Minute)
						},
					),
				)

				persisted := loadTaskSyncJob(t, pgClient, job.ID)
				assert.Equal(t, tt.wantStatus, persisted.Status)
				assert.Nil(t, persisted.ProcessingOwnerToken)
				assert.Nil(t, persisted.StartedAt)
				assert.Equal(
					t,
					tt.wantStatus == coredata.TaskSyncJobStatusFailed,
					persisted.CompletedAt != nil,
				)
				assert.Equal(
					t,
					tt.wantStatus == coredata.TaskSyncJobStatusPending,
					persisted.NextAttemptAt != nil,
				)
				assert.Equal(t, tt.attemptCount+1, persisted.AttemptCount)
			},
		)
	}
}
