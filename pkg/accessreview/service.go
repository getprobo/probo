// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package accessreview

import (
	"context"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"golang.org/x/sync/errgroup"
)

type (
	Service struct {
		pg                *pg.Client
		encryptionKey     cipher.EncryptionKey
		connectorRegistry *connector.ConnectorRegistry
		providerRegistry  *provider.Registry
		logger            *log.Logger

		fetchWorker      *worker.Worker[coredata.AccessReviewCampaignSourceFetchAttempt]
		sourceNameWorker *worker.Worker[coredata.AccessReviewSource]
	}

	Option func(*options)

	options struct {
		fetchInterval time.Duration
	}
)

func WithFetchInterval(interval time.Duration) Option {
	return func(o *options) {
		o.fetchInterval = interval
	}
}

func NewService(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connector.ConnectorRegistry,
	providerRegistry *provider.Registry,
	logger *log.Logger,
	opts ...Option,
) *Service {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	s := &Service{
		pg:                pgClient,
		encryptionKey:     encryptionKey,
		connectorRegistry: connectorRegistry,
		providerRegistry:  providerRegistry,
		logger:            logger,
	}

	var fetchWorkerOpts []worker.Option
	if o.fetchInterval > 0 {
		fetchWorkerOpts = append(fetchWorkerOpts, worker.WithInterval(o.fetchInterval))
	} else {
		fetchWorkerOpts = append(fetchWorkerOpts, worker.WithInterval(30*time.Second))
	}

	fetchWorkerOpts = append(fetchWorkerOpts, worker.WithMaxConcurrency(20))

	s.fetchWorker = NewSourceFetchWorker(
		s,
		pgClient,
		logger,
		fetchWorkerOpts...,
	)
	s.sourceNameWorker = NewSourceNameWorker(
		pgClient,
		encryptionKey,
		connectorRegistry,
		providerRegistry,
		logger.Named("source-name"),
	)

	return s
}

func (s *Service) Run(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error { return s.fetchWorker.Run(gCtx) })
	g.Go(func() error { return s.sourceNameWorker.Run(gCtx) })

	return g.Wait()
}
