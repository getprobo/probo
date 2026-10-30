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
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	// sourceFetchFailureMessage is the generic, user-facing message persisted on a
	// failed fetch attempt. The raw error is only ever written to the logs so that
	// internal connector details are never surfaced through the API or UI.
	sourceFetchFailureMessage = "We couldn't fetch accounts from this source. Verify the source configuration and try again."

	campaignNotInProgressFetchMessage = "The campaign is no longer in progress, so this source was not fetched."

	// maxSourceFetchAttempts bounds the total attempts of a source; past it, a
	// stale attempt is failed instead of re-queued.
	maxSourceFetchAttempts = 3

	// staleFetchRecoveryBatchSize caps the campaigns one stale sweep locks; the
	// rest wait for the next sweep.
	staleFetchRecoveryBatchSize = 100
)

// fetchFailureMessage is the user-facing message for a failed fetch: a refused
// setting's own explanation, which is Probo's fixed text, or the generic one.
func fetchFailureMessage(err error) string {
	if rejected, ok := errors.AsType[*drivers.SettingRejectedError](err); ok && rejected != nil {
		return rejected.Message
	}

	return sourceFetchFailureMessage
}

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

			scope := coredata.NewScope(attempt.TenantID, gid.Nil)
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
			// Campaigns are locked before their attempts, the order
			// DeleteCampaign takes them in.
			var campaigns coredata.AccessReviewCampaigns
			if err := campaigns.LoadWithStaleFetchAttemptsForUpdateSkipLocked(ctx, tx, staleThreshold, staleFetchRecoveryBatchSize); err != nil {
				return fmt.Errorf("cannot lock campaigns with stale fetch attempts: %w", err)
			}

			swept := 0

			for _, campaign := range campaigns {
				count, err := h.recoverStaleCampaignFetches(ctx, tx, campaign, staleThreshold, now)
				if err != nil {
					return fmt.Errorf("cannot recover stale fetches of campaign %s: %w", campaign.ID, err)
				}

				swept += count
			}

			if swept > 0 {
				h.logger.InfoCtx(
					ctx,
					"swept stale fetch attempts",
					log.Int("count", swept),
				)
			}

			return nil
		},
	)
}

// recoverStaleCampaignFetches fails the campaign's stale attempts and queues a
// retry for each one below maxSourceFetchAttempts while the campaign is in
// progress. Once a source exhausts its attempts the campaign is finalized.
func (h *sourceFetchHandler) recoverStaleCampaignFetches(
	ctx context.Context,
	tx pg.Tx,
	campaign *coredata.AccessReviewCampaign,
	staleThreshold time.Time,
	now time.Time,
) (int, error) {
	scope := coredata.NewScopeFromObjectID(campaign.ID)

	var stale coredata.AccessReviewCampaignSourceFetchAttempts
	if err := stale.LoadStaleByCampaignIDForUpdateSkipLocked(ctx, tx, scope, campaign.ID, staleThreshold); err != nil {
		return 0, fmt.Errorf("cannot load stale fetch attempts: %w", err)
	}

	var (
		staleMessage = "fetch timed out"
		exhausted    = false
	)

	for _, attempt := range stale {
		attempt.Status = coredata.AccessReviewCampaignSourceFetchStatusFailed
		attempt.Error = &staleMessage
		attempt.CompletedAt = &now
		attempt.UpdatedAt = now

		if err := attempt.Update(ctx, tx, scope); err != nil {
			return 0, fmt.Errorf("cannot fail stale fetch attempt: %w", err)
		}

		if attempt.AttemptNumber >= maxSourceFetchAttempts {
			h.logger.WarnCtx(
				ctx,
				"fetch attempt exhausted its retries",
				log.String("access_review_campaign_source_id", attempt.AccessReviewCampaignSourceID.String()),
				log.String("fetch_attempt_id", attempt.ID.String()),
			)

			exhausted = true

			continue
		}

		if campaign.Status != coredata.AccessReviewCampaignStatusInProgress {
			continue
		}

		retry := &coredata.AccessReviewCampaignSourceFetchAttempt{
			ID:                           gid.New(scope.GetTenantID(), coredata.AccessReviewCampaignSourceFetchAttemptEntityType),
			OrganizationID:               attempt.OrganizationID,
			AccessReviewCampaignSourceID: attempt.AccessReviewCampaignSourceID,
			Status:                       coredata.AccessReviewCampaignSourceFetchStatusQueued,
			CreatedAt:                    now,
			UpdatedAt:                    now,
		}

		if err := retry.Insert(ctx, tx, scope); err != nil {
			return 0, fmt.Errorf("cannot queue retry fetch attempt: %w", err)
		}
	}

	if exhausted {
		if err := finalizeCampaign(ctx, tx, scope, campaign.ID); err != nil {
			return 0, fmt.Errorf("cannot finalize campaign fetch lifecycle: %w", err)
		}
	}

	return len(stale), nil
}

