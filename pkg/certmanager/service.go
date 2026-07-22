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

package certmanager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"golang.org/x/sync/errgroup"
)

type (
	// Service owns the TLS certificate lifecycle for arbitrary hostnames. It is
	// a generic core service and knows nothing about the resources a
	// certificate protects; callers reference a certificate by its ID.
	Service struct {
		pg                   *pg.Client
		acmeService          *ACMEService
		encryptionKey        cipher.EncryptionKey
		logger               *log.Logger
		beginChallengeWorker *worker.Worker[coredata.Certificate]
		pollOrderWorker      *worker.Worker[coredata.Certificate]
		renewWorker          *worker.Worker[coredata.Certificate]
	}

	// Config holds the SSL provisioning parameters for the service workers.
	Config struct {
		CnameTarget       string
		CAAIssuerDomain   string
		ResolverAddr      string
		ManagedBaseDomain string
		RenewalInterval   time.Duration
		ProvisionInterval time.Duration
	}
)

func NewService(
	pgClient *pg.Client,
	acmeService *ACMEService,
	encryptionKey cipher.EncryptionKey,
	cfg Config,
	logger *log.Logger,
) *Service {
	provisionInterval := cfg.ProvisionInterval
	if provisionInterval <= 0 {
		provisionInterval = 30 * time.Second
	}

	renewalInterval := cfg.RenewalInterval
	if renewalInterval <= 0 {
		renewalInterval = time.Hour
	}

	return &Service{
		pg:            pgClient,
		acmeService:   acmeService,
		encryptionKey: encryptionKey,
		logger:        logger,
		beginChallengeWorker: NewBeginChallengeWorker(
			pgClient,
			acmeService,
			cfg.CnameTarget,
			cfg.CAAIssuerDomain,
			cfg.ResolverAddr,
			cfg.ManagedBaseDomain,
			logger.Named("begin-challenge-worker"),
			worker.WithInterval(provisionInterval),
		),
		pollOrderWorker: NewPollOrderWorker(
			pgClient,
			acmeService,
			encryptionKey,
			logger.Named("poll-order-worker"),
			worker.WithInterval(provisionInterval),
		),
		renewWorker: NewRenewWorker(
			pgClient,
			encryptionKey,
			logger.Named("renew-worker"),
			worker.WithInterval(renewalInterval),
		),
	}
}

func (s *Service) Run(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)

	g.Go(
		func() error {
			return s.beginChallengeWorker.Run(gctx)
		},
	)

	g.Go(
		func() error {
			return s.pollOrderWorker.Run(gctx)
		},
	)

	g.Go(
		func() error {
			return s.renewWorker.Run(gctx)
		},
	)

	return g.Wait()
}

// EnsureCertificate returns the certificate for the given hostname, creating a
// pending one within the given transaction when it does not exist yet. The
// certificate lifecycle is then driven asynchronously by the provision worker.
func (s *Service) EnsureCertificate(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	hostname string,
) (*coredata.Certificate, error) {
	certificate := &coredata.Certificate{}

	err := certificate.LoadByHostname(ctx, tx, scope, hostname)
	if err == nil {
		return certificate, nil
	}

	if !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load certificate: %w", err)
	}

	certificate = coredata.NewCertificate(scope.GetTenantID(), hostname)
	if err := certificate.Insert(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot insert certificate: %w", err)
	}

	return certificate, nil
}

// Get returns a certificate by ID.
func (s *Service) Get(
	ctx context.Context,
	scope coredata.Scoper,
	certificateID gid.GID,
) (*coredata.Certificate, error) {
	certificate := &coredata.Certificate{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return certificate.LoadByID(ctx, conn, scope, certificateID)
		},
	)
	if err != nil {
		return nil, err
	}

	return certificate, nil
}

// GetByHostname returns the certificate matching the given hostname. It is
// unscoped because it powers public host resolution across all tenants.
func (s *Service) GetByHostname(
	ctx context.Context,
	hostname string,
) (*coredata.Certificate, error) {
	certificate := &coredata.Certificate{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return certificate.LoadByHostname(ctx, conn, coredata.NewNoScope(), hostname)
		},
	)
	if err != nil {
		return nil, err
	}

	return certificate, nil
}

// Delete removes a certificate within the given transaction.
func (s *Service) Delete(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	certificateID gid.GID,
) error {
	certificate := &coredata.Certificate{}
	if err := certificate.LoadByID(ctx, tx, scope, certificateID); err != nil {
		return fmt.Errorf("cannot load certificate: %w", err)
	}

	if err := certificate.Delete(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot delete certificate: %w", err)
	}

	return nil
}
