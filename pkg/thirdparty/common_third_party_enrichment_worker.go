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

package thirdparty

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/browser"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
)

const (
	// defaultEnrichmentAgentTimeout caps a single enrichment agent run.
	// It is generous because Agent B browses several pages on top of the
	// LLM round-trips.
	defaultEnrichmentAgentTimeout = 90 * time.Second

	// defaultEnrichmentMaxTurns bounds an agent's reasoning loop (LLM
	// call plus tool round-trips). Agent B may navigate the site footer
	// and trust portal across several turns before synthesizing.
	defaultEnrichmentMaxTurns = 12

	// defaultEnrichmentMaxTokens caps agent output. The structured
	// output is moderate; the budget leaves headroom for reasoning
	// models whose reasoning tokens count against max_tokens.
	defaultEnrichmentMaxTokens = 8192

	// defaultEnrichmentConfidenceThreshold is the floor a resolved value
	// must clear before it is written to its column. Values below it are
	// recorded in the enrichment metadata but not promoted.
	defaultEnrichmentConfidenceThreshold = 0.7

	// defaultEnrichmentStaleAfter is the idle window after which a
	// claimed-but-unfinished enrichment is re-armed.
	defaultEnrichmentStaleAfter = 15 * time.Minute

	// defaultEnrichmentMaxAttempts caps how many times a row is retried
	// before stale recovery leaves it alone, so a permanently failing
	// row does not loop forever.
	defaultEnrichmentMaxAttempts = 3

	enrichmentLogoUserAgent = "Probo-Enricher/1.0"
)

// EnrichmentConfig configures the common-third-party enrichment worker
// and the two agents it runs. The worker no-ops when LLMClient is nil;
// callers gate registration on config presence. Browser tools for Agent
// B are enabled only when ChromeAddr is set; otherwise it relies on
// web_search alone. Logo storage is enabled only when FileManager and
// Bucket are both set.
type EnrichmentConfig struct {
	LLMClient           *llm.Client
	Model               string
	MaxTokens           *int
	Temperature         *float64
	FirecrawlAPIKey     string
	ChromeAddr          string
	AgentTimeout        time.Duration
	MaxTurns            int
	ConfidenceThreshold float64
	StaleAfter          time.Duration
	MaxAttempts         int

	FileManager *filemanager.Service
	Bucket      string
}

func (c EnrichmentConfig) withDefaults() EnrichmentConfig {
	if c.AgentTimeout <= 0 {
		c.AgentTimeout = defaultEnrichmentAgentTimeout
	}

	if c.MaxTurns < 1 {
		c.MaxTurns = defaultEnrichmentMaxTurns
	}

	if c.ConfidenceThreshold <= 0 {
		c.ConfidenceThreshold = defaultEnrichmentConfidenceThreshold
	}

	if c.StaleAfter <= 0 {
		c.StaleAfter = defaultEnrichmentStaleAfter
	}

	if c.MaxAttempts < 1 {
		c.MaxAttempts = defaultEnrichmentMaxAttempts
	}

	return c
}

func resolveEnrichmentMaxTurns(configured int) int {
	if configured > 0 {
		return configured
	}

	return defaultEnrichmentMaxTurns
}

func resolveEnrichmentMaxTokens(configured *int) int {
	if configured != nil && *configured > 0 {
		return *configured
	}

	return defaultEnrichmentMaxTokens
}

type enrichmentHandler struct {
	pg           *pg.Client
	logger       *log.Logger
	cfg          EnrichmentConfig
	companyAgent *agent.Agent
	httpClient   *http.Client
}

// NewCommonThirdPartyEnrichmentWorker builds the worker that enriches
// global common_third_parties rows. It is a system worker: the catalog
// is not tenant-scoped, so a single enrichment benefits all tenants. The
// worker no-ops when no LLM client is configured.
func NewCommonThirdPartyEnrichmentWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	cfg EnrichmentConfig,
	opts ...worker.Option,
) *worker.Worker[coredata.CommonThirdParty] {
	cfg = cfg.withDefaults()

	h := &enrichmentHandler{
		pg:         pgClient,
		logger:     logger,
		cfg:        cfg,
		httpClient: newEnrichmentHTTPClient(),
	}

	// Agent A has no browser, so it is built once and reused. Agent B is
	// built per Process because it needs a per-run browser bound to the
	// process context.
	if cfg.LLMClient != nil {
		h.companyAgent = buildCompanyProfileAgent(cfg, logger)
	}

	return worker.New(
		"common-third-party-enrichment-worker",
		h,
		logger,
		opts...,
	)
}

