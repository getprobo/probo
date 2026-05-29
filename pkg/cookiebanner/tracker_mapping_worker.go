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
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
	"go.probo.inc/probo/pkg/thirdparty"
	"go.probo.inc/probo/pkg/uri"
)

type trackerMappingHandler struct {
	pg                    *pg.Client
	logger                *log.Logger
	mappingAgent          *agent.Agent
	disambiguationAgent   *agent.Agent
	agentTimeout          time.Duration
	disambiguationTimeout time.Duration
}

func NewTrackerMappingWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	mappingCfg TrackerAgentsConfig,
	disambiguationCfg thirdparty.DisambiguationConfig,
	opts ...worker.Option,
) *worker.Worker[coredata.TrackerPattern] {
	agentTimeout := mappingCfg.AgentTimeout
	if agentTimeout <= 0 {
		agentTimeout = defaultAgentTimeout
	}

	h := &trackerMappingHandler{
		pg:                    pgClient,
		logger:                logger,
		agentTimeout:          agentTimeout,
		disambiguationTimeout: disambiguationCfg.Timeout,
	}

	if mappingCfg.LLMClient != nil {
		h.mappingAgent = buildTrackerMappingAgent(mappingCfg, pgClient, logger)
	}

	if disambiguationCfg.LLMClient != nil {
		h.disambiguationAgent = thirdparty.BuildDisambiguationAgent(disambiguationCfg, logger)
	}

	return worker.New(
		"tracker-mapping-worker",
		h,
		logger,
		opts...,
	)
}

func (h *trackerMappingHandler) Claim(ctx context.Context) (coredata.TrackerPattern, error) {
	var tp coredata.TrackerPattern

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := tp.LoadNextForMappingForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			return tp.ClearMappingRequestedAt(ctx, tx)
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.TrackerPattern{}, worker.ErrNoTask
		}

		return coredata.TrackerPattern{}, fmt.Errorf("cannot claim tracker mapping task: %w", err)
	}

	return tp, nil
}

// catalogMatch is the result of a single catalog signal. commonPatternID
// is the catalog row the signal resolved (or backfilled); commonThirdPartyID
// is the catalog third party the signal discovered, when any; thirdPartyID
// is an existing org ThirdParty the signal knows directly (e.g. a sibling
// pattern already promoted in the same organization). A nil *catalogMatch
// means the signal produced nothing.
type catalogMatch struct {
	commonPatternID    *gid.GID
	commonThirdPartyID *gid.GID
	thirdPartyID       *gid.GID
}

