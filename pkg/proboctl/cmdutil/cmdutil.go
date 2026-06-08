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

package cmdutil

import (
	"fmt"
	"os"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/cmd/iostreams"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/proboctl/pgconn"
	"go.probo.inc/probo/pkg/probodconfig"
	"sigs.k8s.io/yaml"
)

type Factory struct {
	IOStreams *iostreams.IOStreams
	Version   string
	PgDSN     string
	CfgFile   string

	pgClient *pg.Client
}

// PgClient returns a shared pg client, building it on first use. The client
// is memoized because pg.NewClient registers Prometheus collectors, so
// constructing it more than once panics with a duplicate registration.
func (f *Factory) PgClient() (*pg.Client, error) {
	if f.pgClient != nil {
		return f.pgClient, nil
	}

	if f.PgDSN == "" {
		return nil, fmt.Errorf("set --pg-dsn or DATABASE_URL")
	}

	client, err := pgconn.NewPgClientFromDSN(f.PgDSN)
	if err != nil {
		return nil, err
	}

	f.pgClient = client

	return f.pgClient, nil
}

// ProbodConfig loads the shared probod configuration file (--cfg-file).
// It reuses the exact file, struct, and json-tagged (un)marshaling probod
// uses, so proboctl and probod stay consistent.
func (f *Factory) ProbodConfig() (probodconfig.Config, error) {
	if f.CfgFile == "" {
		return probodconfig.Config{}, fmt.Errorf("set --cfg-file to the probod config file")
	}

	data, err := os.ReadFile(f.CfgFile)
	if err != nil {
		return probodconfig.Config{}, fmt.Errorf("cannot read config file %q: %w", f.CfgFile, err)
	}

	var full probodconfig.FullConfig
	if err := yaml.Unmarshal(data, &full); err != nil {
		return probodconfig.Config{}, fmt.Errorf("cannot parse config file %q: %w", f.CfgFile, err)
	}

	return full.Probod, nil
}

// TrackerAgentsConfig builds the enrichment and mapping agent configs
// (LLM clients + Firecrawl key) from the shared probod config for
// in-process agent execution, e.g. synchronous common-pattern
// re-enrichment. The enricher runs the enrichment agent and reuses the
// mapping agent to attribute a vendor first, so both configs are
// returned. It errors when no LLM provider is configured.
//
// This duplicates the small wiring probod does in buildTrackerAgents
// rather than sharing a package, keeping the two executables decoupled.
func (f *Factory) TrackerAgentsConfig() (cookiebanner.TrackerEnrichmentAgentConfig, cookiebanner.TrackerMappingAgentConfig, error) {
	cfg, err := f.ProbodConfig()
	if err != nil {
		return cookiebanner.TrackerEnrichmentAgentConfig{}, cookiebanner.TrackerMappingAgentConfig{}, err
	}

	if cfg.Agents.TrackerMapping.Provider == "" {
		return cookiebanner.TrackerEnrichmentAgentConfig{}, cookiebanner.TrackerMappingAgentConfig{}, fmt.Errorf("no LLM provider configured; set llm.tracker-mapping.provider in %q", f.CfgFile)
	}

	logger := log.NewLogger(
		log.WithName("proboctl"),
		log.WithOutput(f.IOStreams.ErrOut),
	)

	firecrawlAPIKey := cfg.Agents.Tools.FirecrawlAPIKey

	mappingAgentCfg, mappingClient, err := resolveAgentClient(cfg.Agents, "tracker-mapping", cfg.Agents.TrackerMapping, logger)
	if err != nil {
		return cookiebanner.TrackerEnrichmentAgentConfig{}, cookiebanner.TrackerMappingAgentConfig{}, fmt.Errorf("cannot build tracker mapping agent: %w", err)
	}

	mappingCfg := cookiebanner.TrackerMappingAgentConfig{
		LLMClient:       mappingClient,
		Model:           mappingAgentCfg.ModelName,
		FirecrawlAPIKey: firecrawlAPIKey,
		MaxTokens:       mappingAgentCfg.MaxTokens,
		Temperature:     mappingAgentCfg.Temperature,
		Timeout:         time.Duration(cfg.TrackerMappingWorker.AgentTimeout) * time.Second,
		MaxTurns:        cfg.TrackerMappingWorker.AgentMaxTurns,
	}

	enrichmentSlot := cfg.Agents.TrackerEnrichment
	if enrichmentSlot.Provider == "" {
		enrichmentSlot = cfg.Agents.TrackerMapping
	}

	enrichmentAgentCfg, enrichmentClient, err := resolveAgentClient(cfg.Agents, "tracker-enrichment", enrichmentSlot, logger)
	if err != nil {
		return cookiebanner.TrackerEnrichmentAgentConfig{}, cookiebanner.TrackerMappingAgentConfig{}, fmt.Errorf("cannot build tracker enrichment agent: %w", err)
	}

	enrichmentCfg := cookiebanner.TrackerEnrichmentAgentConfig{
		LLMClient:       enrichmentClient,
		Model:           enrichmentAgentCfg.ModelName,
		FirecrawlAPIKey: firecrawlAPIKey,
		MaxTokens:       enrichmentAgentCfg.MaxTokens,
		Temperature:     enrichmentAgentCfg.Temperature,
		Timeout:         time.Duration(cfg.CommonPatternEnrichmentWorker.AgentTimeout) * time.Second,
		MaxTurns:        cfg.CommonPatternEnrichmentWorker.AgentMaxTurns,
	}

	return enrichmentCfg, mappingCfg, nil
}
