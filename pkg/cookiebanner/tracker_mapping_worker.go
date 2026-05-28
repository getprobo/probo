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
	"go.probo.inc/probo/pkg/slug"
	"go.probo.inc/probo/pkg/thirdparty"
	"go.probo.inc/probo/pkg/uri"
)

type trackerMappingHandler struct {
	pg                  *pg.Client
	logger              *log.Logger
	mappingAgent        *agent.Agent
	disambiguationAgent *agent.Agent
}

func NewTrackerMappingWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	mappingCfg TrackerMappingConfig,
	disambiguationCfg thirdparty.DisambiguationConfig,
	opts ...worker.Option,
) *worker.Worker[coredata.TrackerPattern] {
	h := &trackerMappingHandler{
		pg:     pgClient,
		logger: logger,
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
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(tp.ID)

			var banner coredata.CookieBanner
			if err := banner.LoadByID(ctx, tx, scope, tp.CookieBannerID); err != nil {
				return fmt.Errorf("cannot load cookie banner for domain filtering: %w", err)
			}

			var (
				commonPatternID    *gid.GID
				commonThirdPartyID *gid.GID
				directThirdPartyID *gid.GID
			)

			if tp.CommonTrackerPatternID != nil {
				commonPatternID = tp.CommonTrackerPatternID

				var commonPattern coredata.CommonTrackerPattern
				if err := commonPattern.LoadByID(ctx, tx, *commonPatternID); err != nil {
					return fmt.Errorf("cannot load linked common tracker pattern: %w", err)
				}

				commonThirdPartyID = commonPattern.CommonThirdPartyID
			} else {
				match, err := h.matchByPattern(ctx, tx, tp)
				if err != nil {
					return fmt.Errorf("cannot match by pattern: %w", err)
				}

				if match != nil {
					commonPatternID = match.commonPatternID
					commonThirdPartyID = match.commonThirdPartyID
				}
			}

			// Whether a catalog third party was already known before the
			// signal pipeline ran this round. Re-enqueuing siblings is
			// only useful when this run is the one that resolves the
			// vendor; a pre-existing link adds no new signal and gating
			// on it keeps cascades finite.
			commonThirdPartyPreexisted := commonThirdPartyID != nil

			var domains []string

			if commonThirdPartyID == nil {
				loaded, err := h.loadInitiatorDomains(ctx, tx, tp)
				if err != nil {
					return err
				}

				domains = loaded

				// Sibling matching is an org-local co-occurrence signal:
				// two patterns served from the same origin on the same
				// banner are likely the same vendor, even when that origin
				// is the site's own (first-party) host — a tracker proxied
				// through first-party still co-occurs with its siblings.
				// So it intentionally runs on the unfiltered domains; the
				// ambiguity guard in resolveThirdPartyFromSiblings prevents
				// grouping unrelated first-party scripts.
				match, err := h.matchBySiblingOrigin(ctx, tx, tp, domains)
				if err != nil {
					return fmt.Errorf("cannot match by sibling origin: %w", err)
				}

				if match != nil {
					commonPatternID = firstNonNil(commonPatternID, match.commonPatternID)
					commonThirdPartyID = match.commonThirdPartyID
					directThirdPartyID = match.thirdPartyID
				}

				if commonThirdPartyID == nil {
					// Domain matching hits the global catalog, so
					// first-party domains must be stripped: a tracker
					// proxied through the site's own host would otherwise
					// match the site owner's own CommonThirdParty entry.
					catalogDomains := uri.FilterFirstPartyDomains(domains, banner.Origin)

					match, err := h.matchByDomain(ctx, tx, tp, catalogDomains)
					if err != nil {
						return fmt.Errorf("cannot match by domain: %w", err)
					}

					if match != nil {
						commonPatternID = firstNonNil(commonPatternID, match.commonPatternID)
						commonThirdPartyID = match.commonThirdPartyID
					}
				}

				if commonThirdPartyID == nil && h.mappingAgent != nil {
					match, err := h.identifyWithAgent(ctx, tx, tp, banner.Origin)
					if err != nil {
						return fmt.Errorf("cannot identify with agent: %w", err)
					}

					if match != nil {
						commonPatternID = firstNonNil(commonPatternID, match.commonPatternID)
						commonThirdPartyID = match.commonThirdPartyID
					}
				}
			}

			if commonPatternID == nil {
				id, err := h.createUnmatchedPattern(ctx, tx, tp)
				if err != nil {
					return fmt.Errorf("cannot create unmatched pattern: %w", err)
				}

				commonPatternID = id
			}

			thirdPartyID := tp.ThirdPartyID

			if thirdPartyID == nil {
				switch {
				case directThirdPartyID != nil:
					thirdPartyID = directThirdPartyID
				case commonThirdPartyID != nil:
					allowCreate, err := h.creationAllowed(ctx, tx, scope, tp)
					if err != nil {
						return err
					}

					resolved, err := h.resolveOrgThirdParty(ctx, tx, tp, *commonThirdPartyID, allowCreate)
					if err != nil {
						return fmt.Errorf("cannot resolve org third party: %w", err)
					}

					thirdPartyID = resolved
				}
			}

			if commonPatternID != nil || thirdPartyID != nil {
				tp.CommonTrackerPatternID = commonPatternID
				tp.ThirdPartyID = thirdPartyID
				tp.UpdatedAt = time.Now()

				if tp.Description == "" && commonPatternID != nil {
					var commonPattern coredata.CommonTrackerPattern
					if err := commonPattern.LoadByID(ctx, tx, *commonPatternID); err == nil && commonPattern.Description != "" {
						tp.Description = commonPattern.Description
					}
				}

				if err := tp.Update(ctx, tx, scope); err != nil {
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
				if commonThirdPartyID != nil && !commonThirdPartyPreexisted {
					if err := h.reenqueueUnmappedSiblings(ctx, tx, tp, domains); err != nil {
						return err
					}
				}
			}

			return nil
		},
	)
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
		Description:        tp.Description,
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

func (h *trackerMappingHandler) identifyWithAgent(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
	siteOrigin string,
) (*catalogMatch, error) {
	var trackers coredata.DetectedTrackers

	domains, err := trackers.LoadInitiatorDomainsByTrackerPatternID(ctx, tx, tp.ID, 5)
	if err != nil {
		h.logger.WarnCtx(ctx, "cannot load initiator domains for agent", log.Error(err))
	}

	domains = uri.FilterFirstPartyDomains(domains, siteOrigin)

	prompt := buildAgentPrompt(tp, domains)

	agentCtx, cancel := context.WithTimeout(ctx, agentTimeout)
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

	if identification.Confidence < agentConfidenceThreshold {
		h.logger.InfoCtx(
			ctx,
			"agent identification below confidence threshold",
			log.String("pattern", tp.Pattern),
			log.Float64("confidence", identification.Confidence),
		)

		return nil, nil
	}

	confidence := float32(identification.Confidence)
	if confidence > agentMaxPatternConfidence {
		confidence = agentMaxPatternConfidence
	}

	var commonThirdPartyID *gid.GID
	if identification.ThirdPartyName != "" {
		commonThirdPartyID, err = h.resolveOrCreateCommonThirdParty(
			ctx,
			tx,
			identification,
			domains,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve or create common third party: %w", err)
		}
	}

	now := time.Now()
	commonPattern := coredata.CommonTrackerPattern{
		ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		CommonThirdPartyID: commonThirdPartyID,
		TrackerType:        tp.TrackerType,
		Pattern:            tp.Pattern,
		MatchType:          tp.MatchType,
		Description:        identification.Description,
		MaxAgeSeconds:      tp.MaxAgeSeconds,
		Confidence:         confidence,
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
		log.String("third_party", identification.ThirdPartyName),
		log.Float64("confidence", identification.Confidence),
	)

	return &catalogMatch{
		commonPatternID:    &commonPattern.ID,
		commonThirdPartyID: commonPattern.CommonThirdPartyID,
	}, nil
}

func (h *trackerMappingHandler) resolveOrCreateCommonThirdParty(
	ctx context.Context,
	tx pg.Tx,
	identification TrackerMappingAgentResult,
	domains []string,
) (*gid.GID, error) {
	var party coredata.CommonThirdParty
	if err := party.LoadByName(ctx, tx, identification.ThirdPartyName); err == nil {
		return &party.ID, nil
	}

	partySlug := slug.Make(identification.ThirdPartyName)
	if partySlug == "" {
		return nil, nil
	}

	if err := party.LoadBySlug(ctx, tx, partySlug); err == nil {
		return &party.ID, nil
	}

	category := identification.Category

	now := time.Now()
	party = coredata.CommonThirdParty{
		ID:             gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
		Name:           identification.ThirdPartyName,
		Slug:           partySlug,
		Category:       category,
		Certifications: []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := party.Insert(ctx, tx); err != nil {
		return nil, fmt.Errorf("cannot create common third party: %w", err)
	}

	for _, domain := range domains {
		domainRecord := coredata.CommonThirdPartyDomain{
			ID:                 gid.New(gid.NilTenant, coredata.CommonThirdPartyDomainEntityType),
			CommonThirdPartyID: party.ID,
			Domain:             domain,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		if _, err := domainRecord.Upsert(ctx, tx); err != nil {
			return nil, fmt.Errorf("cannot create common third party domain: %w", err)
		}
	}

	h.logger.InfoCtx(
		ctx,
		"created common third party from agent identification",
		log.String("name", identification.ThirdPartyName),
		log.String("category", category.String()),
	)

	return &party.ID, nil
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
		Description:        tp.Description,
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
		Description:   tp.Description,
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
	tx pg.Tx,
	tp coredata.TrackerPattern,
	commonThirdPartyID gid.GID,
	allowCreate bool,
) (*gid.GID, error) {
	scope := coredata.NewScopeFromObjectID(tp.ID)

	var existing coredata.ThirdParty

	err := existing.LoadByOrganizationIDAndCommonThirdPartyID(
		ctx,
		tx,
		scope,
		tp.OrganizationID,
		commonThirdPartyID,
	)
	if err == nil {
		return &existing.ID, nil
	}

	if !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load org third party by common id: %w", err)
	}

	var commonParty coredata.CommonThirdParty
	if err := commonParty.LoadByID(ctx, tx, commonThirdPartyID); err != nil {
		return nil, fmt.Errorf("cannot load common third party: %w", err)
	}

	var commonDomains coredata.CommonThirdPartyDomains
	if err := commonDomains.LoadByCommonThirdPartyID(ctx, tx, commonThirdPartyID); err != nil {
		return nil, fmt.Errorf("cannot load common third party domains: %w", err)
	}

	var orgThirdParties coredata.ThirdParties
	if err := orgThirdParties.LoadAllByOrganizationID(ctx, tx, scope, tp.OrganizationID); err != nil {
		return nil, fmt.Errorf("cannot load org third parties: %w", err)
	}

	ranked := thirdparty.RankCandidates(commonParty, commonDomains, orgThirdParties)

	if len(ranked) > 0 && ranked[0].Score >= thirdparty.HighConfidenceScore {
		picked := ranked[0].ThirdParty

		if err := thirdparty.LinkToCommon(ctx, tx, scope, picked, commonThirdPartyID); err != nil {
			return nil, fmt.Errorf("cannot link fuzzy-matched third party to common: %w", err)
		}

		h.logger.InfoCtx(
			ctx,
			"promoted tracker pattern via heuristic match",
			log.String("tracker_pattern_id", tp.ID.String()),
			log.String("third_party_id", picked.ID.String()),
			log.Float64("score", ranked[0].Score),
		)

		return &picked.ID, nil
	}

	agentSet := ranked
	if len(agentSet) > thirdparty.MaxAgentCandidates {
		agentSet = agentSet[:thirdparty.MaxAgentCandidates]
	}

	eligibleForAgent := false

	for _, c := range agentSet {
		if c.Score >= thirdparty.MinAgentScore {
			eligibleForAgent = true

			break
		}
	}

	if eligibleForAgent && h.disambiguationAgent != nil {
		matchedID, err := thirdparty.Disambiguate(
			ctx,
			h.disambiguationAgent,
			h.logger,
			commonParty,
			commonDomains,
			agentSet,
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
			var picked *coredata.ThirdParty

			for _, c := range agentSet {
				if c.ThirdParty.ID == *matchedID {
					picked = c.ThirdParty

					break
				}
			}

			if picked != nil {
				if err := thirdparty.LinkToCommon(ctx, tx, scope, picked, commonThirdPartyID); err != nil {
					return nil, fmt.Errorf("cannot link agent-matched third party to common: %w", err)
				}

				h.logger.InfoCtx(
					ctx,
					"promoted tracker pattern via disambiguation agent",
					log.String("tracker_pattern_id", tp.ID.String()),
					log.String("third_party_id", picked.ID.String()),
				)

				return &picked.ID, nil
			}
		}
	}

	if !allowCreate {
		return nil, nil
	}

	created, err := thirdparty.CreateFromCommon(ctx, tx, scope, tp.OrganizationID, commonParty)
	if err != nil {
		return nil, fmt.Errorf("cannot create third party from common: %w", err)
	}

	h.logger.InfoCtx(
		ctx,
		"promoted tracker pattern by creating org third party from catalog",
		log.String("tracker_pattern_id", tp.ID.String()),
		log.String("third_party_id", created.ID.String()),
		log.String("common_third_party_id", commonThirdPartyID.String()),
	)

	return &created.ID, nil
}