// Process resolves the catalog mapping for a tracker pattern and links it
// to an org ThirdParty. The primary goal is the org ThirdParty link; the
// catalog (common_tracker_patterns -> common_third_parties) is a fast,
// shared lookup layer that gets enriched along the way.
//
// Catalog resolution probes signals in order of confidence (existing
// catalog row, sibling origin, domain overlap, LLM agent) and keeps
// probing until it knows a common third party. Because every signal
// upserts the catalog row keyed by (tracker_type, pattern, max_age), a
// row that was previously unlinked is backfilled in place — this also
// applies on the re-trigger path, where the pattern already carries a
// common_tracker_pattern_id but its catalog row has no common third
// party yet.
//
// Org ThirdParty resolution links to an existing party freely (even for
// uncategorised or extension-sourced patterns); only the creation of a
// brand new org ThirdParty stays gated behind categorisation and a
// non-extension source.
func (h *trackerMappingHandler) Process(ctx context.Context, tp coredata.TrackerPattern) error {
	scope := coredata.NewScopeFromObjectID(tp.ID)

	// Phase 1: deterministic catalog resolution in a short transaction.
	// The existing-link, pattern, sibling, and domain signals (and their
	// idempotent upserts) run here. No LLM or web-search call is made
	// while the transaction — and its FOR UPDATE row lock — is held.
	var det deterministicResult

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var err error

			det, err = h.resolveDeterministic(ctx, tx, tp)

			return err
		},
	); err != nil {
		return err
	}

	commonPatternID := det.commonPatternID
	commonThirdPartyID := det.commonThirdPartyID
	directThirdPartyID := det.directThirdPartyID

	// Phase 2: tracker-mapping agent (no transaction). It runs only when
	// the deterministic signals could not resolve a catalog third party.
	// The LLM and web-search calls happen outside any transaction; the
	// result is persisted in its own short transaction.
	if commonThirdPartyID == nil && h.mappingAgent != nil {
		ident, err := h.identifyWithAgent(ctx, tp, det.origin)
		if err != nil {
			return fmt.Errorf("cannot identify with agent: %w", err)
		}

		if ident != nil {
			if err := h.pg.WithTx(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					match, err := h.persistAgentIdentification(ctx, tx, tp, *ident)
					if err != nil {
						return err
					}

					commonPatternID = firstNonNil(commonPatternID, match.commonPatternID)
					commonThirdPartyID = match.commonThirdPartyID

					return nil
				},
			); err != nil {
				return err
			}
		}
	}

	// Phase 3: org ThirdParty resolution. The heuristic ranking and the
	// disambiguation agent run without a transaction; only the final link
	// or create touches the database (in a short transaction).
	thirdPartyID := tp.ThirdPartyID

	if thirdPartyID == nil {
		switch {
		case directThirdPartyID != nil:
			thirdPartyID = directThirdPartyID
		case commonThirdPartyID != nil:
			resolved, err := h.resolveOrgThirdParty(ctx, tp, *commonThirdPartyID)
			if err != nil {
				return fmt.Errorf("cannot resolve org third party: %w", err)
			}

			thirdPartyID = resolved
		}
	}

	// Phase 4: persist the pattern mapping in a short transaction. The
	// unmatched fallback keeps catalog coverage complete even when no
	// vendor was resolved.
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if commonPatternID == nil {
				id, err := h.createUnmatchedPattern(ctx, tx, tp)
				if err != nil {
					return fmt.Errorf("cannot create unmatched pattern: %w", err)
				}

				commonPatternID = id
			}

			tp.CommonTrackerPatternID = commonPatternID
			tp.ThirdPartyID = thirdPartyID
			tp.UpdatedAt = time.Now()

			// Descriptions are owned by the common-pattern enrichment
			// worker. Here we only propagate: if the linked catalog row
			// is already enriched, copy its description onto this
			// pattern. A pattern linked before enrichment is filled
			// later by the enrichment worker's fan-out instead.
			if commonPatternID != nil && tp.Description == "" {
				var commonPattern coredata.CommonTrackerPattern
				if err := commonPattern.LoadByID(ctx, tx, *commonPatternID); err == nil && commonPattern.Description != "" {
					tp.Description = commonPattern.Description
				}
			}

			if err := tp.UpdateMapping(ctx, tx, scope); err != nil {
				// The pattern can be merged into a glob and deleted by
				// the pattern-analysis worker while this worker holds no
				// row lock (the LLM/web-search phases run between short
				// transactions). A vanished pattern has nothing left to
				// map, so treat the concurrent delete as a no-op instead
				// of failing the task.
				if errors.Is(err, coredata.ErrResourceNotFound) {
					h.logger.InfoCtx(
						ctx,
						"tracker pattern deleted before mapping could be persisted, skipping",
						log.String("tracker_pattern_id", tp.ID.String()),
					)

					return nil
				}

				return fmt.Errorf("cannot update tracker pattern mapping: %w", err)
			}

			h.logger.InfoCtx(
				ctx,
				"mapped tracker pattern",
				log.String("pattern", tp.Pattern),
				log.String("tracker_pattern_id", tp.ID.String()),
			)

			// This run newly resolved a catalog third party, so
			// same-banner siblings that share an initiator domain but
			// were processed earlier and left unmatched can now match
			// against it. Re-arm their mapping so the worker revisits
			// them; the guards keep already-mapped siblings untouched.
			if commonThirdPartyID != nil && !det.commonThirdPartyPreexisted {
				if err := h.reenqueueUnmappedSiblings(ctx, tx, tp, det.domains); err != nil {
					return err
				}
			}

			return nil
		},
	)
}

// deterministicResult carries the outcome of the pure-SQL catalog
// signals (existing link, pattern, sibling origin, domain overlap) from
// the read phase to the agent and persist phases. domains holds the
// observed initiator domains for the pattern with shared-infrastructure
// hosts removed (used by the sibling re-enqueue cascade);
// commonThirdPartyPreexisted records whether a catalog third party was
// already known before this run, so the cascade only fires when this run
// is the one that resolved it.
type deterministicResult struct {
	origin                     string
	commonPatternID            *gid.GID
	commonThirdPartyID         *gid.GID
	directThirdPartyID         *gid.GID
	domains                    []string
	commonThirdPartyPreexisted bool
}

