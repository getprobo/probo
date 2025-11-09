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
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
)

type (
	CacheStore struct {
		pg            *pg.Client
		encryptionKey cipher.EncryptionKey
		logger        *log.Logger
	}
)

func NewCacheStore(
	pg *pg.Client,
	encryptionKey cipher.EncryptionKey,
	logger *log.Logger,
) *CacheStore {
	return &CacheStore{
		pg:            pg,
		encryptionKey: encryptionKey,
		logger:        logger.Named("certmanager.cache-store"),
	}
}

func (w *CacheStore) WarmCache(ctx context.Context) error {
	w.logger.InfoCtx(ctx, "warming certificate cache")
	startTime := time.Now()

	err := w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			domains := coredata.CustomDomains{}
			if err := domains.LoadActiveCertificates(ctx, conn, coredata.NewNoScope(), w.encryptionKey); err != nil {
				return fmt.Errorf("cannot load active certificates: %w", err)
			}

			if len(domains) == 0 {
				w.logger.InfoCtx(ctx, "no active certificates to warm")
				return nil
			}

			w.logger.InfoCtx(ctx, "found active certificates to cache", log.Int("count", len(domains)))

			successCount := 0
			for _, domain := range domains {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				if err := w.warmDomain(ctx, conn, domain); err != nil {
					w.logger.ErrorCtx(ctx, "cannot warm certificate cache for domain", log.String("domain", domain.Domain), log.Error(err))
				} else {
					successCount++
				}
			}

			w.logger.InfoCtx(ctx, "successfully warmed cache", log.Int("success_count", successCount), log.Int("total_count", len(domains)))
			return nil
		},
	)

	if err != nil {
		return fmt.Errorf("cannot warm certificate cache: %w", err)
	}

	w.logger.InfoCtx(ctx, "certificate cache warming completed", log.Duration("duration", time.Since(startTime)))

	return nil
}

func (w *CacheStore) warmDomain(ctx context.Context, conn pg.Conn, domain *coredata.CustomDomain) error {
	var loadedDomain coredata.CustomDomain
	if err := loadedDomain.LoadByID(ctx, conn, coredata.NewNoScope(), w.encryptionKey, domain.ID); err != nil {
		return fmt.Errorf("cannot load domain with decrypted values: %w", err)
	}

	if err := loadedDomain.ParseCertificate(w.encryptionKey); err != nil {
		return fmt.Errorf("cannot parse certificate: %w", err)
	}

	if len(loadedDomain.SSLCertificatePEM) == 0 {
		return fmt.Errorf("domain has no certificate PEM")
	}

	privateKeyPEM, err := loadedDomain.DecryptPrivateKey(w.encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt private key: %w", err)
	}

	if len(privateKeyPEM) == 0 {
		return fmt.Errorf("domain has no private key PEM")
	}

	if loadedDomain.SSLExpiresAt == nil {
		return fmt.Errorf("domain certificate has no expiry date")
	}

	if time.Now().After(*loadedDomain.SSLExpiresAt) {
		return fmt.Errorf("certificate has expired")
	}

	cache := &coredata.CachedCertificate{
		Domain:           loadedDomain.Domain,
		CertificatePEM:   string(loadedDomain.SSLCertificatePEM),
		PrivateKeyPEM:    string(privateKeyPEM),
		CertificateChain: loadedDomain.SSLCertificateChain,
		ExpiresAt:        *loadedDomain.SSLExpiresAt,
		CachedAt:         time.Now(),
		CustomDomainID:   loadedDomain.ID,
	}

	if err := cache.Upsert(ctx, conn); err != nil {
		return fmt.Errorf("cannot upsert cache entry: %w", err)
	}

	return nil
}

func (w *CacheStore) RefreshCache(ctx context.Context) error {
	w.logger.InfoCtx(ctx, "refreshing certificate cache")

	return w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			cachedCertificates := coredata.CachedCertificates{}
			if err := cachedCertificates.CleanExpired(ctx, conn); err != nil {
				w.logger.ErrorCtx(ctx, "cannot clean expired cache", log.Error(err))
			}

			return w.WarmCache(ctx)
		},
	)
}

func (w *CacheStore) WarmSingleDomain(ctx context.Context, domainName string) error {
	return w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var domain coredata.CustomDomain
			if err := domain.LoadByDomain(ctx, conn, coredata.NewNoScope(), w.encryptionKey, domainName); err != nil {
				return fmt.Errorf("cannot load domain: %w", err)
			}

			return w.warmDomain(ctx, conn, &domain)
		},
	)
}
