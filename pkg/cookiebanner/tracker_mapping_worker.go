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
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/browser"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/stringsx"
	"go.probo.inc/probo/pkg/thirdparty"
	"go.probo.inc/probo/pkg/uri"
)

// defaultMappingStaleAfter is the fallback idle window after which a
// claimed-but-unfinished tracker pattern mapping is re-armed. It is
// generous relative to a single Process run (deterministic SQL plus up
// to two bounded agent runs) so an in-flight mapping is never recycled.
const defaultMappingStaleAfter = 10 * time.Minute

type trackerMappingHandler struct {
	pg                    *pg.Client
	logger                *log.Logger
	mappingCfg            TrackerMappingAgentConfig
	mappingEnabled        bool
	disambiguationAgent   *agent.Agent
	agentTimeout          time.Duration
	disambiguationTimeout time.Duration
	staleAfter            time.Duration
}

func NewTrackerMappingWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	mappingCfg TrackerMappingAgentConfig,
	disambiguationCfg thirdparty.DisambiguationAgentConfig,
	staleAfter time.Duration,
	opts ...worker.Option,
) *worker.Worker[coredata.TrackerPattern] {
	agentTimeout := mappingCfg.Timeout
	if agentTimeout <= 0 {
		agentTimeout = defaultAgentTimeout
	}

	if staleAfter <= 0 {
		staleAfter = defaultMappingStaleAfter
	}

	h := &trackerMappingHandler{
		pg:                    pgClient,
		logger:                logger,
		mappingCfg:            mappingCfg,
		mappingEnabled:        mappingCfg.LLMClient != nil,
		agentTimeout:          agentTimeout,
		disambiguationTimeout: disambiguationCfg.Timeout,
		staleAfter:            staleAfter,
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

// RecoverStale re-arms tracker patterns whose mapping was claimed but
// never finished. Claim clears mapping_requested_at up front, so a crash
// or hard failure between phases would otherwise strand the pattern
// unmapped with nothing to re-trigger it. ResetStaleMappings re-queues
// those rows once they have been idle past staleAfter.
func (h *trackerMappingHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := coredata.ResetStaleMappings(ctx, conn, h.staleAfter); err != nil {
				return fmt.Errorf("cannot reset stale tracker pattern mappings: %w", err)
			}

			return nil
		},
	)
}

// catalogMatch is the result of a single catalog signal. commonPatternID
// is the catalog row the signal resolved (or backfilled); commonThirdPartyID
// is the catalog third party the signal discovered, when any; thirdPartyID
// is an existing org ThirdParty the signal knows directly (e.g. a sibling
// pattern already promoted in the same organization). firstParty is set
// when the resolved catalog row carries the terminal FIRST_PARTY verdict.
// untrustedThirdPartyID carries a vendor that was present on the resolved
// row but not adopted because its confidence fell below
// trustedAttributionConfidence; it lets the agent corroborate the prior
// guess. A nil *catalogMatch means the signal produced nothing.
type catalogMatch struct {
	commonPatternID       *gid.GID
	commonThirdPartyID    *gid.GID
	thirdPartyID          *gid.GID
	untrustedThirdPartyID *gid.GID
	firstParty            bool
}

