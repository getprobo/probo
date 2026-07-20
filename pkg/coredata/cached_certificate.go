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

package coredata

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

type (
	CachedCertificate struct {
		Domain           string    `db:"domain"`
		CertificatePEM   string    `db:"certificate_pem"`
		PrivateKeyPEM    string    `db:"private_key_pem"` // Decrypted for fast TLS handshake
		CertificateChain *string   `db:"certificate_chain"`
		ExpiresAt        time.Time `db:"expires_at"`
		CachedAt         time.Time `db:"cached_at"`
		CertificateID    gid.GID   `db:"certificate_id"`
	}

	CachedCertificates []*CachedCertificate
)

func (cc *CachedCertificate) LoadByDomain(ctx context.Context, conn pg.Querier, domain string) error {
	q := `
SELECT
	domain,
	certificate_pem,
	private_key_pem,
	certificate_chain,
	expires_at,
	cached_at,
	certificate_id
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

func (cc *CachedCertificate) Upsert(ctx context.Context, conn pg.Querier) error {
	cc.CachedAt = time.Now()

	q := `
INSERT INTO cached_certificates (
	domain,
	certificate_pem,
	private_key_pem,
	certificate_chain,
	expires_at,
	cached_at,
	certificate_id
) VALUES (
	@domain,
	@certificate_pem,
	@private_key_pem,
	@certificate_chain,
	@expires_at,
	@cached_at,
	@certificate_id
)
ON CONFLICT (domain) DO UPDATE SET
	certificate_pem = EXCLUDED.certificate_pem,
	private_key_pem = EXCLUDED.private_key_pem,
	certificate_chain = EXCLUDED.certificate_chain,
	expires_at = EXCLUDED.expires_at,
	cached_at = NOW(),
	certificate_id = EXCLUDED.certificate_id
`

	args := pgx.NamedArgs{
		"domain":            cc.Domain,
		"certificate_pem":   cc.CertificatePEM,
		"private_key_pem":   cc.PrivateKeyPEM,
		"certificate_chain": cc.CertificateChain,
		"expires_at":        cc.ExpiresAt,
		"cached_at":         cc.CachedAt,
		"certificate_id":    cc.CertificateID,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert certificate cache: %w", err)
	}

	return nil
}

func (cc *CachedCertificate) Delete(ctx context.Context, conn pg.Tx, domain string) error {
	q := `DELETE FROM cached_certificates WHERE domain = @domain`
	args := pgx.NamedArgs{"domain": domain}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete certificate cache: %w", err)
	}

	return nil
}

func (cc *CachedCertificates) CountAll(ctx context.Context, conn pg.Querier) (int, error) {
	q := `SELECT COUNT(*) FROM cached_certificates`

	var count int

	err := conn.QueryRow(ctx, q, pgx.NamedArgs{}).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count certificate cache: %w", err)
	}

	return count, nil
}

func (cc *CachedCertificates) CleanExpired(ctx context.Context, conn pg.Querier) error {
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

// DeleteWhereCertificateIDNotIn removes cache rows whose certificate is not
// among the provided IDs. An empty keep set deletes every cache row.
func (cc *CachedCertificates) DeleteWhereCertificateIDNotIn(
	ctx context.Context,
	conn pg.Querier,
	keepCertificateIDs []gid.GID,
) error {
	q := `
DELETE FROM
	cached_certificates
WHERE
	NOT (certificate_id = ANY(@keep_certificate_ids::text[]))
`

	_, err := conn.Exec(
		ctx,
		q,
		pgx.NamedArgs{"keep_certificate_ids": keepCertificateIDs},
	)
	if err != nil {
		return fmt.Errorf("cannot delete unreferenced certificate cache: %w", err)
	}

	return nil
}

func (cc *CachedCertificate) RefreshFromCertificate(ctx context.Context, conn pg.Querier, certificate *Certificate, encryptionKey cipher.EncryptionKey) error {
	if certificate.SSLCertificate == nil {
		return fmt.Errorf("certificate has no parsed certificate")
	}

	if len(certificate.SSLCertificatePEM) == 0 {
		return fmt.Errorf("certificate has no certificate PEM")
	}

	privateKeyPEM, err := certificate.DecryptPrivateKey(encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt private key: %w", err)
	}

	if len(privateKeyPEM) == 0 {
		return fmt.Errorf("certificate has no private key PEM")
	}

	if certificate.SSLExpiresAt == nil {
		return fmt.Errorf("certificate has no expiry date")
	}

	cache := &CachedCertificate{
		Domain:           certificate.Hostname,
		CertificatePEM:   string(certificate.SSLCertificatePEM),
		PrivateKeyPEM:    string(privateKeyPEM),
		CertificateChain: certificate.SSLCertificateChain,
		ExpiresAt:        *certificate.SSLExpiresAt,
		CachedAt:         time.Now(),
		CertificateID:    certificate.ID,
	}

	return cache.Upsert(ctx, conn)
}
