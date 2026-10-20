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

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/identityfederation"
)

type (
	Service struct {
		pg                *pg.Client
		encryptionKey     cipher.EncryptionKey
		connectorRegistry *connectorprotocol.Registry
		providerRegistry  *provider.Registry
		federation        *identityfederation.Issuer
		logger            *log.Logger

		fetchWorker *worker.Worker[coredata.AccessReviewCampaignSourceFetchAttempt]
	}

	Option func(*options)

	options struct {
		fetchInterval time.Duration
		federation    *identityfederation.Issuer
		registerer    prometheus.Registerer
	}
)

// WithRegisterer supplies the registry the workers publish worker_tasks_total
// and worker_task_duration_seconds to. Without it they register into the
// default registry, which nothing scrapes, and a worker failing every claim is
// visible only as log volume.
func WithRegisterer(registerer prometheus.Registerer) Option {
	return func(o *options) {
		if registerer != nil {
			o.registerer = registerer
		}
	}
}

func WithFetchInterval(interval time.Duration) Option {
	return func(o *options) {
		o.fetchInterval = interval
	}
}

// WithIdentityFederation supplies the issuer that mints the assertion a
// workload identity connector exchanges for cloud credentials. It is nil in a
// deployment that configures no issuer, and the cloud paths report that as a
// failure rather than fetching nothing: an empty account list would read as
// "nobody has access", which is the wrong answer to give a reviewer.
func WithIdentityFederation(issuer *identityfederation.Issuer) Option {
	return func(o *options) {
		o.federation = issuer
	}
}

func NewService(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connectorprotocol.Registry,
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
		federation:        o.federation,
		logger:            logger,
	}

	var fetchWorkerOpts []worker.Option
	if o.fetchInterval > 0 {
		fetchWorkerOpts = append(fetchWorkerOpts, worker.WithInterval(o.fetchInterval))
	} else {
		fetchWorkerOpts = append(fetchWorkerOpts, worker.WithInterval(30*time.Second))
	}

	fetchWorkerOpts = append(fetchWorkerOpts, worker.WithMaxConcurrency(20))

	if o.registerer != nil {
		fetchWorkerOpts = append(fetchWorkerOpts, worker.WithRegisterer(o.registerer))
	}

	s.fetchWorker = NewSourceFetchWorker(
		s,
		pgClient,
		logger,
		fetchWorkerOpts...,
	)

	return s
}

func (s *Service) Run(ctx context.Context) error {
	return s.fetchWorker.Run(ctx)
}