// interpretCatalogRow maps a resolved catalog row onto the mapping
// pipeline's adoption rules. A FIRST_PARTY row is terminal. A vendor is
// adopted only when the row clears trustedAttributionConfidence;
// otherwise the vendor is surfaced as untrusted so the agent can
// corroborate it rather than the pipeline inheriting a low-confidence
// precedent.
func interpretCatalogRow(cp coredata.CommonTrackerPattern) (adopt *gid.GID, untrusted *gid.GID, firstParty bool) {
	if cp.Attribution == coredata.CommonTrackerPatternAttributionFirstParty {
		return nil, nil, true
	}

	if cp.CommonThirdPartyID == nil {
		return nil, nil, false
	}

	if cp.Confidence >= trustedAttributionConfidence {
		return cp.CommonThirdPartyID, nil, false
	}

	return nil, cp.CommonThirdPartyID, false
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
// Org ThirdParty resolution only links to an existing party (even for
// uncategorised or extension-sourced patterns); it never creates a brand
// new org ThirdParty. Creating an org ThirdParty from a catalog vendor is
// done exclusively through the explicit ImportFromCommon action.
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
	firstParty := det.firstParty

	// Phase 2: tracker-mapping agent (no transaction). It runs only when
	// the deterministic signals could not resolve a catalog third party.
	// The LLM and web-search calls happen outside any transaction; the
	// result is persisted in its own short transaction. Patterns whose
	// source is PRE_EXISTING are skipped: that source is the low-signal
	// catch-all (storage enumerated at SDK init, which bundles extension
	// state and prior-session artifacts), so a speculative agent run on it
	// is more likely to invent a vendor than to find a real one. The
	// deterministic catalog match still applies above, so a known cookie
	// still maps; and a later SCRIPT/EXTENSION detection upgrades the
	// source and re-arms mapping, giving the agent a better-grounded run.
	if commonThirdPartyID == nil && h.mappingEnabled && !det.firstParty && !isPreExistingSource(tp) {
		ident, err := h.identifyWithAgent(ctx, tp, det.origin)
		if err != nil {
			return fmt.Errorf("cannot identify with agent: %w", err)
		}

		if ident != nil {
			if err := h.pg.WithTx(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					var match *catalogMatch

					if ident.firstParty {
						match, err = h.persistFirstPartyVerdict(ctx, tx, tp)
					} else {
						match, err = h.persistAgentIdentification(ctx, tx, tp, *ident, det.untrustedThirdPartyID)
					}

					if err != nil {
						return err
					}

					commonPatternID = firstNonNil(commonPatternID, match.commonPatternID)
					commonThirdPartyID = match.commonThirdPartyID
					firstParty = match.firstParty

					return nil
				},
			); err != nil {
				return err
			}
		}
	}

	// Phase 3: org ThirdParty resolution. The heuristic ranking and the
	// disambiguation agent run without a transaction; only the final link
	// touches the database (in a short transaction).
	thirdPartyID := tp.ThirdPartyID

	// A first-party verdict is terminal: the artifact has no vendor, so
	// any org ThirdParty link a prior mapping run left on the pattern is
	// stale and must be cleared.
	if firstParty {
		thirdPartyID = nil
	} else if thirdPartyID == nil {
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
	mapped := true

	if err := h.pg.WithTx(
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

					mapped = false

					return nil
				}

				return fmt.Errorf("cannot update tracker pattern mapping: %w", err)
			}

			h.logger.DebugCtx(
				ctx,
				"mapped tracker pattern",
				log.String("pattern", tp.Pattern),
				log.String("tracker_pattern_id", tp.ID.String()),
			)

			return nil
		},
	); err != nil {
		return err
	}

	// Phase 5: re-arm same-banner siblings in a separate short
	// transaction, after the mapping above has committed. This run newly
	// resolved a catalog third party, so siblings that share an initiator
	// domain but were processed earlier and left unmatched can now match
	// against it. Re-arm their mapping so the worker revisits them; the
	// guards keep already-mapped siblings untouched.
	//
	// The re-enqueue must not run inside the Phase 4 transaction: that
	// transaction holds the row lock on tp, and the sibling UPDATE then
	// takes locks on other tracker_patterns rows while holding it. Two
	// workers mapping sibling patterns on the same banner would acquire
	// those row locks in opposite orders and deadlock. Committing Phase 4
	// first releases tp's lock, and RequestMappingForUnmappedSiblings
	// takes its locks in a deterministic id order, so the two can no
	// longer cycle.
	if mapped && commonThirdPartyID != nil && !det.commonThirdPartyPreexisted {
		if err := h.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return h.reenqueueUnmappedSiblings(ctx, tx, tp, det.domains)
			},
		); err != nil {
			return err
		}
	}

	return nil
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
	untrustedThirdPartyID      *gid.GID
	domains                    []string
	commonThirdPartyPreexisted bool
	firstParty                 bool
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

		res.commonThirdPartyID, res.untrustedThirdPartyID, res.firstParty = interpretCatalogRow(commonPattern)
	} else {
		match, err := h.matchByPattern(ctx, tx, tp)
		if err != nil {
			return res, fmt.Errorf("cannot match by pattern: %w", err)
		}

		if match != nil {
			res.commonPatternID = match.commonPatternID
			res.commonThirdPartyID = match.commonThirdPartyID
			res.untrustedThirdPartyID = match.untrustedThirdPartyID
			res.firstParty = match.firstParty
		}
	}

	// A terminal FIRST_PARTY verdict short-circuits every remaining
	// signal: the artifact has no third party, so neither the heuristic
	// matches nor the agent should run, and no org party is linked.
	if res.firstParty {
		return res, nil
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
		h.logger.DebugCtx(
			ctx,
			"re-enqueued unmapped sibling tracker patterns",
			log.String("tracker_pattern_id", tp.ID.String()),
			log.Int64("count", count),
		)
	}

	return nil
}

