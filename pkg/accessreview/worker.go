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
	"go.probo.inc/probo/pkg/gid"
)

// sourceFetchFailureMessage is the generic, user-facing message persisted on a
// failed fetch attempt. The raw error is only ever written to the logs so that
// internal connector details are never surfaced through the API or UI.
const sourceFetchFailureMessage = "We couldn't fetch accounts from this source. Verify the source configuration and try again."

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
) *worker.Worker[coredata.AccessReviewCampaignSourceFetchAttempt] {
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

func (h *sourceFetchHandler) Claim(ctx context.Context) (coredata.AccessReviewCampaignSourceFetchAttempt, error) {
	var attempt coredata.AccessReviewCampaignSourceFetchAttempt

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := attempt.LoadNextQueuedForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			attempt.Status = coredata.AccessReviewCampaignSourceFetchStatusFetching
			attempt.Error = nil
			attempt.StartedAt = &now
			attempt.CompletedAt = nil
			attempt.UpdatedAt = now

			scope := coredata.NewScope(attempt.TenantID)
			if err := attempt.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update fetch attempt status: %w", err)
			}

			return nil
		},
	); err != nil {
		if errors.Is(err, coredata.ErrNoAccessReviewCampaignSourceFetchAttemptAvailable) {
			return coredata.AccessReviewCampaignSourceFetchAttempt{}, worker.ErrNoTask
		}

		return coredata.AccessReviewCampaignSourceFetchAttempt{}, fmt.Errorf("cannot claim fetch attempt: %w", err)
	}

	return attempt, nil
}

func (h *sourceFetchHandler) Process(ctx context.Context, attempt coredata.AccessReviewCampaignSourceFetchAttempt) error {
	return h.handle(ctx, &attempt)
}

func (h *sourceFetchHandler) RecoverStale(ctx context.Context) error {
	now := time.Now()
	staleThreshold := now.Add(-h.staleAfter)

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var attempts coredata.AccessReviewCampaignSourceFetchAttempts

			count, err := attempts.RecoverStale(ctx, tx, staleThreshold, now)
			if err != nil {
				return fmt.Errorf("cannot recover stale fetch attempts: %w", err)
			}

			if count > 0 {
				h.logger.InfoCtx(
					ctx,
					"recovered stale fetch attempts",
					log.Int("count", count),
				)
			}

			return nil
		},
	)
}

func (h *sourceFetchHandler) handle(
	ctx context.Context,
	attempt *coredata.AccessReviewCampaignSourceFetchAttempt,
) error {
	scope := coredata.NewScope(attempt.TenantID)

	campaignSource := &coredata.AccessReviewCampaignSource{}
	if err := h.loadCampaignSource(ctx, scope, attempt.AccessReviewCampaignSourceID, campaignSource); err != nil {
		commitErr := h.commitFailedSourceFetch(ctx, attempt, fmt.Errorf("cannot load campaign source: %w", err))
		if commitErr != nil {
			return fmt.Errorf("cannot load campaign source: %w, and cannot commit failed fetch attempt: %w", err, commitErr)
		}

		return fmt.Errorf("cannot load campaign source: %w", err)
	}

	campaign, err := h.svc.GetCampaign(ctx, scope, campaignSource.AccessReviewCampaignID)
	if err != nil {
		commitErr := h.commitFailedSourceFetch(ctx, attempt, fmt.Errorf("cannot load campaign: %w", err))
		if commitErr != nil {
			return fmt.Errorf("cannot load campaign: %w, and cannot commit failed fetch attempt: %w", err, commitErr)
		}

		return fmt.Errorf("cannot load campaign: %w", err)
	}

	count, err := h.svc.FetchSource(ctx, scope, campaign, campaignSource)
	if err != nil {
		if commitErr := h.commitFailedSourceFetch(ctx, attempt, err); commitErr != nil {
			return fmt.Errorf("cannot fetch source: %w, and cannot commit failed fetch attempt: %w", err, commitErr)
		}

		if finalizeErr := h.finalizeCampaignFetchLifecycle(ctx, attempt.TenantID, campaignSource.AccessReviewCampaignID); finalizeErr != nil {
			return fmt.Errorf("cannot finalize campaign after failed fetch attempt: %w", finalizeErr)
		}

		return nil
	}

	if err := h.commitSuccessfulSourceFetch(ctx, attempt, count); err != nil {
		return fmt.Errorf("cannot commit successful fetch attempt: %w", err)
	}

	if err := h.finalizeCampaignFetchLifecycle(ctx, attempt.TenantID, campaignSource.AccessReviewCampaignID); err != nil {
		return fmt.Errorf("cannot finalize campaign fetch lifecycle: %w", err)
	}

	return nil
}

func (h *sourceFetchHandler) loadCampaignSource(
	ctx context.Context,
	scope coredata.Scoper,
	campaignSourceID gid.GID,
	campaignSource *coredata.AccessReviewCampaignSource,
) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return campaignSource.LoadByID(ctx, conn, scope, campaignSourceID)
		},
	)
}

// commitFailedSourceFetch marks the in-flight attempt as failed with a generic,
// user-facing message and logs the raw error so the internal detail stays in the
// logs only.
func (h *sourceFetchHandler) commitFailedSourceFetch(
	ctx context.Context,
	attempt *coredata.AccessReviewCampaignSourceFetchAttempt,
	failureErr error,
) error {
	h.logger.WarnCtx(
		ctx,
		"source fetch failed but campaign can continue",
		log.String("access_review_campaign_source_id", attempt.AccessReviewCampaignSourceID.String()),
		log.String("fetch_attempt_id", attempt.ID.String()),
		log.Error(failureErr),
	)

	var (
		now    = time.Now()
		errMsg = sourceFetchFailureMessage
		scope  = coredata.NewScope(attempt.TenantID)
	)

	attempt.Status = coredata.AccessReviewCampaignSourceFetchStatusFailed
	attempt.Error = &errMsg
	attempt.CompletedAt = &now
	attempt.UpdatedAt = now

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return attempt.Update(ctx, tx, scope)
		},
	)
}

func (h *sourceFetchHandler) commitSuccessfulSourceFetch(
	ctx context.Context,
	attempt *coredata.AccessReviewCampaignSourceFetchAttempt,
	fetchedAccountsCount int,
) error {
	var (
		now   = time.Now()
		scope = coredata.NewScope(attempt.TenantID)
	)

	attempt.Status = coredata.AccessReviewCampaignSourceFetchStatusSuccess
	attempt.FetchedAccountsCount = fetchedAccountsCount
	attempt.Error = nil
	attempt.CompletedAt = &now
	attempt.UpdatedAt = now

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return attempt.Update(ctx, tx, scope)
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

			latest := coredata.AccessReviewCampaignSourceFetchAttempts{}
			if err := latest.LoadLatestByCampaignID(ctx, tx, scope, campaignID); err != nil {
				return fmt.Errorf("cannot load latest fetch attempts: %w", err)
			}

			if len(latest) == 0 {
				return nil
			}

			for _, attempt := range latest {
				if !attempt.Status.IsTerminal() {
					return nil
				}
			}

			campaign.Status = coredata.AccessReviewCampaignStatusPendingActions
			campaign.UpdatedAt = time.Now()

			return campaign.Update(ctx, tx, scope)
		},
	)
}
