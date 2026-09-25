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
	"sync"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

const (
	outboundHeartbeatInterval   = 30 * time.Second
	maxOutboundRevisionAttempts = 5
)

type (
	outboundHandler struct {
		svc               *Service
		pg                *pg.Client
		logger            *log.Logger
		staleAfter        time.Duration
		heartbeatInterval time.Duration
	}
)

func NewOutboundWorker(
	svc *Service,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.TaskSyncJob] {
	staleAfter := 5 * time.Minute
	h := &outboundHandler{
		svc:               svc,
		pg:                svc.pg,
		logger:            logger,
		staleAfter:        staleAfter,
		heartbeatInterval: outboundLeaseHeartbeatInterval(staleAfter, outboundHeartbeatInterval),
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
			return item.ClaimNextForUpdateSkipLocked(ctx, tx, time.Now())
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
	job := *h
	job.logger = h.logger.With(log.String("task_sync_job_id", item.ID.String()))

	runCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	stopHeartbeat := job.startHeartbeat(runCtx, cancel, item)
	defer stopHeartbeat()

	if err := job.handle(runCtx, &item); err != nil {
		if processingLeaseLost(runCtx, err) {
			job.logger.InfoCtx(ctx, "lost task sync job processing lease")

			return nil
		}

		if failErr := job.fail(ctx, &item, err); failErr != nil {
			if errors.Is(failErr, coredata.ErrProcessingLeaseLost) {
				job.logger.InfoCtx(ctx, "lost task sync job processing lease")

				return nil
			}

			job.logger.ErrorCtx(ctx, "cannot fail task sync job", log.Error(failErr))
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

	client, err := h.linearClient(ctx, scope, payload.ConnectorID)
	if err != nil {
		return err
	}

	switch payload.Action {
	case SyncActionCommentUpsert:
		return h.processCommentUpsert(ctx, item, payload, scope, client)
	case SyncActionCommentDelete:
		return h.processCommentDelete(ctx, item, payload, client)
	}

	states, err := client.ListWorkflowStates(ctx, payload.TeamID)
	if err != nil {
		return fmt.Errorf("cannot list Linear workflow states: %w", err)
	}

	switch payload.Action {
	case SyncActionCancel:
		return h.processCancel(ctx, item, payload, scope, client, states)
	case SyncActionUpdate:
		return h.processUpdate(ctx, item, payload, scope, client, states)
	default:
		return fmt.Errorf("cannot process task sync job: unknown action %q", payload.Action)
	}
}

func (h *outboundHandler) linearClient(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (*linear.Client, error) {
	var dbConnector *coredata.Connector

	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			dbConnector = &coredata.Connector{}
			if err := dbConnector.LoadByID(
				ctx,
				conn,
				scope,
				connectorID,
				h.svc.encryptionKey,
			); err != nil {
				return fmt.Errorf("cannot load Linear connector: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	client, _, err := h.svc.linearClientForConnector(ctx, scope, dbConnector)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (h *outboundHandler) processCancel(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload JobPayload,
	scope coredata.Scoper,
	client *linear.Client,
	states []linear.WorkflowState,
) error {
	skip, err := h.hasNewerOutboundJob(ctx, item, payload.TaskID, scope)
	if err != nil {
		return err
	}

	if skip {
		return h.succeed(ctx, item)
	}

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

	return h.succeed(ctx, item)
}

func (h *outboundHandler) processUpdate(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload JobPayload,
	scope coredata.Scoper,
	client *linear.Client,
	states []linear.WorkflowState,
) error {
	sent := false

	for range maxOutboundRevisionAttempts {
		task, destination, skip, err := h.loadOutboundUpdate(ctx, item, payload, scope, sent)
		if err != nil {
			return err
		}

		if skip {
			return h.succeed(ctx, item)
		}

		markdown, sentHash, err := outboundTaskHash(task)
		if err != nil {
			return err
		}

		stateID, err := PickWorkflowStateID(states, task.State)
		if err != nil {
			return err
		}

		priority, err := outboundLinearPriority(task.Priority, destination)
		if err != nil {
			return err
		}

		dueDate := DeadlineToLinearDate(task.Deadline)

		issue, err := client.UpdateIssue(
			ctx,
			payload.ExternalID,
			linear.IssueUpdateInput{
				Title:       &task.Name,
				Description: &markdown,
				StateID:     &stateID,
				Priority:    priority,
				DueDate:     dueDate,
				DueDateSet:  true,
				AssigneeID:  h.svc.linearAssigneeID(ctx, scope, client, task.AssignedToID),
			},
		)
		if err != nil {
			return fmt.Errorf("cannot update Linear issue: %w", err)
		}

		sent = true

		stable, err := h.finishOutboundUpdate(
			ctx,
			payload,
			scope,
			sentHash,
			issue,
			priority != nil,
		)
		if err != nil {
			return err
		}

		if stable {
			return h.succeed(ctx, item)
		}
	}

	return fmt.Errorf("task changed during Linear update")
}

func (h *outboundHandler) loadOutboundUpdate(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload JobPayload,
	scope coredata.Scoper,
	sent bool,
) (*coredata.Task, json.RawMessage, bool, error) {
	var (
		task        *coredata.Task
		destination json.RawMessage
		skip        bool
	)

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			task = &coredata.Task{}
			if err := task.LoadByID(ctx, tx, scope, payload.TaskID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					skip = true

					return nil
				}

				return fmt.Errorf("cannot load task %q: %w", payload.TaskID, err)
			}

			link, err := loadCurrentOutboundLink(ctx, tx, scope, payload.TaskID)
			if err != nil {
				return err
			}

			if !outboundLinkMatchesJob(link, payload) {
				skip = true

				return nil
			}

			destination = link.Destination

			if sent {
				return nil
			}

			newer, err := item.HasNewerOutboundJob(ctx, tx, scope, payload.TaskID, commentSyncActions())
			if err != nil {
				return err
			}

			skip = newer

			return nil
		},
	)
	if err != nil {
		return nil, nil, false, err
	}

	return task, destination, skip, nil
}

func (h *outboundHandler) finishOutboundUpdate(
	ctx context.Context,
	payload JobPayload,
	scope coredata.Scoper,
	sentHash string,
	issue *linear.Issue,
	sentPriority bool,
) (bool, error) {
	var stable bool

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			task := &coredata.Task{}
			if err := task.LoadByID(ctx, tx, scope, payload.TaskID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					stable = true

					return nil
				}

				return fmt.Errorf("cannot load task %q: %w", payload.TaskID, err)
			}

			_, currentHash, err := outboundTaskHash(task)
			if err != nil {
				return err
			}

			if currentHash != sentHash {
				return nil
			}

			link, err := loadCurrentOutboundLink(ctx, tx, scope, payload.TaskID)
			if err != nil {
				return err
			}

			if !outboundLinkMatchesJob(link, payload) {
				stable = true

				return nil
			}

			link.ContentHash = &sentHash
			link.RemoteUpdatedAt = &issue.UpdatedAt
			link.ExternalIdentifier = issue.Identifier
			link.ExternalURL = issue.URL

			if sentPriority {
				destination, err := clearLinearNoneFromDestination(link.Destination)
				if err != nil {
					return err
				}

				link.Destination = destination
			}

			link.UpdatedAt = time.Now()

			if err := link.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update task external link: %w", err)
			}

			stable = true

			return nil
		},
	)
	if err != nil {
		return false, err
	}

	return stable, nil
}

