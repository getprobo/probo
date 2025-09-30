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
	"crypto/tls"
	"fmt"
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/crypto/cipher"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	CustomDomain struct {
		ID                         gid.GID                        `db:"id"`
		OrganizationID             gid.GID                        `db:"organization_id"`
		Domain                     string                         `db:"domain"`
		VerificationStatus         CustomDomainVerificationStatus `db:"verification_status"`
		VerificationMethod         *string                        `db:"verification_method"`
		VerificationToken          []byte                         `db:"-"` // Decrypted value
		EncryptedVerificationToken []byte                         `db:"encrypted_verification_token"`
		AcmeChallengeRecord        *string                        `db:"acme_challenge_record"`
		SSLCertificate             *tls.Certificate               `db:"-"` // Parsed certificate
		SSLCertificatePEM          []byte                         `db:"-"` // Decrypted PEM
		EncryptedSSLCertificate    []byte                         `db:"encrypted_ssl_certificate"`
		SSLPrivateKeyPEM           []byte                         `db:"-"` // Decrypted PEM
		EncryptedSSLPrivateKey     []byte                         `db:"encrypted_ssl_private_key"`
		SSLCertificateChain        *string                        `db:"ssl_certificate_chain"`
		SSLStatus                  *CustomDomainSSLStatus         `db:"ssl_status"`
		SSLExpiresAt               *time.Time                     `db:"ssl_expires_at"`
		IsActive                   bool                           `db:"is_active"`
		CreatedAt                  time.Time                      `db:"created_at"`
		UpdatedAt                  time.Time                      `db:"updated_at"`
		VerifiedAt                 *time.Time                     `db:"verified_at"`
		LastVerificationAttempt    *time.Time                     `db:"last_verification_attempt"`
		VerificationAttempts       int                            `db:"verification_attempts"`
	}

	CustomDomains []*CustomDomain
)

func NewCustomDomain(orgID gid.GID, domain string) *CustomDomain {
	now := time.Now()
	return &CustomDomain{
		ID:                   gid.New(orgID.TenantID(), CustomDomainEntityType),
		OrganizationID:       orgID,
		Domain:               domain,
		VerificationStatus:   CustomDomainVerificationStatusPending,
		IsActive:             false,
		VerificationAttempts: 0,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

func (cd *CustomDomain) CursorKey(orderBy CustomDomainOrderField) page.CursorKey {
	switch orderBy {
	case CustomDomainOrderFieldCreatedAt:
		return page.NewCursorKey(cd.ID, cd.CreatedAt)
	case CustomDomainOrderFieldDomain:
		return page.NewCursorKey(cd.ID, cd.Domain)
	default:
		panic(fmt.Sprintf("unsupported order by: %s", orderBy))
	}
}

func (cd *CustomDomain) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
	domainID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	verification_status,
	verification_method,
	encrypted_verification_token,
	acme_challenge_record,
	encrypted_ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	is_active,
	created_at,
	updated_at,
	verified_at,
	last_verification_attempt,
	verification_attempts
FROM
	custom_domains
WHERE
	%s
	AND id = @domain_id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"domain_id": domainID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query custom domain: %w", err)
	}

	customDomain, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CustomDomain])
	if err != nil {
		return fmt.Errorf("cannot collect custom domain: %w", err)
	}

	*cd = customDomain

	// Decrypt verification token
	if len(cd.EncryptedVerificationToken) > 0 {
		decrypted, err := cipher.Decrypt(cd.EncryptedVerificationToken, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt verification token: %w", err)
		}
		cd.VerificationToken = decrypted
	}

	// Decrypt SSL certificate
	if len(cd.EncryptedSSLCertificate) > 0 {
		decrypted, err := cipher.Decrypt(cd.EncryptedSSLCertificate, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt SSL certificate: %w", err)
		}
		cd.SSLCertificatePEM = decrypted
	}

	// Decrypt SSL private key
	if len(cd.EncryptedSSLPrivateKey) > 0 {
		decrypted, err := cipher.Decrypt(cd.EncryptedSSLPrivateKey, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt SSL private key: %w", err)
		}
		cd.SSLPrivateKeyPEM = decrypted
	}

	// Parse certificate and key into tls.Certificate if both are present
	if cd.SSLCertificatePEM != nil && cd.SSLPrivateKeyPEM != nil {
		// Build full certificate PEM with chain if present
		fullCertPEM := string(cd.SSLCertificatePEM)
		if cd.SSLCertificateChain != nil && *cd.SSLCertificateChain != "" {
			fullCertPEM += "\n" + *cd.SSLCertificateChain
		}

		tlsCert, err := tls.X509KeyPair([]byte(fullCertPEM), cd.SSLPrivateKeyPEM)
		if err != nil {
			return fmt.Errorf("cannot parse certificate and key: %w", err)
		}
		cd.SSLCertificate = &tlsCert
	}

	return nil
}

