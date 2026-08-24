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

package cookiebanner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/browser"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
	"go.probo.inc/probo/pkg/thirdparty"
)

// CommonPatternEnricher fills descriptions on common_tracker_patterns
// using an agent with web search, optionally attributing a vendor first,
// then fans the result out to every linked tracker pattern. It holds the
// agent dependencies so the enrichment logic can run from either the
// background worker (one claimed row at a time) or synchronously over a
// known set of ids (e.g. proboctl). It is enrichment's single source of
// truth - the worker is a thin queue poller that delegates here.
type CommonPatternEnricher struct {
	pg                *pg.Client
	logger            *log.Logger
	enrichmentCfg     TrackerEnrichmentAgentConfig
	mappingCfg        TrackerMappingAgentConfig
	enrichmentEnabled bool
	mappingEnabled    bool
	enrichmentTimeout time.Duration
	mappingTimeout    time.Duration
}

// NewCommonPatternEnricher builds the enricher from the enrichment and
// mapping agent configs. It runs the enrichment agent to research a
// description and reuses the mapping agent to attribute a vendor first,
// so it needs both configs. When the enrichment config has no LLM client
// the agents are left nil and Enabled reports false; callers must gate on
// Enabled before running.
func NewCommonPatternEnricher(
	pgClient *pg.Client,
	logger *log.Logger,
	enrichmentCfg TrackerEnrichmentAgentConfig,
	mappingCfg TrackerMappingAgentConfig,
) *CommonPatternEnricher {
	enrichmentTimeout := enrichmentCfg.Timeout
	if enrichmentTimeout <= 0 {
		enrichmentTimeout = defaultAgentTimeout
	}

	mappingTimeout := mappingCfg.Timeout
	if mappingTimeout <= 0 {
		mappingTimeout = defaultAgentTimeout
	}

	e := &CommonPatternEnricher{
		pg:                pgClient,
		logger:            logger,
		enrichmentCfg:     enrichmentCfg,
		mappingCfg:        mappingCfg,
		enrichmentEnabled: enrichmentCfg.LLMClient != nil,
		mappingEnabled:    mappingCfg.LLMClient != nil,
		enrichmentTimeout: enrichmentTimeout,
		mappingTimeout:    mappingTimeout,
	}

	return e
}

// Enabled reports whether an LLM-backed enrichment agent is configured.
func (e *CommonPatternEnricher) Enabled() bool {
	return e.enrichmentEnabled
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

	// Build one per-run browser shared by both agent sub-runs when a
	// Chrome endpoint is configured. The browser lets the agents open
	// cookie-database and cookie-policy pages to read the true setter and
	// ground a description; it is closed when this run returns.
	var browserTools []agent.Tool

	if e.enrichmentCfg.ChromeAddr != "" {
		webBrowser := browser.NewBrowser(ctx, e.enrichmentCfg.ChromeAddr)
		defer webBrowser.Close()

		browserTools = browser.NewReadOnlyToolset(webBrowser).Tools()
	}

	// Map before enriching: an unlinked pattern is run through the
	// mapping agent first so a confident vendor both seeds the enrichment
	// prompt and gets linked. Attribution stays the mapping pipeline's
	// job; the enricher only reuses it. An already-linked pattern skips
	// this entirely.
	var attribution *agentIdentification

	if cp.CommonThirdPartyID == nil {
		attribution, err = e.identifyThirdParty(ctx, cp, browserTools)
		if err != nil {
			return err
		}

		if attribution != nil && !attribution.firstParty {
			thirdPartyName = attribution.result.ThirdPartyName
		}
	}

	firstParty := attribution != nil && attribution.firstParty

	// A first-party verdict discards the description, so researching one
	// would burn an agent run and its web searches for nothing.
	var description string

	if !firstParty {
		description, err = e.research(ctx, cp, thirdPartyName, browserTools)
		if err != nil {
			return fmt.Errorf("cannot research tracker description: %w", err)
		}
	}

	if firstParty {
		return e.persistTerminalVerdict(ctx, cp, attribution)
	}

	alreadyLinked := cp.CommonThirdPartyID != nil

	return e.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			// Resolve or create the catalog vendor only for an unlinked
			// pattern; the mapping pipeline owns creation, so we reuse its
			// name+slug dedup and never create a duplicate or override an
			// existing link.
			var thirdPartyID *gid.GID

			if attribution != nil && cp.CommonThirdPartyID == nil {
				thirdPartyID, err = thirdparty.ResolveOrCreateCommonThirdParty(ctx, tx, e.logger, attribution.result.ThirdPartyName, attribution.result.Category)
				if err != nil {
					return fmt.Errorf("cannot resolve or create common third party: %w", err)
				}
			}

			linked := alreadyLinked || thirdPartyID != nil

			var attributionResult *TrackerMappingAgentResult

			if attribution != nil {
				attributionResult = &attribution.result
			}

			meta := buildCommonPatternEnrichmentMetadata(
				e.enrichmentCfg.Model,
				description,
				attributionResult,
				alreadyLinked,
				linked,
				time.Now(),
			)

			payload, err := json.Marshal(meta)
			if err != nil {
				return fmt.Errorf("cannot marshal common tracker pattern enrichment metadata: %w", err)
			}

			if err := cp.UpdateEnrichment(ctx, tx, description, thirdPartyID, payload); err != nil {
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
				log.Bool("third_party_linked", linked),
				log.Int("enrichment_attempts", cp.EnrichmentAttempts),
				log.Int64("backfilled_tracker_patterns", backfilled),
			)

			return nil
		},
	)
}