// resolveDeterministic runs the catalog signals that need no network
// call (existing link, pattern, sibling origin, domain overlap) inside a
// single short transaction and reports what they resolved. It never
// invokes the mapping agent; the caller runs that outside any
// transaction.
func (h *trackerMappingHandler) resolveDeterministic(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
) (deterministicResult, error) {
	scope := coredata.NewScopeFromObjectID(tp.ID)

	var res deterministicResult

	var banner coredata.CookieBanner
	if err := banner.LoadByID(ctx, tx, scope, tp.CookieBannerID); err != nil {
		return res, fmt.Errorf("cannot load cookie banner for domain filtering: %w", err)
	}

	res.origin = banner.Origin

	if tp.CommonTrackerPatternID != nil {
		res.commonPatternID = tp.CommonTrackerPatternID

		var commonPattern coredata.CommonTrackerPattern
		if err := commonPattern.LoadByID(ctx, tx, *res.commonPatternID); err != nil {
			return res, fmt.Errorf("cannot load linked common tracker pattern: %w", err)
		}

		res.commonThirdPartyID = commonPattern.CommonThirdPartyID
	} else {
		match, err := h.matchByPattern(ctx, tx, tp)
		if err != nil {
			return res, fmt.Errorf("cannot match by pattern: %w", err)
		}

		if match != nil {
			res.commonPatternID = match.commonPatternID
			res.commonThirdPartyID = match.commonThirdPartyID
		}
	}

	res.commonThirdPartyPreexisted = res.commonThirdPartyID != nil

	if res.commonThirdPartyID != nil {
		return res, nil
	}

	loaded, err := h.loadInitiatorDomains(ctx, tx, tp)
	if err != nil {
		return res, err
	}

	// Shared tracker-delivery infrastructure (tag managers, CDPs, generic
	// CDNs) initiates trackers for many unrelated vendors, so a shared
	// initiator domain among them is not a same-vendor signal. Strip them
	// once here so no downstream domain-overlap heuristic (sibling
	// grouping, catalog domain match, or the sibling re-enqueue cascade)
	// can group unrelated trackers on, say, a common googletagmanager.com.
	res.domains = uri.FilterSharedInfrastructureDomains(loaded)

	// Sibling matching is an org-local co-occurrence signal: two
	// patterns served from the same origin on the same banner are likely
	// the same vendor, even when that origin is the site's own
	// (first-party) host — a tracker proxied through first-party still
	// co-occurs with its siblings. So it intentionally keeps first-party
	// domains (shared infrastructure was already removed above); the
	// ambiguity guard in resolveThirdPartyFromSiblings prevents grouping
	// unrelated first-party scripts.
	siblingMatch, err := h.matchBySiblingOrigin(ctx, tx, tp, res.domains)
	if err != nil {
		return res, fmt.Errorf("cannot match by sibling origin: %w", err)
	}

	if siblingMatch != nil {
		res.commonPatternID = firstNonNil(res.commonPatternID, siblingMatch.commonPatternID)
		res.commonThirdPartyID = siblingMatch.commonThirdPartyID
		res.directThirdPartyID = siblingMatch.thirdPartyID
	}

	if res.commonThirdPartyID != nil {
		return res, nil
	}

	// Domain matching hits the global catalog, so first-party domains
	// must be stripped: a tracker proxied through the site's own host
	// would otherwise match the site owner's own CommonThirdParty entry.
	catalogDomains := uri.FilterFirstPartyDomains(res.domains, banner.Origin)

	domainMatch, err := h.matchByDomain(ctx, tx, tp, catalogDomains)
	if err != nil {
		return res, fmt.Errorf("cannot match by domain: %w", err)
	}

	if domainMatch != nil {
		res.commonPatternID = firstNonNil(res.commonPatternID, domainMatch.commonPatternID)
		res.commonThirdPartyID = domainMatch.commonThirdPartyID
	}

	return res, nil
}

// reenqueueUnmappedSiblings re-arms mapping_requested_at on same-banner
// siblings sharing an initiator domain with tp that are still unpromoted,
// so the worker re-evaluates them now that tp resolved a vendor.
func (h *trackerMappingHandler) reenqueueUnmappedSiblings(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
	domains []string,
) error {
	scope := coredata.NewScopeFromObjectID(tp.ID)

	var patterns coredata.TrackerPatterns

	count, err := patterns.RequestMappingForUnmappedSiblings(
		ctx,
		tx,
		scope,
		tp.CookieBannerID,
		tp.ID,
		domains,
	)
	if err != nil {
		return fmt.Errorf("cannot re-enqueue unmapped siblings: %w", err)
	}

	if count > 0 {
		h.logger.InfoCtx(
			ctx,
			"re-enqueued unmapped sibling tracker patterns",
			log.String("tracker_pattern_id", tp.ID.String()),
			log.Int64("count", count),
		)
	}

	return nil
}

