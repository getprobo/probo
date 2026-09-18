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
	"context"
	"fmt"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/timespan"
)

func shouldCloneRecurringTask(oldState, newState coredata.TaskState, task *coredata.Task) bool {
	return oldState != coredata.TaskStateDone &&
		newState == coredata.TaskStateDone &&
		task.Recurrence != nil &&
		task.Deadline != nil
}

func insertNextRecurringTask(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	completed *coredata.Task,
	now time.Time,
) (*coredata.Task, error) {
	if completed.Recurrence == nil || completed.Deadline == nil {
		return nil, fmt.Errorf("cannot clone a task without a recurrence interval and deadline")
	}

	interval := *completed.Recurrence

	deadline, err := nextRecurrenceDeadline(*completed.Deadline, interval, now)
	if err != nil {
		return nil, fmt.Errorf("cannot compute next deadline of task %q: %w", completed.ID, err)
	}

	referenceID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot generate reference id: %w", err)
	}

	next := &coredata.Task{
		ID:             gid.New(completed.OrganizationID.TenantID(), coredata.TaskEntityType),
		OrganizationID: completed.OrganizationID,
		MeasureID:      completed.MeasureID,
		Name:           completed.Name,
		Content:        completed.Content,
		Priority:       completed.Priority,
		ReferenceID:    "custom-task-" + referenceID.String(),
		TimeEstimate:   completed.TimeEstimate,
		AssignedToID:   completed.AssignedToID,
		Deadline:       &deadline,
		Recurrence:     &interval,
		State:          coredata.TaskStateTodo,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := next.Insert(ctx, conn, scope); err != nil {
		return nil, fmt.Errorf("cannot insert next recurring task: %w", err)
	}

	return next, nil
}

// nextRecurrenceDeadline advances deadline by at least one interval, then
// repeats until the result is after now. Completing early still moves the
// series forward one cycle; completing late collapses missed cycles into a
// single upcoming occurrence instead of backfilling one task per cycle.
//
// Clock intervals are jumped in O(1). Calendar intervals (months) use a
// bounded binary search so a 1µs series cannot spin for billions of cycles.
func nextRecurrenceDeadline(
	deadline time.Time,
	interval timespan.TimeSpan,
	now time.Time,
) (time.Time, error) {
	if interval.Compare(timespan.Zero) <= 0 {
		return time.Time{}, fmt.Errorf("recurrence interval must be greater than 0, got %s", interval)
	}

	first := interval.AddTo(deadline)
	if !first.After(deadline) {
		return time.Time{}, fmt.Errorf("recurrence interval must advance the deadline, got %s", interval)
	}

	if first.After(now) {
		return first, nil
	}

	if duration, ok := interval.ClockDuration(); ok {
		return nextClockRecurrenceDeadline(deadline, duration, now)
	}

	return nextCalendarRecurrenceDeadline(deadline, interval, now)
}

func nextClockRecurrenceDeadline(
	deadline time.Time,
	interval time.Duration,
	now time.Time,
) (time.Time, error) {
	if interval <= 0 {
		return time.Time{}, fmt.Errorf("recurrence interval must be greater than 0, got %s", interval)
	}

	elapsed := now.Sub(deadline)
	cycles := elapsed / interval

	next := deadline.Add(cycles * interval).Add(interval)
	if !next.After(now) {
		return time.Time{}, fmt.Errorf("recurrence interval must advance the deadline, got %s", interval)
	}

	return next, nil
}

const maxCalendarCatchUp = 1_000_000

func nextCalendarRecurrenceDeadline(
	deadline time.Time,
	interval timespan.TimeSpan,
	now time.Time,
) (time.Time, error) {
	hi := calendarCatchUpBound(deadline, interval, now)

	high, err := occurrenceAfterCycles(deadline, interval, hi)
	if err != nil {
		return time.Time{}, err
	}

	if !high.After(now) {
		return time.Time{}, fmt.Errorf("recurrence interval must advance the deadline, got %s", interval)
	}

	lo := 1
	for lo+1 < hi {
		mid := lo + (hi-lo)/2

		candidate, err := occurrenceAfterCycles(deadline, interval, mid)
		if err != nil {
			hi = mid

			continue
		}

		if candidate.After(now) {
			hi = mid
			high = candidate

			continue
		}

		lo = mid
	}

	return high, nil
}

func occurrenceAfterCycles(
	deadline time.Time,
	interval timespan.TimeSpan,
	n int,
) (time.Time, error) {
	scaled, err := interval.Times(n)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot advance recurrence interval: %w", err)
	}

	next := scaled.AddTo(deadline)

	if n > 1 {
		previous, err := interval.Times(n - 1)
		if err != nil {
			return time.Time{}, fmt.Errorf("cannot advance recurrence interval: %w", err)
		}

		if !next.After(previous.AddTo(deadline)) {
			return time.Time{}, fmt.Errorf("recurrence interval must advance the deadline, got %s", interval)
		}
	}

	return next, nil
}

func calendarCatchUpBound(deadline time.Time, interval timespan.TimeSpan, now time.Time) int {
	months := (now.Year()-deadline.Year())*12 + int(now.Month()-deadline.Month()) + 2
	if interval.Months > 0 {
		months = months/int(interval.Months) + 4
	}

	if months < 4 {
		months = 4
	}

	if months > maxCalendarCatchUp {
		return maxCalendarCatchUp
	}

	return months
}

// recurrenceAdvancesDeadline reports whether adding interval to deadline
// produces a later instant. RangeDuration treats a month as 30 days, so a
// mixed-sign span can look positive while calendar-month clamping moves
// the deadline backward (January 31 plus P1M-29D is January 30).
func recurrenceAdvancesDeadline(deadline time.Time, interval timespan.TimeSpan) bool {
	return interval.AddTo(deadline).After(deadline)
}