// persistTerminalVerdict records a no-vendor verdict, clears the
// vendor-naming description, and re-queues linked org patterns. The
// enrichment payload is still written so the stale-recovery sweep does
// not re-queue the row.
func (e *CommonPatternEnricher) persistTerminalVerdict(
	ctx context.Context,
	cp coredata.CommonTrackerPattern,
	attribution *agentIdentification,
) error {
	verdict := attribution.terminalVerdict
	if verdict == "" {
		verdict = coredata.CommonTrackerPatternAttributionFirstParty
	}

	meta := buildCommonPatternTerminalMetadata(e.enrichmentCfg.Model, verdict, time.Now())

	payload, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("cannot marshal common tracker pattern enrichment metadata: %w", err)
	}

	return e.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			ids := []gid.GID{cp.ID}

			var patterns coredata.CommonTrackerPatterns

			if _, err := patterns.SetAttributionByIDs(
				ctx,
				tx,
				ids,
				verdict,
			); err != nil {
				return fmt.Errorf("cannot record terminal verdict: %w", err)
			}

			// Blank the description and write the payload in one statement.
			// SetAttributionByIDs already nulled the vendor link, and the
			// nil third party here leaves that null in place.
			if err := cp.UpdateEnrichment(ctx, tx, "", nil, payload); err != nil {
				return fmt.Errorf("cannot set common tracker pattern enriched: %w", err)
			}

			var orgPatterns coredata.TrackerPatterns

			remapped, err := orgPatterns.RequestMappingForUncategorisedByCommonTrackerPatternIDs(ctx, tx, ids)
			if err != nil {
				return err
			}

			cleared, err := orgPatterns.ClearDescriptionForUncategorisedByCommonTrackerPatternIDs(ctx, tx, ids)
			if err != nil {
				return err
			}

			e.logger.InfoCtx(
				ctx,
				"recorded terminal verdict on common tracker pattern",
				log.String("common_tracker_pattern_id", cp.ID.String()),
				log.String("pattern", cp.Pattern),
				log.String("attribution", string(verdict)),
				log.Int("enrichment_attempts", cp.EnrichmentAttempts),
				log.Int64("remapped_tracker_patterns", remapped),
				log.Int64("cleared_tracker_patterns", cleared),
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
	browserTools []agent.Tool,
) (string, error) {
	prompt := buildEnrichmentPrompt(cp, thirdPartyName)

	enrichmentAgent := buildCommonPatternEnrichmentAgent(e.enrichmentCfg, e.pg, e.logger, browserTools)

	agentCtx, cancel := context.WithTimeout(ctx, e.enrichmentTimeout)
	defer cancel()

	result, err := agent.RunTyped[CommonPatternEnrichmentResult](
		agentCtx,
		enrichmentAgent,
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

// identifyThirdParty reuses the tracker-mapping agent to judge an unlinked
// catalog pattern. It performs no DB writes: it returns a guard-accepted
// vendor attribution, a terminal first-party verdict, or nil when the
// agent is unsure, leaving the caller to persist. A failed agent run is
// best-effort and non-fatal, mirroring the mapping worker's
// identifyWithAgent.
func (e *CommonPatternEnricher) identifyThirdParty(
	ctx context.Context,
	cp coredata.CommonTrackerPattern,
	browserTools []agent.Tool,
) (*agentIdentification, error) {
	if !e.mappingEnabled {
		return nil, nil
	}

	prompt := buildCommonPatternIdentificationPrompt(cp)

	mappingAgent := buildTrackerMappingAgent(e.mappingCfg, e.pg, e.logger, browserTools)

	agentCtx, cancel := context.WithTimeout(ctx, e.mappingTimeout)
	defer cancel()

	result, err := agent.RunTyped[TrackerMappingAgentResult](
		agentCtx,
		mappingAgent,
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

	ident, rejection := interpretEnrichmentAttribution(result.Output)
	if rejection != attributionAccepted {
		e.logger.InfoCtx(
			ctx,
			"discarded agent third-party attribution during enrichment",
			log.String("common_tracker_pattern_id", cp.ID.String()),
			log.String("pattern", cp.Pattern),
			log.String("reason", string(rejection)),
			log.String("third_party_name", result.Output.ThirdPartyName),
			log.Float64("third_party_confidence", result.Output.ThirdPartyConfidence),
			log.String("evidence_source", result.Output.EvidenceSource),
		)
	}

	return ident, nil
}

// interpretEnrichmentAttribution applies the shared catalog acceptance bar to
// a mapping-agent output and returns a vendor attribution, a terminal
// first-party verdict, or nil. Applying the same bar as the mapping worker is
// what stops an attribution the worker would discard entering the catalog
// through enrichment instead; only the scanned-site backstop is skipped, since
// a catalog pattern belongs to no site.
//
// A rejected vendor still yields a first-party verdict when the agent declared
// one: the guards judge the proposed name, and an artifact with no external
// recipient is a separate, terminal answer.
//
// Split from identifyThirdParty so the decision is testable without an LLM.
// The rejection is for logging only.
func interpretEnrichmentAttribution(
	out TrackerMappingAgentResult,
) (*agentIdentification, attributionRejection) {
	out.ThirdPartyName = strings.TrimSpace(out.ThirdPartyName)

	rejection := rejectVendorAttribution(out, attributionContext{})
	if rejection == attributionAccepted {
		return &agentIdentification{result: out}, attributionAccepted
	}

	if verdict := terminalVerdictFor(out); verdict != "" {
		return &agentIdentification{firstParty: true, terminalVerdict: verdict}, rejection
	}

	return nil, rejection
}
