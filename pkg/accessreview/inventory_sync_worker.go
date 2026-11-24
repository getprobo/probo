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

package accessreview

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	staleInventorySyncMessage = "inventory sync timed out"

	staleInventorySyncRecoveryBatchSize = 100
)

type inventorySyncHandler struct {
	svc        *Service
	pg         *pg.Client
	logger     *log.Logger
	staleAfter time.Duration
}

func NewInventorySyncWorker(
	svc *Service,
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.AccessSourceSyncRun] {
	h := &inventorySyncHandler{
		svc:    svc,
		pg:     pgClient,
		logger: logger,
		// ListAccounts may take up to sourceFetchTimeout (2m); the roster
		// write can add more. Heartbeat after ListAccounts renews the lease;
		// this bound covers the remaining write without mid-txn pulses.
		staleAfter: 15 * time.Minute,
	}

	return worker.New(
		"inventory-sync-worker",
		h,
		logger,
		opts...,
	)
}

func (h *inventorySyncHandler) Claim(ctx context.Context) (coredata.AccessSourceSyncRun, error) {
	var run coredata.AccessSourceSyncRun

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := run.LoadNextQueuedForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			scope := coredata.NewScopeFromObjectID(run.ID)

			run.Status = coredata.AccessSourceSyncRunStatusRunning
			run.Error = nil
			run.StartedAt = &now
			run.CompletedAt = nil
			run.UpdatedAt = now

			if err := run.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update sync run status: %w", err)
			}

			return nil
		},
	); err != nil {
		if errors.Is(err, coredata.ErrNoAccessSourceSyncRunAvailable) {
			return coredata.AccessSourceSyncRun{}, worker.ErrNoTask
		}

		return coredata.AccessSourceSyncRun{}, fmt.Errorf("cannot claim sync run: %w", err)
	}

	return run, nil
}

func (h *inventorySyncHandler) Process(ctx context.Context, run coredata.AccessSourceSyncRun) error {
	return h.handle(ctx, &run)
}

func (h *inventorySyncHandler) RecoverStale(ctx context.Context) error {
	now := time.Now()
	staleThreshold := now.Add(-h.staleAfter)

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var stale coredata.AccessSourceSyncRuns
			if err := stale.LoadStaleRunningForUpdateSkipLocked(
				ctx,
				tx,
				staleThreshold,
				staleInventorySyncRecoveryBatchSize,
			); err != nil {
				return fmt.Errorf("cannot load stale inventory sync runs: %w", err)
			}

			for _, run := range stale {
				scope := coredata.NewScopeFromObjectID(run.ID)
				message := staleInventorySyncMessage

				run.Status = coredata.AccessSourceSyncRunStatusFailed
				run.GrantsUpsertedCount = 0
				run.Error = &message
				run.CompletedAt = &now
				run.UpdatedAt = now

				if err := run.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot fail stale sync run %s: %w", run.ID, err)
				}
			}

			if len(stale) > 0 {
				h.logger.InfoCtx(
					ctx,
					"swept stale inventory sync runs",
					log.Int("count", len(stale)),
				)
			}

			return nil
		},
	)
}

func (h *inventorySyncHandler) handle(
	ctx context.Context,
	run *coredata.AccessSourceSyncRun,
) error {
	scope := coredata.NewScopeFromObjectID(run.ID)

	if err := h.svc.SyncInventorySource(ctx, scope, run); err != nil {
		if commitErr := h.commitFailedSync(ctx, run, err); commitErr != nil {
			return fmt.Errorf("cannot sync inventory: %w, and cannot commit failed sync run: %w", err, commitErr)
		}

		return nil
	}

	return nil
}

func (h *inventorySyncHandler) commitFailedSync(
	ctx context.Context,
	run *coredata.AccessSourceSyncRun,
	failureErr error,
) error {
	h.logger.WarnCtx(
		ctx,
		"inventory sync failed",
		log.String("access_review_source_id", run.AccessReviewSourceID.String()),
		log.String("sync_run_id", run.ID.String()),
		log.Error(failureErr),
	)

	return h.failSync(ctx, run, fetchFailureMessage(failureErr))
}

func (h *inventorySyncHandler) failSync(
	ctx context.Context,
	run *coredata.AccessSourceSyncRun,
	message string,
) error {
	var (
		now   = time.Now()
		scope = coredata.NewScopeFromObjectID(run.ID)
	)

	run.Status = coredata.AccessSourceSyncRunStatusFailed
	run.GrantsUpsertedCount = 0
	run.Error = &message
	run.CompletedAt = &now
	run.UpdatedAt = now

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := run.Update(ctx, tx, scope); err != nil {
				return err
			}

			source := &coredata.AccessReviewSource{}
			if err := source.LoadByID(ctx, tx, scope, run.AccessReviewSourceID); err != nil {
				return fmt.Errorf("cannot load access source for sync backoff: %w", err)
			}

			return h.svc.scheduleInventorySyncAfterFailure(ctx, tx, scope, source, now)
		},
	)
}
