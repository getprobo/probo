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

package cookiebanner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/llm"
)

const defaultEnrichmentStaleAfter = 10 * time.Minute

type commonPatternEnrichmentHandler struct {
	pg              *pg.Client
	logger          *log.Logger
	enrichmentAgent *agent.Agent
	staleAfter      time.Duration
	agentTimeout    time.Duration
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
	cfg TrackerAgentsConfig,
	staleAfter time.Duration,
	opts ...worker.Option,
) *worker.Worker[coredata.CommonTrackerPattern] {
	if staleAfter <= 0 {
		staleAfter = defaultEnrichmentStaleAfter
	}

	agentTimeout := cfg.AgentTimeout
	if agentTimeout <= 0 {
		agentTimeout = defaultAgentTimeout
	}

	h := &commonPatternEnrichmentHandler{
		pg:           pgClient,
		logger:       logger,
		staleAfter:   staleAfter,
		agentTimeout: agentTimeout,
	}

	if cfg.LLMClient != nil {
		h.enrichmentAgent = buildCommonPatternEnrichmentAgent(cfg, pgClient, logger)
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
	if h.enrichmentAgent == nil {
		return nil
	}

	thirdPartyName, err := h.loadThirdPartyName(ctx, cp)
	if err != nil {
		return err
	}

	description, err := h.research(ctx, cp, thirdPartyName)
	if err != nil {
		return fmt.Errorf("cannot research tracker description: %w", err)
	}

	if description == "" {
		return fmt.Errorf("enrichment produced empty description for pattern %q", cp.Pattern)
	}

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := cp.SetEnriched(ctx, tx, description); err != nil {
				return fmt.Errorf("cannot set common tracker pattern enriched: %w", err)
			}

			var patterns coredata.TrackerPatterns

			count, err := patterns.BackfillDescriptionByCommonTrackerPatternID(ctx, tx, cp.ID, description)
			if err != nil {
				return err
			}

			h.logger.InfoCtx(
				ctx,
				"enriched common tracker pattern",
				log.String("common_tracker_pattern_id", cp.ID.String()),
				log.String("pattern", cp.Pattern),
				log.Int64("backfilled_tracker_patterns", count),
			)

			return nil
		},
	)
}

func (h *commonPatternEnrichmentHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := coredata.ResetStaleEnrichments(ctx, conn, h.staleAfter); err != nil {
				return fmt.Errorf("cannot reset stale common tracker pattern enrichments: %w", err)
			}

			return nil
		},
	)
}

func (h *commonPatternEnrichmentHandler) loadThirdPartyName(
	ctx context.Context,
	cp coredata.CommonTrackerPattern,
) (string, error) {
	if cp.CommonThirdPartyID == nil {
		return "", nil
	}

	var name string

	if err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var party coredata.CommonThirdParty
			if err := party.LoadByID(ctx, conn, *cp.CommonThirdPartyID); err != nil {
				return err
			}

			name = party.Name

			return nil
		},
	); err != nil {
		return "", fmt.Errorf("cannot load common third party for enrichment: %w", err)
	}

	return name, nil
}

func (h *commonPatternEnrichmentHandler) research(
	ctx context.Context,
	cp coredata.CommonTrackerPattern,
	thirdPartyName string,
) (string, error) {
	prompt := buildEnrichmentPrompt(cp, thirdPartyName)

	agentCtx, cancel := context.WithTimeout(ctx, h.agentTimeout)
	defer cancel()

	result, err := agent.RunTyped[CommonPatternEnrichmentResult](
		agentCtx,
		h.enrichmentAgent,
		[]llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: prompt}},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("enrichment agent run failed: %w", err)
	}

	return strings.TrimSpace(result.Output.Description), nil
}
