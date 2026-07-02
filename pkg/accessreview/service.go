// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