func (h *sourceFetchHandler) handle(
	ctx context.Context,
	attempt *coredata.AccessReviewCampaignSourceFetchAttempt,
) error {
	scope := coredata.NewScope(attempt.TenantID, gid.Nil)

	campaignSource := &coredata.AccessReviewCampaignSource{}
	if err := h.loadCampaignSource(ctx, scope, attempt.AccessReviewCampaignSourceID, campaignSource); err != nil {
		commitErr := h.commitFailedSourceFetch(ctx, attempt, nil, fmt.Errorf("cannot load campaign source: %w", err))
		if commitErr != nil {
			return fmt.Errorf("cannot load campaign source: %w, and cannot commit failed fetch attempt: %w", err, commitErr)
		}

		return fmt.Errorf("cannot load campaign source: %w", err)
	}

	campaign, err := h.svc.GetCampaign(ctx, scope, campaignSource.AccessReviewCampaignID)
	if err != nil {
		commitErr := h.commitFailedSourceFetch(ctx, attempt, campaignSource, fmt.Errorf("cannot load campaign: %w", err))
		if commitErr != nil {
			return fmt.Errorf("cannot load campaign: %w, and cannot commit failed fetch attempt: %w", err, commitErr)
		}

		return fmt.Errorf("cannot load campaign: %w", err)
	}

	count, err := h.svc.FetchSource(ctx, scope, campaign, campaignSource)
	if errors.Is(err, errCampaignNotInProgress) {
		if err := h.failSourceFetch(ctx, attempt, campaignNotInProgressFetchMessage); err != nil {
			return fmt.Errorf("cannot commit skipped fetch attempt: %w", err)
		}

		return nil
	}

	if err != nil {
		if commitErr := h.commitFailedSourceFetch(ctx, attempt, campaignSource, err); commitErr != nil {
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

// commitFailedSourceFetch marks the in-flight attempt as failed with a
// user-facing message and logs the raw error so the internal detail stays in the
// logs only.
func (h *sourceFetchHandler) commitFailedSourceFetch(
	ctx context.Context,
	attempt *coredata.AccessReviewCampaignSourceFetchAttempt,
	campaignSource *coredata.AccessReviewCampaignSource,
	failureErr error,
) error {
	// campaignSource is nil when the snapshot itself failed to load.
	fields := []log.Attr{
		log.String("access_review_campaign_source_id", attempt.AccessReviewCampaignSourceID.String()),
		log.String("fetch_attempt_id", attempt.ID.String()),
		log.Error(failureErr),
	}
	if campaignSource != nil && campaignSource.ConnectorID != nil {
		fields = append(fields, log.String("connector_id", campaignSource.ConnectorID.String()))
	}

	h.logger.WarnCtx(
		ctx,
		"source fetch failed but campaign can continue",
		fields...,
	)

	return h.failSourceFetch(ctx, attempt, fetchFailureMessage(failureErr))
}

func (h *sourceFetchHandler) failSourceFetch(
	ctx context.Context,
	attempt *coredata.AccessReviewCampaignSourceFetchAttempt,
	message string,
) error {
	var (
		now   = time.Now()
		scope = coredata.NewScope(attempt.TenantID, gid.Nil)
	)

	attempt.Status = coredata.AccessReviewCampaignSourceFetchStatusFailed
	attempt.Error = &message
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
		scope = coredata.NewScope(attempt.TenantID, gid.Nil)
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
	scope := coredata.NewScope(tenantID, gid.Nil)

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return finalizeCampaign(ctx, tx, scope, campaignID)
		},
	)
}

// finalizeCampaign moves an in-progress campaign to PENDING_ACTIONS once the
// latest attempt of every source is terminal.
func finalizeCampaign(ctx context.Context, tx pg.Tx, scope coredata.Scoper, campaignID gid.GID) error {
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
}
