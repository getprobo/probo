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

// Process resolves the catalog mapping (when missing) and then promotes
// the pattern to an org ThirdParty (when eligible). Promotion is
// skipped for patterns still in the uncategorised category — the user
// must categorize a tracker before it creates or links an org
// ThirdParty. When a pattern is re-triggered by a manual move (it
// already carries a common_tracker_pattern_id), we MUST NOT re-resolve
// the catalog: the existing link is preserved and we jump straight to
// third-party promotion.
func (h *trackerMappingHandler) Process(ctx context.Context, tp coredata.TrackerPattern) error {
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var (
				commonPatternID *gid.GID
				err             error
			)

			if tp.CommonTrackerPatternID != nil {
				commonPatternID = tp.CommonTrackerPatternID
			} else {
				commonPatternID, err = h.matchByPattern(ctx, tx, tp)
				if err != nil {
					return fmt.Errorf("cannot match by pattern: %w", err)
				}

				if commonPatternID == nil {
					commonPatternID, err = h.matchByDomain(ctx, tx, tp)
					if err != nil {
						return fmt.Errorf("cannot match by domain: %w", err)
					}
				}

				if commonPatternID == nil && h.mappingAgent != nil {
					commonPatternID, err = h.identifyWithAgent(ctx, tx, tp)
					if err != nil {
						return fmt.Errorf("cannot identify with agent: %w", err)
					}
				}

				if commonPatternID == nil {
					commonPatternID, err = h.createUnmatchedPattern(ctx, tx, tp)
					if err != nil {
						return fmt.Errorf("cannot create unmatched pattern: %w", err)
					}
				}
			}

			thirdPartyID := tp.ThirdPartyID

			if thirdPartyID == nil &&
				commonPatternID != nil &&
				(tp.Source == nil || *tp.Source != coredata.CookieSourceExtension) {
				scope := coredata.NewScopeFromObjectID(tp.ID)

				var category coredata.CookieCategory
				if err := category.LoadByID(ctx, tx, scope, tp.CookieCategoryID); err != nil {
					return fmt.Errorf("cannot load cookie category: %w", err)
				}

				if category.Kind != coredata.CookieCategoryKindUncategorised {
					promoted, err := h.promoteThirdParty(ctx, tx, tp, *commonPatternID)
					if err != nil {
						return fmt.Errorf("cannot promote third party: %w", err)
					}

					thirdPartyID = promoted
				}
			}

			if commonPatternID != nil || thirdPartyID != nil {
				if err := tp.UpdateMapping(ctx, tx, commonPatternID, thirdPartyID); err != nil {
					return fmt.Errorf("cannot update tracker pattern mapping: %w", err)
				}

				h.logger.InfoCtx(
					ctx,
					"mapped tracker pattern",
					log.String("pattern", tp.Pattern),
					log.String("tracker_pattern_id", tp.ID.String()),
				)
			}

			return nil
		},
	)
}

// matchByPattern looks for a catalog row with the same pattern. It now
// only returns the catalog ID; third-party resolution happens later in
// promoteThirdParty.
func (h *trackerMappingHandler) matchByPattern(
	ctx context.Context,
	conn pg.Querier,
	tp coredata.TrackerPattern,
) (*gid.GID, error) {
	var commonPattern coredata.CommonTrackerPattern
	if err := commonPattern.LoadByPattern(ctx, conn, tp.TrackerType, tp.Pattern, tp.MaxAgeSeconds); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load common tracker pattern: %w", err)
	}

	return &commonPattern.ID, nil
}

// matchByDomain finds a CommonThirdParty whose registered domains
// overlap the pattern's observed initiator domains, and upserts a
// CommonTrackerPattern linking the two. As with matchByPattern,
// third-party resolution is deferred to promoteThirdParty.
func (h *trackerMappingHandler) matchByDomain(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
) (*gid.GID, error) {
	var trackers coredata.DetectedTrackers

	domains, err := trackers.LoadInitiatorDomainsByTrackerPatternID(ctx, tx, tp.ID, 10)
	if err != nil {
		return nil, fmt.Errorf("cannot load initiator domains: %w", err)
	}

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

	return &commonPattern.ID, nil
}

func (h *trackerMappingHandler) identifyWithAgent(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
) (*gid.GID, error) {
	var trackers coredata.DetectedTrackers

	domains, err := trackers.LoadInitiatorDomainsByTrackerPatternID(ctx, tx, tp.ID, 5)
	if err != nil {
		h.logger.WarnCtx(ctx, "cannot load initiator domains for agent", log.Error(err))
	}

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

	return &commonPattern.ID, nil
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

// promoteThirdParty resolves an org ThirdParty for the given pattern
// once the catalog mapping is known. The resolution order is:
//
//  1. Exact link by common_third_party_id (O(1)).
//  2. Heuristic match against the org's existing ThirdParty rows
//     (lowercased name, suffix-stripped name, slug, website host,
//     CommonThirdPartyDomain overlap).
//  3. Agent disambiguation when the heuristic is ambiguous.
//  4. Fallback create from CommonThirdParty.
//
// A confident heuristic/agent match is auto-tagged with
// common_third_party_id so subsequent promotions hit the exact-link
// path in O(1). Returns (nil, nil) when the catalog row has no
// CommonThirdPartyID — there is nothing to promote to.
func (h *trackerMappingHandler) promoteThirdParty(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
	commonPatternID gid.GID,
) (*gid.GID, error) {
	var commonPattern coredata.CommonTrackerPattern
	if err := commonPattern.LoadByID(ctx, tx, commonPatternID); err != nil {
		return nil, fmt.Errorf("cannot load common tracker pattern: %w", err)
	}

	if commonPattern.CommonThirdPartyID == nil {
		return nil, nil
	}

	commonThirdPartyID := *commonPattern.CommonThirdPartyID
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
