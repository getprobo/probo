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
	"sync"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
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

	// defaultEnrichmentDomainConfidenceThreshold is the ownership-confidence
	// floor a discovered domain must clear before it is written to
	// common_third_party_domains. It is stricter than the field threshold
	// because a domain feeds cross-tenant tracker attribution: a wrong
	// domain mis-attributes every tracker served from it.
	defaultEnrichmentDomainConfidenceThreshold = 0.85

	// defaultEnrichmentStaleAfter is the idle window after which a
	// claimed-but-unfinished enrichment is re-armed.
	defaultEnrichmentStaleAfter = 15 * time.Minute

	// defaultEnrichmentMaxAttempts caps how many times a row is retried
	// before stale recovery leaves it alone, so a permanently failing
	// row does not loop forever.
	defaultEnrichmentMaxAttempts = 3

	enrichmentLogoUserAgent = "Probo-Enricher/1.0"

	// maxEnrichmentErrorLen bounds the per-agent error text persisted in
	// enrichment metadata so a verbose or hostile agent/tool error cannot
	// leak unbounded internal detail into the column.
	maxEnrichmentErrorLen = 500
)

// sanitizeAgentError reduces a raw agent or tool error to a single bounded
// line safe to persist in enrichment metadata: whitespace runs (including
// embedded newlines) collapse to single spaces and the result is truncated
// to maxEnrichmentErrorLen runes.
func sanitizeAgentError(err error) string {
	msg := strings.Join(strings.Fields(err.Error()), " ")

	if runes := []rune(msg); len(runes) > maxEnrichmentErrorLen {
		msg = string(runes[:maxEnrichmentErrorLen]) + "…"
	}

	return msg
}

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
	pg         *pg.Client
	logger     *log.Logger
	cfg        EnrichmentConfig
	httpClient *http.Client
	profiler   *Profiler
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
		profiler:   NewProfiler(cfg, logger),
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
// (company profile) first, then Agent B (compliance docs), Agent C
// (owned domains), and the deterministic logo step concurrently, all
// outside any transaction. Agent A runs first because the others depend
// on the website it resolves. The merged result and per-field provenance
// are persisted in a single final transaction. Process always writes an
// enrichment payload, even on a no-result run, so stale recovery does not
// re-queue the row.
func (h *enrichmentHandler) Process(ctx context.Context, party coredata.CommonThirdParty) error {
	if h.cfg.LLMClient == nil {
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
	company, err := h.profiler.runCompanyProfile(ctx, generalInfoInputFromParty(party))
	if err != nil {
		h.logger.WarnCtx(ctx, "company profile agent failed", log.Error(err), log.String("common_third_party_id", party.ID.String()))
		runErrors = append(runErrors, "company_profile: "+sanitizeAgentError(err))
	} else {
		anySuccess = true
	}

	website := effectiveWebsiteURL(party, company, h.cfg.ConfidenceThreshold)

	meta := make(map[string]EnrichmentFieldMeta)

	// Website is the hard precondition: Agent B and the logo step both
	// depend on it, and an Agent B run without a domain scope produces
	// inconsistent cross-domain results. When no website is resolved,
	// persist what Agent A found and stop here rather than running Agent
	// B blind.
	if website == "" {
		for _, field := range scalarFields(company, ComplianceDocsResult{}) {
			applyScalarField(&party, meta, prior, field, h.cfg.ConfidenceThreshold, now)
		}

		applyCertifications(&party, meta, prior, CertificationsField{}, h.cfg.ConfidenceThreshold, now)
		applyLegalNameFallback(&party, meta, now)

		runErrors = append(runErrors, "website_url unresolved: skipped compliance docs")

		payload := EnrichmentMetadata{
			Model:       h.cfg.Model,
			AttemptedAt: now,
			Status:      enrichmentStatusFailed,
			Error:       strings.Join(runErrors, "; "),
			Fields:      meta,
		}

		return h.persist(ctx, party, payload, nil, nil, now)
	}

	legalName := effectiveLegalName(party, company, h.cfg.ConfidenceThreshold)

	// Agent B (compliance docs), Agent C (owned domains), and the
	// deterministic logo step all depend only on the resolved website
	// (Agent B also on the legal name) and are independent of each other.
	// Run them concurrently so wall time is the slowest of the three
	// rather than their sum. Each builds its own per-run browser and
	// writes only into its own locals; the shared LLM/HTTP/FileManager
	// clients are safe for concurrent use and the database is not touched
	// until persist below. Results are merged after Wait in a fixed order
	// to keep runErrors and log output deterministic.
	var (
		compliance    ComplianceDocsResult
		complianceErr error

		domainsResult DomainsResult
		domainsErr    error

		logoFile *coredata.File
	)

	var wg sync.WaitGroup

	wg.Go(func() {
		compliance, complianceErr = h.profiler.runComplianceDocs(ctx, party.Name, website, legalName)
	})

	wg.Go(func() {
		domainsResult, domainsErr = h.profiler.runDomains(ctx, party.Name, website)
	})

	wg.Go(func() {
		logoFile = h.prepareLogo(ctx, party, website)
	})

	wg.Wait()

	// Agent B: compliance documents and trust pages.
	if complianceErr != nil {
		h.logger.WarnCtx(ctx, "compliance docs agent failed", log.Error(complianceErr), log.String("common_third_party_id", party.ID.String()))
		runErrors = append(runErrors, "compliance_docs: "+sanitizeAgentError(complianceErr))
	} else {
		anySuccess = true
	}

	// Agent C: domains the vendor owns and operates. Anchored on the
	// resolved website, so it runs only on this website-resolved path.
	var owned []ownedDomain

	if domainsErr != nil {
		h.logger.WarnCtx(ctx, "domains agent failed", log.Error(domainsErr), log.String("common_third_party_id", party.ID.String()))
		runErrors = append(runErrors, "domains: "+sanitizeAgentError(domainsErr))
	} else {
		anySuccess = true
		owned = resolveOwnedDomains(party.Name, website, domainsResult, defaultEnrichmentDomainConfidenceThreshold)
	}

	for _, field := range scalarFields(company, compliance) {
		applyScalarField(&party, meta, prior, field, h.cfg.ConfidenceThreshold, now)
	}

	applyCertifications(&party, meta, prior, compliance.Certifications, h.cfg.ConfidenceThreshold, now)
	applyLegalNameFallback(&party, meta, now)

	domainRows := make([]coredata.CommonThirdPartyDomain, 0, len(owned))
	domainMeta := make([]EnrichmentDomainMeta, 0, len(owned))

	for _, d := range owned {
		domainRows = append(domainRows, coredata.CommonThirdPartyDomain{
			ID:                 gid.New(gid.NilTenant, coredata.CommonThirdPartyDomainEntityType),
			CommonThirdPartyID: party.ID,
			Domain:             d.Domain,
			CreatedAt:          now,
			UpdatedAt:          now,
		})

		domainMeta = append(domainMeta, EnrichmentDomainMeta{
			Domain:     d.Domain,
			Confidence: d.Confidence,
			SourceURL:  d.SourceURL,
			UpdatedAt:  now,
		})
	}

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
		Domains:     domainMeta,
	}

	return h.persist(ctx, party, payload, logoFile, domainRows, now)
}

