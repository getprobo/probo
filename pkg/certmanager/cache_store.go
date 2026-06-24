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
		func(ctx context.Context, conn pg.Querier) error {
			domains := coredata.CustomDomains{}
			if err := domains.LoadActiveCertificates(ctx, conn, coredata.NewNoScope()); err != nil {
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

func (w *CacheStore) warmDomain(ctx context.Context, conn pg.Querier, domain *coredata.CustomDomain) error {
	var loadedDomain coredata.CustomDomain
	if err := loadedDomain.LoadByID(ctx, conn, coredata.NewNoScope(), domain.ID); err != nil {
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
