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

package coredata

import (
	"context"
	"fmt"
	"time"

	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	CachedCertificate struct {
		Domain           string    `db:"domain"`
		CertificatePEM   string    `db:"certificate_pem"`
		PrivateKeyPEM    string    `db:"private_key_pem"` // Decrypted for fast TLS handshake
		CertificateChain *string   `db:"certificate_chain"`
		ExpiresAt        time.Time `db:"expires_at"`
		CachedAt         time.Time `db:"cached_at"`
		CustomDomainID   gid.GID   `db:"custom_domain_id"`
	}

	CachedCertificates []*CachedCertificate
)

func NewCachedCertificate(domain string, domainID gid.GID) *CachedCertificate {
	now := time.Now()
	return &CachedCertificate{
		Domain:         domain,
		CustomDomainID: domainID,
		CachedAt:       now,
	}
}

func (cc *CachedCertificate) LoadByDomain(ctx context.Context, conn pg.Conn, domain string) error {
	q := `
SELECT
	domain,
	certificate_pem,
	private_key_pem,
	certificate_chain,
	expires_at,
	cached_at,
	custom_domain_id
FROM
	cached_certificates
WHERE
	domain = @domain
	AND expires_at > NOW()
LIMIT 1
`

	args := pgx.NamedArgs{"domain": domain}
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query certificate cache: %w", err)
	}

	cache, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CachedCertificate])
	if err != nil {
		return fmt.Errorf("cannot collect certificate cache: %w", err)
	}

	*cc = cache
	return nil
}

func (cc *CachedCertificate) Upsert(ctx context.Context, conn pg.Conn) error {
	cc.CachedAt = time.Now()

	q := `
INSERT INTO cached_certificates (
	domain,
	certificate_pem,
	private_key_pem,
	certificate_chain,
	expires_at,
	cached_at,
	custom_domain_id
) VALUES (
	@domain,
	@certificate_pem,
	@private_key_pem,
	@certificate_chain,
	@expires_at,
	@cached_at,
	@custom_domain_id
)
ON CONFLICT (domain) DO UPDATE SET
	certificate_pem = EXCLUDED.certificate_pem,
	private_key_pem = EXCLUDED.private_key_pem,
	certificate_chain = EXCLUDED.certificate_chain,
	expires_at = EXCLUDED.expires_at,
	cached_at = NOW(),
	custom_domain_id = EXCLUDED.custom_domain_id
`

	args := pgx.NamedArgs{
		"domain":            cc.Domain,
		"certificate_pem":   cc.CertificatePEM,
		"private_key_pem":   cc.PrivateKeyPEM,
		"certificate_chain": cc.CertificateChain,
		"expires_at":        cc.ExpiresAt,
		"cached_at":         cc.CachedAt,
		"custom_domain_id":  cc.CustomDomainID,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert certificate cache: %w", err)
	}

	return nil
}

func (cc *CachedCertificate) Delete(ctx context.Context, conn pg.Conn, domain string) error {
	q := `DELETE FROM cached_certificates WHERE domain = @domain`
	args := pgx.NamedArgs{"domain": domain}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete certificate cache: %w", err)
	}

	return nil
}

func (cc *CachedCertificates) CountAll(ctx context.Context, conn pg.Conn) (int, error) {
	q := `SELECT COUNT(*) FROM cached_certificates`

	var count int
	err := conn.QueryRow(ctx, q, pgx.NamedArgs{}).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count certificate cache: %w", err)
	}

	return count, nil
}

func (cc *CachedCertificates) CleanExpired(ctx context.Context, conn pg.Conn) error {
	q := `
DELETE
FROM
	cached_certificates
WHERE
	expires_at < NOW() - INTERVAL '30 days'
`

	_, err := conn.Exec(ctx, q, pgx.NamedArgs{})
	if err != nil {
		return fmt.Errorf("cannot clean expired cache: %w", err)
	}

	return nil
}

func (cc *CachedCertificate) RefreshFromDomain(ctx context.Context, conn pg.Conn, domain *CustomDomain, encryptionKey cipher.EncryptionKey) error {
	if domain.SSLCertificate == nil {
		return fmt.Errorf("domain has no parsed certificate")
	}

	if len(domain.SSLCertificatePEM) == 0 {
		return fmt.Errorf("domain has no certificate PEM")
	}

	privateKeyPEM, err := domain.DecryptPrivateKey(encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt private key: %w", err)
	}

	if len(privateKeyPEM) == 0 {
		return fmt.Errorf("domain has no private key PEM")
	}

	if domain.SSLExpiresAt == nil {
		return fmt.Errorf("domain certificate has no expiry date")
	}

	cache := &CachedCertificate{
		Domain:           domain.Domain,
		CertificatePEM:   string(domain.SSLCertificatePEM),
		PrivateKeyPEM:    string(privateKeyPEM),
		CertificateChain: domain.SSLCertificateChain,
		ExpiresAt:        *domain.SSLExpiresAt,
		CachedAt:         time.Now(),
		CustomDomainID:   domain.ID,
	}

	return cache.Upsert(ctx, conn)
}
