// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
