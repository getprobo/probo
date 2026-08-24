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

// defaultMappingStaleAfter is the idle window after which a claimed
// mapping is re-armed. Sized so an in-flight Process is never recycled.
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

// RecoverStale re-queues patterns whose mapping was claimed but never
// finished. Claim clears mapping_requested_at, so a crash between phases
// would otherwise strand the row.
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

// catalogMatch is one catalog signal's result. A nil pointer means the
// signal produced nothing. firstParty is any terminal no-vendor verdict;
// untrustedThirdPartyID is a vendor below trustedAttributionConfidence
// that the agent may corroborate.
type catalogMatch struct {
	commonPatternID       *gid.GID
	commonThirdPartyID    *gid.GID
	thirdPartyID          *gid.GID
	untrustedThirdPartyID *gid.GID
	firstParty            bool
}

// interpretCatalogRow applies adoption rules: a terminal row stops the
// pipeline; a vendor below trustedAttributionConfidence is surfaced as
// untrusted for the agent to corroborate.
func interpretCatalogRow(cp coredata.CommonTrackerPattern) (adopt *gid.GID, untrusted *gid.GID, firstParty bool) {
	// Any terminal verdict stops attribution, not only FIRST_PARTY.
	if cp.Attribution.IsTerminal() {
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

// Process maps a tracker pattern onto the catalog and links an existing
// org ThirdParty. Signals run in confidence order and upsert the catalog
// row in place. New org ThirdParties are created only by ImportFromCommon.
func (h *trackerMappingHandler) Process(ctx context.Context, tp coredata.TrackerPattern) error {
	scope := coredata.NewScopeFromObjectID(tp.ID)

	// Phase 1: deterministic catalog signals. No LLM while the row is locked.
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

	// A rejected catalog row still matches by name. Apply the review's
	// verdict instead of linking the vendor; the agent would re-derive
	// the same wrong attribution.
	if commonThirdPartyID != nil {
		var rejected, gone bool

		// Read and persist under one lock so a concurrent review cannot
		// land between them. Nothing is written unless the row is rejected.
		if err := h.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				verdict, missing, err := h.rejectedVerdictFor(ctx, tx, *commonThirdPartyID)
				if err != nil {
					return err
				}

				gone = missing

				if verdict == nil {
					return nil
				}

				rejected = true

				match, err := h.persistTerminalVerdict(ctx, tx, tp, *verdict)
				if err != nil {
					return err
				}

				if match != nil {
					commonPatternID = firstNonNil(commonPatternID, match.commonPatternID)
				}

				return nil
			},
		); err != nil {
			return fmt.Errorf("cannot persist verdict from a rejected catalog row: %w", err)
		}

		switch {
		case rejected:
			commonThirdPartyID = nil
			directThirdPartyID = nil
			firstParty = true
		case gone:
			// Gone since phase one. Drop the id so phase three does not
			// resolve a vendor that no longer exists.
			commonThirdPartyID = nil
		}
	}

	// Phase 2: mapping agent, outside any transaction. Skipped for
	// PRE_EXISTING (low signal) and EXTENSION (visitor-installed). Gated
	// on the local firstParty so a rejected-row verdict is not overwritten.
	if commonThirdPartyID == nil && h.mappingEnabled && !firstParty &&
		!isPreExistingSource(tp) && !isExtensionSource(tp) {
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
						match, err = h.persistTerminalVerdict(ctx, tx, tp, ident.terminalVerdict)
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

	// Phase 3: link an existing org ThirdParty. Ranking and the
	// disambiguation agent run without a transaction.
	thirdPartyID := tp.ThirdPartyID

	// A terminal verdict has no vendor; clear any stale org link.
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

	// Phase 4: persist the mapping. The unmatched fallback keeps catalog
	// coverage when no vendor was resolved.
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

			// Copy an already-enriched catalog description. Later
			// enrichment fans out to patterns linked before then.
			if commonPatternID != nil && tp.Description == "" {
				var commonPattern coredata.CommonTrackerPattern
				if err := commonPattern.LoadByID(ctx, tx, *commonPatternID); err == nil && commonPattern.Description != "" {
					tp.Description = commonPattern.Description
				}
			}

			if err := tp.UpdateMapping(ctx, tx, scope); err != nil {
				// The pattern can be glob-merged and deleted between
				// phases. Nothing left to map; treat it as a no-op.
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

	// Phase 5: re-arm unmatched siblings that share an initiator
	// domain, now that this run resolved a catalog vendor. Own
	// transaction so two workers mapping siblings cannot deadlock on
	// opposite lock orders.
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

// deterministicResult is the pure-SQL catalog outcome. domains are the
// initiator hosts with shared infrastructure stripped;
// commonThirdPartyPreexisted is true when a catalog vendor was already
// known, so the sibling cascade does not fire.
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

// resolveDeterministic runs the no-network catalog signals in one
// short transaction. The caller runs the mapping agent outside any
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

	// A terminal verdict has no vendor; skip the remaining signals.
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

	// Tag managers and CDNs initiate many unrelated vendors. Strip them
	// so no downstream domain-overlap heuristic groups on a shared host.
	res.domains = uri.FilterSharedInfrastructureDomains(loaded)

	// Sibling matching is org-local: keep first-party hosts (a proxied
	// tracker still co-occurs with its siblings). Shared infrastructure
	// was already stripped; the ambiguity guard stops unrelated grouping.
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

	// Domain matching hits the global catalog, so strip first-party
	// hosts or a proxied tracker matches the site owner.
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

// reenqueueUnmappedSiblings re-arms unpromoted same-banner siblings that
// share an initiator domain, now that tp resolved a vendor.
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

// isPreExistingSource reports the low-signal SDK-init catch-all. The
// mapping agent is not run for it; deterministic catalog signals still
// apply.
func isPreExistingSource(tp coredata.TrackerPattern) bool {
	return tp.Source != nil && *tp.Source == coredata.CookieSourcePreExisting
}

// isExtensionSource reports a write whose stack carried an extension
// frame. That settles attribution: visitor-installed software, no vendor.
func isExtensionSource(tp coredata.TrackerPattern) bool {
	return tp.Source != nil && *tp.Source == coredata.CookieSourceExtension
}

// rejectedVerdictFor returns the terminal verdict a rejected catalog
// row carries, or nil when the row was not rejected. gone means the id
// from an earlier transaction was pruned or merged; the caller must
// drop it rather than treat it as a vendor.
func (h *trackerMappingHandler) rejectedVerdictFor(
	ctx context.Context,
	tx pg.Tx,
	commonThirdPartyID gid.GID,
) (verdict *coredata.CommonTrackerPatternAttribution, gone bool, err error) {
	var party coredata.CommonThirdParty

	review, stored, err := party.LoadReviewForUpdate(ctx, tx, commonThirdPartyID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, true, nil
		}

		return nil, false, err
	}

	if review != coredata.CommonThirdPartyReviewRejected {
		return nil, false, nil
	}

	// A rejected row without a verdict must not fall through to a vendor.
	if stored == nil || !stored.IsTerminal() {
		return nil, false, fmt.Errorf(
			"rejected common third party %s carries no terminal verdict",
			commonThirdPartyID,
		)
	}

	return stored, false, nil
}

// firstNonNil returns a when set, otherwise b. Keeps the first catalog
// row id the pipeline resolved.
func firstNonNil(a, b *gid.GID) *gid.GID {
	if a != nil {
		return a
	}

	return b
}

// loadInitiatorDomains returns the raw initiator domains. Catalog
// matching must strip first-party hosts; sibling matching keeps them.
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

// matchByPattern looks up the catalog row for this pattern so the
// caller can adopt its vendor or keep probing.
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

// matchByDomain upserts a catalog row for a CommonThirdParty whose
// registered domains overlap the initiator hosts. The caller must strip
// first-party domains, or a proxied tracker matches the site owner.
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

// agentIdentification is the mapping agent's verdict. firstParty is any
// terminal no-vendor outcome; otherwise result is a defensible vendor.
type agentIdentification struct {
	result TrackerMappingAgentResult

	// firstParty marks any terminal verdict. terminalVerdict says which
	// one to persist, so an extension is not recorded as first-party.
	firstParty      bool
	terminalVerdict coredata.CommonTrackerPatternAttribution
}

// terminalVerdictFor maps the agent's flags to an attribution, or "" if
// it settled nothing. NOT_ATTRIBUTABLE wins: an extension key looks
// first-party but is not the operator's code.
func terminalVerdictFor(r TrackerMappingAgentResult) coredata.CommonTrackerPatternAttribution {
	switch {
	case r.IsNotAttributable:
		return coredata.CommonTrackerPatternAttributionNotAttributable
	case r.IsFirstParty:
		return coredata.CommonTrackerPatternAttributionFirstParty
	default:
		return ""
	}
}

// identifyWithAgent runs the mapping agent outside any transaction and
// returns a confident identification or nil. It writes nothing.
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

	// Per-run agent so a configured Chrome endpoint can read cookie
	// databases; the browser is closed when this run returns.
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

	// No defensible vendor. A terminal verdict stops retrying; otherwise
	// leave the pattern undetermined.
	if verdict := terminalVerdictFor(identification); verdict != "" {
		h.logger.InfoCtx(
			ctx,
			"agent declared tracker terminal",
			log.String("pattern", tp.Pattern),
			log.String("attribution", string(verdict)),
		)

		return &agentIdentification{firstParty: true, terminalVerdict: verdict}, nil
	}

	return nil, nil
}

// vendorAttributionRejected reports whether the agent's vendor must be
// discarded. The bar lives in rejectVendorAttribution; this wrapper
// supplies the scanned site and logs the reason as a field.
func (h *trackerMappingHandler) vendorAttributionRejected(
	ctx context.Context,
	tp coredata.TrackerPattern,
	identification TrackerMappingAgentResult,
	siteOrigin string,
) bool {
	var actx attributionContext
	if siteOrigin != "" {
		actx.SiteOrigin = &siteOrigin
	}

	rejection := rejectVendorAttribution(identification, actx)
	if rejection == attributionAccepted {
		return false
	}

	h.logger.InfoCtx(
		ctx,
		"discarded agent third-party attribution",
		log.String("pattern", tp.Pattern),
		log.String("reason", string(rejection)),
		log.String("third_party_name", identification.ThirdPartyName),
		log.Float64("third_party_confidence", identification.ThirdPartyConfidence),
		log.String("evidence_source", identification.EvidenceSource),
	)

	return true
}

// nameMatchesSiteDomain reports whether a vendor name is the scanned
// site itself. Compared alphanumerically against the eTLD+1 and its
// primary label so overlapping names are not suppressed.
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

// cookieDatabaseAggregators are directory operators that catalog cookies
// but never set one. Consent-management vendors that do set cookies are
// excluded; the prompt handles their directory pages.
var cookieDatabaseAggregators = map[string]struct{}{
	"cookifi":        {},
	"cookiepedia":    {},
	"cookiedatabase": {},
	"cookieserve":    {},
}

// nameIsCookieDatabaseAggregator reports a known cookie-database
// directory. Matches brand names and domain forms by looking up both
// the normalised name and its primary label.
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

// persistAgentIdentification writes a confident agent identification
// inside the caller's transaction. When the agent lands on the same
// vendor as priorUntrustedThirdPartyID, the row is promoted to the
// trusted tier.
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

// persistTerminalVerdict upserts a catalog row with no vendor and the
// given terminal attribution. Later automated runs preserve it. Runs
// inside the caller's transaction.
func (h *trackerMappingHandler) persistTerminalVerdict(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
	verdict coredata.CommonTrackerPatternAttribution,
) (*catalogMatch, error) {
	if verdict == "" {
		verdict = coredata.CommonTrackerPatternAttributionFirstParty
	}

	now := time.Now()
	commonPattern := coredata.CommonTrackerPattern{
		ID:            gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		TrackerType:   tp.TrackerType,
		Pattern:       tp.Pattern,
		MatchType:     tp.MatchType,
		MaxAgeSeconds: tp.MaxAgeSeconds,
		Confidence:    agentSourceConfidence,
		Attribution:   verdict,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if _, err := commonPattern.Upsert(ctx, tx); err != nil {
		return nil, fmt.Errorf("cannot upsert terminal common tracker pattern: %w", err)
	}

	h.logger.InfoCtx(
		ctx,
		"recorded terminal tracker verdict",
		log.String("pattern", tp.Pattern),
		log.String("tracker_pattern_id", tp.ID.String()),
		log.String("attribution", string(verdict)),
	)

	return &catalogMatch{
		commonPatternID: &commonPattern.ID,
		firstParty:      true,
	}, nil
}

// matchBySiblingOrigin finds same-banner patterns that share an
// initiator domain. A single shared org ThirdParty is returned
// directly; otherwise the catalog vendor is upserted onto the row.
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

	// No catalog vendor. Surface a known org ThirdParty so promotion
	// can still link; leave catalog creation to a later signal.
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

// resolveThirdPartyFromSiblings returns a direct org ThirdParty when
// siblings share one, and a single unambiguous catalog vendor for
// backfill. Disagreement on the catalog vendor resolves it to nothing.
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

	// A single shared org ThirdParty is the strongest same-org signal.
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

		// Siblings disagree on the catalog vendor; do not guess. A
		// shared org ThirdParty is still a safe direct link.
		if len(commonIDs) > 1 {
			return nil, thirdPartyID, nil
		}
	}

	// Fall back to siblings that only carry a catalog pattern, or whose
	// org ThirdParty is not itself linked to the catalog.
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

// resolveOrgThirdParty links an existing org ThirdParty: exact
// common-id, then heuristic, then agent disambiguation. It never
// creates one. A confident match is tagged so later runs hit the
// exact-link path.
func (h *trackerMappingHandler) resolveOrgThirdParty(
	ctx context.Context,
	tp coredata.TrackerPattern,
	commonThirdPartyID gid.GID,
) (*gid.GID, error) {
	scope := coredata.NewScopeFromObjectID(tp.ID)

	// Read phase: exact link, ranking, eligibility. No write or LLM.
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

	// Agent phase: disambiguate when no heuristic scored high enough.
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

	// Nothing to link. Org ThirdParties are created only by ImportFromCommon.
	if picked == nil {
		return nil, nil
	}

	// Write phase: link the picked candidate to the catalog entry.
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

// orgThirdPartyPrep is the read-phase outcome. existingID is an exact
// common-id link; otherwise highConfidence or agentSet hold the
// heuristic result.
type orgThirdPartyPrep struct {
	existingID       *gid.GID
	commonParty      coredata.CommonThirdParty
	commonDomains    coredata.CommonThirdPartyDomains
	agentSet         []thirdparty.ScoredCandidate
	highConfidence   *coredata.ThirdParty
	highScore        float64
	eligibleForAgent bool
}

// prepareOrgThirdParty is the read-only org ThirdParty lookup: exact
// common-id, then rank the org's existing parties. No writes or LLM.
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
			if err := batch.LoadByOrganizationID(ctx, conn, scope, tp.OrganizationID, cursor, coredata.NewThirdPartyFilter(&firstLevel, nil, nil, nil)); err != nil {
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
