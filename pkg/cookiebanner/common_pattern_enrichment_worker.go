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

package cookiebanner

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
	defaultEnrichmentStaleAfter = 10 * time.Minute

	// defaultEnrichmentMaxAttempts caps how many times a row is retried
	// before stale recovery leaves it alone, so a permanently failing row
	// does not loop forever.
	defaultEnrichmentMaxAttempts = 3
)

// commonPatternEnrichmentHandler is the queue poller for common tracker
// pattern enrichment. It owns only the claim/dequeue and stale-recovery
// mechanics; the enrichment work itself lives in CommonPatternEnricher so
// it can also run synchronously from operator tooling.
type commonPatternEnrichmentHandler struct {
	pg          *pg.Client
	logger      *log.Logger
	enricher    *CommonPatternEnricher
	staleAfter  time.Duration
	maxAttempts int
}

// NewCommonPatternEnrichmentWorker builds the worker that fills
// descriptions on common_tracker_patterns using an agent with web
// search, then fans the result out to every linked tracker pattern. It is
// a global system worker: common_tracker_patterns is not tenant-scoped,
// so a single enrichment benefits all tenants. The worker no-ops when no
// LLM client is configured; callers should gate registration on config
// presence.
func NewCommonPatternEnrichmentWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	enrichmentCfg TrackerEnrichmentAgentConfig,
	mappingCfg TrackerMappingAgentConfig,
	staleAfter time.Duration,
	maxAttempts int,
	opts ...worker.Option,
) *worker.Worker[coredata.CommonTrackerPattern] {
	if staleAfter <= 0 {
		staleAfter = defaultEnrichmentStaleAfter
	}

	if maxAttempts <= 0 {
		maxAttempts = defaultEnrichmentMaxAttempts
	}

	h := &commonPatternEnrichmentHandler{
		pg:          pgClient,
		logger:      logger,
		enricher:    NewCommonPatternEnricher(pgClient, logger, enrichmentCfg, mappingCfg),
		staleAfter:  staleAfter,
		maxAttempts: maxAttempts,
	}

	return worker.New(
		"common-pattern-enrichment-worker",
		h,
		logger,
		opts...,
	)
}

func (h *commonPatternEnrichmentHandler) Claim(ctx context.Context) (coredata.CommonTrackerPattern, error) {
	var cp coredata.CommonTrackerPattern

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := cp.LoadNextForEnrichmentForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			return cp.ClearEnrichmentRequestedAt(ctx, tx)
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.CommonTrackerPattern{}, worker.ErrNoTask
		}

		return coredata.CommonTrackerPattern{}, fmt.Errorf("cannot claim common tracker pattern enrichment task: %w", err)
	}

	return cp, nil
}

func (h *commonPatternEnrichmentHandler) Process(ctx context.Context, cp coredata.CommonTrackerPattern) error {
	if !h.enricher.Enabled() {
		return nil
	}

	return h.enricher.EnrichPattern(ctx, cp)
}

func (h *commonPatternEnrichmentHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := coredata.ResetStaleEnrichments(ctx, conn, h.staleAfter, h.maxAttempts); err != nil {
				return fmt.Errorf("cannot reset stale common tracker pattern enrichments: %w", err)
			}

			return nil
		},
	)
}
