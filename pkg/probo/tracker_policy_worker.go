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

package probo

import (
	"context"
	"errors"
	"fmt"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

type trackerPolicyHandler struct {
	generatedDocuments *GeneratedDocumentService
	pg                 *pg.Client
	logger             *log.Logger
}

// NewTrackerPolicyWorker returns a worker that regenerates a banner's cookie
// and tracking technologies policy document whenever a banner version is
// published. Publishing sets policy_generation_requested_at on the banner; this
// worker claims those banners, clears the flag, and rebuilds the policy from
// the latest published snapshot.
func NewTrackerPolicyWorker(
	generatedDocuments *GeneratedDocumentService,
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.CookieBanner] {
	h := &trackerPolicyHandler{
		generatedDocuments: generatedDocuments,
		pg:                 pgClient,
		logger:             logger,
	}

	return worker.New(
		"tracker-policy-worker",
		h,
		logger,
		opts...,
	)
}

func (h *trackerPolicyHandler) Claim(ctx context.Context) (coredata.CookieBanner, error) {
	var banner coredata.CookieBanner

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := banner.LoadNextForPolicyGenerationForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			return banner.ClearPolicyGenerationRequestedAt(ctx, tx)
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.CookieBanner{}, worker.ErrNoTask
		}

		return coredata.CookieBanner{}, fmt.Errorf("cannot claim tracker policy task: %w", err)
	}

	return banner, nil
}

func (h *trackerPolicyHandler) Process(ctx context.Context, banner coredata.CookieBanner) error {
	scope := coredata.NewScopeFromObjectID(banner.ID)

	if err := h.generatedDocuments.PublishTrackerPolicy(ctx, scope, banner.ID); err != nil {
		// A banner can lose its published version between the publish that
		// armed the flag and this run (e.g. it was deleted). There is nothing
		// to generate in that case, so skip rather than fail the task.
		if errors.Is(err, coredata.ErrResourceNotFound) {
			h.logger.InfoCtx(
				ctx,
				"skipping tracker policy generation: no published version",
				log.String("banner_id", banner.ID.String()),
			)

			return nil
		}

		return fmt.Errorf("cannot generate tracker policy: %w", err)
	}

	h.logger.InfoCtx(
		ctx,
		"generated tracker policy document",
		log.String("banner_id", banner.ID.String()),
	)

	return nil
}
