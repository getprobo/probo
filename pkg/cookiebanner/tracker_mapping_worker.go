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
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/search"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
	"go.probo.inc/probo/pkg/slug"
)

const (
	agentTimeout              = 60 * time.Second
	agentMaxTurns             = 5
	agentConfidenceThreshold  = 0.6
	agentMaxPatternConfidence = 0.8
)

//go:embed prompts/tracker_identification.txt
var trackerIdentificationPrompt string

type trackerMappingHandler struct {
	pg     *pg.Client
	logger *log.Logger
	agent  *agent.Agent
}

type TrackerMappingConfig struct {
	LLMClient         *llm.Client
	Model             string
	SearchEndpoint    string
	FirecrawlEndpoint string
	FirecrawlAPIKey   string
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

func buildTrackerMappingAgent(
	cfg TrackerMappingConfig,
	pgClient *pg.Client,
	logger *log.Logger,
) *agent.Agent {
	tools := []agent.Tool{
		searchTrackerPatternsTool(pgClient),
		searchThirdPartiesTool(pgClient),
	}

	if cfg.FirecrawlEndpoint != "" && cfg.FirecrawlAPIKey != "" {
		tools = append(tools, search.FirecrawlSearchTool(cfg.FirecrawlEndpoint, cfg.FirecrawlAPIKey))
	} else if cfg.SearchEndpoint != "" {
		tools = append(tools, search.WebSearchTool(cfg.SearchEndpoint))
	}

	outputType, err := agent.NewOutputType[TrackerIdentification]("tracker_identification")
	if err != nil {
		panic(fmt.Sprintf("cookiebanner: cannot build tracker identification output type: %s", err))
	}

	return agent.New(
		"tracker-mapping",
		cfg.LLMClient,
		agent.WithInstructions(trackerIdentificationPrompt),
		agent.WithModel(cfg.Model),
		agent.WithTools(tools...),
		agent.WithOutputType(outputType),
		agent.WithMaxTurns(agentMaxTurns),
		agent.WithLogger(logger),
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
			var commonPatternID *gid.GID
			var thirdPartyID *gid.GID

			commonPatternID, thirdPartyID = h.matchByPattern(ctx, tx, tp)

			if commonPatternID == nil {
				commonPatternID, thirdPartyID = h.matchByDomain(ctx, tx, tp)
			}

			if commonPatternID == nil && h.agent != nil {
				commonPatternID, thirdPartyID = h.identifyWithAgent(ctx, tx, tp)
			}

			if commonPatternID == nil {
				commonPatternID = h.createUnmatchedPattern(ctx, tx, tp)
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
) (*gid.GID, *gid.GID) {
	var commonPattern coredata.CommonTrackerPattern
	if err := commonPattern.LoadByPattern(ctx, conn, tp.TrackerType, tp.Pattern, tp.MaxAgeSeconds); err != nil {
		if !errors.Is(err, coredata.ErrResourceNotFound) {
			h.logger.ErrorCtx(ctx, "cannot load common tracker pattern", log.Error(err))
		}
		return nil, nil
	}

	var thirdPartyID *gid.GID
	if commonPattern.CommonThirdPartyID != nil {
		var err error
		thirdPartyID, err = h.resolveThirdParty(ctx, conn, tp, &commonPattern)
		if err != nil {
			h.logger.ErrorCtx(ctx, "cannot resolve third party from pattern match", log.Error(err))
			return nil, nil
		}
	}

	return &commonPattern.ID, thirdPartyID
}

func (h *trackerMappingHandler) matchByDomain(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
) (*gid.GID, *gid.GID) {
	var trackers coredata.DetectedTrackers
	commonThirdPartyID, err := trackers.LoadCommonThirdPartyIDByDomainMatch(ctx, tx, tp.ID)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot load common third party ID from domain", log.Error(err))
		return nil, nil
	}

	if commonThirdPartyID == nil {
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

	actualID, _, err := commonPattern.Upsert(ctx, tx)
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot upsert common tracker pattern from domain match",
			log.Error(err),
		)
		return nil, nil
	}

	commonPattern.ID = actualID
	thirdPartyID, err := h.resolveThirdParty(ctx, tx, tp, &commonPattern)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot resolve third party from domain match", log.Error(err))
		return &commonPattern.ID, nil
	}

	return &commonPattern.ID, thirdPartyID
}

func (h *trackerMappingHandler) identifyWithAgent(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
) (*gid.GID, *gid.GID) {
	var trackers coredata.DetectedTrackers
	domains, err := trackers.LoadInitiatorDomainsByTrackerPatternID(ctx, tx, tp.ID)
	if err != nil {
		h.logger.WarnCtx(ctx, "cannot load initiator domains for agent", log.Error(err))
	}

	prompt := buildAgentPrompt(tp, domains)

	agentCtx, cancel := context.WithTimeout(ctx, agentTimeout)
	defer cancel()

	result, err := h.agent.Run(
		agentCtx,
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

	var identification TrackerIdentification
	if err := json.Unmarshal([]byte(result.FinalMessage().Text()), &identification); err != nil {
		h.logger.WarnCtx(
			ctx,
			"cannot parse agent identification output",
			log.Error(err),
			log.String("pattern", tp.Pattern),
		)
		return nil, nil
	}

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
		commonThirdPartyID = h.resolveOrCreateCommonThirdParty(
			ctx,
			tx,
			identification,
			domains,
		)
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

	actualID, _, err := commonPattern.Upsert(ctx, tx)
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot upsert common tracker pattern from agent",
			log.Error(err),
			log.String("pattern", tp.Pattern),
		)
		return nil, nil
	}

	commonPattern.ID = actualID
	thirdPartyID, err := h.resolveThirdParty(ctx, tx, tp, &commonPattern)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot resolve third party from agent match", log.Error(err))
		return &commonPattern.ID, nil
	}

	h.logger.InfoCtx(
		ctx,
		"agent identified tracker pattern",
		log.String("pattern", tp.Pattern),
		log.String("third_party", identification.ThirdPartyName),
		log.Float64("confidence", identification.Confidence),
	)

	return &commonPattern.ID, thirdPartyID
}

