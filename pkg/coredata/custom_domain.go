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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	CustomDomain struct {
		ID                     gid.GID               `db:"id"`
		OrganizationID         gid.GID               `db:"organization_id"`
		Domain                 string                `db:"domain"`
		HTTPChallengeToken     *string               `db:"http_challenge_token"`
		HTTPChallengeKeyAuth   *string               `db:"http_challenge_key_auth"`
		HTTPChallengeURL       *string               `db:"http_challenge_url"`
		HTTPOrderURL           *string               `db:"http_order_url"`
		SSLCertificate         *tls.Certificate      `db:"-"`
		SSLCertificatePEM      []byte                `db:"ssl_certificate"`
		EncryptedSSLPrivateKey []byte                `db:"encrypted_ssl_private_key"`
		SSLCertificateChain    *string               `db:"ssl_certificate_chain"`
		SSLStatus              CustomDomainSSLStatus `db:"ssl_status"`
		SSLExpiresAt           *time.Time            `db:"ssl_expires_at"`
		SSLRetryCount          int                   `db:"ssl_retry_count"`
		SSLLastAttemptAt       *time.Time            `db:"ssl_last_attempt_at"`
		CreatedAt              time.Time             `db:"created_at"`
		UpdatedAt              time.Time             `db:"updated_at"`
	}

	CustomDomains []*CustomDomain
)

func NewCustomDomain(tenantID gid.TenantID, domain string) *CustomDomain {
	now := time.Now()
	return &CustomDomain{
		ID:        gid.New(tenantID, CustomDomainEntityType),
		SSLStatus: CustomDomainSSLStatusPending,
		Domain:    domain,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (cd *CustomDomain) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM custom_domains WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, cd.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query custom domain authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (cd *CustomDomain) CursorKey(field CustomDomainOrderField) page.CursorKey {
	switch field {
	case CustomDomainOrderFieldCreatedAt:
		return page.NewCursorKey(cd.ID, cd.CreatedAt)
	case CustomDomainOrderFieldDomain:
		return page.NewCursorKey(cd.ID, cd.Domain)
	case CustomDomainOrderFieldUpdatedAt:
		return page.NewCursorKey(cd.ID, cd.UpdatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (cd *CustomDomain) DecryptPrivateKey(encryptionKey cipher.EncryptionKey) ([]byte, error) {
	if len(cd.EncryptedSSLPrivateKey) == 0 {
		return nil, nil
	}

	decrypted, err := cipher.Decrypt(cd.EncryptedSSLPrivateKey, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt SSL private key: %w", err)
	}

	return decrypted, nil
}

func (cd *CustomDomain) EncryptPrivateKey(privateKeyPEM []byte, encryptionKey cipher.EncryptionKey) error {
	if len(privateKeyPEM) == 0 {
		cd.EncryptedSSLPrivateKey = nil
		return nil
	}

	encrypted, err := cipher.Encrypt(privateKeyPEM, encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt SSL private key: %w", err)
	}

	cd.EncryptedSSLPrivateKey = encrypted
	return nil
}

func (cd *CustomDomain) ParseCertificate(encryptionKey cipher.EncryptionKey) error {
	if len(cd.SSLCertificatePEM) == 0 {
		return fmt.Errorf("no certificate PEM data")
	}

	privateKeyPEM, err := cd.DecryptPrivateKey(encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt private key: %w", err)
	}

	if len(privateKeyPEM) == 0 {
		return fmt.Errorf("no private key data")
	}

	fullCertPEM := string(cd.SSLCertificatePEM)
	if cd.SSLCertificateChain != nil && *cd.SSLCertificateChain != "" {
		fullCertPEM += "\n" + *cd.SSLCertificateChain
	}

	tlsCert, err := tls.X509KeyPair([]byte(fullCertPEM), privateKeyPEM)
	if err != nil {
		return fmt.Errorf("cannot parse certificate and key: %w", err)
	}

	cd.SSLCertificate = &tlsCert
	return nil
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
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND id = @id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"id": domainID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query custom domain: %w", err)
	}

	customDomain, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CustomDomain])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect custom domain: %w", err)
	}

	*cd = customDomain

	return nil
}

func (cd *CustomDomain) LoadByIDForUpdateSkipLocked(
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
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND id = @id
LIMIT 1
FOR UPDATE SKIP LOCKED
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"id": domainID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query custom domain for update: %w", err)
	}

	customDomain, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CustomDomain])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect custom domain: %w", err)
	}

	*cd = customDomain

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
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND domain = @domain
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
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect custom domain: %w", err)
	}

	*cd = customDomain

	return nil
}

func (cd *CustomDomain) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
	organizationID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND organization_id = @organization_id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query custom domain: %w", err)
	}

	customDomain, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CustomDomain])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect custom domain: %w", err)
	}

	*cd = customDomain

	return nil
}

