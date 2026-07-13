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

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
)

type Runner struct {
	pg                *pg.Client
	encryptionKey     cipher.EncryptionKey
	connectorRegistry *connector.ConnectorRegistry
	providerRegistry  *provider.Registry
	synthesizer       Synthesizer
	logger            *log.Logger
}

func NewRunner(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connector.ConnectorRegistry,
	providerRegistry *provider.Registry,
	synthesizer Synthesizer,
	logger *log.Logger,
) *Runner {
	if synthesizer == nil {
		synthesizer = DeterministicSynthesizer{}
	}

	return &Runner{
		pg:                pgClient,
		encryptionKey:     encryptionKey,
		connectorRegistry: connectorRegistry,
		providerRegistry:  providerRegistry,
		synthesizer:       synthesizer,
		logger:            logger,
	}
}

func (r *Runner) Run(ctx context.Context, run *coredata.AgentRun) (*RunResult, error) {
	var input RunInput
	if err := json.Unmarshal(run.InputMessages, &input); err != nil {
		return nil, fmt.Errorf("cannot parse discovery run input: %w", err)
	}

	scope := coredata.NewScope(run.OrganizationID.TenantID())

	httpClient, connector, err := connectorHTTPClient(
		ctx,
		r.pg,
		scope,
		r.encryptionKey,
		r.connectorRegistry,
		r.providerRegistry,
		input.ConnectorID,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build github connector HTTP client: %w", err)
	}

	githubOrg, err := githubOrganizationFromConnector(connector)
	if err != nil {
		return nil, err
	}

	sheet, err := newOrgScanner(httpClient, githubOrg, r.logger).scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot scan github organization: %w", err)
	}

	thirdParty, err := EnsureThirdParty(ctx, r.pg, scope, run.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot ensure github third party: %w", err)
	}

	var existing []ExistingMeasure

	err = r.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			loaded, loadErr := loadGitHubLinkedMeasures(ctx, conn, scope, thirdParty.ID)
			if loadErr != nil {
				return loadErr
			}

			existing = loaded

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load github-linked measures: %w", err)
	}

	plan, err := r.synthesizer.Synthesize(ctx, sheet, existing)
	if err != nil {
		r.logger.WarnCtx(ctx, "llm synthesis failed, falling back to deterministic materialization", log.Error(err))

		plan, err = MaterializeFromFacts(sheet, existing)
		if err != nil {
			return nil, fmt.Errorf("cannot materialize measure plan: %w", err)
		}
	}

	stats, err := applyMeasurePlan(
		ctx,
		r.pg,
		scope,
		persistInput{
			plan:           plan,
			factSheet:      sheet,
			thirdPartyID:   thirdParty.ID,
			organizationID: run.OrganizationID,
			agentRunID:     run.ID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot apply measure plan: %w", err)
	}

	return &RunResult{
		Integration:      "github",
		ThirdPartyID:     thirdParty.ID,
		GitHubOrg:        sheet.GitHubOrg,
		CompletedAt:      time.Now().UTC().Format(time.RFC3339),
		Limitations:      sheet.Limitations,
		ReposScanned:     sheet.ReposScanned,
		MeasuresUpserted: stats.upserted,
		Summary:          stats.summary,
	}, nil
}

func githubOrganizationFromConnector(connector *coredata.Connector) (string, error) {
	settings, err := coredata.ConnectorSettings[coredata.GitHubConnectorSettings](connector)
	if err != nil {
		return "", fmt.Errorf("cannot read github connector settings: %w", err)
	}

	if settings.Organization == "" {
		return "", fmt.Errorf("github connector is missing organization settings")
	}

	return settings.Organization, nil
}
