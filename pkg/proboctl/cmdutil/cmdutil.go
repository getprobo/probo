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

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.opentelemetry.io/otel/trace/noop"
	"go.probo.inc/probo/pkg/agentsbuild"
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
}

func (f *Factory) PgClient() (*pg.Client, error) {
	if f.PgDSN == "" {
		return nil, fmt.Errorf("set --pg-dsn or DATABASE_URL")
	}

	return pgconn.NewPgClientFromDSN(f.PgDSN)
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

// TrackerAgentsConfig builds the tracker-agents config (LLM client +
// Firecrawl key) from the shared probod config for in-process agent
// execution, e.g. synchronous common-pattern re-enrichment. It errors
// when no LLM provider is configured.
func (f *Factory) TrackerAgentsConfig() (cookiebanner.TrackerAgentsConfig, error) {
	cfg, err := f.ProbodConfig()
	if err != nil {
		return cookiebanner.TrackerAgentsConfig{}, err
	}

	logger := log.NewLogger(
		log.WithName("proboctl"),
		log.WithOutput(f.IOStreams.ErrOut),
	)

	trackerCfg, _, err := agentsbuild.BuildTrackerAgentsConfig(
		cfg,
		logger,
		noop.NewTracerProvider(),
		prometheus.NewRegistry(),
	)
	if err != nil {
		return cookiebanner.TrackerAgentsConfig{}, fmt.Errorf("cannot build tracker agents config: %w", err)
	}

	if trackerCfg.LLMClient == nil {
		return cookiebanner.TrackerAgentsConfig{}, fmt.Errorf("no LLM provider configured; set llm.tracker-mapping.provider in %q", f.CfgFile)
	}

	return trackerCfg, nil
}
