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

package probo

import (
	"context"
	"errors"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
)

type (
	AgentRunSupervisor struct {
		handler *agentRunHandler
		worker  *worker.Worker[coredata.AgentRun]
	}

	AgentRunSupervisorOption func(*agentRunSupervisorConfig)

	agentRunSupervisorConfig struct {
		interval       time.Duration
		leaseDuration  time.Duration
		maxConcurrency int
	}
)

var (
	ErrAgentRunHeartbeatFailed = errors.New("agent run heartbeat failed")
	ErrAgentRunLeaseLost       = errors.New("agent run lease lost")
)

func WithAgentRunSupervisorInterval(d time.Duration) AgentRunSupervisorOption {
	return func(c *agentRunSupervisorConfig) {
		if d > 0 {
			c.interval = d
		}
	}
}

func WithAgentRunSupervisorLeaseDuration(d time.Duration) AgentRunSupervisorOption {
	return func(c *agentRunSupervisorConfig) {
		if d > 0 {
			c.leaseDuration = d
		}
	}
}

func WithAgentRunSupervisorMaxConcurrency(n int) AgentRunSupervisorOption {
	return func(c *agentRunSupervisorConfig) {
		if n > 0 {
			c.maxConcurrency = n
		}
	}
}

func NewAgentRunSupervisor(
	pgClient *pg.Client,
	store *coredata.PGCheckpointer,
	registry agent.AgentRegistry,
	logger *log.Logger,
	opts ...AgentRunSupervisorOption,
) *AgentRunSupervisor {
	cfg := agentRunSupervisorConfig{
		interval:       10 * time.Second,
		leaseDuration:  5 * time.Minute,
		maxConcurrency: 5,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	h := &agentRunHandler{
		pg:            pgClient,
		store:         store,
		registry:      registry,
		logger:        logger,
		leaseDuration: cfg.leaseDuration,
		workerID:      uuid.MustNewV4().String(),
		shutdownCh:    make(chan struct{}),
	}

	w := worker.New(
		"agent-run-supervisor",
		h,
		logger,
		worker.WithInterval(cfg.interval),
		worker.WithMaxConcurrency(cfg.maxConcurrency),
	)

	return &AgentRunSupervisor{handler: h, worker: w}
}

// Run starts the supervisor loop. It blocks until ctx is cancelled, then
// closes the shutdown broadcast channel so in-flight Process calls can
// checkpoint and exit, and waits for all of them to drain before
// returning.
func (s *AgentRunSupervisor) Run(ctx context.Context) error {
	stop := context.AfterFunc(ctx, s.handler.signalShutdown)
	defer stop()
	return s.worker.Run(ctx)
}