func (cd *CustomDomain) LoadByDomain(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
	domain string,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	verification_status,
	verification_method,
	encrypted_verification_token,
	acme_challenge_record,
	encrypted_ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	is_active,
	created_at,
	updated_at,
	verified_at,
	last_verification_attempt,
	verification_attempts
FROM
	custom_domains
WHERE
	%s
	domain = @domain
LIMIT 1
	`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"domain": domain}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query custom domain: %w", err)
	}

	customDomain, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CustomDomain])
	if err != nil {
		return fmt.Errorf("cannot collect custom domain: %w", err)
	}

	*cd = customDomain

	// Decrypt verification token
	if len(cd.EncryptedVerificationToken) > 0 {
		decrypted, err := cipher.Decrypt(cd.EncryptedVerificationToken, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt verification token: %w", err)
		}
		cd.VerificationToken = decrypted
	}

	// Decrypt SSL certificate
	if len(cd.EncryptedSSLCertificate) > 0 {
		decrypted, err := cipher.Decrypt(cd.EncryptedSSLCertificate, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt SSL certificate: %w", err)
		}
		cd.SSLCertificatePEM = decrypted
	}

	// Decrypt SSL private key
	if len(cd.EncryptedSSLPrivateKey) > 0 {
		decrypted, err := cipher.Decrypt(cd.EncryptedSSLPrivateKey, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt SSL private key: %w", err)
		}
		cd.SSLPrivateKeyPEM = decrypted
	}

	// Parse certificate and key into tls.Certificate if both are present
	if cd.SSLCertificatePEM != nil && cd.SSLPrivateKeyPEM != nil {
		// Build full certificate PEM with chain if present
		fullCertPEM := string(cd.SSLCertificatePEM)
		if cd.SSLCertificateChain != nil && *cd.SSLCertificateChain != "" {
			fullCertPEM += "\n" + *cd.SSLCertificateChain
		}

		tlsCert, err := tls.X509KeyPair([]byte(fullCertPEM), cd.SSLPrivateKeyPEM)
		if err != nil {
			return fmt.Errorf("cannot parse certificate and key: %w", err)
		}
		cd.SSLCertificate = &tlsCert
	}

	return nil
}

func (cd *CustomDomain) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
	verificationToken []byte,
) error {
	encryptedToken, err := cipher.Encrypt(verificationToken, encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt verification token: %w", err)
	}

	var encryptedCert []byte
	if len(cd.SSLCertificatePEM) > 0 {
		encryptedCert, err = cipher.Encrypt(cd.SSLCertificatePEM, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot encrypt SSL certificate: %w", err)
		}
	}

	var encryptedKey []byte
	if len(cd.SSLPrivateKeyPEM) > 0 {
		encryptedKey, err = cipher.Encrypt(cd.SSLPrivateKeyPEM, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot encrypt SSL private key: %w", err)
		}
	}

	q := `
INSERT INTO custom_domains (
	id,
	tenant_id,
	organization_id,
	domain,
	verification_status,
	verification_method,
	encrypted_verification_token,
	acme_challenge_record,
	encrypted_ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	is_active,
	created_at,
	updated_at,
	verified_at,
	last_verification_attempt,
	verification_attempts
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@domain,
	@verification_status,
	@verification_method,
	@encrypted_verification_token,
	@acme_challenge_record,
	@encrypted_ssl_certificate,
	@encrypted_ssl_private_key,
	@ssl_certificate_chain,
	@ssl_status,
	@ssl_expires_at,
	@is_active,
	@created_at,
	@updated_at,
	@verified_at,
	@last_verification_attempt,
	@verification_attempts
)
`

	args := pgx.NamedArgs{
		"id":                           cd.ID,
		"tenant_id":                    scope.GetTenantID(),
		"organization_id":              cd.OrganizationID,
		"domain":                       cd.Domain,
		"verification_status":          cd.VerificationStatus,
		"verification_method":          cd.VerificationMethod,
		"encrypted_verification_token": encryptedToken,
		"acme_challenge_record":        cd.AcmeChallengeRecord,
		"encrypted_ssl_certificate":    encryptedCert,
		"encrypted_ssl_private_key":    encryptedKey,
		"ssl_certificate_chain":        cd.SSLCertificateChain,
		"ssl_status":                   cd.SSLStatus,
		"ssl_expires_at":               cd.SSLExpiresAt,
		"is_active":                    cd.IsActive,
		"created_at":                   cd.CreatedAt,
		"updated_at":                   cd.UpdatedAt,
		"verified_at":                  cd.VerifiedAt,
		"last_verification_attempt":    cd.LastVerificationAttempt,
		"verification_attempts":        cd.VerificationAttempts,
	}

	_, err = conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert custom domain: %w", err)
	}

	cd.VerificationToken = verificationToken
	cd.EncryptedVerificationToken = encryptedToken
	cd.EncryptedSSLCertificate = encryptedCert
	cd.EncryptedSSLPrivateKey = encryptedKey

	return nil
}

func (cd *CustomDomain) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
) error {
	var encryptedToken []byte
	if len(cd.VerificationToken) > 0 {
		var err error
		encryptedToken, err = cipher.Encrypt(cd.VerificationToken, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot encrypt verification token: %w", err)
		}
	}

	var encryptedCert []byte
	if len(cd.SSLCertificatePEM) > 0 {
		var err error
		encryptedCert, err = cipher.Encrypt(cd.SSLCertificatePEM, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot encrypt SSL certificate: %w", err)
		}
	}

	var encryptedKey []byte
	if len(cd.SSLPrivateKeyPEM) > 0 {
		var err error
		encryptedKey, err = cipher.Encrypt(cd.SSLPrivateKeyPEM, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot encrypt SSL private key: %w", err)
		}
	}

	q := `
UPDATE
	custom_domains
SET
	verification_status = @verification_status,
	verification_method = @verification_method,
	encrypted_verification_token = @encrypted_verification_token,
	acme_challenge_record = @acme_challenge_record,
	encrypted_ssl_certificate = @encrypted_ssl_certificate,
	encrypted_ssl_private_key = @encrypted_ssl_private_key,
	ssl_certificate_chain = @ssl_certificate_chain,
	ssl_status = @ssl_status,
	ssl_expires_at = @ssl_expires_at,
	is_active = @is_active,
	updated_at = @updated_at,
	verified_at = @verified_at,
	last_verification_attempt = @last_verification_attempt,
	verification_attempts = @verification_attempts
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"id":                           cd.ID,
		"verification_status":          cd.VerificationStatus,
		"verification_method":          cd.VerificationMethod,
		"encrypted_verification_token": encryptedToken,
		"acme_challenge_record":        cd.AcmeChallengeRecord,
		"encrypted_ssl_certificate":    encryptedCert,
		"encrypted_ssl_private_key":    encryptedKey,
		"ssl_certificate_chain":        cd.SSLCertificateChain,
		"ssl_status":                   cd.SSLStatus,
		"ssl_expires_at":               cd.SSLExpiresAt,
		"is_active":                    cd.IsActive,
		"updated_at":                   cd.UpdatedAt,
		"verified_at":                  cd.VerifiedAt,
		"last_verification_attempt":    cd.LastVerificationAttempt,
		"verification_attempts":        cd.VerificationAttempts,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update custom domain: %w", err)
	}

	cd.EncryptedVerificationToken = encryptedToken
	cd.EncryptedSSLCertificate = encryptedCert
	cd.EncryptedSSLPrivateKey = encryptedKey

	return nil
}

