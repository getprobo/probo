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

package vetting

import (
	"context"
	_ "embed"
	"fmt"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/browser"
	"go.probo.inc/probo/pkg/llm"
)

//go:embed extraction_prompt.txt
var extractionPrompt string

type (
	Config struct {
		Client     *llm.Client
		Model      string
		ChromeAddr string
	}

	Assessor struct {
		cfg Config
	}

	VendorInfo struct {
		Name                          string   `json:"name"`
		Description                   string   `json:"description"`
		Category                      string   `json:"category"`
		HeadquarterAddress            string   `json:"headquarter_address"`
		LegalName                     string   `json:"legal_name"`
		PrivacyPolicyURL              string   `json:"privacy_policy_url"`
		ServiceLevelAgreementURL      string   `json:"service_level_agreement_url"`
		DataProcessingAgreementURL    string   `json:"data_processing_agreement_url"`
		BusinessAssociateAgreementURL string   `json:"business_associate_agreement_url"`
		SubprocessorsListURL          string   `json:"subprocessors_list_url"`
		SecurityPageURL               string   `json:"security_page_url"`
		TrustPageURL                  string   `json:"trust_page_url"`
		TermsOfServiceURL             string   `json:"terms_of_service_url"`
		StatusPageURL                 string   `json:"status_page_url"`
		BugBountyURL                  string   `json:"bug_bounty_url"`
		DataLocations                 []string `json:"data_locations"`
		Certifications                []string `json:"certifications"`
	}

	Result struct {
		Document string
		Info     VendorInfo
	}
)

func NewAssessor(cfg Config) *Assessor {
	return &Assessor{cfg: cfg}
}

func (a *Assessor) Assess(ctx context.Context, websiteURL string, reporter agent.ProgressReporter) (*Result, error) {
	b := browser.NewBrowser(ctx, a.cfg.ChromeAddr)
	defer b.Close()

	orchestrator, err := newOrchestratorAgent(a.cfg.Client, a.cfg.Model, b, reporter)
	if err != nil {
		return nil, fmt.Errorf("cannot create orchestrator agent: %w", err)
	}

	result, err := orchestrator.Run(
		ctx,
		[]llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: websiteURL}},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot assess vendor: %w", err)
	}

	document := result.FinalMessage().Text()

	if reporter != nil {
		reporter(
			ctx,
			agent.ProgressEvent{
				Type:    agent.ProgressEventStepStarted,
				Step:    "extract_vendor_info",
				Message: "Extracting vendor information from assessment",
			},
		)
	}

	info, err := a.extractVendorInfo(ctx, document)
	if err != nil {
		if reporter != nil {
			reporter(
				ctx,
				agent.ProgressEvent{
					Type:    agent.ProgressEventStepFailed,
					Step:    "extract_vendor_info",
					Message: fmt.Sprintf("Failed to extract vendor information: %s", err),
				},
			)
		}
		return nil, fmt.Errorf("cannot extract vendor info: %w", err)
	}

	if reporter != nil {
		reporter(
			ctx,
			agent.ProgressEvent{
				Type:    agent.ProgressEventStepCompleted,
				Step:    "extract_vendor_info",
				Message: "Extracting vendor information from assessment",
			},
		)
	}

	return &Result{
		Document: document,
		Info:     *info,
	}, nil
}

func (a *Assessor) extractVendorInfo(ctx context.Context, document string) (*VendorInfo, error) {
	extractor := agent.New(
		"vendor_info_extractor",
		a.cfg.Client,
		agent.WithInstructions(extractionPrompt),
		agent.WithModel(a.cfg.Model),
	)

	typedResult, err := agent.RunTyped[VendorInfo](
		ctx,
		extractor,
		[]llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: document}},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot extract vendor info: %w", err)
	}

	return &typedResult.Output, nil
}