func (cd *CustomDomain) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
) error {
	var encryptedKey []byte
	if len(cd.EncryptedSSLPrivateKey) > 0 {
		encryptedKey = cd.EncryptedSSLPrivateKey
	}

	q := `
INSERT INTO custom_domains (
	id,
	tenant_id,
	organization_id,
	domain,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@domain,
	@http_challenge_token,
	@http_challenge_key_auth,
	@http_challenge_url,
	@http_order_url,
	@ssl_certificate,
	@encrypted_ssl_private_key,
	@ssl_certificate_chain,
	@ssl_status,
	@ssl_expires_at,
	@ssl_retry_count,
	@ssl_last_attempt_at,
	@created_at,
	@updated_at
)
`

	args := pgx.NamedArgs{
		"id":                        cd.ID,
		"tenant_id":                 scope.GetTenantID(),
		"organization_id":           cd.OrganizationID,
		"domain":                    cd.Domain,
		"http_challenge_token":      cd.HTTPChallengeToken,
		"http_challenge_key_auth":   cd.HTTPChallengeKeyAuth,
		"http_challenge_url":        cd.HTTPChallengeURL,
		"http_order_url":            cd.HTTPOrderURL,
		"ssl_certificate":           cd.SSLCertificatePEM,
		"encrypted_ssl_private_key": encryptedKey,
		"ssl_certificate_chain":     cd.SSLCertificateChain,
		"ssl_status":                cd.SSLStatus,
		"ssl_expires_at":            cd.SSLExpiresAt,
		"ssl_retry_count":           cd.SSLRetryCount,
		"ssl_last_attempt_at":       cd.SSLLastAttemptAt,
		"created_at":                cd.CreatedAt,
		"updated_at":                cd.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "custom_domains_domain_key" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert custom domain: %w", err)
	}

	cd.EncryptedSSLPrivateKey = encryptedKey

	return nil
}

func (cd *CustomDomain) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
) error {
	var encryptedKey []byte
	if len(cd.EncryptedSSLPrivateKey) > 0 {
		encryptedKey = cd.EncryptedSSLPrivateKey
	}

	q := `
UPDATE
	custom_domains
SET
	http_challenge_token = @http_challenge_token,
	http_challenge_key_auth = @http_challenge_key_auth,
	http_challenge_url = @http_challenge_url,
	http_order_url = @http_order_url,
	ssl_certificate = @ssl_certificate,
	encrypted_ssl_private_key = @encrypted_ssl_private_key,
	ssl_certificate_chain = @ssl_certificate_chain,
	ssl_status = @ssl_status,
	ssl_expires_at = @ssl_expires_at,
	ssl_retry_count = @ssl_retry_count,
	ssl_last_attempt_at = @ssl_last_attempt_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"id":                        cd.ID,
		"http_challenge_token":      cd.HTTPChallengeToken,
		"http_challenge_key_auth":   cd.HTTPChallengeKeyAuth,
		"http_challenge_url":        cd.HTTPChallengeURL,
		"http_order_url":            cd.HTTPOrderURL,
		"ssl_certificate":           cd.SSLCertificatePEM,
		"encrypted_ssl_private_key": encryptedKey,
		"ssl_certificate_chain":     cd.SSLCertificateChain,
		"ssl_status":                cd.SSLStatus,
		"ssl_expires_at":            cd.SSLExpiresAt,
		"ssl_retry_count":           cd.SSLRetryCount,
		"ssl_last_attempt_at":       cd.SSLLastAttemptAt,
		"updated_at":                time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update custom domain: %w", err)
	}

	cd.EncryptedSSLPrivateKey = encryptedKey

	return nil
}

func (cd *CustomDomain) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM
	custom_domains
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"id": cd.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete custom domain: %w", err)
	}

	return nil
}

func (cd *CustomDomain) LoadByHTTPChallengeToken(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
	token string,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND http_challenge_token = @token
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"token": token}
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

	return nil
}

func (domains *CustomDomains) ListDomainsForRenewal(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND ssl_status = @status
	AND ssl_expires_at <= CURRENT_TIMESTAMP + INTERVAL '30 days'
ORDER BY
	ssl_expires_at ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"status": string(CustomDomainSSLStatusActive)}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query custom domains for renewal: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CustomDomain])
	if err != nil {
		return fmt.Errorf("cannot collect custom domains: %w", err)
	}

	*domains = result
	return nil
}

func (domains *CustomDomains) ListDomainsWithPendingHTTPChallenges(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND ssl_status = ANY(@statuses)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"statuses": []string{
			string(CustomDomainSSLStatusPending),
			string(CustomDomainSSLStatusProvisioning),
			string(CustomDomainSSLStatusRenewing),
		},
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query custom domains with pending challenges: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CustomDomain])
	if err != nil {
		return fmt.Errorf("cannot collect custom domains: %w", err)
	}

	*domains = result
	return nil
}

func (domains *CustomDomains) LoadActiveCertificates(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND ssl_status = @status
	AND ssl_certificate IS NOT NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"status": string(CustomDomainSSLStatusActive)}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query active certificates: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CustomDomain])
	if err != nil {
		return fmt.Errorf("cannot collect custom domains: %w", err)
	}

	*domains = result
	return nil
}

func (domains *CustomDomains) ListStaleProvisioningDomains(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND (
		(ssl_status IN (@provisioning_status, @renewing_status) AND updated_at < CURRENT_TIMESTAMP - INTERVAL '4 hours')
		OR
		(ssl_retry_count > 0 AND ssl_last_attempt_at < CURRENT_TIMESTAMP - INTERVAL '24 hours')
	)
	AND ssl_status != @failed_status
	AND ssl_status != @active_status
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"provisioning_status": string(CustomDomainSSLStatusProvisioning),
		"renewing_status":     string(CustomDomainSSLStatusRenewing),
		"failed_status":       string(CustomDomainSSLStatusFailed),
		"active_status":       string(CustomDomainSSLStatusActive),
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query stale provisioning domains: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CustomDomain])
	if err != nil {
		return fmt.Errorf("cannot collect stale provisioning domains: %w", err)
	}

	*domains = result
	return nil
}
