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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type recurringTaskHandler struct {
	service *Service
	logger  *log.Logger
}

var _ worker.Handler[coredata.Task] = (*recurringTaskHandler)(nil)

// NewRecurringTaskWorker builds the worker that advances recurring tasks past
// their deadline: it detaches the overdue task from its recurrence series and
// spawns the next occurrence, so the series keeps moving forward on its own
// schedule regardless of whether the overdue task was ever completed.
func NewRecurringTaskWorker(
	service *Service,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.Task] {
	h := &recurringTaskHandler{
		service: service,
		logger:  logger,
	}

	return worker.New(
		"recurring-task-worker",
		h,
		logger,
		opts...,
	)
}

func (h *recurringTaskHandler) Claim(ctx context.Context) (coredata.Task, error) {
	var nextTask coredata.Task

	err := h.service.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			task := &coredata.Task{}
			now := time.Now()

			if err := task.LoadNextOverdueRecurringForUpdateSkipLocked(ctx, tx, now); err != nil {
				return err
			}

			unit := *task.RecurrenceIntervalUnit
			count := *task.RecurrenceIntervalCount

			deadline, err := nextRecurrenceDeadline(*task.Deadline, unit, count, now)
			if err != nil {
				return fmt.Errorf("cannot compute next deadline of task %q: %w", task.ID, err)
			}

			task.RecurrenceIntervalUnit = nil
			task.RecurrenceIntervalCount = nil
			task.UpdatedAt = now

			if err := task.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot detach recurring task: %w", err)
			}

			referenceID, err := uuid.NewV4()
			if err != nil {
				return fmt.Errorf("cannot generate reference id: %w", err)
			}

			nextTask = coredata.Task{
				ID:                      gid.New(task.OrganizationID.TenantID(), coredata.TaskEntityType),
				OrganizationID:          task.OrganizationID,
				MeasureID:               task.MeasureID,
				Name:                    task.Name,
				Description:             task.Description,
				Priority:                task.Priority,
				ReferenceID:             "custom-task-" + referenceID.String(),
				TimeEstimate:            task.TimeEstimate,
				AssignedToID:            task.AssignedToID,
				Deadline:                &deadline,
				RecurrenceIntervalUnit:  &unit,
				RecurrenceIntervalCount: &count,
				State:                   coredata.TaskStateTodo,
				CreatedAt:               now,
				UpdatedAt:               now,
			}

			scope := coredata.NewScopeFromObjectID(task.OrganizationID)

			if err := nextTask.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert next recurring task: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.Task{}, worker.ErrNoTask
		}

		return coredata.Task{}, err
	}

	return nextTask, nil
}

func (h *recurringTaskHandler) Process(ctx context.Context, task coredata.Task) error {
	h.logger.InfoCtx(
		ctx,
		"generated next occurrence of recurring task",
		log.String("task_id", task.ID.String()),
		log.Time("deadline", *task.Deadline),
	)

	return nil
}

// nextRecurrenceDeadline advances deadline by count units of unit, repeating
// until the result is after now. This collapses any number of missed cycles
// (e.g. after a long outage) into a single upcoming occurrence instead of
// backfilling one task per missed cycle.
//
// It errors on any recurrence that cannot move a deadline forward — a
// non-positive count or a unit this function does not handle — because such a
// recurrence would otherwise spin the loop forever.
func nextRecurrenceDeadline(
	deadline time.Time,
	unit coredata.TaskRecurrenceIntervalUnit,
	count int,
	now time.Time,
) (time.Time, error) {
	if count < 1 {
		return time.Time{}, fmt.Errorf("recurrence interval count must be at least 1, got %d", count)
	}

	next := deadline

	for !next.After(now) {
		switch unit {
		case coredata.TaskRecurrenceIntervalUnitDay:
			next = next.AddDate(0, 0, count)
		case coredata.TaskRecurrenceIntervalUnitWeek:
			next = next.AddDate(0, 0, 7*count)
		case coredata.TaskRecurrenceIntervalUnitMonth:
			next = next.AddDate(0, count, 0)
		case coredata.TaskRecurrenceIntervalUnitYear:
			next = next.AddDate(count, 0, 0)
		default:
			return time.Time{}, fmt.Errorf("unhandled recurrence interval unit %q", unit)
		}
	}

	return next, nil
}