func (cd *CustomDomain) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `DELETE FROM custom_domains WHERE %s AND id = @id`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"id": cd.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete custom domain: %w", err)
	}

	return nil
}

func (cds *CustomDomains) ListCustomDomainsByOrganization(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[CustomDomainOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	verification_status,
	verification_method,
	encrypted_verification_token,
	acme_challenge_record,
	encrypted_ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	is_active,
	created_at,
	updated_at,
	verified_at,
	last_verification_attempt,
	verification_attempts
FROM
	custom_domains
WHERE
	%s 
	AND organization_id = @organization_id
	AND %s
	`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query custom domains: %w", err)
	}

	domains, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CustomDomain])
	if err != nil {
		return fmt.Errorf("cannot collect custom domains: %w", err)
	}

	*cds = domains

	return nil
}

func (cds *CustomDomains) ListDomainsForRenewal(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	verification_status,
	verification_method,
	encrypted_verification_token,
	acme_challenge_record,
	encrypted_ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	is_active,
	created_at,
	updated_at,
	verified_at,
	last_verification_attempt,
	verification_attempts
FROM
	custom_domains
WHERE
	%s
	is_active = true
	AND ssl_status = @ssl_status
	AND ssl_expires_at < NOW() + INTERVAL '30 days'
ORDER BY
	ssl_expires_at ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"ssl_status": CustomDomainSSLStatusActive}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query domains for renewal: %w", err)
	}

	customDomains, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CustomDomain])
	if err != nil {
		return fmt.Errorf("cannot collect domains: %w", err)
	}

	*cds = customDomains

	return nil
}

func (cds *CustomDomains) LoadActiveCertificates(ctx context.Context, conn pg.Conn, scope Scoper) error {
	q := `
SELECT
	id,
	tenant_id,
	organization_id,
	domain,
	verification_status,
	verification_method,
	encrypted_verification_token,
	acme_challenge_record,
	encrypted_ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	is_active,
	created_at,
	updated_at,
	verified_at,
	last_verification_attempt,
	verification_attempts
FROM
	custom_domains
WHERE
	%s
	is_active = @is_active
	AND encrypted_ssl_certificate IS NOT NULL
	AND ssl_status = @ssl_status
ORDER BY
	ssl_expires_at DESC
	`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"is_active":  true,
		"ssl_status": CustomDomainSSLStatusActive,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query active certificates: %w", err)
	}

	customDomains, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CustomDomain])
	if err != nil {
		return fmt.Errorf("cannot collect active certificates: %w", err)
	}

	*cds = customDomains

	return nil
}
