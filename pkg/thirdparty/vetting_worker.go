// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package thirdparty

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
	"go.probo.inc/probo/pkg/vetting"
)

type (
	vettingHandler struct {
		pg         *pg.Client
		vetter     Vetter
		logger     *log.Logger
		staleAfter time.Duration
	}

	VettingWorkerConfig struct {
		StaleAfter time.Duration
	}
)

var (
	_ worker.Handler[coredata.ThirdParty] = (*vettingHandler)(nil)
	_ worker.StaleRecoverer               = (*vettingHandler)(nil)
)

func NewVettingWorker(
	pgClient *pg.Client,
	vetter Vetter,
	logger *log.Logger,
	cfg VettingWorkerConfig,
	opts ...worker.Option,
) *worker.Worker[coredata.ThirdParty] {
	staleAfter := cfg.StaleAfter
	if staleAfter <= 0 {
		staleAfter = 25 * time.Minute
	}

	h := &vettingHandler{
		pg:         pgClient,
		vetter:     vetter,
		logger:     logger,
		staleAfter: staleAfter,
	}

	return worker.New(
		"vetting-worker",
		h,
		logger,
		opts...,
	)
}

func (h *vettingHandler) Claim(ctx context.Context) (coredata.ThirdParty, error) {
	var thirdParty coredata.ThirdParty

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := thirdParty.LoadNextPendingVettingForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			processing := coredata.ThirdPartyVettingStatusProcessing

			thirdParty.VettingStatus = &processing
			thirdParty.VettingProcessingStartedAt = &now
			thirdParty.VettingErrorMessage = nil
			thirdParty.UpdatedAt = now

			if err := thirdParty.Update(ctx, tx, coredata.NewNoPredicate()); err != nil {
				return fmt.Errorf("cannot update third party: %w", err)
			}

			return nil
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.ThirdParty{}, worker.ErrNoTask
		}

		return coredata.ThirdParty{}, err
	}

	return thirdParty, nil
}

func (h *vettingHandler) Process(ctx context.Context, thirdParty coredata.ThirdParty) error {
	if err := h.processThirdParty(ctx, &thirdParty); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"vetting worker failure",
			log.Error(err),
			log.String("third_party_id", thirdParty.ID.String()),
		)

		if failErr := h.failThirdParty(ctx, &thirdParty, err); failErr != nil {
			h.logger.ErrorCtx(ctx, "cannot mark third party vetting as failed", log.Error(failErr))
		}

		return err
	}

	return nil
}

func (h *vettingHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := coredata.ResetStaleVettingProcessing(ctx, conn, h.staleAfter); err != nil {
				return fmt.Errorf("cannot reset stale vetting processing: %w", err)
			}

			return nil
		},
	)
}

func (h *vettingHandler) processThirdParty(
	ctx context.Context,
	thirdParty *coredata.ThirdParty,
) error {
	if thirdParty.VettingWebsiteURL == nil {
		return fmt.Errorf("third party %s has no vetting website URL", thirdParty.ID)
	}

	procedure := ""
	if thirdParty.VettingProcedure != nil {
		procedure = *thirdParty.VettingProcedure
	}

	pc := &vetting.PersistenceContext{
		PG:             h.pg,
		ThirdPartyID:   thirdParty.ID,
		OrganizationID: thirdParty.OrganizationID,
		WebsiteURL:     *thirdParty.VettingWebsiteURL,
	}

	// Assessment runs outside any database transaction. Persistence tools
	// are not passed in so the agent cannot open DB transactions during
	// the long LLM/browser phase; results are written afterward.
	result, err := h.vetter.Assess(
		ctx,
		*thirdParty.VettingWebsiteURL,
		procedure,
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("cannot vet third party: %w", err)
	}

	if err := vetting.PersistAssessmentResult(ctx, pc, *result); err != nil {
		return fmt.Errorf("cannot persist vetting results: %w", err)
	}

	return h.commitVettingOutcome(
		ctx,
		thirdParty.ID,
		func(fresh *coredata.ThirdParty) {
			completed := coredata.ThirdPartyVettingStatusCompleted

			fresh.VettingStatus = &completed
			fresh.VettingProcessingStartedAt = nil
			fresh.VettingErrorMessage = nil
		},
	)
}

func (h *vettingHandler) failThirdParty(
	ctx context.Context,
	thirdParty *coredata.ThirdParty,
	reason error,
) error {
	errMsg := sanitizeVettingError(reason)

	return h.commitVettingOutcome(
		ctx,
		thirdParty.ID,
		func(fresh *coredata.ThirdParty) {
			failed := coredata.ThirdPartyVettingStatusFailed

			fresh.VettingStatus = &failed
			fresh.VettingProcessingStartedAt = nil
			fresh.VettingErrorMessage = &errMsg
		},
	)
}

func (h *vettingHandler) commitVettingOutcome(
	ctx context.Context,
	thirdPartyID gid.GID,
	apply func(*coredata.ThirdParty),
) error {
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			thirdParty := &coredata.ThirdParty{}

			if err := thirdParty.LoadByID(ctx, tx, coredata.NewNoPredicate(), thirdPartyID); err != nil {
				return fmt.Errorf("cannot reload third party: %w", err)
			}

			apply(thirdParty)
			thirdParty.UpdatedAt = time.Now()

			if err := thirdParty.Update(ctx, tx, coredata.NewNoPredicate()); err != nil {
				return fmt.Errorf("cannot update third party: %w", err)
			}

			return nil
		},
	)
}