// isPreExistingSource reports whether the org tracker pattern's source is
// PRE_EXISTING. That source is the low-signal catch-all enumerated from
// storage at SDK init (it bundles browser-extension state and
// prior-session artifacts), so the speculative mapping agent is not run
// for it; the deterministic catalog signals still apply.
func isPreExistingSource(tp coredata.TrackerPattern) bool {
	return tp.Source != nil && *tp.Source == coredata.CookieSourcePreExisting
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

	adopt, untrusted, firstParty := interpretCatalogRow(commonPattern)

	return &catalogMatch{
		commonPatternID:       &commonPattern.ID,
		commonThirdPartyID:    adopt,
		untrustedThirdPartyID: untrusted,
		firstParty:            firstParty,
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
		Attribution:        coredata.CommonTrackerPatternAttributionThirdParty,
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

// agentIdentification carries a tracker-mapping agent verdict from the
// no-tx agent phase to the short transaction that persists it. Exactly
// one outcome is meaningful: firstParty set means the agent declared a
// terminal no-third-party verdict; otherwise result holds a defensible
// vendor attribution.
type agentIdentification struct {
	result     TrackerMappingAgentResult
	firstParty bool
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

	siteDomain := uri.ExtractDomain(siteOrigin)

	prompt := buildAgentPrompt(tp, domains, siteDomain)

	// Build the mapping agent per run so it can carry a per-run browser
	// when a Chrome endpoint is configured. The browser lets the agent
	// open cookie-database and cookie-policy pages to read the true
	// setter; it is closed when this run returns.
	var browserTools []agent.Tool

	if h.mappingCfg.ChromeAddr != "" {
		webBrowser := browser.NewBrowser(ctx, h.mappingCfg.ChromeAddr)
		defer webBrowser.Close()

		browserTools = browser.NewReadOnlyToolset(webBrowser).Tools()
	}

	mappingAgent := buildTrackerMappingAgent(h.mappingCfg, h.pg, h.logger, browserTools)

	agentCtx, cancel := context.WithTimeout(ctx, h.agentTimeout)
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
		h.logger.WarnCtx(
			ctx,
			"agent identification failed",
			log.Error(err),
			log.String("pattern", tp.Pattern),
		)

		return nil, nil
	}

	identification := result.Output

	// A defensible vendor attribution wins: record it for the catalog.
	if !h.vendorAttributionRejected(ctx, tp, identification, siteOrigin) {
		return &agentIdentification{result: identification}, nil
	}

	// No defensible vendor. An explicit first-party declaration is a
	// terminal verdict: persist it so the pipeline stops retrying this
	// artifact. Otherwise leave the pattern undetermined for a later,
	// better-informed attempt (the unmatched fallback records it with no
	// third party).
	if identification.IsFirstParty {
		h.logger.InfoCtx(
			ctx,
			"agent declared tracker first-party",
			log.String("pattern", tp.Pattern),
		)

		return &agentIdentification{firstParty: true}, nil
	}

	return nil, nil
}

// vendorAttributionRejected reports whether the agent's vendor
// attribution must be discarded, logging the reason. It enforces, in
// order: a confident attribution, a concrete evidence source (no
// general-knowledge guesses), the scanned-site backstop, and the
// cookie-database-aggregator backstop.
func (h *trackerMappingHandler) vendorAttributionRejected(
	ctx context.Context,
	tp coredata.TrackerPattern,
	identification TrackerMappingAgentResult,
	siteOrigin string,
) bool {
	// The agent's confidence gauges the attribution (who set the
	// tracker), not whether the artifact is a meaningful tracker. Without
	// a confident vendor there is nothing to catalog here.
	if identification.ThirdPartyName == "" || identification.ThirdPartyConfidence < agentThirdPartyConfidenceThreshold {
		h.logger.InfoCtx(
			ctx,
			"agent third-party attribution below confidence threshold",
			log.String("pattern", tp.Pattern),
			log.Float64("third_party_confidence", identification.ThirdPartyConfidence),
		)

		return true
	}

	// Evidence guard: a vendor is attributed only on concrete evidence (a
	// database match, a meaningful naming convention, or a web/browser
	// result that names the setter). An attribution with no evidence
	// source is a general-knowledge guess and is discarded, so a wrong
	// precedent never enters the catalog.
	if !evidenceSupportsAttribution(identification.EvidenceSource) {
		h.logger.InfoCtx(
			ctx,
			"agent attribution lacks concrete evidence, discarding",
			log.String("pattern", tp.Pattern),
			log.String("evidence_source", identification.EvidenceSource),
		)

		return true
	}

	// Backstop for the prompt rule that the scanned site is never a third
	// party of itself: a pattern that embeds the site's own domain (e.g.
	// an "ethereum-https://example.com" wallet-extension key, or an
	// owner-set tracker) can lead the agent to attribute the site's own
	// brand. Discard such attributions outright so the pattern falls
	// through to the unmatched fallback instead of being mapped to the
	// site owner.
	if nameMatchesSiteDomain(identification.ThirdPartyName, siteOrigin) {
		h.logger.InfoCtx(
			ctx,
			"agent attributed scanned site as third party, discarding",
			log.String("pattern", tp.Pattern),
		)

		return true
	}

	// Cookie-database and cookie-banner directory sites (Cookifi,
	// Cookiepedia, cookiedatabase.org, ...) rank highly in web search
	// only because they catalog cookies, not because they set them. A
	// web result hosted on one can lead the agent to attribute the
	// tracker to the directory operator itself. Discard such an
	// attribution so the pattern falls through to the unmatched fallback
	// instead of being mapped to a database aggregator. The denylist is
	// scoped to pure aggregators, so a CMP's own product cookie (e.g.
	// OptanonConsent -> OneTrust) is still attributed normally.
	if nameIsCookieDatabaseAggregator(identification.ThirdPartyName) {
		h.logger.InfoCtx(
			ctx,
			"agent attributed cookie-database aggregator as third party, discarding",
			log.String("pattern", tp.Pattern),
		)

		return true
	}

	return false
}

// nameMatchesSiteDomain reports whether a candidate vendor name refers to
// the scanned site itself. The site owner is never a third party of its
// own site, so an attribution whose name resolves to the site's own
// domain must be rejected. The comparison is alphanumeric-normalised and
// conservative (equality against the eTLD+1 and its primary label) to
// avoid suppressing unrelated vendors whose name merely overlaps.
func nameMatchesSiteDomain(name, siteOrigin string) bool {
	domain := uri.ExtractDomain(siteOrigin)
	if domain == "" {
		return false
	}

	normalizedName := stringsx.NormalizeAlnum(name)
	if normalizedName == "" {
		return false
	}

	label, _, _ := strings.Cut(domain, ".")

	return normalizedName == stringsx.NormalizeAlnum(domain) ||
		normalizedName == stringsx.NormalizeAlnum(label)
}

// cookieDatabaseAggregators holds alphanumeric-normalised names of pure
// cookie-database / cookie-banner directory operators that catalog
// cookies but never legitimately set one on a third-party site. They
// surface in web search only because they host pattern databases, so an
// attribution to one is always search-database noise. The set is kept
// deliberately narrow: consent-management vendors that DO set their own
// product cookies (Cookie-Script, OneTrust, Cookiebot, CookieYes) are
// excluded so the backstop never suppresses a legitimate own-cookie
// attribution — the prompt handles their directory pages instead.
var cookieDatabaseAggregators = map[string]struct{}{
	"cookifi":        {},
	"cookiepedia":    {},
	"cookiedatabase": {},
	"cookieserve":    {},
}

// nameIsCookieDatabaseAggregator reports whether a candidate vendor name
// is a known cookie-database directory operator that must never be
// attributed a tracker. The agent may return either a brand name
// ("Cookiepedia") or a domain form ("cookiedatabase.org"); the latter
// would survive a plain normalised lookup because NormalizeAlnum folds
// the eTLD into the key (e.g. "cookiedatabaseorg"). To catch both forms
// the candidate is also reduced to its primary domain label before the
// alphanumeric-normalised lookup. The comparison is alphanumeric-
// normalised so spacing, punctuation, and casing differences do not
// matter.
func nameIsCookieDatabaseAggregator(name string) bool {
	if _, ok := cookieDatabaseAggregators[stringsx.NormalizeAlnum(name)]; ok {
		return true
	}

	label := stringsx.NormalizeAlnum(uri.DomainLabel(name))
	if label == "" {
		return false
	}

	_, ok := cookieDatabaseAggregators[label]

	return ok
}

// persistAgentIdentification writes a confident agent identification:
// it resolves or creates the catalog third party and upserts the
// catalog pattern row that links to it. It runs inside the caller's
// short transaction.
//
// priorUntrustedThirdPartyID, when set, is the vendor an existing
// catalog row carried but that was too low-confidence to adopt
// deterministically. When the agent independently lands on the same
// vendor, that is corroboration: the row is promoted to the trusted tier
// so subsequent patterns adopt it without re-running the agent.
func (h *trackerMappingHandler) persistAgentIdentification(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
	ident agentIdentification,
	priorUntrustedThirdPartyID *gid.GID,
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

	confidence := float32(agentSourceConfidence)

	corroborated := priorUntrustedThirdPartyID != nil &&
		commonThirdPartyID != nil &&
		*priorUntrustedThirdPartyID == *commonThirdPartyID
	if corroborated {
		confidence = trustedAttributionConfidence
	}

	now := time.Now()
	commonPattern := coredata.CommonTrackerPattern{
		ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		CommonThirdPartyID: commonThirdPartyID,
		TrackerType:        tp.TrackerType,
		Pattern:            tp.Pattern,
		MatchType:          tp.MatchType,
		MaxAgeSeconds:      tp.MaxAgeSeconds,
		Confidence:         confidence,
		Attribution:        coredata.CommonTrackerPatternAttributionThirdParty,
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
		log.Bool("corroborated_prior_attribution", corroborated),
	)

	return &catalogMatch{
		commonPatternID:    &commonPattern.ID,
		commonThirdPartyID: commonPattern.CommonThirdPartyID,
	}, nil
}

// persistFirstPartyVerdict records the agent's terminal first-party
// verdict on the catalog: it upserts the row with no vendor and the
// FIRST_PARTY attribution, which the upsert preserves on later automated
// runs. Any stray low-confidence vendor a prior run left on the row is
// cleared. It runs inside the caller's short transaction.
func (h *trackerMappingHandler) persistFirstPartyVerdict(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
) (*catalogMatch, error) {
	now := time.Now()
	commonPattern := coredata.CommonTrackerPattern{
		ID:            gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		TrackerType:   tp.TrackerType,
		Pattern:       tp.Pattern,
		MatchType:     tp.MatchType,
		MaxAgeSeconds: tp.MaxAgeSeconds,
		Confidence:    agentSourceConfidence,
		Attribution:   coredata.CommonTrackerPatternAttributionFirstParty,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if _, err := commonPattern.Upsert(ctx, tx); err != nil {
		return nil, fmt.Errorf("cannot upsert first-party common tracker pattern: %w", err)
	}

	h.logger.InfoCtx(
		ctx,
		"recorded first-party tracker verdict",
		log.String("pattern", tp.Pattern),
		log.String("tracker_pattern_id", tp.ID.String()),
	)

	return &catalogMatch{
		commonPatternID: &commonPattern.ID,
		firstParty:      true,
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
		Attribution:        coredata.CommonTrackerPatternAttributionThirdParty,
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
		Attribution:   coredata.CommonTrackerPatternAttributionUndetermined,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if _, err := commonPattern.Upsert(ctx, tx); err != nil {
		return nil, fmt.Errorf("cannot upsert unmatched common tracker pattern: %w", err)
	}

	return &commonPattern.ID, nil
}

// resolveOrgThirdParty resolves an org ThirdParty for the given pattern
// from a known catalog third party by linking to an existing party. The
// resolution order is:
//
//  1. Exact link by common_third_party_id (O(1)).
//  2. Heuristic match against the org's existing ThirdParty rows
//     (lowercased name, suffix-stripped name, slug, website host,
//     CommonThirdPartyDomain overlap).
//  3. Agent disambiguation when the heuristic is ambiguous.
//
// When none of these resolve an existing party, the function returns
// (nil, nil); it never creates a brand new org ThirdParty (that happens
// only through the explicit ImportFromCommon action). A confident
// heuristic/agent match is auto-tagged with common_third_party_id so
// subsequent resolutions hit the exact-link path in O(1).
func (h *trackerMappingHandler) resolveOrgThirdParty(
	ctx context.Context,
	tp coredata.TrackerPattern,
	commonThirdPartyID gid.GID,
) (*gid.GID, error) {
	scope := coredata.NewScopeFromObjectID(tp.ID)

	// Read phase: exact link, candidate ranking, and eligibility. No
	// write or LLM call happens here.
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

	// Nothing to link: leave the pattern without an org third party. An
	// org ThirdParty is created only through the explicit ImportFromCommon
	// action, never here.
	if picked == nil {
		return nil, nil
	}

	// Write phase: link the picked candidate to the catalog entry in a
	// short transaction.
	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
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

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &picked.ID, nil
}

// orgThirdPartyPrep is the read-phase outcome for org ThirdParty
// resolution. existingID is set when an exact common-id link already
// exists (the other fields are then unused). Otherwise highConfidence
// holds a heuristic match at or above HighConfidenceScore (with
// highScore), or agentSet/eligibleForAgent describe the disambiguation
// candidates.
type orgThirdPartyPrep struct {
	existingID       *gid.GID
	commonParty      coredata.CommonThirdParty
	commonDomains    coredata.CommonThirdPartyDomains
	agentSet         []thirdparty.ScoredCandidate
	highConfidence   *coredata.ThirdParty
	highScore        float64
	eligibleForAgent bool
}

// prepareOrgThirdParty performs the read-only work for org ThirdParty
// resolution: it checks for an exact common-id link, loads the catalog
// entry and the org's existing third parties, and ranks the candidates.
// It makes no writes and no LLM call.
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

	firstLevel := 1

	orgThirdParties, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.ThirdPartyOrderField]{
			Field:     coredata.ThirdPartyOrderFieldName,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.ThirdPartyOrderField]) ([]*coredata.ThirdParty, error) {
			var batch coredata.ThirdParties
			if err := batch.LoadByOrganizationID(ctx, conn, scope, tp.OrganizationID, cursor, coredata.NewThirdPartyFilter(nil, &firstLevel, nil)); err != nil {
				return nil, fmt.Errorf("cannot load org third parties: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return prep, err
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

	return prep, nil
}
