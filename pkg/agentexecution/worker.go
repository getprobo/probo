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

package agentexecution

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
)

type (
	Worker struct {
		handler   *handler
		kitWorker *worker.Worker[coredata.AgentExecution]
	}

	WorkerOption func(*workerConfig)

	workerConfig struct {
		interval          time.Duration
		maxConcurrency    int
		heartbeatInterval time.Duration
		staleAfter        time.Duration
		retryBase         time.Duration
		retryMax          time.Duration
		preparer          ExecutionPreparer
		registerer        prometheus.Registerer
		tracerProvider    trace.TracerProvider
	}

	// ExecutionPreparer is the provider-neutral extension point for attaching
	// trusted run context and decorating an execution-specific agent registry.
	ExecutionPreparer interface {
		Prepare(
			ctx context.Context,
			execution *coredata.AgentExecution,
			registry agent.AgentRegistry,
			input *coredata.AgentInput,
		) (context.Context, agent.AgentRegistry, error)
	}

	defaultExecutionPreparer struct{}
)

func (defaultExecutionPreparer) Prepare(
	ctx context.Context,
	_ *coredata.AgentExecution,
	registry agent.AgentRegistry,
	_ *coredata.AgentInput,
) (context.Context, agent.AgentRegistry, error) {
	return ctx, registry, nil
}

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

func WithWorkerHeartbeatInterval(d time.Duration) WorkerOption {
	return func(c *workerConfig) {
		if d > 0 {
			c.heartbeatInterval = d
		}
	}
}

func WithWorkerStaleAfter(d time.Duration) WorkerOption {
	return func(c *workerConfig) {
		if d > 0 {
			c.staleAfter = d
		}
	}
}

func WithWorkerRetryBackoff(base time.Duration, maxDelay time.Duration) WorkerOption {
	return func(c *workerConfig) {
		if base > 0 {
			c.retryBase = base
		}

		if maxDelay > 0 {
			c.retryMax = maxDelay
		}
	}
}

func WithExecutionPreparer(preparer ExecutionPreparer) WorkerOption {
	return func(c *workerConfig) {
		if preparer != nil {
			c.preparer = preparer
		}
	}
}

func WithWorkerRegisterer(registerer prometheus.Registerer) WorkerOption {
	return func(c *workerConfig) {
		if registerer != nil {
			c.registerer = registerer
		}
	}
}

func WithWorkerTracerProvider(provider trace.TracerProvider) WorkerOption {
	return func(c *workerConfig) {
		if provider != nil {
			c.tracerProvider = provider
		}
	}
}

func NewWorker(
	pgClient *pg.Client,
	registry agent.AgentRegistry,
	logger *log.Logger,
	opts ...WorkerOption,
) *Worker {
	cfg := workerConfig{
		interval:          10 * time.Second,
		maxConcurrency:    5,
		heartbeatInterval: 30 * time.Second,
		staleAfter:        2 * time.Minute,
		retryBase:         time.Second,
		retryMax:          5 * time.Minute,
		preparer:          defaultExecutionPreparer{},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.heartbeatInterval >= cfg.staleAfter {
		cfg.heartbeatInterval = max(cfg.staleAfter/2, time.Millisecond)
	}

	if cfg.retryMax < cfg.retryBase {
		cfg.retryMax = cfg.retryBase
	}

	h := &handler{
		pg:                pgClient,
		registry:          registry,
		preparer:          cfg.preparer,
		logger:            logger,
		heartbeatInterval: cfg.heartbeatInterval,
		staleAfter:        cfg.staleAfter,
		retryBase:         cfg.retryBase,
		retryMax:          cfg.retryMax,
		shutdownCh:        make(chan struct{}),
		now:               time.Now,
	}

	kitOpts := []worker.Option{
		worker.WithInterval(cfg.interval),
		worker.WithMaxConcurrency(cfg.maxConcurrency),
	}
	if cfg.registerer != nil {
		kitOpts = append(kitOpts, worker.WithRegisterer(cfg.registerer))
	}

	if cfg.tracerProvider != nil {
		kitOpts = append(kitOpts, worker.WithTracerProvider(cfg.tracerProvider))
	}

	w := worker.New(
		"agent-execution-worker",
		h,
		logger,
		kitOpts...,
	)

	return &Worker{handler: h, kitWorker: w}
}

func (w *Worker) Run(ctx context.Context) error {
	context.AfterFunc(ctx, w.handler.signalShutdown)

	return w.kitWorker.Run(ctx)
}
