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
	"crypto/tls"
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
		func(conn pg.Conn) error {
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

func (s *Selector) rebuildCacheEntry(ctx context.Context, conn pg.Conn, domain string) error {
	var customDomain coredata.CustomDomain
	if err := customDomain.LoadByDomain(ctx, conn, coredata.NewNoScope(), s.encryptionKey, domain); err != nil {
		return fmt.Errorf("cannot load domain: %w", err)
	}

	if customDomain.SSLStatus != coredata.CustomDomainSSLStatusActive {
		return fmt.Errorf("domain does not have active SSL certificate")
	}

	if err := customDomain.ParseCertificate(s.encryptionKey); err != nil {
		return fmt.Errorf("cannot parse certificate: %w", err)
	}

	if len(customDomain.SSLCertificatePEM) == 0 {
		return fmt.Errorf("domain has no certificate PEM data")
	}

	if len(customDomain.EncryptedSSLPrivateKey) == 0 {
		return fmt.Errorf("domain has no encrypted private key data")
	}

	privateKeyPEM, err := customDomain.DecryptPrivateKey(s.encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt private key: %w", err)
	}

	s.cache.Store(domain, customDomain.SSLCertificate)

	cache := &coredata.CachedCertificate{
		Domain:           customDomain.Domain,
		CertificatePEM:   string(customDomain.SSLCertificatePEM),
		PrivateKeyPEM:    string(privateKeyPEM),
		CertificateChain: customDomain.SSLCertificateChain,
		ExpiresAt:        *customDomain.SSLExpiresAt,
		CachedAt:         time.Now(),
		CustomDomainID:   customDomain.ID,
	}

	if err := cache.Upsert(ctx, conn); err != nil {
		return fmt.Errorf("cannot insert cache entry: %w", err)
	}

	return nil
}

func (s *Selector) ClearCache() {
	s.cache.Range(
		func(key, _ any) bool {
			s.cache.Delete(key)
			return true
		},
	)
}