// firstNonNil returns a when it is set, otherwise b. It keeps the first
// catalog row id resolved by the pipeline stable: later signals upsert
// the same row (same key) and return the same id, but the explicit guard
// documents that the original match wins.
func firstNonNil(a, b *gid.GID) *gid.GID {
	if a != nil {
		return a
	}

	return b
}

// loadInitiatorDomains loads the distinct initiator domains observed for
// the pattern's detected trackers. The raw, unfiltered set is returned:
// callers matching against the global catalog must strip first-party
// domains themselves (uri.FilterFirstPartyDomains), but sibling matching
// deliberately keeps them, since co-occurrence on the site's own origin
// is still a valid same-vendor signal within a single banner.
func (h *trackerMappingHandler) loadInitiatorDomains(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
) ([]string, error) {
	var trackers coredata.DetectedTrackers

	domains, err := trackers.LoadInitiatorDomainsByTrackerPatternID(ctx, tx, tp.ID, 10)
	if err != nil {
		return nil, fmt.Errorf("cannot load initiator domains: %w", err)
	}

	return domains, nil
}

// creationAllowed reports whether the pattern is eligible for creating a
// brand new org ThirdParty. Extension-sourced patterns are never allowed
// to create one, and a pattern must be categorized first.
func (h *trackerMappingHandler) creationAllowed(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	tp coredata.TrackerPattern,
) (bool, error) {
	if tp.Source != nil && *tp.Source == coredata.CookieSourceExtension {
		return false, nil
	}

	var category coredata.CookieCategory
	if err := category.LoadByID(ctx, conn, scope, tp.CookieCategoryID); err != nil {
		return false, fmt.Errorf("cannot load cookie category: %w", err)
	}

	return category.Kind != coredata.CookieCategoryKindUncategorised, nil
}

// matchByPattern looks for a catalog row with the same pattern and
// surfaces both the row id and the common third party it points at (when
// set), so the caller can short-circuit promotion or keep probing for a
// common third party to backfill an unlinked row.
func (h *trackerMappingHandler) matchByPattern(
	ctx context.Context,
	conn pg.Querier,
	tp coredata.TrackerPattern,
) (*catalogMatch, error) {
	var commonPattern coredata.CommonTrackerPattern
	if err := commonPattern.LoadByPattern(ctx, conn, tp.TrackerType, tp.Pattern, tp.MaxAgeSeconds); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load common tracker pattern: %w", err)
	}

	return &catalogMatch{
		commonPatternID:    &commonPattern.ID,
		commonThirdPartyID: commonPattern.CommonThirdPartyID,
	}, nil
}