func (h *enrichmentHandler) Claim(ctx context.Context) (coredata.CommonThirdParty, error) {
	var party coredata.CommonThirdParty

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := party.LoadNextForEnrichmentForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			return party.ClearEnrichmentRequestedAt(ctx, tx)
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.CommonThirdParty{}, worker.ErrNoTask
		}

		return coredata.CommonThirdParty{}, fmt.Errorf("cannot claim common third party enrichment task: %w", err)
	}

	return party, nil
}

// Process runs the enrichment pipeline for one catalog row: Agent A
// (company profile) first, then Agent B (compliance docs) and the
// deterministic logo step, all outside any transaction. The merged
// result and per-field provenance are persisted in a single final
// transaction. Process always writes an enrichment payload, even on a
// no-result run, so stale recovery does not re-queue the row.
func (h *enrichmentHandler) Process(ctx context.Context, party coredata.CommonThirdParty) error {
	if h.companyAgent == nil {
		return nil
	}

	now := time.Now()
	prior := parseEnrichmentFields(party.Enrichment)

	var (
		runErrors  []string
		anySuccess bool
	)

	// Agent A first: it resolves website_url, which Agent B and the logo
	// step depend on.
	company, err := h.runCompanyProfile(ctx, party)
	if err != nil {
		h.logger.WarnCtx(ctx, "company profile agent failed", log.Error(err), log.String("common_third_party_id", party.ID.String()))
		runErrors = append(runErrors, "company_profile: "+err.Error())
	} else {
		anySuccess = true
	}

	website := effectiveWebsiteURL(party, company, h.cfg.ConfidenceThreshold)
	legalName := effectiveLegalName(party, company, h.cfg.ConfidenceThreshold)

	// Agent B: compliance documents and trust pages.
	compliance, err := h.runComplianceDocs(ctx, party.Name, website, legalName)
	if err != nil {
		h.logger.WarnCtx(ctx, "compliance docs agent failed", log.Error(err), log.String("common_third_party_id", party.ID.String()))
		runErrors = append(runErrors, "compliance_docs: "+err.Error())
	} else {
		anySuccess = true
	}

	// Deterministic logo step (no LLM). Uploads to S3 outside the final
	// transaction; the File row is inserted below.
	logoFile := h.prepareLogo(ctx, party, website)

	meta := make(map[string]EnrichmentFieldMeta)

	for _, field := range scalarFields(company, compliance) {
		applyScalarField(&party, meta, prior, field, h.cfg.ConfidenceThreshold, now)
	}

	applyCertifications(&party, meta, prior, compliance.Certifications, h.cfg.ConfidenceThreshold, now)

	status := enrichmentStatusDone
	switch {
	case !anySuccess:
		status = enrichmentStatusFailed
	case len(runErrors) > 0:
		status = enrichmentStatusPartial
	}

	payload := EnrichmentMetadata{
		Model:       h.cfg.Model,
		AttemptedAt: now,
		Status:      status,
		Error:       strings.Join(runErrors, "; "),
		Fields:      meta,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot marshal enrichment metadata: %w", err)
	}

	party.Enrichment = raw
	party.UpdatedAt = now

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if logoFile != nil {
				if err := logoFile.Insert(ctx, tx, coredata.NewScope(gid.NilTenant)); err != nil {
					return fmt.Errorf("cannot insert common third party logo file: %w", err)
				}

				party.LogoFileID = &logoFile.ID

				if err := party.UpdateLogoFileID(ctx, tx); err != nil {
					return fmt.Errorf("cannot update common third party logo: %w", err)
				}
			}

			if err := party.UpdateEnrichment(ctx, tx); err != nil {
				return fmt.Errorf("cannot persist common third party enrichment: %w", err)
			}

			h.logger.InfoCtx(
				ctx,
				"enriched common third party",
				log.String("common_third_party_id", party.ID.String()),
				log.String("name", party.Name),
				log.String("status", status),
				log.Bool("logo_stored", logoFile != nil),
			)

			return nil
		},
	)
}

