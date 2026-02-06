// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

type (
	Renewer struct {
		pg            *pg.Client
		acmeService   *ACMEService
		encryptionKey cipher.EncryptionKey
		interval      time.Duration
		logger        *log.Logger
	}
)

func NewRenewer(
	pg *pg.Client,
	acmeService *ACMEService,
	encryptionKey cipher.EncryptionKey,
	interval time.Duration,
	logger *log.Logger,
) *Renewer {
	return &Renewer{
		pg:            pg,
		acmeService:   acmeService,
		encryptionKey: encryptionKey,
		interval:      interval,
		logger:        logger.Named("certmanager.renewer"),
	}
}

func (r *Renewer) Run(ctx context.Context) error {
	r.logger.InfoCtx(ctx, "certificate renewer starting")

	if err := r.checkAndRenew(ctx); err != nil {
		r.logger.ErrorCtx(ctx, "cannot perform initial renewal check", log.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			r.logger.InfoCtx(ctx, "certificate renewer shutting down")
			return ctx.Err()
		case <-time.After(r.interval):
			if err := r.checkAndRenew(ctx); err != nil {
				r.logger.ErrorCtx(ctx, "cannot perform renewal check", log.Error(err))
			}
		}
	}
}

func (r *Renewer) checkAndRenew(ctx context.Context) error {
	return r.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			var caches coredata.CachedCertificates
			cacheCount, err := caches.CountAll(ctx, tx)
			if err != nil {
				r.logger.ErrorCtx(ctx, "cannot count certificate cache", log.Error(err))
			} else if cacheCount == 0 {
				r.logger.InfoCtx(ctx, "certificate cache is empty, rebuilding from custom_domains")

				warmer := NewCacheStore(r.pg, r.encryptionKey, r.logger)
				if err := warmer.WarmCache(ctx); err != nil {
					r.logger.ErrorCtx(ctx, "cannot rebuild certificate cache", log.Error(err))
				} else {
					r.logger.InfoCtx(ctx, "certificate cache rebuilt successfully")
				}
			}

			if err := caches.CleanExpired(ctx, tx); err != nil {
				r.logger.ErrorCtx(ctx, "cannot clean certificate cache", log.Error(err))
			}

			domains := coredata.CustomDomains{}
			scope := coredata.NewNoScope()
			if err := domains.ListDomainsForRenewal(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot list domains for renewal: %w", err)
			}

			if len(domains) == 0 {
				return nil
			}

			r.logger.InfoCtx(ctx, "found domains needing renewal", log.Int("count", len(domains)))

			for _, domain := range domains {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				r.logger.InfoCtx(ctx, "renewing certificate for domain", log.String("domain", domain.Domain))
				if err := r.renewDomain(ctx, tx, domain.ID); err != nil {
					r.logger.ErrorCtx(ctx, "cannot renew certificate", log.String("domain", domain.Domain), log.Error(err))
				} else {
					r.logger.InfoCtx(ctx, "successfully renewed certificate", log.String("domain", domain.Domain))
				}
			}

			return nil
		},
	)
}

func (r *Renewer) renewDomain(ctx context.Context, tx pg.Conn, domainID gid.GID) error {
	domain := &coredata.CustomDomain{}
	if err := domain.LoadByIDForUpdateSkipLocked(ctx, tx, coredata.NewNoScope(), r.encryptionKey, domainID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

		return fmt.Errorf("cannot lock domain for renewal: %w", err)
	}

	if domain.SSLStatus != coredata.CustomDomainSSLStatusActive {
		r.logger.InfoCtx(
			ctx,
			"domain status changed, skipping renewal",
			log.String("domain", domain.Domain),
		)

		return nil
	}

	domain.SSLStatus = coredata.CustomDomainSSLStatusRenewing
	if err := domain.Update(ctx, tx, coredata.NewNoScope(), r.encryptionKey); err != nil {
		return fmt.Errorf("cannot update domain status: %w", err)
	}

	return nil
}