// matchByDomain finds a CommonThirdParty whose registered domains
// overlap the pattern's observed initiator domains, and upserts a
// CommonTrackerPattern linking the two. The upsert is keyed by
// (tracker_type, pattern, max_age), so it backfills a previously
// unlinked catalog row in place.
//
// The caller is responsible for loading and filtering the domains
// (removing first-party domains). Tracker scripts loaded through a
// first-party proxy (e.g. t.probo.com proxying PostHog on a probo.com
// site) would otherwise match the site owner's own CommonThirdParty
// entry.
func (h *trackerMappingHandler) matchByDomain(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
	domains []string,
) (*catalogMatch, error) {
	if len(domains) == 0 {
		return nil, nil
	}

	filter := coredata.NewCommonThirdPartyDomainFilter(domains)

	var matchedDomains coredata.CommonThirdPartyDomains
	if err := matchedDomains.Load(ctx, tx, 1, filter); err != nil {
		return nil, fmt.Errorf("cannot load common third party domain by domain match: %w", err)
	}

	if len(matchedDomains) == 0 {
		return nil, nil
	}

	commonThirdPartyID := matchedDomains[0].CommonThirdPartyID

	now := time.Now()
	commonPattern := coredata.CommonTrackerPattern{
		ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		CommonThirdPartyID: &commonThirdPartyID,
		TrackerType:        tp.TrackerType,
		Pattern:            tp.Pattern,
		MatchType:          tp.MatchType,
		MaxAgeSeconds:      tp.MaxAgeSeconds,
		Confidence:         0.7,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if _, err := commonPattern.Upsert(ctx, tx); err != nil {
		return nil, fmt.Errorf("cannot upsert common tracker pattern from domain match: %w", err)
	}

	return &catalogMatch{
		commonPatternID:    &commonPattern.ID,
		commonThirdPartyID: commonPattern.CommonThirdPartyID,
	}, nil
}

// agentIdentification carries a confident tracker-mapping agent result
// from the no-tx agent phase to the short transaction that persists it.
type agentIdentification struct {
	result TrackerMappingAgentResult
}

// identifyWithAgent runs the tracker-mapping agent outside any
// transaction. It loads the observed initiator domains with a
// short-lived connection, calls the LLM (and any web-search tool), and
// returns a confident identification or nil. It performs no writes; the
// caller persists the result via persistAgentIdentification.
func (h *trackerMappingHandler) identifyWithAgent(
	ctx context.Context,
	tp coredata.TrackerPattern,
	siteOrigin string,
) (*agentIdentification, error) {
	var domains []string

	if err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var trackers coredata.DetectedTrackers

			loaded, err := trackers.LoadInitiatorDomainsByTrackerPatternID(ctx, conn, tp.ID, 5)
			if err != nil {
				return err
			}

			domains = loaded

			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("cannot load initiator domains for agent: %w", err)
	}

	domains = uri.FilterFirstPartyDomains(domains, siteOrigin)

	prompt := buildAgentPrompt(tp, domains)

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
			"agent identification failed",
			log.Error(err),
			log.String("pattern", tp.Pattern),
		)

		return nil, nil
	}

	identification := result.Output

	// The agent's confidence gauges the attribution (who set the
	// tracker), not whether the artifact is a meaningful tracker. Without
	// a confident vendor there is nothing to catalog here; the unmatched
	// fallback records the pattern with no third party instead.
	if identification.ThirdPartyName == "" || identification.ThirdPartyConfidence < agentThirdPartyConfidenceThreshold {
		h.logger.InfoCtx(
			ctx,
			"agent third-party attribution below confidence threshold",
			log.String("pattern", tp.Pattern),
			log.Float64("third_party_confidence", identification.ThirdPartyConfidence),
		)

		return nil, nil
	}

	return &agentIdentification{
		result: identification,
	}, nil
}