// RecoverStale re-arms enrichment for rows whose run was claimed but
// never finished. Claim clears enrichment_requested_at up front, so a
// crash between phases would otherwise strand the row.
func (h *enrichmentHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := coredata.ResetStaleCommonThirdPartyEnrichments(ctx, conn, h.cfg.StaleAfter, h.cfg.MaxAttempts); err != nil {
				return fmt.Errorf("cannot reset stale common third party enrichments: %w", err)
			}

			return nil
		},
	)
}

func (h *enrichmentHandler) runCompanyProfile(
	ctx context.Context,
	party coredata.CommonThirdParty,
) (CompanyProfileResult, error) {
	prompt := buildCompanyProfilePrompt(party)

	agentCtx, cancel := context.WithTimeout(ctx, h.cfg.AgentTimeout)
	defer cancel()

	result, err := agent.RunTyped[CompanyProfileResult](
		agentCtx,
		h.companyAgent,
		[]llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: prompt}},
			},
		},
	)
	if err != nil {
		return CompanyProfileResult{}, fmt.Errorf("company profile agent run failed: %w", err)
	}

	return result.Output, nil
}

// runComplianceDocs builds Agent B with a per-run browser when a Chrome
// endpoint is configured, then runs it. The browser is closed when the
// run returns. The browser is intentionally not pinned to the vendor
// domain so the agent can follow links to hosted trust portals (Vanta,
// SafeBase, etc.); SSRF protection still blocks non-public hosts.
func (h *enrichmentHandler) runComplianceDocs(
	ctx context.Context,
	name string,
	website string,
	legalName string,
) (ComplianceDocsResult, error) {
	var browserTools []agent.Tool

	if h.cfg.ChromeAddr != "" {
		webBrowser := browser.NewBrowser(ctx, h.cfg.ChromeAddr)
		defer webBrowser.Close()

		browserTools = browser.NewReadOnlyToolset(webBrowser).Tools()
	}

	complianceAgent := buildComplianceDocsAgent(h.cfg, h.logger, browserTools)

	prompt := buildComplianceDocsPrompt(name, website, legalName)

	agentCtx, cancel := context.WithTimeout(ctx, h.cfg.AgentTimeout)
	defer cancel()

	result, err := agent.RunTyped[ComplianceDocsResult](
		agentCtx,
		complianceAgent,
		[]llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: prompt}},
			},
		},
	)
	if err != nil {
		return ComplianceDocsResult{}, fmt.Errorf("compliance docs agent run failed: %w", err)
	}

	return result.Output, nil
}

// effectiveWebsiteURL is the website passed to Agent B and the logo step.
// A curated value already on the row wins (seed data and human edits are
// trusted); otherwise Agent A's value is used when it clears the
// confidence threshold.
func effectiveWebsiteURL(
	party coredata.CommonThirdParty,
	company CompanyProfileResult,
	threshold float64,
) string {
	if party.WebsiteURL != nil {
		if v := strings.TrimSpace(*party.WebsiteURL); v != "" {
			return v
		}
	}

	if v := strings.TrimSpace(company.WebsiteURL.Value); v != "" && company.WebsiteURL.Confidence >= threshold {
		return v
	}

	return ""
}

// effectiveLegalName is the legal name hint passed to Agent B, resolved
// the same way as effectiveWebsiteURL.
func effectiveLegalName(
	party coredata.CommonThirdParty,
	company CompanyProfileResult,
	threshold float64,
) string {
	if party.LegalName != nil {
		if v := strings.TrimSpace(*party.LegalName); v != "" {
			return v
		}
	}

	if v := strings.TrimSpace(company.LegalName.Value); v != "" && company.LegalName.Confidence >= threshold {
		return v
	}

	return ""
}

type userAgentRoundTripper struct {
	next http.RoundTripper
}

func (t *userAgentRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	r2.Header.Set("User-Agent", enrichmentLogoUserAgent)

	return t.next.RoundTrip(r2)
}

// newEnrichmentHTTPClient builds the SSRF-protected client used by the
// deterministic logo step.
func newEnrichmentHTTPClient() *http.Client {
	client := httpclient.DefaultPooledClient(httpclient.WithSSRFProtection())
	client.Timeout = 20 * time.Second
	client.Transport = &userAgentRoundTripper{next: client.Transport}

	return client
}
