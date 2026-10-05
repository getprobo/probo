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

package task

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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/timespan"
)

func TestNextRecurrenceDeadline(t *testing.T) {
	t.Parallel()

	t.Run(
		"advances from the deadline",
		func(t *testing.T) {
			t.Parallel()

			deadline := time.Date(2027, 1, 15, 0, 0, 0, 0, time.UTC)
			now := deadline.Add(48 * time.Hour)

			next, err := nextRecurrenceDeadline(deadline, mustInterval(t, "P21D"), now)
			require.NoError(t, err)
			assert.Equal(t, time.Date(2027, 2, 5, 0, 0, 0, 0, time.UTC), next)
		},
	)

	t.Run(
		"repeats when the deadline is this instant",
		func(t *testing.T) {
			t.Parallel()

			deadline := time.Date(2027, 1, 15, 0, 0, 0, 0, time.UTC)

			next, err := nextRecurrenceDeadline(deadline, mustInterval(t, "P21D"), deadline)
			require.NoError(t, err)
			assert.Equal(t, time.Date(2027, 2, 5, 0, 0, 0, 0, time.UTC), next)
		},
	)

	t.Run(
		"collapses missed clock cycles",
		func(t *testing.T) {
			t.Parallel()

			deadline := time.Date(2027, 1, 15, 0, 0, 0, 0, time.UTC)
			now := time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)

			next, err := nextRecurrenceDeadline(deadline, mustInterval(t, "P21D"), now)
			require.NoError(t, err)
			assert.Equal(t, time.Date(2027, 3, 19, 0, 0, 0, 0, time.UTC), next)
			assert.True(t, next.After(now))
		},
	)

	t.Run(
		"advances a calendar month",
		func(t *testing.T) {
			t.Parallel()

			deadline := time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC)
			now := time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC)

			next, err := nextRecurrenceDeadline(deadline, mustInterval(t, "P1M"), now)
			require.NoError(t, err)
			assert.Equal(t, time.Date(2027, 2, 28, 0, 0, 0, 0, time.UTC), next)
		},
	)

	t.Run(
		"rejects an interval that does not advance this deadline",
		func(t *testing.T) {
			t.Parallel()

			deadline := time.Date(2000, 1, 31, 0, 0, 0, 0, time.UTC)

			_, err := nextRecurrenceDeadline(deadline, mustInterval(t, "P1M-29D"), deadline)
			require.ErrorIs(t, err, errRecurrenceCannotAdvance)
		},
	)

	t.Run(
		"accepts that interval on a day it can advance",
		func(t *testing.T) {
			t.Parallel()

			deadline := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

			next, err := nextRecurrenceDeadline(deadline, mustInterval(t, "P1M-29D"), deadline)
			require.NoError(t, err)
			assert.True(t, next.After(deadline))
		},
	)

	t.Run(
		"computes a later deadline when the catch-up bound overflows",
		func(t *testing.T) {
			t.Parallel()

			// One cycle is one month plus 2148 days, about six years, so the
			// due cycle still fits in int32. This deadline caps the catch-up
			// bound at 1,000,000, and 2148 days times that bound overflows.
			deadline := time.Date(-84000, 1, 1, 0, 0, 0, 0, time.UTC)
			now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
			interval := mustInterval(t, "P1M2148D")

			_, scaleErr := interval.Times(calendarCatchUpBound(deadline, interval, now))
			require.ErrorContains(t, scaleErr, "overflow")

			next, err := nextRecurrenceDeadline(deadline, interval, now)
			require.NoError(t, err)
			assert.True(t, next.After(now))
		},
	)

	t.Run(
		"returns timespan overflow when no representable cycle is due",
		func(t *testing.T) {
			t.Parallel()

			// Six million years needs more cycles than int32 days can scale,
			// and every representable cycle is still before now.
			deadline := time.Date(-6000000, 1, 31, 0, 0, 0, 0, time.UTC)
			now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

			_, err := nextRecurrenceDeadline(deadline, mustInterval(t, "P72M-2148D"), now)
			require.Error(t, err)
			assert.NotErrorIs(t, err, errRecurrenceCannotAdvance)
			assert.ErrorContains(t, err, "overflow")
		},
	)
}