// persistAgentIdentification writes a confident agent identification:
// it resolves or creates the catalog third party and upserts the
// catalog pattern row that links to it. It runs inside the caller's
// short transaction.
func (h *trackerMappingHandler) persistAgentIdentification(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
	ident agentIdentification,
) (*catalogMatch, error) {
	commonThirdPartyID, err := thirdparty.ResolveOrCreateCommonThirdParty(
		ctx,
		tx,
		h.logger,
		ident.result.ThirdPartyName,
		ident.result.Category,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve or create common third party: %w", err)
	}

	now := time.Now()
	commonPattern := coredata.CommonTrackerPattern{
		ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		CommonThirdPartyID: commonThirdPartyID,
		TrackerType:        tp.TrackerType,
		Pattern:            tp.Pattern,
		MatchType:          tp.MatchType,
		MaxAgeSeconds:      tp.MaxAgeSeconds,
		Confidence:         agentSourceConfidence,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if _, err := commonPattern.Upsert(ctx, tx); err != nil {
		return nil, fmt.Errorf("cannot upsert common tracker pattern from agent: %w", err)
	}

	h.logger.InfoCtx(
		ctx,
		"agent identified tracker pattern",
		log.String("pattern", tp.Pattern),
		log.String("third_party", ident.result.ThirdPartyName),
		log.Float64("third_party_confidence", ident.result.ThirdPartyConfidence),
	)

	return &catalogMatch{
		commonPatternID:    &commonPattern.ID,
		commonThirdPartyID: commonPattern.CommonThirdPartyID,
	}, nil
}

// matchBySiblingOrigin finds other tracker patterns on the same banner
// that share initiator domains with the current pattern. Sharing an
// origin across multiple detected patterns is a strong indicator of the
// same third party. When the siblings resolve to a single existing org
// ThirdParty, that id is returned directly so promotion can link to it
// without re-running heuristics; otherwise the resolved common third
// party is upserted onto the catalog row.
func (h *trackerMappingHandler) matchBySiblingOrigin(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
	domains []string,
) (*catalogMatch, error) {
	if len(domains) == 0 {
		return nil, nil
	}

	var trackers coredata.DetectedTrackers

	siblingIDs, err := trackers.LoadSiblingPatternIDsByInitiatorDomains(
		ctx,
		tx,
		tp.CookieBannerID,
		domains,
		tp.ID,
		20,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load sibling pattern ids: %w", err)
	}

	if len(siblingIDs) == 0 {
		return nil, nil
	}

	scope := coredata.NewScopeFromObjectID(tp.ID)

	commonThirdPartyID, thirdPartyID, err := h.resolveThirdPartyFromSiblings(ctx, tx, scope, siblingIDs)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve third party from siblings: %w", err)
	}

	// No catalog third party to record: surface a directly-known org
	// third party (if any) so promotion can still link to it, and leave
	// catalog creation to a later signal or the unmatched fallback.
	if commonThirdPartyID == nil {
		if thirdPartyID != nil {
			return &catalogMatch{thirdPartyID: thirdPartyID}, nil
		}

		return nil, nil
	}

	now := time.Now()
	commonPattern := coredata.CommonTrackerPattern{
		ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		CommonThirdPartyID: commonThirdPartyID,
		TrackerType:        tp.TrackerType,
		Pattern:            tp.Pattern,
		MatchType:          tp.MatchType,
		MaxAgeSeconds:      tp.MaxAgeSeconds,
		Confidence:         0.7,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if _, err := commonPattern.Upsert(ctx, tx); err != nil {
		return nil, fmt.Errorf("cannot upsert common tracker pattern from sibling origin: %w", err)
	}

	h.logger.InfoCtx(
		ctx,
		"matched tracker pattern via sibling origin",
		log.String("pattern", tp.Pattern),
		log.String("tracker_pattern_id", tp.ID.String()),
		log.String("common_third_party_id", commonThirdPartyID.String()),
	)

	return &catalogMatch{
		commonPatternID:    &commonPattern.ID,
		commonThirdPartyID: commonPattern.CommonThirdPartyID,
		thirdPartyID:       thirdPartyID,
	}, nil
}

// resolveThirdPartyFromSiblings inspects sibling patterns to resolve a
// third party. It returns two independent signals: a direct org
// ThirdParty (set only when the siblings share a single one — the
// strongest, same-org signal), and a single unambiguous catalog third
// party for backfill. The catalog third party is resolved first from the
// siblings' org ThirdParties, then, when those carry none, from siblings'
// common_tracker_pattern rows. Either signal may be nil; siblings that
// disagree on the catalog third party resolve it to nothing.
func (h *trackerMappingHandler) resolveThirdPartyFromSiblings(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	siblingIDs []gid.GID,
) (commonThirdPartyID *gid.GID, thirdPartyID *gid.GID, err error) {
	var patterns coredata.TrackerPatterns

	thirdPartyIDs, err := patterns.LoadDistinctThirdPartyIDsByIDs(ctx, conn, scope, siblingIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot load distinct third party ids from siblings: %w", err)
	}

	// A single org third party shared across the siblings is the
	// strongest, same-org signal: link to it directly. This is resolved
	// independently from the catalog third party used for backfill.
	if len(thirdPartyIDs) == 1 {
		directID := thirdPartyIDs[0]
		thirdPartyID = &directID
	}

	if len(thirdPartyIDs) > 0 {
		commonIDs := make(map[gid.GID]struct{})

		for _, tpID := range thirdPartyIDs {
			var t coredata.ThirdParty
			if err := t.LoadByID(ctx, conn, scope, tpID); err != nil {
				continue
			}

			if t.CommonThirdPartyID != nil {
				commonIDs[*t.CommonThirdPartyID] = struct{}{}
			}
		}

		if len(commonIDs) == 1 {
			for id := range commonIDs {
				return &id, thirdPartyID, nil
			}
		}

		// Siblings are promoted to several different catalog third
		// parties: do not guess one. A single shared org third party (if
		// any) is still a safe direct link.
		if len(commonIDs) > 1 {
			return nil, thirdPartyID, nil
		}
	}

	// Fall back to siblings carrying only a common_tracker_pattern_id, or
	// whose org ThirdParty is not itself linked to the catalog. This is
	// reached when the org-third-party scan above found no catalog third
	// party, so it must not be short-circuited by a direct match.
	commonPatternIDs, err := patterns.LoadDistinctCommonTrackerPatternIDsByIDs(ctx, conn, scope, siblingIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot load distinct common tracker pattern ids from siblings: %w", err)
	}

	if len(commonPatternIDs) == 0 {
		return nil, thirdPartyID, nil
	}

	commonIDs := make(map[gid.GID]struct{})

	for _, cpID := range commonPatternIDs {
		var cp coredata.CommonTrackerPattern
		if err := cp.LoadByID(ctx, conn, cpID); err != nil {
			continue
		}

		if cp.CommonThirdPartyID != nil {
			commonIDs[*cp.CommonThirdPartyID] = struct{}{}
		}
	}

	if len(commonIDs) == 1 {
		for id := range commonIDs {
			return &id, thirdPartyID, nil
		}
	}

	return nil, thirdPartyID, nil
}

func (h *trackerMappingHandler) createUnmatchedPattern(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
) (*gid.GID, error) {
	now := time.Now()
	commonPattern := coredata.CommonTrackerPattern{
		ID:            gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		TrackerType:   tp.TrackerType,
		Pattern:       tp.Pattern,
		MatchType:     tp.MatchType,
		MaxAgeSeconds: tp.MaxAgeSeconds,
		Confidence:    0.5,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if _, err := commonPattern.Upsert(ctx, tx); err != nil {
		return nil, fmt.Errorf("cannot upsert unmatched common tracker pattern: %w", err)
	}

	return &commonPattern.ID, nil
}

// resolveOrgThirdParty resolves an org ThirdParty for the given pattern
// from a known catalog third party. The resolution order is:
//
//  1. Exact link by common_third_party_id (O(1)).
//  2. Heuristic match against the org's existing ThirdParty rows
//     (lowercased name, suffix-stripped name, slug, website host,
//     CommonThirdPartyDomain overlap).
//  3. Agent disambiguation when the heuristic is ambiguous.
//  4. Fallback create from CommonThirdParty — only when allowCreate.
//
// Linking to an existing org ThirdParty (steps 1-3) is always allowed.
// Creating a brand new org ThirdParty (step 4) is gated by allowCreate:
// when false, the function returns (nil, nil) rather than creating one.
// A confident heuristic/agent match is auto-tagged with
// common_third_party_id so subsequent resolutions hit the exact-link
// path in O(1).
func (h *trackerMappingHandler) resolveOrgThirdParty(
	ctx context.Context,
	tp coredata.TrackerPattern,
	commonThirdPartyID gid.GID,
) (*gid.GID, error) {
	scope := coredata.NewScopeFromObjectID(tp.ID)

	// Read phase: exact link, candidate ranking, eligibility, and
	// creation gating. No write or LLM call happens here.
	var prep orgThirdPartyPrep

	if err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			prep, err = h.prepareOrgThirdParty(ctx, conn, scope, tp, commonThirdPartyID)

			return err
		},
	); err != nil {
		return nil, err
	}

	if prep.existingID != nil {
		return prep.existingID, nil
	}

	picked := prep.highConfidence
	viaAgent := false

	// Agent phase (no transaction): disambiguate among the heuristic
	// candidates when none scored high enough on its own.
	if picked == nil && prep.eligibleForAgent && h.disambiguationAgent != nil {
		matchedID, err := thirdparty.Disambiguate(
			ctx,
			h.disambiguationAgent,
			h.logger,
			prep.commonParty,
			prep.commonDomains,
			prep.agentSet,
			h.disambiguationTimeout,
		)
		if err != nil {
			h.logger.WarnCtx(
				ctx,
				"third-party disambiguation agent failed",
				log.Error(err),
				log.String("tracker_pattern_id", tp.ID.String()),
			)
		}

		if matchedID != nil {
			for _, c := range prep.agentSet {
				if c.ThirdParty.ID == *matchedID {
					picked = c.ThirdParty
					viaAgent = true

					break
				}
			}
		}
	}

	// Nothing to link and creation is not allowed: leave the pattern
	// without an org third party.
	if picked == nil && !prep.allowCreate {
		return nil, nil
	}

	// Write phase: link the picked candidate or create a new org third
	// party from the catalog entry, in a short transaction.
	var resolved *gid.GID

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if picked != nil {
				if err := thirdparty.LinkToCommon(ctx, tx, scope, picked, commonThirdPartyID); err != nil {
					return fmt.Errorf("cannot link third party to common: %w", err)
				}

				if viaAgent {
					h.logger.InfoCtx(
						ctx,
						"promoted tracker pattern via disambiguation agent",
						log.String("tracker_pattern_id", tp.ID.String()),
						log.String("third_party_id", picked.ID.String()),
					)
				} else {
					h.logger.InfoCtx(
						ctx,
						"promoted tracker pattern via heuristic match",
						log.String("tracker_pattern_id", tp.ID.String()),
						log.String("third_party_id", picked.ID.String()),
						log.Float64("score", prep.highScore),
					)
				}

				resolved = &picked.ID

				return nil
			}

			created, err := thirdparty.CreateFromCommon(ctx, tx, scope, tp.OrganizationID, prep.commonParty)
			if err != nil {
				return fmt.Errorf("cannot create third party from common: %w", err)
			}

			h.logger.InfoCtx(
				ctx,
				"promoted tracker pattern by creating org third party from catalog",
				log.String("tracker_pattern_id", tp.ID.String()),
				log.String("third_party_id", created.ID.String()),
				log.String("common_third_party_id", commonThirdPartyID.String()),
			)

			resolved = &created.ID

			return nil
		},
	); err != nil {
		return nil, err
	}

	return resolved, nil
}

