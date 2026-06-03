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
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
	"go.probo.inc/probo/pkg/thirdparty"
	"golang.org/x/sync/errgroup"
)

// CommonPatternEnricher fills descriptions on common_tracker_patterns
// using an agent with web search, optionally attributing a vendor first,
// then fans the result out to every linked tracker pattern. It holds the
// agent dependencies so the enrichment logic can run from either the
// background worker (one claimed row at a time) or synchronously over a
// known set of ids (e.g. proboctl). It is enrichment's single source of
// truth - the worker is a thin queue poller that delegates here.
type CommonPatternEnricher struct {
	pg              *pg.Client
	logger          *log.Logger
	enrichmentAgent *agent.Agent
	mappingAgent    *agent.Agent
	agentTimeout    time.Duration
}

// NewCommonPatternEnricher builds the enricher from the shared tracker
// agents config. When no LLM client is configured the agents are left nil
// and Enabled reports false; callers must gate on Enabled before running.
func NewCommonPatternEnricher(
	pgClient *pg.Client,
	logger *log.Logger,
	cfg TrackerAgentsConfig,
) *CommonPatternEnricher {
	agentTimeout := cfg.AgentTimeout
	if agentTimeout <= 0 {
		agentTimeout = defaultAgentTimeout
	}

	e := &CommonPatternEnricher{
		pg:           pgClient,
		logger:       logger,
		agentTimeout: agentTimeout,
	}

	if cfg.LLMClient != nil {
		e.enrichmentAgent = buildCommonPatternEnrichmentAgent(cfg, pgClient, logger)
		e.mappingAgent = buildTrackerMappingAgent(cfg, pgClient, logger)
	}

	return e
}

// Enabled reports whether an LLM-backed enrichment agent is configured.
func (e *CommonPatternEnricher) Enabled() bool {
	return e.enrichmentAgent != nil
}

// EnrichByIDs enriches each common tracker pattern id, with bounded
// concurrency, and returns the number successfully enriched. It is the
// synchronous entry point used by operator tooling: it completes when the
// work is done rather than arming the async queue. A concurrency <= 0 is
// treated as 1.
func (e *CommonPatternEnricher) EnrichByIDs(
	ctx context.Context,
	ids []gid.GID,
	concurrency int,
) (int, error) {
	if !e.Enabled() {
		return 0, fmt.Errorf("common pattern enricher is disabled: no LLM client configured")
	}

	if concurrency <= 0 {
		concurrency = 1
	}

	var enriched atomic.Int64

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	for _, id := range ids {
		g.Go(func() error {
			var cp coredata.CommonTrackerPattern

			if err := e.pg.WithConn(
				gctx,
				func(ctx context.Context, conn pg.Querier) error {
					return cp.LoadByID(ctx, conn, id)
				},
			); err != nil {
				return fmt.Errorf("cannot load common tracker pattern %s: %w", id, err)
			}

			if err := e.EnrichPattern(gctx, cp); err != nil {
				return fmt.Errorf("cannot enrich common tracker pattern %s: %w", id, err)
			}

			enriched.Add(1)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return int(enriched.Load()), err
	}

	return int(enriched.Load()), nil
}

// EnrichPattern researches a description for one common tracker pattern
// (attributing a vendor first when unlinked), records it, and fans it out
// to linked org patterns. A blank description is a terminal-for-now state:
// the row is marked enriched so stale recovery never re-queues it, while a
// later third-party link re-arms a vendor-informed second attempt.
func (e *CommonPatternEnricher) EnrichPattern(ctx context.Context, cp coredata.CommonTrackerPattern) error {
	if !e.Enabled() {
		return nil
	}

	thirdPartyName, err := e.loadThirdPartyName(ctx, cp)
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
		attribution, err = e.identifyThirdParty(ctx, cp)
		if err != nil {
			return err
		}

		if attribution != nil {
			thirdPartyName = attribution.ThirdPartyName
		}
	}

	description, err := e.research(ctx, cp, thirdPartyName)
	if err != nil {
		return fmt.Errorf("cannot research tracker description: %w", err)
	}

	return e.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			// Resolve or create the catalog vendor only for an unlinked
			// pattern; the mapping pipeline owns creation, so we reuse its
			// name+slug dedup and never create a duplicate or override an
			// existing link.
			var thirdPartyID *gid.GID

			if attribution != nil && cp.CommonThirdPartyID == nil {
				thirdPartyID, err = thirdparty.ResolveOrCreateCommonThirdParty(ctx, tx, e.logger, attribution.ThirdPartyName, attribution.Category)
				if err != nil {
					return fmt.Errorf("cannot resolve or create common third party: %w", err)
				}
			}

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

			e.logger.InfoCtx(
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

func (e *CommonPatternEnricher) loadThirdPartyName(
	ctx context.Context,
	cp coredata.CommonTrackerPattern,
) (string, error) {
	if cp.CommonThirdPartyID == nil {
		return "", nil
	}

	var name string

	if err := e.pg.WithConn(
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

func (e *CommonPatternEnricher) research(
	ctx context.Context,
	cp coredata.CommonTrackerPattern,
	thirdPartyName string,
) (string, error) {
	prompt := buildEnrichmentPrompt(cp, thirdPartyName)

	agentCtx, cancel := context.WithTimeout(ctx, e.agentTimeout)
	defer cancel()

	result, err := agent.RunTyped[CommonPatternEnrichmentResult](
		agentCtx,
		e.enrichmentAgent,
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
func (e *CommonPatternEnricher) identifyThirdParty(
	ctx context.Context,
	cp coredata.CommonTrackerPattern,
) (*TrackerMappingAgentResult, error) {
	if e.mappingAgent == nil {
		return nil, nil
	}

	prompt := buildCommonPatternIdentificationPrompt(cp)

	agentCtx, cancel := context.WithTimeout(ctx, e.agentTimeout)
	defer cancel()

	result, err := agent.RunTyped[TrackerMappingAgentResult](
		agentCtx,
		e.mappingAgent,
		[]llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: prompt}},
			},
		},
	)
	if err != nil {
		e.logger.WarnCtx(
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
