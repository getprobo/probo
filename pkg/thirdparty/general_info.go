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

package thirdparty

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/browser"
	"go.probo.inc/probo/pkg/llm"
)

type (
	// GeneralInfoInput is the org-agnostic input to the general-info
	// agents. Name is required; WebsiteURL and LegalName are optional
	// hints. It carries no catalog/org identity so the same fact-finder
	// serves both the global catalog and org third parties.
	GeneralInfoInput struct {
		Name       string
		WebsiteURL string
		LegalName  string
	}

	// GeneralInfoProfile is the merged factual profile from the
	// general-info agents. WebsiteURL is what the profiler scoped Agents B
	// and C to, resolved from the input hint or Agent A's output.
	GeneralInfoProfile struct {
		Company    CompanyProfileResult
		Compliance ComplianceDocsResult
		Domains    DomainsResult
		WebsiteURL string
	}

	// Profiler runs the general-info agents against a GeneralInfoInput and
	// owns no persistence: callers decide what to do with the profile.
	Profiler struct {
		cfg    EnrichmentConfig
		logger *log.Logger
	}
)

// NewProfiler builds a Profiler from the enrichment worker's config.
func NewProfiler(cfg EnrichmentConfig, logger *log.Logger) *Profiler {
	return &Profiler{
		cfg:    cfg.withDefaults(),
		logger: logger,
	}
}

// Profile runs the pipeline best-effort. Agent A runs first because the
// website it resolves scopes Agents B and C; without a website, B and C
// are skipped. Agent A failing is fatal; B or C failing is only logged.
func (p *Profiler) Profile(ctx context.Context, in GeneralInfoInput) (GeneralInfoProfile, error) {
	profile := GeneralInfoProfile{}

	company, err := p.runCompanyProfile(ctx, in)
	if err != nil {
		return profile, fmt.Errorf("cannot run company profile agent: %w", err)
	}

	profile.Company = company

	website := strings.TrimSpace(in.WebsiteURL)
	if website == "" {
		if v := strings.TrimSpace(company.WebsiteURL.Value); v != "" && company.WebsiteURL.Confidence >= p.cfg.ConfidenceThreshold {
			website = v
		}
	}

	profile.WebsiteURL = website

	// B and C need a resolved website to scope to the vendor.
	if website == "" {
		return profile, nil
	}

	legalName := strings.TrimSpace(in.LegalName)
	if legalName == "" {
		if v := strings.TrimSpace(company.LegalName.Value); v != "" && company.LegalName.Confidence >= p.cfg.ConfidenceThreshold {
			legalName = v
		}
	}

	var (
		compliance    ComplianceDocsResult
		complianceErr error

		domains    DomainsResult
		domainsErr error
	)

	var wg sync.WaitGroup

	wg.Go(func() {
		compliance, complianceErr = p.runComplianceDocs(ctx, in.Name, website, legalName)
	})

	wg.Go(func() {
		domains, domainsErr = p.runDomains(ctx, in.Name, website)
	})

	wg.Wait()

	if complianceErr != nil {
		p.logger.WarnCtx(ctx, "compliance docs agent failed during profiling", log.Error(complianceErr))
	} else {
		profile.Compliance = compliance
	}

	if domainsErr != nil {
		p.logger.WarnCtx(ctx, "domains agent failed during profiling", log.Error(domainsErr))
	} else {
		profile.Domains = domains
	}

	return profile, nil
}

// runCompanyProfile runs Agent A with a per-run browser (when ChromeAddr
// is set). Not pinned to a domain so it can follow a product site to the
// corporate domain where the legal name/address live; SSRF still blocks
// non-public hosts.
func (p *Profiler) runCompanyProfile(
	ctx context.Context,
	in GeneralInfoInput,
) (CompanyProfileResult, error) {
	var browserTools []agent.Tool

	if p.cfg.ChromeAddr != "" {
		webBrowser := browser.NewBrowser(ctx, p.cfg.ChromeAddr)
		defer webBrowser.Close()

		browserTools = browser.NewReadOnlyToolset(webBrowser).Tools()
	}

	companyAgent := buildCompanyProfileAgent(p.cfg, p.logger, browserTools)

	prompt := buildCompanyProfilePrompt(in)

	agentCtx, cancel := context.WithTimeout(ctx, p.cfg.AgentTimeout)
	defer cancel()

	result, err := agent.RunTyped[CompanyProfileResult](
		agentCtx,
		companyAgent,
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

// runComplianceDocs runs Agent B with a per-run browser (when ChromeAddr
// is set). Not pinned to the vendor domain so it can follow links to
// hosted trust portals (Vanta, SafeBase, etc.); SSRF still blocks
// non-public hosts.
func (p *Profiler) runComplianceDocs(
	ctx context.Context,
	name string,
	website string,
	legalName string,
) (ComplianceDocsResult, error) {
	var browserTools []agent.Tool

	if p.cfg.ChromeAddr != "" {
		webBrowser := browser.NewBrowser(ctx, p.cfg.ChromeAddr)
		defer webBrowser.Close()

		browserTools = browser.NewReadOnlyToolset(webBrowser).Tools()
	}

	complianceAgent := buildComplianceDocsAgent(p.cfg, p.logger, browserTools)

	prompt := buildComplianceDocsPrompt(name, website, legalName)

	agentCtx, cancel := context.WithTimeout(ctx, p.cfg.AgentTimeout)
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

// runDomains runs Agent C with a per-run browser (when ChromeAddr is
// set). Not pinned to the vendor domain so it can follow links to the
// vendor's other owned domains; SSRF still blocks non-public hosts.
func (p *Profiler) runDomains(
	ctx context.Context,
	name string,
	website string,
) (DomainsResult, error) {
	var browserTools []agent.Tool

	if p.cfg.ChromeAddr != "" {
		webBrowser := browser.NewBrowser(ctx, p.cfg.ChromeAddr)
		defer webBrowser.Close()

		browserTools = browser.NewReadOnlyToolset(webBrowser).Tools()
	}

	domainsAgent := buildCommonThirdPartyDomainsAgent(p.cfg, p.logger, browserTools)

	prompt := buildCommonThirdPartyDomainsPrompt(name, website)

	agentCtx, cancel := context.WithTimeout(ctx, p.cfg.AgentTimeout)
	defer cancel()

	result, err := agent.RunTyped[DomainsResult](
		agentCtx,
		domainsAgent,
		[]llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: prompt}},
			},
		},
	)
	if err != nil {
		return DomainsResult{}, fmt.Errorf("domains agent run failed: %w", err)
	}

	return result.Output, nil
}
