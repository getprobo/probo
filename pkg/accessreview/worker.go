// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package accessreview

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	SourceFetchWorker struct {
		svc            *Service
		pg             *pg.Client
		logger         *log.Logger
		interval       time.Duration
		staleAfter     time.Duration
		maxConcurrency int
	}

	SourceFetchWorkerOption func(*SourceFetchWorker)
)

func WithSourceFetchWorkerIntervalDuration(interval time.Duration) SourceFetchWorkerOption {
	return func(w *SourceFetchWorker) {
		w.interval = interval
	}
}

func WithSourceFetchWorkerStaleAfter(staleAfter time.Duration) SourceFetchWorkerOption {
	return func(w *SourceFetchWorker) {
		w.staleAfter = staleAfter
	}
}

func WithSourceFetchWorkerMaxConcurrency(maxConcurrency int) SourceFetchWorkerOption {
	return func(w *SourceFetchWorker) {
		w.maxConcurrency = maxConcurrency
	}
}

func NewSourceFetchWorker(
	svc *Service,
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...SourceFetchWorkerOption,
) *SourceFetchWorker {
	w := &SourceFetchWorker{
		svc:            svc,
		pg:             pgClient,
		logger:         logger,
		interval:       30 * time.Second,
		staleAfter:     5 * time.Minute,
		maxConcurrency: 20,
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

func (w *SourceFetchWorker) Run(ctx context.Context) error {
	var (
		wg     sync.WaitGroup
		sem    = make(chan struct{}, w.maxConcurrency)
		ticker = time.NewTicker(w.interval)
	)
	defer ticker.Stop()
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			nonCancelableCtx := context.WithoutCancel(ctx)
			w.recoverStaleRows(nonCancelableCtx)
			for {
				if err := w.processNext(ctx, sem, &wg); err != nil {
					if !errors.Is(err, coredata.ErrNoAccessReviewCampaignSourceFetchAvailable) {
						w.logger.ErrorCtx(nonCancelableCtx, "cannot claim item", log.Error(err))
					}
					break
				}
			}
		}
	}
}

func (w *SourceFetchWorker) processNext(
	ctx context.Context,
	sem chan struct{},
	wg *sync.WaitGroup,
) error {
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	var (
		sourceFetch      coredata.AccessReviewCampaignSourceFetch
		now              = time.Now()
		nonCancelableCtx = context.WithoutCancel(ctx)
	)

	if err := w.pg.WithTx(
		nonCancelableCtx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := sourceFetch.LoadNextQueuedForUpdateSkipLocked(nonCancelableCtx, tx); err != nil {
				return err // sentinel errors checked by caller
			}

			sourceFetch.Status = coredata.AccessReviewCampaignSourceFetchStatusFetching
			sourceFetch.AttemptCount++
			sourceFetch.LastError = nil
			sourceFetch.StartedAt = new(now)
			sourceFetch.CompletedAt = nil
			sourceFetch.UpdatedAt = now

			scope := coredata.NewScope(sourceFetch.TenantID)
			if err := sourceFetch.Update(nonCancelableCtx, tx, scope); err != nil {
				return fmt.Errorf("cannot update source fetch status: %w", err)
			}
			return nil
		},
	); err != nil {
		<-sem
		return fmt.Errorf("cannot claim source fetch: %w", err)
	}

	wg.Add(1)
	go func(sourceFetch coredata.AccessReviewCampaignSourceFetch) {
		defer wg.Done()
		defer func() { <-sem }()

		if err := w.handle(nonCancelableCtx, &sourceFetch); err != nil {
			w.logger.ErrorCtx(nonCancelableCtx, "cannot process source fetch", log.Error(err))
		}
	}(sourceFetch)

	return nil
}

