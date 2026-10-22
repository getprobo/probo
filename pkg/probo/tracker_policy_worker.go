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