func TestRepeatLockedRecurringTask(t *testing.T) {
	t.Parallel()

	t.Run(
		"repeats a done task whose deadline has passed",
		func(t *testing.T) {
			t.Parallel()

			client := test.PGClient(t)
			now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
			deadline := now.Add(-time.Hour)
			source := insertRecurringTask(
				t,
				client,
				coredata.TaskStateDone,
				deadline,
				mustInterval(t, "P21D"),
			)
			scope := coredata.NewScope(source.OrganizationID.TenantID())

			var next *coredata.Task

			err := client.WithTx(
				t.Context(),
				func(ctx context.Context, tx pg.Tx) error {
					locked := &coredata.Task{}
					if err := locked.LoadByIDForUpdate(ctx, tx, scope, source.ID); err != nil {
						return err
					}

					cloned, err := repeatLockedRecurringTask(ctx, tx, locked, now)
					next = cloned

					return err
				},
			)
			require.NoError(t, err)
			require.NotNil(t, next)

			stored := loadTask(t, client, scope, source.ID)
			assert.Equal(t, coredata.TaskStateDone, stored.State)
			assert.Nil(t, stored.Recurrence)
			require.NotNil(t, stored.Deadline)
			assert.True(t, stored.Deadline.Equal(deadline))

			cloned := loadTask(t, client, scope, next.ID)
			assert.NotEqual(t, source.ID, cloned.ID)
			assert.Equal(t, source.Name, cloned.Name)
			assert.Equal(t, coredata.TaskStateTodo, cloned.State)
			require.NotNil(t, cloned.Recurrence)
			assert.Equal(t, mustInterval(t, "P21D"), *cloned.Recurrence)
			require.NotNil(t, cloned.Deadline)
			assert.True(t, cloned.Deadline.Equal(deadline.Add(21*24*time.Hour)))
			assert.Equal(t, 1, countTaskActivities(t, client, cloned.ID, coredata.TaskActivityTypeCreated))
		},
	)

	t.Run(
		"repeats a canceled task whose deadline has passed",
		func(t *testing.T) {
			t.Parallel()

			client := test.PGClient(t)
			now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
			source := insertRecurringTask(
				t,
				client,
				coredata.TaskStateCanceled,
				now.Add(-time.Minute),
				mustInterval(t, "P7D"),
			)
			scope := coredata.NewScope(source.OrganizationID.TenantID())

			var next *coredata.Task

			err := client.WithTx(
				t.Context(),
				func(ctx context.Context, tx pg.Tx) error {
					locked := &coredata.Task{}
					if err := locked.LoadByIDForUpdate(ctx, tx, scope, source.ID); err != nil {
						return err
					}

					cloned, err := repeatLockedRecurringTask(ctx, tx, locked, now)
					next = cloned

					return err
				},
			)
			require.NoError(t, err)
			require.NotNil(t, next)

			stored := loadTask(t, client, scope, source.ID)
			assert.Equal(t, coredata.TaskStateCanceled, stored.State)
			assert.Nil(t, stored.Recurrence)

			cloned := loadTask(t, client, scope, next.ID)
			assert.Equal(t, coredata.TaskStateTodo, cloned.State)
			require.NotNil(t, cloned.Recurrence)
			assert.Equal(t, mustInterval(t, "P7D"), *cloned.Recurrence)
		},
	)

	t.Run(
		"rejects a task whose deadline is still ahead",
		func(t *testing.T) {
			t.Parallel()

			client := test.PGClient(t)
			now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
			source := insertRecurringTask(
				t,
				client,
				coredata.TaskStateTodo,
				now.Add(24*time.Hour),
				mustInterval(t, "P21D"),
			)
			scope := coredata.NewScope(source.OrganizationID.TenantID())

			err := client.WithTx(
				t.Context(),
				func(ctx context.Context, tx pg.Tx) error {
					locked := &coredata.Task{}
					if err := locked.LoadByIDForUpdate(ctx, tx, scope, source.ID); err != nil {
						return err
					}

					_, err := repeatLockedRecurringTask(ctx, tx, locked, now)

					return err
				},
			)
			require.Error(t, err)
			assert.ErrorContains(t, err, "deadline")

			stored := loadTask(t, client, scope, source.ID)
			require.NotNil(t, stored.Recurrence)
			assert.Equal(t, mustInterval(t, "P21D"), *stored.Recurrence)
			assert.Equal(t, coredata.TaskStateTodo, stored.State)
		},
	)

	t.Run(
		"advances a monthly deadline to the next calendar month",
		func(t *testing.T) {
			t.Parallel()

			client := test.PGClient(t)
			deadline := time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC)
			now := time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC)
			source := insertRecurringTask(t, client, coredata.TaskStateTodo, deadline, mustInterval(t, "P1M"))
			scope := coredata.NewScope(source.OrganizationID.TenantID())

			var next *coredata.Task

			err := client.WithTx(
				t.Context(),
				func(ctx context.Context, tx pg.Tx) error {
					locked := &coredata.Task{}
					if err := locked.LoadByIDForUpdate(ctx, tx, scope, source.ID); err != nil {
						return err
					}

					cloned, err := repeatLockedRecurringTask(ctx, tx, locked, now)
					next = cloned

					return err
				},
			)
			require.NoError(t, err)

			cloned := loadTask(t, client, scope, next.ID)
			require.NotNil(t, cloned.Deadline)
			assert.True(t, cloned.Deadline.Equal(time.Date(2027, 2, 28, 0, 0, 0, 0, time.UTC)))
			require.NotNil(t, cloned.Recurrence)
			assert.Equal(t, mustInterval(t, "P1M"), *cloned.Recurrence)
		},
	)
}