func (w *SourceFetchWorker) handle(
	ctx context.Context,
	sourceFetch *coredata.AccessReviewCampaignSourceFetch,
) error {
	scope := coredata.NewScope(sourceFetch.TenantID)

	campaign, err := w.svc.Campaigns(scope).Get(ctx, sourceFetch.AccessReviewCampaignID)
	if err != nil {
		commitErr := w.commitFailedSourceFetch(
			ctx,
			sourceFetch,
			fmt.Errorf("cannot load campaign: %w", err),
		)
		if commitErr != nil {
			return fmt.Errorf("cannot load campaign: %w, and cannot commit failed source fetch: %w", err, commitErr)
		}
		return fmt.Errorf("cannot load campaign: %w", err)
	}

	count, err := w.svc.Engine(scope).FetchSource(ctx, campaign, sourceFetch.AccessSourceID)
	if err != nil {
		commitErr := w.commitFailedSourceFetch(ctx, sourceFetch, err)
		if commitErr != nil {
			return fmt.Errorf("cannot fetch source: %w, and cannot commit failed source fetch: %w", err, commitErr)
		}

		if finalizeErr := w.finalizeCampaignFetchLifecycle(ctx, sourceFetch.TenantID, sourceFetch.AccessReviewCampaignID); finalizeErr != nil {
			return fmt.Errorf("cannot finalize campaign after failed source fetch: %w", finalizeErr)
		}
		return fmt.Errorf("cannot fetch source: %w", err)
	}

	if err := w.commitSuccessfulSourceFetch(ctx, sourceFetch, count); err != nil {
		return fmt.Errorf("cannot commit successful source fetch: %w", err)
	}

	if err := w.finalizeCampaignFetchLifecycle(ctx, sourceFetch.TenantID, sourceFetch.AccessReviewCampaignID); err != nil {
		return fmt.Errorf("cannot finalize campaign fetch lifecycle: %w", err)
	}

	return nil
}

func (w *SourceFetchWorker) recoverStaleRows(ctx context.Context) {
	now := time.Now()
	staleThreshold := now.Add(-w.staleAfter)

	err := w.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var fetches coredata.AccessReviewCampaignSourceFetches
			count, err := fetches.RecoverStale(ctx, tx, staleThreshold, now)
			if err != nil {
				return fmt.Errorf("cannot recover stale source fetches: %w", err)
			}

			if count > 0 {
				w.logger.InfoCtx(
					ctx,
					"recovered stale source fetches",
					log.Int64("count", count),
				)
			}

			return nil
		},
	)
	if err != nil {
		w.logger.ErrorCtx(ctx, "cannot recover stale rows", log.Error(err))
	}
}

func (w *SourceFetchWorker) commitFailedSourceFetch(
	ctx context.Context,
	sourceFetch *coredata.AccessReviewCampaignSourceFetch,
	failureErr error,
) error {
	var (
		now    = time.Now()
		errMsg = failureErr.Error()
		scope  = coredata.NewScopeFromObjectID(sourceFetch.AccessReviewCampaignID)
	)

	sourceFetch.Status = coredata.AccessReviewCampaignSourceFetchStatusFailed
	sourceFetch.LastError = &errMsg
	sourceFetch.CompletedAt = new(now)
	sourceFetch.UpdatedAt = now

	return w.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return sourceFetch.Update(ctx, tx, scope)
		},
	)
}

func (w *SourceFetchWorker) commitSuccessfulSourceFetch(
	ctx context.Context,
	sourceFetch *coredata.AccessReviewCampaignSourceFetch,
	fetchedAccountsCount int,
) error {
	var (
		now   = time.Now()
		scope = coredata.NewScopeFromObjectID(sourceFetch.AccessReviewCampaignID)
	)

	sourceFetch.Status = coredata.AccessReviewCampaignSourceFetchStatusSuccess
	sourceFetch.FetchedAccountsCount = fetchedAccountsCount
	sourceFetch.LastError = nil
	sourceFetch.CompletedAt = new(now)
	sourceFetch.UpdatedAt = now

	return w.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return sourceFetch.Update(ctx, tx, scope)
		},
	)
}

func (w *SourceFetchWorker) finalizeCampaignFetchLifecycle(
	ctx context.Context,
	tenantID gid.TenantID,
	campaignID gid.GID,
) error {
	scope := coredata.NewScope(tenantID)

	return w.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, tx, scope, campaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			campaign := &coredata.AccessReviewCampaign{}
			if err := campaign.LoadByID(ctx, tx, scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusInProgress {
				return nil
			}

			fetches := coredata.AccessReviewCampaignSourceFetches{}
			if err := fetches.LoadByCampaignID(ctx, tx, scope, campaignID); err != nil {
				return fmt.Errorf("cannot load source fetches: %w", err)
			}

			if len(fetches) == 0 {
				return nil
			}

			hasFailure := false
			for _, fetch := range fetches {
				if !fetch.Status.IsTerminal() {
					return nil
				}
				if fetch.Status == coredata.AccessReviewCampaignSourceFetchStatusFailed {
					hasFailure = true
				}
			}

			if hasFailure {
				campaign.Status = coredata.AccessReviewCampaignStatusFailed
			} else {
				campaign.Status = coredata.AccessReviewCampaignStatusPendingActions
			}

			campaign.UpdatedAt = time.Now()
			return campaign.Update(ctx, tx, scope)
		},
	)
}
