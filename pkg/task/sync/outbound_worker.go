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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

type outboundHandler struct {
	svc        *Service
	pg         *pg.Client
	logger     *log.Logger
	staleAfter time.Duration
}

func NewOutboundWorker(
	svc *Service,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.TaskSyncJob] {
	h := &outboundHandler{
		svc:        svc,
		pg:         svc.pg,
		logger:     logger,
		staleAfter: 5 * time.Minute,
	}

	return worker.New(
		"task-sync-outbound",
		h,
		logger,
		opts...,
	)
}

func (h *outboundHandler) Claim(ctx context.Context) (coredata.TaskSyncJob, error) {
	var item coredata.TaskSyncJob

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := item.LoadNextPendingForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			item.Status = coredata.TaskSyncJobStatusProcessing
			item.StartedAt = &now
			item.UpdatedAt = now

			return item.Update(ctx, tx, coredata.NewNoScope())
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.TaskSyncJob{}, worker.ErrNoTask
		}

		return coredata.TaskSyncJob{}, err
	}

	return item, nil
}

func (h *outboundHandler) Process(ctx context.Context, item coredata.TaskSyncJob) error {
	if err := h.handle(ctx, &item); err != nil {
		if failErr := h.fail(ctx, &item, err); failErr != nil {
			h.logger.ErrorCtx(ctx, "cannot fail task sync job", log.Error(failErr))
		}

		return err
	}

	return nil
}

func (h *outboundHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleTaskSyncJobs(ctx, conn, h.staleAfter)
		},
	)
}

func (h *outboundHandler) handle(ctx context.Context, item *coredata.TaskSyncJob) error {
	var payload JobPayload
	if err := json.Unmarshal(item.Payload, &payload); err != nil {
		return fmt.Errorf("cannot unmarshal task sync payload: %w", err)
	}

	scope := coredata.NewScopeFromObjectID(item.ID)

	var (
		client *linear.Client
		task   *coredata.Task
	)

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			dbConnector := &coredata.Connector{}
			if err := dbConnector.LoadByID(ctx, tx, scope, payload.ConnectorID, h.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot load Linear connector: %w", err)
			}

			var err error

			client, _, err = h.svc.linearClientForConnector(ctx, tx, scope, dbConnector)
			if err != nil {
				return err
			}

			if payload.Action != SyncActionUpdate {
				return nil
			}

			task = &coredata.Task{}
			if err := task.LoadByID(ctx, tx, scope, payload.TaskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", payload.TaskID, err)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	states, err := client.ListWorkflowStates(ctx, payload.TeamID)
	if err != nil {
		return fmt.Errorf("cannot list Linear workflow states: %w", err)
	}

	switch payload.Action {
	case SyncActionCancel:
		stateID, err := PickWorkflowStateID(states, coredata.TaskStateCanceled)
		if err != nil {
			return err
		}

		_, err = client.UpdateIssue(
			ctx,
			payload.ExternalID,
			linear.IssueUpdateInput{StateID: &stateID},
		)
		if err != nil {
			return fmt.Errorf("cannot cancel Linear issue: %w", err)
		}
	case SyncActionUpdate:
		markdown, err := ContentToMarkdown(task.Content)
		if err != nil {
			return err
		}

		stateID, err := PickWorkflowStateID(states, task.State)
		if err != nil {
			return err
		}

		priority := TaskPriorityToLinear(task.Priority)
		dueDate := DeadlineToLinearDate(task.Deadline)

		issue, err := client.UpdateIssue(
			ctx,
			payload.ExternalID,
			linear.IssueUpdateInput{
				Title:       &task.Name,
				Description: &markdown,
				StateID:     &stateID,
				Priority:    &priority,
				DueDate:     dueDate,
				DueDateSet:  true,
			},
		)
		if err != nil {
			return fmt.Errorf("cannot update Linear issue: %w", err)
		}

		err = h.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				link := &coredata.TaskExternalLink{}
				if err := link.LoadByTaskID(ctx, tx, scope, payload.TaskID); err != nil {
					if errors.Is(err, coredata.ErrResourceNotFound) {
						return nil
					}

					return fmt.Errorf("cannot load task external link: %w", err)
				}

				hash := ContentHash(task.Name, markdown, task.State, task.Priority, task.Deadline)
				link.ContentHash = &hash
				link.RemoteUpdatedAt = &issue.UpdatedAt
				link.ExternalIdentifier = issue.Identifier
				link.ExternalURL = issue.URL
				link.UpdatedAt = time.Now()

				if err := link.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update task external link: %w", err)
				}

				return nil
			},
		)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot process task sync job: unknown action %q", payload.Action)
	}

	now := time.Now()
	item.Status = coredata.TaskSyncJobStatusSucceeded
	item.CompletedAt = &now
	item.UpdatedAt = now
	item.Error = nil

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return item.Update(ctx, tx, coredata.NewNoScope())
		},
	)
}

func (h *outboundHandler) fail(ctx context.Context, item *coredata.TaskSyncJob, processErr error) error {
	now := time.Now()
	message := processErr.Error()
	item.Error = &message
	item.UpdatedAt = now

	item.AttemptCount++
	if item.AttemptCount < coredata.TaskSyncJobDefaultMaxAttempts {
		next := now.Add(outboundRetryDelay(item.AttemptCount))
		item.Status = coredata.TaskSyncJobStatusPending
		item.NextAttemptAt = &next
		item.StartedAt = nil
		item.CompletedAt = nil
	} else {
		item.Status = coredata.TaskSyncJobStatusFailed
		item.CompletedAt = &now
	}

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return item.Update(ctx, tx, coredata.NewNoScope())
		},
	)
}

func outboundRetryDelay(attemptCount int) time.Duration {
	shift := min(max(attemptCount-1, 0), 5)

	return time.Duration(1<<uint(shift)) * time.Minute
}