func (h *trackerMappingHandler) resolveOrCreateCommonThirdParty(
	ctx context.Context,
	tx pg.Tx,
	identification TrackerIdentification,
	domains []string,
) *gid.GID {
	var party coredata.CommonThirdParty
	if err := party.LoadByName(ctx, tx, identification.ThirdPartyName); err == nil {
		return &party.ID
	}

	partySlug := slug.Make(identification.ThirdPartyName)
	if partySlug == "" {
		return nil
	}

	if err := party.LoadBySlug(ctx, tx, partySlug); err == nil {
		return &party.ID
	}

	category := coredata.ThirdPartyCategoryOther
	if parsed := parseThirdPartyCategory(identification.Category); parsed != "" {
		category = parsed
	}

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
		h.logger.WarnCtx(
			ctx,
			"cannot create common third party from agent",
			log.Error(err),
			log.String("name", identification.ThirdPartyName),
		)
		return nil
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
			h.logger.WarnCtx(
				ctx,
				"cannot create common third party domain from agent",
				log.Error(err),
				log.String("domain", domain),
			)
		}
	}

	h.logger.InfoCtx(
		ctx,
		"created common third party from agent identification",
		log.String("name", identification.ThirdPartyName),
		log.String("category", string(category)),
	)

	return &party.ID
}

func (h *trackerMappingHandler) createUnmatchedPattern(
	ctx context.Context,
	tx pg.Tx,
	tp coredata.TrackerPattern,
) *gid.GID {
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

	actualID, _, err := commonPattern.Upsert(ctx, tx)
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot upsert common tracker pattern for unmatched pattern",
			log.Error(err),
		)
		return nil
	}

	return &actualID
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

func buildAgentPrompt(tp coredata.TrackerPattern, domains []string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Identify the following tracker:\n\n")
	fmt.Fprintf(&b, "- Pattern: %s\n", tp.Pattern)
	fmt.Fprintf(&b, "- Type: %s\n", tp.TrackerType)
	fmt.Fprintf(&b, "- Match type: %s\n", tp.MatchType)

	if tp.MaxAgeSeconds != nil {
		fmt.Fprintf(&b, "- Max age: %d seconds\n", *tp.MaxAgeSeconds)
	} else {
		fmt.Fprintf(&b, "- Max age: session\n")
	}

	if len(domains) > 0 {
		fmt.Fprintf(&b, "- Observed on domains: %s\n", strings.Join(domains, ", "))
	}

	return b.String()
}

func parseThirdPartyCategory(s string) coredata.ThirdPartyCategory {
	switch s {
	case "ANALYTICS":
		return coredata.ThirdPartyCategoryAnalytics
	case "ADVERTISING":
		return coredata.ThirdPartyCategoryMarketing
	case "CLOUD_MONITORING":
		return coredata.ThirdPartyCategoryCloudMonitoring
	case "CLOUD_PROVIDER":
		return coredata.ThirdPartyCategoryCloudProvider
	case "COLLABORATION":
		return coredata.ThirdPartyCategoryCollaboration
	case "CUSTOMER_SUPPORT":
		return coredata.ThirdPartyCategoryCustomerSupport
	case "DATA_STORAGE_AND_PROCESSING":
		return coredata.ThirdPartyCategoryDataStorageAndProcessing
	case "DOCUMENT_MANAGEMENT":
		return coredata.ThirdPartyCategoryDocumentManagement
	case "EMPLOYEE_MANAGEMENT":
		return coredata.ThirdPartyCategoryEmployeeManagement
	case "ENGINEERING":
		return coredata.ThirdPartyCategoryEngineering
	case "FINANCE":
		return coredata.ThirdPartyCategoryFinance
	case "IDENTITY_PROVIDER":
		return coredata.ThirdPartyCategoryIdentityProvider
	case "IT":
		return coredata.ThirdPartyCategoryIT
	case "MARKETING":
		return coredata.ThirdPartyCategoryMarketing
	case "OFFICE_OPERATIONS":
		return coredata.ThirdPartyCategoryOfficeOperations
	case "OTHER":
		return coredata.ThirdPartyCategoryOther
	case "PASSWORD_MANAGEMENT":
		return coredata.ThirdPartyCategoryPasswordManagement
	case "PRODUCT_AND_DESIGN":
		return coredata.ThirdPartyCategoryProductAndDesign
	case "PROFESSIONAL_SERVICES":
		return coredata.ThirdPartyCategoryProfessionalServices
	case "RECRUITING":
		return coredata.ThirdPartyCategoryRecruiting
	case "SALES":
		return coredata.ThirdPartyCategorySales
	case "SECURITY":
		return coredata.ThirdPartyCategorySecurity
	case "VERSION_CONTROL":
		return coredata.ThirdPartyCategoryVersionControl
	default:
		return ""
	}
}
