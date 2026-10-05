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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/timespan"
)

// errRecurrenceCannotAdvance means this deadline and interval can never
// produce a later occurrence. Retrying the same row would block the queue.
var errRecurrenceCannotAdvance = errors.New("recurrence interval cannot advance the deadline")

func repeatLockedRecurringTask(
	ctx context.Context,
	conn pg.Tx,
	source *coredata.Task,
	now time.Time,
) (*coredata.Task, error) {
	if source.Recurrence == nil || source.Deadline == nil {
		return nil, fmt.Errorf("cannot repeat task %q without a recurrence interval and deadline", source.ID)
	}

	if source.Deadline.After(now) {
		return nil, fmt.Errorf("cannot repeat task %q before its deadline", source.ID)
	}

	previous := *source
	scope := coredata.NewScopeFromObjectID(source.ID)

	next, err := insertNextRecurringTask(ctx, conn, scope, source, now)
	if err != nil {
		return nil, err
	}

	source.Recurrence = nil
	source.UpdatedAt = now

	if err := source.Update(ctx, conn, scope); err != nil {
		return nil, fmt.Errorf("cannot clear recurrence on task %q: %w", source.ID, err)
	}

	if err := InsertCreatedActivity(ctx, conn, scope, next, nil, now); err != nil {
		return nil, fmt.Errorf("cannot record next task created event: %w", err)
	}

	if err := emitTaskCreated(ctx, conn, scope, next); err != nil {
		return nil, fmt.Errorf("cannot emit next task created webhook: %w", err)
	}

	if err := emitTaskUpdated(ctx, conn, scope, &previous, source, nil); err != nil {
		return nil, fmt.Errorf("cannot emit recurring task updated webhook: %w", err)
	}

	return next, nil
}

func insertNextRecurringTask(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	source *coredata.Task,
	now time.Time,
) (*coredata.Task, error) {
	if source.Recurrence == nil || source.Deadline == nil {
		return nil, fmt.Errorf("cannot clone a task without a recurrence interval and deadline")
	}

	interval := *source.Recurrence

	deadline, err := nextRecurrenceDeadline(*source.Deadline, interval, now)
	if err != nil {
		return nil, fmt.Errorf("cannot compute next deadline of task %q: %w", source.ID, err)
	}

	referenceID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot generate reference id: %w", err)
	}

	next := &coredata.Task{
		ID:             gid.New(source.OrganizationID.TenantID(), coredata.TaskEntityType),
		OrganizationID: source.OrganizationID,
		Name:           source.Name,
		Content:        source.Content,
		Priority:       source.Priority,
		ReferenceID:    "custom-task-" + referenceID.String(),
		TimeEstimate:   source.TimeEstimate,
		AssignedToID:   source.AssignedToID,
		Deadline:       &deadline,
		Recurrence:     &interval,
		State:          coredata.TaskStateTodo,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := next.Insert(ctx, conn, scope); err != nil {
		return nil, fmt.Errorf("cannot insert next recurring task: %w", err)
	}

	if err := (coredata.InternalControlTask{}).CopyFromTask(ctx, conn, scope, source.ID, next.ID, next.ReferenceID, now); err != nil {
		return nil, fmt.Errorf("cannot copy task internal controls: %w", err)
	}

	return next, nil
}

// Late runs collapse missed cycles into one occurrence. Clock intervals
// jump in O(1); calendar months binary-search to avoid spinning.
func nextRecurrenceDeadline(
	deadline time.Time,
	interval timespan.TimeSpan,
	now time.Time,
) (time.Time, error) {
	if interval.Compare(timespan.Zero) <= 0 {
		return time.Time{}, unrepeatableRecurrence(interval)
	}

	first := interval.AddTo(deadline)
	if !first.After(deadline) {
		return time.Time{}, unrepeatableRecurrence(interval)
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
		return time.Time{}, fmt.Errorf("%w: got %s", errRecurrenceCannotAdvance, interval)
	}

	elapsed := now.Sub(deadline)
	cycles := elapsed / interval

	next := deadline.Add(cycles * interval).Add(interval)
	if !next.After(now) {
		return time.Time{}, fmt.Errorf("%w: got %s", errRecurrenceCannotAdvance, interval)
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
		if errors.Is(err, errRecurrenceCannotAdvance) {
			return time.Time{}, fmt.Errorf("cannot repeat a deadline that stops advancing: %w", err)
		}

		// Scaling the catch-up bound can overflow a mixed-sign interval that
		// still has a representable later deadline. That stays a computation
		// failure so the worker does not clear the recurrence.
		overflow := err

		high, hi, err = latestComputableOccurrence(deadline, interval, hi)
		if err != nil {
			return time.Time{}, fmt.Errorf("cannot find a representable recurrence cycle: %w", err)
		}

		if !high.After(now) {
			return time.Time{}, fmt.Errorf(
				"cannot compute a recurrence deadline within a representable timespan: %w",
				overflow,
			)
		}
	}

	if !high.After(now) {
		return time.Time{}, unrepeatableRecurrence(interval)
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
			return time.Time{}, unrepeatableRecurrence(interval)
		}
	}

	return next, nil
}

// latestComputableOccurrence returns the greatest cycle count below limit
// that TimeSpan.Times can represent. limit itself already failed.
func latestComputableOccurrence(
	deadline time.Time,
	interval timespan.TimeSpan,
	limit int,
) (time.Time, int, error) {
	var (
		best    time.Time
		bestN   int
		lastErr error
	)

	lo := 1
	hi := limit

	for lo < hi {
		mid := lo + (hi-lo)/2

		candidate, err := occurrenceAfterCycles(deadline, interval, mid)
		if err != nil {
			lastErr = err
			hi = mid

			continue
		}

		best = candidate
		bestN = mid
		lo = mid + 1
	}

	if bestN == 0 {
		if lastErr == nil {
			return time.Time{}, 0, fmt.Errorf("cannot scale recurrence interval by %d", limit)
		}

		return time.Time{}, 0, fmt.Errorf("cannot scale recurrence interval: %w", lastErr)
	}

	return best, bestN, nil
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

// A month is 30 days in RangeDuration, so P1M-29D looks positive while
// January 31 plus that span clamps backward to January 30.
func recurrenceAdvancesDeadline(deadline time.Time, interval timespan.TimeSpan) bool {
	return interval.AddTo(deadline).After(deadline)
}

func unrepeatableRecurrence(interval timespan.TimeSpan) error {
	return fmt.Errorf("%w: got %s", errRecurrenceCannotAdvance, interval)
}
