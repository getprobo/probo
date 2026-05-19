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
)

type trackerMappingHandler struct {
	pg     *pg.Client
	logger *log.Logger
	agent  *agent.Agent
}

func NewTrackerMappingWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	cfg TrackerMappingConfig,
	opts ...worker.Option,
) *worker.Worker[coredata.TrackerPattern] {
	h := &trackerMappingHandler{
		pg:     pgClient,
		logger: logger,
	}

	if cfg.LLMClient != nil {
		h.agent = buildTrackerMappingAgent(cfg, pgClient, logger)
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

func (h *trackerMappingHandler) Process(ctx context.Context, tp coredata.TrackerPattern) error {
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var (
				commonPatternID *gid.GID
				thirdPartyID    *gid.GID
				err             error
			)

			commonPatternID, thirdPartyID, err = h.matchByPattern(ctx, tx, tp)
			if err != nil {
				return fmt.Errorf("cannot match by pattern: %w", err)
			}

			if commonPatternID == nil {
				commonPatternID, thirdPartyID, err = h.matchByDomain(ctx, tx, tp)
				if err != nil {
					return fmt.Errorf("cannot match by domain: %w", err)
				}
			}

			if commonPatternID == nil && h.agent != nil {
				commonPatternID, thirdPartyID, err = h.identifyWithAgent(ctx, tx, tp)
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

func (h *trackerMappingHandler) matchByPattern(
	ctx context.Context,
	conn pg.Querier,
	tp coredata.TrackerPattern,
) (*gid.GID, *gid.GID, error) {
	var commonPattern coredata.CommonTrackerPattern
	if err := commonPattern.LoadByPattern(ctx, conn, tp.TrackerType, tp.Pattern, tp.MaxAgeSeconds); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil, nil
		}

		return nil, nil, fmt.Errorf("cannot load common tracker pattern: %w", err)
	}

	var thirdPartyID *gid.GID

	if commonPattern.CommonThirdPartyID != nil {
		var err error

		thirdPartyID, err = h.resolveThirdParty(ctx, conn, tp, &commonPattern)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot resolve third party from pattern match: %w", err)
		}
	}

	return &commonPattern.ID, thirdPartyID, nil
}

func (h *trackerMappingHandler) matchByDomain(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
) (*gid.GID, *gid.GID, error) {
	var trackers coredata.DetectedTrackers

	domains, err := trackers.LoadInitiatorDomainsByTrackerPatternID(ctx, tx, tp.ID, 10)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot load initiator domains: %w", err)
	}

	if len(domains) == 0 {
		return nil, nil, nil
	}

	filter := coredata.NewCommonThirdPartyDomainFilter(domains)

	var matchedDomains coredata.CommonThirdPartyDomains
	if err := matchedDomains.Load(ctx, tx, 1, filter); err != nil {
		return nil, nil, fmt.Errorf("cannot load common third party domain by domain match: %w", err)
	}

	if len(matchedDomains) == 0 {
		return nil, nil, nil
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
		return nil, nil, fmt.Errorf("cannot upsert common tracker pattern from domain match: %w", err)
	}

	thirdPartyID, err := h.resolveThirdParty(ctx, tx, tp, &commonPattern)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot resolve third party from domain match: %w", err)
	}

	return &commonPattern.ID, thirdPartyID, nil
}

func (h *trackerMappingHandler) identifyWithAgent(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
) (*gid.GID, *gid.GID, error) {
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
		h.agent,
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

		return nil, nil, nil
	}

	identification := result.Output

	if identification.Confidence < agentConfidenceThreshold {
		h.logger.InfoCtx(
			ctx,
			"agent identification below confidence threshold",
			log.String("pattern", tp.Pattern),
			log.Float64("confidence", identification.Confidence),
		)

		return nil, nil, nil
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
			return nil, nil, fmt.Errorf("cannot resolve or create common third party: %w", err)
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
		return nil, nil, fmt.Errorf("cannot upsert common tracker pattern from agent: %w", err)
	}

	thirdPartyID, err := h.resolveThirdParty(ctx, tx, tp, &commonPattern)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot resolve third party from agent match: %w", err)
	}

	h.logger.InfoCtx(
		ctx,
		"agent identified tracker pattern",
		log.String("pattern", tp.Pattern),
		log.String("third_party", identification.ThirdPartyName),
		log.Float64("confidence", identification.Confidence),
	)

	return &commonPattern.ID, thirdPartyID, nil
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

func (h *trackerMappingHandler) resolveThirdParty(
	ctx context.Context,
	conn pg.Querier,
	tp coredata.TrackerPattern,
	commonPattern *coredata.CommonTrackerPattern,
) (*gid.GID, error) {
	if commonPattern.CommonThirdPartyID == nil {
		return nil, nil
	}

	scope := coredata.NewScopeFromObjectID(tp.ID)

	var t coredata.ThirdParty
	if err := t.LoadByOrganizationIDAndCommonThirdPartyID(
		ctx,
		conn,
		scope,
		tp.OrganizationID,
		*commonPattern.CommonThirdPartyID,
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot resolve third party: %w", err)
	}

	return &t.ID, nil
}
