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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
	"go.probo.inc/probo/pkg/thirdparty"
)

const defaultEnrichmentStaleAfter = 10 * time.Minute

type commonPatternEnrichmentHandler struct {
	pg              *pg.Client
	logger          *log.Logger
	enrichmentAgent *agent.Agent
	mappingAgent    *agent.Agent
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
		h.mappingAgent = buildTrackerMappingAgent(cfg, pgClient, logger)
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

	// Map before enriching: an unlinked pattern is run through the
	// mapping agent first so a confident vendor both seeds the enrichment
	// prompt and gets linked. Attribution stays the mapping pipeline's
	// job; the enricher only reuses it. An already-linked pattern skips
	// this entirely.
	var attribution *TrackerMappingAgentResult

	if cp.CommonThirdPartyID == nil {
		attribution, err = h.identifyThirdParty(ctx, cp)
		if err != nil {
			return err
		}

		if attribution != nil {
			thirdPartyName = attribution.ThirdPartyName
		}
	}

	description, err := h.research(ctx, cp, thirdPartyName)
	if err != nil {
		return fmt.Errorf("cannot research tracker description: %w", err)
	}

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			// Resolve or create the catalog vendor only for an unlinked
			// pattern; the mapping pipeline owns creation, so we reuse its
			// name+slug dedup and never create a duplicate or override an
			// existing link.
			var thirdPartyID *gid.GID

			if attribution != nil && cp.CommonThirdPartyID == nil {
				thirdPartyID, err = thirdparty.ResolveOrCreateCommonThirdParty(ctx, tx, h.logger, attribution.ThirdPartyName, attribution.Category)
				if err != nil {
					return fmt.Errorf("cannot resolve or create common third party: %w", err)
				}
			}

			// A blank description is recorded as a terminal-for-now state:
			// the row is marked enriched so the stale-recovery loop never
			// re-queues it, but a later third-party link (mapping worker)
			// re-arms enrichment for a vendor-informed second attempt.
			if err := cp.SetEnriched(ctx, tx, description, thirdPartyID); err != nil {
				return fmt.Errorf("cannot set common tracker pattern enriched: %w", err)
			}

			var backfilled int64

			if description != "" {
				var patterns coredata.TrackerPatterns

				backfilled, err = patterns.BackfillDescriptionByCommonTrackerPatternID(ctx, tx, cp.ID, description)
				if err != nil {
					return err
				}
			}

			h.logger.InfoCtx(
				ctx,
				"enriched common tracker pattern",
				log.String("common_tracker_pattern_id", cp.ID.String()),
				log.String("pattern", cp.Pattern),
				log.Bool("described", description != ""),
				log.Bool("third_party_linked", thirdPartyID != nil),
				log.Int64("backfilled_tracker_patterns", backfilled),
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

// identifyThirdParty reuses the tracker-mapping agent to attribute a
// vendor to an unlinked catalog pattern. It performs no DB writes: it
// returns the confident attribution (name, category, confidence) or nil
// when the agent is unsure, leaving the caller to resolve or create the
// catalog row. A failed agent run is best-effort and non-fatal,
// mirroring the mapping worker's identifyWithAgent.
func (h *commonPatternEnrichmentHandler) identifyThirdParty(
	ctx context.Context,
	cp coredata.CommonTrackerPattern,
) (*TrackerMappingAgentResult, error) {
	if h.mappingAgent == nil {
		return nil, nil
	}

	prompt := buildCommonPatternIdentificationPrompt(cp)

	agentCtx, cancel := context.WithTimeout(ctx, h.agentTimeout)
	defer cancel()

	result, err := agent.RunTyped[TrackerMappingAgentResult](
		agentCtx,
		h.mappingAgent,
		[]llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: prompt}},
			},
		},
	)
	if err != nil {
		h.logger.WarnCtx(
			ctx,
			"mapping agent identification failed during enrichment",
			log.Error(err),
			log.String("pattern", cp.Pattern),
		)

		return nil, nil
	}

	out := result.Output
	out.ThirdPartyName = strings.TrimSpace(out.ThirdPartyName)

	if out.ThirdPartyName == "" || out.ThirdPartyConfidence < agentThirdPartyConfidenceThreshold {
		return nil, nil
	}

	return &out, nil
}
