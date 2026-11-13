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
	asanasvc "go.probo.inc/probo/pkg/task/sync/asana"
	"go.probo.inc/probo/pkg/task/sync/internal"
	"go.probo.inc/probo/pkg/task/sync/job"
	linearsvc "go.probo.inc/probo/pkg/task/sync/linear"
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
		heartbeatInterval: internal.OutboundLeaseHeartbeatInterval(staleAfter, outboundHeartbeatInterval),
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
		if internal.ProcessingLeaseLost(runCtx, err) {
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
	var payload job.JobPayload
	if err := json.Unmarshal(item.Payload, &payload); err != nil {
		return fmt.Errorf("cannot unmarshal task sync payload: %w", err)
	}

	if payload.Provider == coredata.ConnectorProviderAsana {
		return asanasvc.HandleOutbound(
			h.svc.Asana,
			asanasvc.OutboundHooks{
				Newer: h.hasNewerOutboundJob,
				Done:  h.succeed,
			},
			ctx,
			item,
			payload,
		)
	}

	return linearsvc.HandleOutbound(
		h.svc.Linear,
		linearsvc.OutboundHooks{
			Newer: h.hasNewerOutboundJob,
			Done:  h.succeed,
		},
		ctx,
		item,
		payload,
	)
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

			newer, err = item.HasNewerOutboundJob(ctx, tx, scope, taskID, job.CommentSyncActions())

			return err
		},
	)
	if err != nil {
		return false, err
	}

	return newer, nil
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
		panic("task sync job touch requires an owner token")
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
						return lease.Touch(
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
					cancel(fmt.Errorf("cannot touch task sync job: %w", err))
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
