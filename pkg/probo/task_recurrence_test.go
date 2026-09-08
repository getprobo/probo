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

package probo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/timespan"
)

func TestNextRecurrenceDeadline(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 28, 12, 0, 0, 0, time.UTC)
	oneDayOverdue := now.AddDate(0, 0, -1)

	tests := []struct {
		name     string
		deadline time.Time
		interval timespan.TimeSpan
		want     time.Time
	}{
		{
			name:     "every 2 days, completed a day late",
			deadline: oneDayOverdue,
			interval: timespan.TimeSpan{Days: 2},
			want:     oneDayOverdue.AddDate(0, 0, 2),
		},
		{
			name:     "every 3 weeks, completed a day late",
			deadline: oneDayOverdue,
			interval: timespan.TimeSpan{Days: 21},
			want:     oneDayOverdue.AddDate(0, 0, 21),
		},
		{
			name:     "every month, completed a day late",
			deadline: oneDayOverdue,
			interval: timespan.TimeSpan{Months: 1},
			want:     timespan.TimeSpan{Months: 1}.AddTo(oneDayOverdue),
		},
		{
			name:     "completed before the deadline still advances one interval",
			deadline: now.AddDate(0, 1, 0),
			interval: timespan.TimeSpan{Days: 90},
			want:     now.AddDate(0, 1, 0).AddDate(0, 0, 90),
		},
		{
			name:     "every year, completed a day late",
			deadline: oneDayOverdue,
			interval: timespan.TimeSpan{Months: 12},
			want:     timespan.TimeSpan{Months: 12}.AddTo(oneDayOverdue),
		},
		{
			name:     "long outage collapses to a single future occurrence",
			deadline: now.AddDate(-1, 0, 0),
			interval: timespan.TimeSpan{Days: 2},
			want:     now.AddDate(-1, 0, 0).AddDate(0, 0, 366),
		},
		{
			name:     "monthly from january 31 skips to the next calendar month after now",
			deadline: time.Date(2026, time.January, 31, 12, 0, 0, 0, time.UTC),
			interval: timespan.TimeSpan{Months: 1},
			want:     time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "one-microsecond catch-up jumps to the next instant after now",
			deadline: now.Add(-24 * time.Hour),
			interval: timespan.TimeSpan{Microseconds: 1},
			want:     now.Add(time.Microsecond),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := nextRecurrenceDeadline(tt.deadline, tt.interval, now)

			require.NoError(t, err)
			assert.True(t, got.After(now), "next deadline must be strictly after now")
			assert.True(t, got.Equal(tt.want), "got %v, want %v", got, tt.want)
		})
	}
}

func TestNextRecurrenceDeadline_InvalidRecurrence(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 28, 12, 0, 0, 0, time.UTC)
	oneDayOverdue := now.AddDate(0, 0, -1)
	january31 := time.Date(2026, time.January, 31, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		deadline time.Time
		interval timespan.TimeSpan
	}{
		{
			name:     "zero interval never advances",
			deadline: oneDayOverdue,
			interval: timespan.Zero,
		},
		{
			name:     "negative interval advances backwards",
			deadline: oneDayOverdue,
			interval: timespan.TimeSpan{Microseconds: -int64(time.Hour / time.Microsecond)},
		},
		{
			name:     "mixed-sign month does not advance january 31",
			deadline: january31,
			interval: timespan.TimeSpan{Months: 1, Days: -29},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := nextRecurrenceDeadline(tt.deadline, tt.interval, now)

			assert.Error(t, err)
		})
	}
}

func TestCreateTaskRequest_Validate_RecurringDone(t *testing.T) {
	t.Parallel()

	interval := timespan.TimeSpan{Days: 7}
	deadline := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	state := coredata.TaskStateDone

	err := (&CreateTaskRequest{
		OrganizationID:     gid.New(gid.NewTenantID(), coredata.OrganizationEntityType),
		Name:               "Recurring",
		Priority:           coredata.TaskPriorityMedium,
		Deadline:           &deadline,
		RecurrenceInterval: &interval,
		State:              &state,
	}).Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "state")
}

func TestCreateTaskRequest_Validate_RecurrenceMustAdvanceDeadline(t *testing.T) {
	t.Parallel()

	january31 := time.Date(2026, time.January, 31, 12, 0, 0, 0, time.UTC)
	january15 := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)
	mixedSign := timespan.TimeSpan{Months: 1, Days: -29}
	monthly := timespan.TimeSpan{Months: 1}

	t.Run("mixed-sign interval that clamps backward is rejected", func(t *testing.T) {
		t.Parallel()

		err := newRecurringCreateRequest(january31, mixedSign).Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "recurrence_interval")
		assert.Contains(t, err.Error(), "deadline")
	})

	t.Run("the same mixed-sign interval is accepted when it still advances", func(t *testing.T) {
		t.Parallel()

		err := newRecurringCreateRequest(january15, mixedSign).Validate()

		assert.NoError(t, err)
	})

	t.Run("a whole month from january 31 is accepted", func(t *testing.T) {
		t.Parallel()

		err := newRecurringCreateRequest(january31, monthly).Validate()

		assert.NoError(t, err)
	})
}

func newRecurringCreateRequest(deadline time.Time, interval timespan.TimeSpan) *CreateTaskRequest {
	return &CreateTaskRequest{
		OrganizationID:     gid.New(gid.NewTenantID(), coredata.OrganizationEntityType),
		Name:               "Recurring",
		Priority:           coredata.TaskPriorityMedium,
		Deadline:           &deadline,
		RecurrenceInterval: &interval,
	}
}

func TestShouldCloneRecurringTask(t *testing.T) {
	t.Parallel()

	interval := timespan.TimeSpan{Days: 7}
	deadline := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	recurring := &coredata.Task{
		Recurrence: &interval,
		Deadline:   &deadline,
	}

	t.Run("completing a recurring task clones", func(t *testing.T) {
		t.Parallel()
		assert.True(t, shouldCloneRecurringTask(coredata.TaskStateTodo, coredata.TaskStateDone, recurring))
	})

	t.Run("already done does not clone again", func(t *testing.T) {
		t.Parallel()
		assert.False(t, shouldCloneRecurringTask(coredata.TaskStateDone, coredata.TaskStateDone, recurring))
	})

	t.Run("canceling does not clone", func(t *testing.T) {
		t.Parallel()
		assert.False(t, shouldCloneRecurringTask(coredata.TaskStateTodo, coredata.TaskStateCanceled, recurring))
	})

	t.Run("completing a non-recurring task does not clone", func(t *testing.T) {
		t.Parallel()
		assert.False(t, shouldCloneRecurringTask(coredata.TaskStateTodo, coredata.TaskStateDone, &coredata.Task{
			Deadline: &deadline,
		}))
	})
}
