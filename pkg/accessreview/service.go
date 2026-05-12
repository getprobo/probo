// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"golang.org/x/sync/errgroup"
)

type (
	Service struct {
		pg                *pg.Client
		encryptionKey     cipher.EncryptionKey
		connectorRegistry *connector.ConnectorRegistry
		logger            *log.Logger

		fetchWorker      *worker.Worker[coredata.AccessReviewCampaignSourceFetch]
		sourceNameWorker *worker.Worker[coredata.AccessSource]
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
		logger.Named("source-name"),
	)

	return s
}

// Sources returns a tenant-scoped AccessSourceService.
func (s *Service) Sources(scope coredata.Scoper) *AccessSourceService {
	return &AccessSourceService{
		pg:                s.pg,
		scope:             scope,
		encryptionKey:     s.encryptionKey,
		connectorRegistry: s.connectorRegistry,
	}
}

// Campaigns returns a tenant-scoped CampaignService.
func (s *Service) Campaigns(scope coredata.Scoper) *CampaignService {
	return NewCampaignService(s.pg, scope)
}

// Entries returns a tenant-scoped AccessEntryService.
func (s *Service) Entries(scope coredata.Scoper) *AccessEntryService {
	return &AccessEntryService{pg: s.pg, scope: scope}
}

// Engine returns a tenant-scoped ReviewEngine.
func (s *Service) Engine(scope coredata.Scoper) *ReviewEngine {
	return NewReviewEngine(
		s.pg,
		scope,
		s.encryptionKey,
		s.connectorRegistry,
		s.logger.Named("review_engine"),
	)
}

// ResolveEntryOrganizationID resolves the organization ID for an access entry.
// This is unscoped because it is used by resolvers before authorization to
// find the organization from an entry ID.
func (s *Service) ResolveEntryOrganizationID(ctx context.Context, entryID gid.GID) (gid.GID, error) {
	var organizationID gid.GID

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error
			entry := &coredata.AccessEntry{}
			organizationID, err = entry.LoadOrganizationID(ctx, conn, entryID)
			if err != nil {
				return fmt.Errorf("cannot load organization id: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return gid.GID{}, fmt.Errorf("cannot resolve organization id: %w", err)
	}

	return organizationID, nil
}

func (s *Service) Run(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error { return s.fetchWorker.Run(gCtx) })
	g.Go(func() error { return s.sourceNameWorker.Run(gCtx) })

	return g.Wait()
}
