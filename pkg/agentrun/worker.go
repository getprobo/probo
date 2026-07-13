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

package agentrun

import (
	"context"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

type (
	Worker struct {
		handler   *handler
		kitWorker *worker.Worker[coredata.AgentRun]
	}

	WorkerOption func(*workerConfig)

	workerConfig struct {
		interval       time.Duration
		maxConcurrency int
	}
)

func WithWorkerInterval(d time.Duration) WorkerOption {
	return func(c *workerConfig) {
		if d > 0 {
			c.interval = d
		}
	}
}

func WithWorkerMaxConcurrency(n int) WorkerOption {
	return func(c *workerConfig) {
		if n > 0 {
			c.maxConcurrency = n
		}
	}
}

func NewWorker(
	pgClient *pg.Client,
	store *coredata.PGCheckpointer,
	registry *Registry,
	logger *log.Logger,
	opts ...WorkerOption,
) *Worker {
	cfg := workerConfig{
		interval:       10 * time.Second,
		maxConcurrency: 5,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	h := &handler{
		pg:         pgClient,
		store:      store,
		registry:   registry,
		logger:     logger,
		shutdownCh: make(chan struct{}),
	}

	w := worker.New(
		"agent-run-worker",
		h,
		logger,
		worker.WithInterval(cfg.interval),
		worker.WithMaxConcurrency(cfg.maxConcurrency),
	)

	return &Worker{handler: h, kitWorker: w}
}

// Run starts the worker loop. It blocks until ctx is cancelled, then
// closes the shutdown broadcast channel so in-flight Process calls can
// checkpoint and exit, and waits for all of them to drain before
// returning.
//
// signalShutdown is registered without a stop hook because it is
// idempotent (sync.Once) and we want it to fire on every ctx
// cancellation, even one that races with kitWorker.Run returning.
func (w *Worker) Run(ctx context.Context) error {
	context.AfterFunc(ctx, w.handler.signalShutdown)
	return w.kitWorker.Run(ctx)
}
