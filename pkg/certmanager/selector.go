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
	"crypto/tls"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
)

type (
	Selector struct {
		pg            *pg.Client
		cache         sync.Map
		encryptionKey cipher.EncryptionKey
	}

	// NoSNIError is returned when a TLS client doesn't provide SNI (Server Name Indication)
	NoSNIError struct{}
)

func (e *NoSNIError) Error() string {
	return "no SNI provided"
}

func NewSelector(
	pg *pg.Client,
	encryptionKey cipher.EncryptionKey,
) *Selector {
	return &Selector{
		pg:            pg,
		encryptionKey: encryptionKey,
	}
}

func (s *Selector) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	domain := hello.ServerName

	// Empty domain, return error
	if domain == "" {
		return nil, &NoSNIError{}
	}

	if cached, ok := s.cache.Load(domain); ok {
		if cert, ok := cached.(*tls.Certificate); ok {
			return cert, nil
		}
	}

	cert, err := s.loadFromDatabase(domain)
	if err != nil {
		return nil, fmt.Errorf("cannot load certificate from database: %w", err)
	}

	s.cache.Store(domain, cert)

	return cert, nil
}

func (s *Selector) loadFromDatabase(domain string) (*tls.Certificate, error) {
	ctx := context.Background()

	var cert *tls.Certificate

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := requireRoutableDomain(ctx, conn, domain); err != nil {
				return err
			}

			var cache coredata.CachedCertificate
			if err := cache.LoadByDomain(ctx, conn, domain); err != nil {
				if err := s.rebuildCacheEntry(ctx, conn, domain); err != nil {
					return fmt.Errorf("cannot rebuild cache entry: %w", err)
				}

				if err := cache.LoadByDomain(ctx, conn, domain); err != nil {
					return fmt.Errorf("cannot load certificate from cache after rebuild: %w", err)
				}
			}

			fullCertPEM := cache.CertificatePEM
			if cache.CertificateChain != nil {
				fullCertPEM += "\n" + *cache.CertificateChain
			}

			tlsCert, err := tls.X509KeyPair([]byte(fullCertPEM), []byte(cache.PrivateKeyPEM))
			if err != nil {
				return fmt.Errorf("cannot parse certificate: %w", err)
			}

			cert = &tlsCert

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (s *Selector) rebuildCacheEntry(ctx context.Context, conn pg.Querier, domain string) error {
	if err := requireRoutableDomain(ctx, conn, domain); err != nil {
		return err
	}

	var certificate coredata.Certificate
	if err := certificate.LoadByHostname(ctx, conn, coredata.NewNoScope(), domain); err != nil {
		return fmt.Errorf("cannot load certificate: %w", err)
	}

	if certificate.Status != coredata.CertificateStatusActive {
		return fmt.Errorf("hostname does not have active SSL certificate")
	}

	if err := certificate.ParseCertificate(s.encryptionKey); err != nil {
		return fmt.Errorf("cannot parse certificate: %w", err)
	}

	if len(certificate.SSLCertificatePEM) == 0 {
		return fmt.Errorf("certificate has no certificate PEM data")
	}

	if len(certificate.EncryptedSSLPrivateKey) == 0 {
		return fmt.Errorf("certificate has no encrypted private key data")
	}

	privateKeyPEM, err := certificate.DecryptPrivateKey(s.encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt private key: %w", err)
	}

	s.cache.Store(domain, certificate.SSLCertificate)

	cache := &coredata.CachedCertificate{
		Domain:           certificate.Hostname,
		CertificatePEM:   string(certificate.SSLCertificatePEM),
		PrivateKeyPEM:    string(privateKeyPEM),
		CertificateChain: certificate.SSLCertificateChain,
		ExpiresAt:        *certificate.SSLExpiresAt,
		CachedAt:         time.Now(),
		CertificateID:    certificate.ID,
	}

	if err := cache.Upsert(ctx, conn); err != nil {
		return fmt.Errorf("cannot insert cache entry: %w", err)
	}

	return nil
}

// requireRoutableDomain ensures the SNI hostname still maps to a custom domain
// row. Orphaned certificates left after domain deletion must not be served.
func requireRoutableDomain(ctx context.Context, conn pg.Querier, domain string) error {
	var customDomain coredata.CustomDomain
	if err := customDomain.LoadByDomain(ctx, conn, coredata.NewNoScope(), domain); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return err
		}

		return fmt.Errorf("cannot load custom domain: %w", err)
	}

	if customDomain.CertificateID == nil {
		return coredata.ErrResourceNotFound
	}

	return nil
}