// persist marshals the enrichment payload onto the row and writes it in a
// single transaction, inserting the logo File row and linking it when one
// was prepared and upserting the discovered owned domains. It always
// writes an enrichment payload, even on a no-result run, so stale
// recovery does not re-queue the row. Domain upserts are idempotent
// against the (common_third_party_id, domain) unique index, so re-runs
// are safe and never conflict with curated seed rows.
func (h *enrichmentHandler) persist(
	ctx context.Context,
	party coredata.CommonThirdParty,
	payload EnrichmentMetadata,
	logoFile *coredata.File,
	domains []coredata.CommonThirdPartyDomain,
	now time.Time,
) error {
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

			var (
				newDomains []string
				remapped   int64
			)

			for i := range domains {
				inserted, err := domains[i].Upsert(ctx, tx)
				if err != nil {
					return fmt.Errorf("cannot upsert common third party domain: %w", err)
				}

				if inserted {
					newDomains = append(newDomains, domains[i].Domain)
				}
			}

			// A newly-discovered owned domain can resolve org tracker
			// patterns that were detected and left unmapped before the
			// domain was known. Re-arm those still-unmapped patterns so
			// the mapping worker re-resolves them via its domain-overlap
			// signal. Only newly-inserted domains trigger this: a re-run
			// that rediscovers existing domains cascaded them already.
			if len(newDomains) > 0 {
				var patterns coredata.TrackerPatterns

				n, err := patterns.RequestMappingForUnmappedByInitiatorDomains(ctx, tx, newDomains)
				if err != nil {
					return fmt.Errorf("cannot re-arm mapping for unmapped tracker patterns: %w", err)
				}

				remapped = n
			}

			h.logger.InfoCtx(
				ctx,
				"enriched common third party",
				log.String("common_third_party_id", party.ID.String()),
				log.String("name", party.Name),
				log.String("status", payload.Status),
				log.Bool("logo_stored", logoFile != nil),
				log.Int("domains_stored", len(domains)),
				log.Int("domains_new", len(newDomains)),
				log.Int64("tracker_patterns_remapped", remapped),
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

// generalInfoInputFromParty maps a catalog row to the general-info input.
func generalInfoInputFromParty(party coredata.CommonThirdParty) GeneralInfoInput {
	in := GeneralInfoInput{Name: party.Name}

	if party.WebsiteURL != nil {
		in.WebsiteURL = strings.TrimSpace(*party.WebsiteURL)
	}

	if party.LegalName != nil {
		in.LegalName = strings.TrimSpace(*party.LegalName)
	}

	return in
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
