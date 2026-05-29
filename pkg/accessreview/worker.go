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
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	maxAllowedFailedSourceFetches = 1
)

type sourceFetchHandler struct {
	svc        *Service
	pg         *pg.Client
	logger     *log.Logger
	staleAfter time.Duration
}

func NewSourceFetchWorker(
	svc *Service,
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.AccessReviewCampaignSourceFetch] {
	h := &sourceFetchHandler{
		svc:        svc,
		pg:         pgClient,
		logger:     logger,
		staleAfter: 5 * time.Minute,
	}

	return worker.New(
		"source-fetch-worker",
		h,
		logger,
		opts...,
	)
}

func (h *sourceFetchHandler) Claim(ctx context.Context) (coredata.AccessReviewCampaignSourceFetch, error) {
	var sourceFetch coredata.AccessReviewCampaignSourceFetch

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := sourceFetch.LoadNextQueuedForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			sourceFetch.Status = coredata.AccessReviewCampaignSourceFetchStatusFetching
			sourceFetch.AttemptCount++
			sourceFetch.LastError = nil
			sourceFetch.StartedAt = new(now)
			sourceFetch.CompletedAt = nil
			sourceFetch.UpdatedAt = now

			scope := coredata.NewScope(sourceFetch.TenantID)
			if err := sourceFetch.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update source fetch status: %w", err)
			}

			return nil
		},
	); err != nil {
		if errors.Is(err, coredata.ErrNoAccessReviewCampaignSourceFetchAvailable) {
			return coredata.AccessReviewCampaignSourceFetch{}, worker.ErrNoTask
		}

		return coredata.AccessReviewCampaignSourceFetch{}, fmt.Errorf("cannot claim source fetch: %w", err)
	}

	return sourceFetch, nil
}

func (h *sourceFetchHandler) Process(ctx context.Context, sourceFetch coredata.AccessReviewCampaignSourceFetch) error {
	return h.handle(ctx, &sourceFetch)
}

func (h *sourceFetchHandler) RecoverStale(ctx context.Context) error {
	now := time.Now()
	staleThreshold := now.Add(-h.staleAfter)

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var fetches coredata.AccessReviewCampaignSourceFetches

			count, err := fetches.RecoverStale(ctx, tx, staleThreshold, now)
			if err != nil {
				return fmt.Errorf("cannot recover stale source fetches: %w", err)
			}

			if count > 0 {
				h.logger.InfoCtx(
					ctx,
					"recovered stale source fetches",
					log.Int64("count", count),
				)
			}

			return nil
		},
	)
}

func (h *sourceFetchHandler) handle(
	ctx context.Context,
	sourceFetch *coredata.AccessReviewCampaignSourceFetch,
) error {
	scope := coredata.NewScope(sourceFetch.TenantID)

	campaign, err := h.svc.Campaigns(scope).Get(ctx, sourceFetch.AccessReviewCampaignID)
	if err != nil {
		commitErr := h.commitFailedSourceFetch(
			ctx,
			sourceFetch,
			fmt.Errorf("cannot load campaign: %w", err),
		)
		if commitErr != nil {
			return fmt.Errorf("cannot load campaign: %w, and cannot commit failed source fetch: %w", err, commitErr)
		}

		return fmt.Errorf("cannot load campaign: %w", err)
	}

	count, err := h.svc.Engine(scope).FetchSource(ctx, campaign, sourceFetch.AccessSourceID)
	if err != nil {
		commitErr := h.commitFailedSourceFetch(ctx, sourceFetch, err)
		if commitErr != nil {
			return fmt.Errorf("cannot fetch source: %w, and cannot commit failed source fetch: %w", err, commitErr)
		}

		if finalizeErr := h.finalizeCampaignFetchLifecycle(ctx, sourceFetch.TenantID, sourceFetch.AccessReviewCampaignID); finalizeErr != nil {
			return fmt.Errorf("cannot finalize campaign after failed source fetch: %w", finalizeErr)
		}

		failedSourceFetchCount, countErr := h.failedSourceFetchCount(ctx, sourceFetch.TenantID, sourceFetch.AccessReviewCampaignID)
		if countErr != nil {
			return fmt.Errorf("cannot count failed source fetches: %w", countErr)
		}

		if isSourceFetchFailureTolerated(failedSourceFetchCount) {
			h.logger.WarnCtx(
				ctx,
				"source fetch failed but campaign can continue",
				log.String("campaign_id", sourceFetch.AccessReviewCampaignID.String()),
				log.String("access_source_id", sourceFetch.AccessSourceID.String()),
				log.Int("failed_source_fetch_count", failedSourceFetchCount),
				log.Int("max_allowed_failed_source_fetches", maxAllowedFailedSourceFetches),
				log.Error(err),
			)

			return nil
		}

		return fmt.Errorf("cannot fetch source: %w", err)
	}

	if err := h.commitSuccessfulSourceFetch(ctx, sourceFetch, count); err != nil {
		return fmt.Errorf("cannot commit successful source fetch: %w", err)
	}

	if err := h.finalizeCampaignFetchLifecycle(ctx, sourceFetch.TenantID, sourceFetch.AccessReviewCampaignID); err != nil {
		return fmt.Errorf("cannot finalize campaign fetch lifecycle: %w", err)
	}

	return nil
}

func (h *sourceFetchHandler) commitFailedSourceFetch(
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

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return sourceFetch.Update(ctx, tx, scope)
		},
	)
}

func (h *sourceFetchHandler) commitSuccessfulSourceFetch(
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

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return sourceFetch.Update(ctx, tx, scope)
		},
	)
}

func (h *sourceFetchHandler) finalizeCampaignFetchLifecycle(
	ctx context.Context,
	tenantID gid.TenantID,
	campaignID gid.GID,
) error {
	scope := coredata.NewScope(tenantID)

	return h.pg.WithTx(
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

			for _, fetch := range fetches {
				if !fetch.Status.IsTerminal() {
					return nil
				}
			}

			campaign.Status = coredata.AccessReviewCampaignStatusPendingActions
			campaign.UpdatedAt = time.Now()

			return campaign.Update(ctx, tx, scope)
		},
	)
}

func (h *sourceFetchHandler) failedSourceFetchCount(
	ctx context.Context,
	tenantID gid.TenantID,
	campaignID gid.GID,
) (int, error) {
	scope := coredata.NewScope(tenantID)
	failedSourceFetchCount := 0

	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			fetches := coredata.AccessReviewCampaignSourceFetches{}
			if err := fetches.LoadByCampaignID(ctx, conn, scope, campaignID); err != nil {
				return fmt.Errorf("cannot load source fetches: %w", err)
			}

			for _, fetch := range fetches {
				if fetch.Status == coredata.AccessReviewCampaignSourceFetchStatusFailed {
					failedSourceFetchCount++
				}
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return failedSourceFetchCount, nil
}

func isSourceFetchFailureTolerated(failedSourceFetchCount int) bool {
	return failedSourceFetchCount <= maxAllowedFailedSourceFetches
}