func (h *outboundHandler) hasNewerOutboundJob(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	taskID gid.GID,
	scope coredata.Scoper,
) (bool, error) {
	var newer bool

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var err error

			newer, err = item.HasNewerOutboundJob(ctx, tx, scope, taskID, commentSyncActions())

			return err
		},
	)
	if err != nil {
		return false, err
	}

	return newer, nil
}

func outboundTaskHash(task *coredata.Task) (string, string, error) {
	markdown, err := ContentToMarkdown(task.Content)
	if err != nil {
		return "", "", err
	}

	return markdown, ContentHash(
		task.Name,
		markdown,
		task.State,
		task.Priority,
		task.Deadline,
		task.AssignedToID,
	), nil
}

func (h *outboundHandler) succeed(ctx context.Context, item *coredata.TaskSyncJob) error {
	now := time.Now()
	item.Status = coredata.TaskSyncJobStatusSucceeded
	item.CompletedAt = &now
	item.UpdatedAt = now
	item.Error = nil

	return h.pg.WithTx(
		context.WithoutCancel(ctx),
		func(ctx context.Context, tx pg.Tx) error {
			return item.UpdateProcessingState(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(item.ID),
			)
		},
	)
}

func (h *outboundHandler) startHeartbeat(
	ctx context.Context,
	cancel context.CancelCauseFunc,
	item coredata.TaskSyncJob,
) func() {
	if item.ProcessingOwnerToken == nil || *item.ProcessingOwnerToken == "" {
		panic("task sync job heartbeat requires an owner token")
	}

	stop := make(chan struct{})
	exited := make(chan struct{})

	var once sync.Once

	lease := coredata.TaskSyncJob{
		ID:                   item.ID,
		ProcessingOwnerToken: item.ProcessingOwnerToken,
	}

	go func() {
		defer close(exited)

		ticker := time.NewTicker(h.heartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := h.pg.WithConn(
					context.WithoutCancel(ctx),
					func(ctx context.Context, conn pg.Querier) error {
						return lease.TouchLease(
							ctx,
							conn,
							coredata.NewScopeFromObjectID(lease.ID),
							time.Now(),
						)
					},
				)
				if err == nil {
					continue
				}

				select {
				case <-stop:
					return
				default:
				}

				if errors.Is(err, coredata.ErrProcessingLeaseLost) {
					cancel(coredata.ErrProcessingLeaseLost)
				} else {
					cancel(fmt.Errorf("cannot heartbeat task sync job: %w", err))
				}

				return
			}
		}
	}()

	return func() {
		once.Do(
			func() {
				close(stop)
				<-exited
			},
		)
	}
}

func processingLeaseLost(ctx context.Context, err error) bool {
	return errors.Is(err, coredata.ErrProcessingLeaseLost) ||
		errors.Is(context.Cause(ctx), coredata.ErrProcessingLeaseLost)
}

func loadCurrentOutboundLink(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	taskID gid.GID,
) (*coredata.TaskExternalLink, error) {
	link := &coredata.TaskExternalLink{}
	if err := link.LoadByTaskID(ctx, conn, scope, taskID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load task external link: %w", err)
	}

	return link, nil
}

func outboundLinkMatchesJob(link *coredata.TaskExternalLink, payload JobPayload) bool {
	return link != nil && link.ExternalID == payload.ExternalID
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
		context.WithoutCancel(ctx),
		func(ctx context.Context, tx pg.Tx) error {
			return item.UpdateProcessingState(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(item.ID),
			)
		},
	)
}

func outboundRetryDelay(attemptCount int) time.Duration {
	shift := min(max(attemptCount-1, 0), 5)

	return time.Duration(1<<uint(shift)) * time.Minute
}

func outboundLeaseHeartbeatInterval(staleAfter, interval time.Duration) time.Duration {
	if interval >= staleAfter {
		return max(staleAfter/2, time.Millisecond)
	}

	return interval
}