// orgThirdPartyPrep is the read-phase outcome for org ThirdParty
// resolution. existingID is set when an exact common-id link already
// exists (the other fields are then unused). Otherwise highConfidence
// holds a heuristic match at or above HighConfidenceScore (with
// highScore), or agentSet/eligibleForAgent describe the disambiguation
// candidates. allowCreate gates falling back to creating a new org
// ThirdParty from the catalog entry.
type orgThirdPartyPrep struct {
	existingID       *gid.GID
	commonParty      coredata.CommonThirdParty
	commonDomains    coredata.CommonThirdPartyDomains
	agentSet         []thirdparty.ScoredCandidate
	highConfidence   *coredata.ThirdParty
	highScore        float64
	eligibleForAgent bool
	allowCreate      bool
}

// prepareOrgThirdParty performs the read-only work for org ThirdParty
// resolution: it checks for an exact common-id link, loads the catalog
// entry and the org's existing third parties, ranks the candidates, and
// computes creation eligibility. It makes no writes and no LLM call.
func (h *trackerMappingHandler) prepareOrgThirdParty(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	tp coredata.TrackerPattern,
	commonThirdPartyID gid.GID,
) (orgThirdPartyPrep, error) {
	var prep orgThirdPartyPrep

	var existing coredata.ThirdParty

	err := existing.LoadByOrganizationIDAndCommonThirdPartyID(
		ctx,
		conn,
		scope,
		tp.OrganizationID,
		commonThirdPartyID,
	)
	if err == nil {
		id := existing.ID
		prep.existingID = &id

		return prep, nil
	}

	if !errors.Is(err, coredata.ErrResourceNotFound) {
		return prep, fmt.Errorf("cannot load org third party by common id: %w", err)
	}

	if err := prep.commonParty.LoadByID(ctx, conn, commonThirdPartyID); err != nil {
		return prep, fmt.Errorf("cannot load common third party: %w", err)
	}

	if err := prep.commonDomains.LoadByCommonThirdPartyID(ctx, conn, commonThirdPartyID); err != nil {
		return prep, fmt.Errorf("cannot load common third party domains: %w", err)
	}

	var orgThirdParties coredata.ThirdParties
	if err := orgThirdParties.LoadAllByOrganizationID(ctx, conn, scope, tp.OrganizationID); err != nil {
		return prep, fmt.Errorf("cannot load org third parties: %w", err)
	}

	ranked := thirdparty.RankCandidates(prep.commonParty, prep.commonDomains, orgThirdParties)

	if len(ranked) > 0 && ranked[0].Score >= thirdparty.HighConfidenceScore {
		prep.highConfidence = ranked[0].ThirdParty
		prep.highScore = ranked[0].Score
	} else {
		prep.agentSet = ranked
		if len(prep.agentSet) > thirdparty.MaxAgentCandidates {
			prep.agentSet = prep.agentSet[:thirdparty.MaxAgentCandidates]
		}

		for _, c := range prep.agentSet {
			if c.Score >= thirdparty.MinAgentScore {
				prep.eligibleForAgent = true

				break
			}
		}
	}

	allowCreate, err := h.creationAllowed(ctx, conn, scope, tp)
	if err != nil {
		return prep, err
	}

	prep.allowCreate = allowCreate

	return prep, nil
}