func TestRecurrenceWorkerClaim(t *testing.T) {
	t.Parallel()

	t.Run(
		"skips an interval that cannot advance and repeats the next task",
		func(t *testing.T) {
			t.Parallel()

			client := test.PGClient(t)
			poison := insertRecurringTask(
				t,
				client,
				coredata.TaskStateTodo,
				time.Date(2000, 1, 31, 0, 0, 0, 0, time.UTC),
				mustInterval(t, "P1M-29D"),
			)

			good := insertRecurringTask(
				t,
				client,
				coredata.TaskStateTodo,
				time.Date(2000, 2, 1, 0, 0, 0, 0, time.UTC),
				mustInterval(t, "P21D"),
			)

			insertTaskWebhookSubscription(
				t,
				client,
				poison.OrganizationID,
				coredata.WebhookEventTypeTaskUpdated,
			)

			handler := &recurrenceHandler{
				pg:     client,
				logger: log.NewLogger(log.WithOutput(io.Discard)),
			}

			got, err := handler.Claim(t.Context())
			require.NoError(t, err)
			assert.Equal(t, good.OrganizationID, got.OrganizationID)
			assert.NotEqual(t, good.ID, got.ID)
			assert.Equal(t, coredata.TaskStateTodo, got.State)

			stopped := loadTask(
				t,
				client,
				coredata.NewScope(poison.OrganizationID.TenantID()),
				poison.ID,
			)
			assert.Nil(t, stopped.Recurrence)
			require.NotNil(t, stopped.Deadline)
			assert.True(t, stopped.Deadline.Equal(time.Date(2000, 1, 31, 0, 0, 0, 0, time.UTC)))

			payloads := loadTaskWebhookPayloads(
				t,
				client,
				poison.OrganizationID,
				coredata.WebhookEventTypeTaskUpdated,
			)
			require.Len(t, payloads, 1)
			assert.Equal(t, poison.ID.String(), payloads[0]["id"])
			assert.Nil(t, payloads[0]["recurrence"])

			updatedFrom := loadTaskWebhookUpdatedFrom(
				t,
				client,
				poison.OrganizationID,
				coredata.WebhookEventTypeTaskUpdated,
			)
			assert.Equal(t, "P1M-29D", updatedFrom["recurrence"])

			source := loadTask(
				t,
				client,
				coredata.NewScope(good.OrganizationID.TenantID()),
				good.ID,
			)
			assert.Nil(t, source.Recurrence)
		},
	)
}

func mustInterval(t *testing.T, raw string) timespan.TimeSpan {
	t.Helper()

	interval, err := timespan.Parse(raw)
	require.NoError(t, err)

	return interval
}

func insertRecurringTask(
	t *testing.T,
	client *pg.Client,
	state coredata.TaskState,
	deadline time.Time,
	interval timespan.TimeSpan,
) *coredata.Task {
	t.Helper()

	now := time.Now()
	tenantID := gid.NewTenantID()
	organization := coredata.Organization{
		ID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TenantID:  tenantID,
		Name:      "recurrence-" + tenantID.String(),
		CreatedAt: now,
		UpdatedAt: now,
	}

	task := &coredata.Task{
		ID:             gid.New(tenantID, coredata.TaskEntityType),
		OrganizationID: organization.ID,
		Name:           "Recurring task",
		Content:        `{"type":"doc","content":[]}`,
		State:          state,
		Priority:       coredata.TaskPriorityMedium,
		ReferenceID:    "custom-task-" + gid.New(tenantID, coredata.TaskEntityType).String(),
		Deadline:       &deadline,
		Recurrence:     &interval,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	scope := coredata.NewScope(tenantID)

	require.NoError(
		t,
		client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				if err := organization.Insert(ctx, tx); err != nil {
					return err
				}

				return task.Insert(ctx, tx, scope)
			},
		),
	)

	t.Cleanup(
		func() {
			_ = client.WithTx(
				context.Background(),
				func(ctx context.Context, tx pg.Tx) error {
					return organization.Delete(ctx, tx, organization.ID)
				},
			)
		},
	)

	return task
}

func loadTask(
	t *testing.T,
	client *pg.Client,
	scope coredata.Scoper,
	taskID gid.GID,
) *coredata.Task {
	t.Helper()

	task := &coredata.Task{}

	require.NoError(
		t,
		client.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return task.LoadByID(ctx, conn, scope, taskID)
			},
		),
	)

	return task
}

func countTaskActivities(
	t *testing.T,
	client *pg.Client,
	taskID gid.GID,
	activityType coredata.TaskActivityType,
) int {
	t.Helper()

	var count int

	require.NoError(
		t,
		client.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return conn.QueryRow(
					ctx,
					`SELECT COUNT(*) FROM task_activities WHERE task_id = $1 AND activity_type = $2`,
					taskID.String(),
					activityType.String(),
				).Scan(&count)
			},
		),
	)

	return count
}
