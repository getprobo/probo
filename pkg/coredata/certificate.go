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
)

type (
	// Certificate is a generic TLS certificate together with its ACME
	// provisioning lifecycle for a single hostname. It carries no knowledge of
	// the resource it protects; the hostname is globally unique.
	Certificate struct {
		ID                     gid.GID           `db:"id"`
		Hostname               string            `db:"hostname"`
		HTTPChallengeToken     *string           `db:"http_challenge_token"`
		HTTPChallengeKeyAuth   *string           `db:"http_challenge_key_auth"`
		HTTPChallengeURL       *string           `db:"http_challenge_url"`
		HTTPOrderURL           *string           `db:"http_order_url"`
		SSLCertificate         *tls.Certificate  `db:"-"`
		SSLCertificatePEM      []byte            `db:"ssl_certificate"`
		EncryptedSSLPrivateKey []byte            `db:"encrypted_ssl_private_key"`
		SSLCertificateChain    *string           `db:"ssl_certificate_chain"`
		Status                 CertificateStatus `db:"status"`
		SSLExpiresAt           *time.Time        `db:"ssl_expires_at"`
		SSLRetryCount          int               `db:"ssl_retry_count"`
		SSLLastAttemptAt       *time.Time        `db:"ssl_last_attempt_at"`
		ProvisioningError      *string           `db:"provisioning_error"`
		CreatedAt              time.Time         `db:"created_at"`
		UpdatedAt              time.Time         `db:"updated_at"`
	}

	Certificates []*Certificate
)

func NewCertificate(
	tenantID gid.TenantID,
	hostname string,
) *Certificate {
	now := time.Now()

	return &Certificate{
		ID:        gid.New(tenantID, CertificateEntityType),
		Hostname:  hostname,
		Status:    CertificateStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Certificate) DecryptPrivateKey(encryptionKey cipher.EncryptionKey) ([]byte, error) {
	if len(c.EncryptedSSLPrivateKey) == 0 {
		return nil, nil
	}

	decrypted, err := cipher.Decrypt(c.EncryptedSSLPrivateKey, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt SSL private key: %w", err)
	}

	return decrypted, nil
}

func (c *Certificate) EncryptPrivateKey(privateKeyPEM []byte, encryptionKey cipher.EncryptionKey) error {
	if len(privateKeyPEM) == 0 {
		c.EncryptedSSLPrivateKey = nil
		return nil
	}

	encrypted, err := cipher.Encrypt(privateKeyPEM, encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt SSL private key: %w", err)
	}

	c.EncryptedSSLPrivateKey = encrypted

	return nil
}

func (c *Certificate) ParseCertificate(encryptionKey cipher.EncryptionKey) error {
	if len(c.SSLCertificatePEM) == 0 {
		return fmt.Errorf("no certificate PEM data")
	}

	privateKeyPEM, err := c.DecryptPrivateKey(encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt private key: %w", err)
	}

	if len(privateKeyPEM) == 0 {
		return fmt.Errorf("no private key data")
	}

	fullCertPEM := string(c.SSLCertificatePEM)
	if c.SSLCertificateChain != nil && *c.SSLCertificateChain != "" {
		fullCertPEM += "\n" + *c.SSLCertificateChain
	}

	tlsCert, err := tls.X509KeyPair([]byte(fullCertPEM), privateKeyPEM)
	if err != nil {
		return fmt.Errorf("cannot parse certificate and key: %w", err)
	}

	c.SSLCertificate = &tlsCert

	return nil
}

func (c *Certificate) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	certificateID gid.GID,
) error {
	q := `
SELECT
	id,
	hostname,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	provisioning_error,
	created_at,
	updated_at
FROM
	certificates
WHERE
	%s
	AND id = @id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"id": certificateID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query certificate: %w", err)
	}

	certificate, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Certificate])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect certificate: %w", err)
	}

	*c = certificate

	return nil
}

func (c *Certificate) LoadByIDForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	certificateID gid.GID,
) error {
	q := `
SELECT
	id,
	hostname,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	provisioning_error,
	created_at,
	updated_at
FROM
	certificates
WHERE
	%s
	AND id = @id
LIMIT 1
FOR UPDATE SKIP LOCKED
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": certificateID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query certificate for update: %w", err)
	}

	certificate, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Certificate])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect certificate: %w", err)
	}

	*c = certificate

	return nil
}

// LoadByIDForUpdate locks the row with a blocking FOR UPDATE. Prefer this for
// write-backs that must not be dropped (e.g. persisting a freshly issued cert).
// Use LoadByIDForUpdateSkipLocked when skipping a locked row is acceptable.
func (c *Certificate) LoadByIDForUpdate(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	certificateID gid.GID,
) error {
	q := `
SELECT
	id,
	hostname,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	provisioning_error,
	created_at,
	updated_at
FROM
	certificates
WHERE
	%s
	AND id = @id
LIMIT 1
FOR UPDATE
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": certificateID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query certificate for update: %w", err)
	}

	certificate, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Certificate])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect certificate: %w", err)
	}

	*c = certificate

	return nil
}

func (c *Certificate) LoadByHostname(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	hostname string,
) error {
	q := `
SELECT
	id,
	hostname,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	provisioning_error,
	created_at,
	updated_at
FROM
	certificates
WHERE
	%s
	AND hostname = @hostname
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"hostname": hostname}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query certificate: %w", err)
	}

	certificate, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Certificate])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect certificate: %w", err)
	}

	*c = certificate

	return nil
}

func (c *Certificate) LoadByHTTPChallengeToken(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	token string,
) error {
	q := `
SELECT
	id,
	hostname,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	provisioning_error,
	created_at,
	updated_at
FROM
	certificates
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
		return fmt.Errorf("cannot query certificate: %w", err)
	}

	certificate, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Certificate])
	if err != nil {
		return fmt.Errorf("cannot collect certificate: %w", err)
	}

	*c = certificate

	return nil
}

func (certificates *Certificates) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ids []gid.GID,
) error {
	q := `
SELECT
	id,
	hostname,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	provisioning_error,
	created_at,
	updated_at
FROM
	certificates
WHERE
	%s
	AND id = ANY(@ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"ids": ids}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query certificates: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Certificate])
	if err != nil {
		return fmt.Errorf("cannot collect certificates: %w", err)
	}

	*certificates = result

	return nil
}

func (c *Certificate) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	var encryptedKey []byte
	if len(c.EncryptedSSLPrivateKey) > 0 {
		encryptedKey = c.EncryptedSSLPrivateKey
	}

	q := `
INSERT INTO certificates (
	id,
	tenant_id,
	hostname,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	provisioning_error,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@hostname,
	@http_challenge_token,
	@http_challenge_key_auth,
	@http_challenge_url,
	@http_order_url,
	@ssl_certificate,
	@encrypted_ssl_private_key,
	@ssl_certificate_chain,
	@status,
	@ssl_expires_at,
	@ssl_retry_count,
	@ssl_last_attempt_at,
	@provisioning_error,
	@created_at,
	@updated_at
)
`

	args := pgx.NamedArgs{
		"id":                        c.ID,
		"tenant_id":                 scope.GetTenantID(),
		"hostname":                  c.Hostname,
		"http_challenge_token":      c.HTTPChallengeToken,
		"http_challenge_key_auth":   c.HTTPChallengeKeyAuth,
		"http_challenge_url":        c.HTTPChallengeURL,
		"http_order_url":            c.HTTPOrderURL,
		"ssl_certificate":           c.SSLCertificatePEM,
		"encrypted_ssl_private_key": encryptedKey,
		"ssl_certificate_chain":     c.SSLCertificateChain,
		"status":                    c.Status,
		"ssl_expires_at":            c.SSLExpiresAt,
		"ssl_retry_count":           c.SSLRetryCount,
		"ssl_last_attempt_at":       c.SSLLastAttemptAt,
		"provisioning_error":        c.ProvisioningError,
		"created_at":                c.CreatedAt,
		"updated_at":                c.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "certificates_hostname_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert certificate: %w", err)
	}

	c.EncryptedSSLPrivateKey = encryptedKey

	return nil
}

func (c *Certificate) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	var encryptedKey []byte
	if len(c.EncryptedSSLPrivateKey) > 0 {
		encryptedKey = c.EncryptedSSLPrivateKey
	}

	q := `
UPDATE
	certificates
SET
	hostname = @hostname,
	http_challenge_token = @http_challenge_token,
	http_challenge_key_auth = @http_challenge_key_auth,
	http_challenge_url = @http_challenge_url,
	http_order_url = @http_order_url,
	ssl_certificate = @ssl_certificate,
	encrypted_ssl_private_key = @encrypted_ssl_private_key,
	ssl_certificate_chain = @ssl_certificate_chain,
	status = @status,
	ssl_expires_at = @ssl_expires_at,
	ssl_retry_count = @ssl_retry_count,
	ssl_last_attempt_at = @ssl_last_attempt_at,
	provisioning_error = @provisioning_error,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"id":                        c.ID,
		"hostname":                  c.Hostname,
		"http_challenge_token":      c.HTTPChallengeToken,
		"http_challenge_key_auth":   c.HTTPChallengeKeyAuth,
		"http_challenge_url":        c.HTTPChallengeURL,
		"http_order_url":            c.HTTPOrderURL,
		"ssl_certificate":           c.SSLCertificatePEM,
		"encrypted_ssl_private_key": encryptedKey,
		"ssl_certificate_chain":     c.SSLCertificateChain,
		"status":                    c.Status,
		"ssl_expires_at":            c.SSLExpiresAt,
		"ssl_retry_count":           c.SSLRetryCount,
		"ssl_last_attempt_at":       c.SSLLastAttemptAt,
		"provisioning_error":        c.ProvisioningError,
		"updated_at":                time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update certificate: %w", err)
	}

	c.EncryptedSSLPrivateKey = encryptedKey

	return nil
}

func (c *Certificate) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM
	certificates
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete certificate: %w", err)
	}

	return nil
}

func (certificates *Certificates) ListForRenewal(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
SELECT
	id,
	hostname,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	provisioning_error,
	created_at,
	updated_at
FROM
	certificates
WHERE
	%s
	AND status = @status
	AND ssl_expires_at <= CURRENT_TIMESTAMP + INTERVAL '30 days'
ORDER BY
	ssl_expires_at ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"status": string(CertificateStatusActive)}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query certificates for renewal: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Certificate])
	if err != nil {
		return fmt.Errorf("cannot collect certificates: %w", err)
	}

	*certificates = result

	return nil
}

func (certificates *Certificates) ListWithPendingHTTPChallenges(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
SELECT
	id,
	hostname,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	provisioning_error,
	created_at,
	updated_at
FROM
	certificates
WHERE
	%s
	AND status = ANY(@statuses)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"statuses": []string{
			string(CertificateStatusPending),
			string(CertificateStatusProvisioning),
			string(CertificateStatusRenewing),
		},
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query certificates with pending challenges: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Certificate])
	if err != nil {
		return fmt.Errorf("cannot collect certificates: %w", err)
	}

	*certificates = result

	return nil
}

// LoadActiveReferenced loads active certificates that are still referenced by
// at least one custom domain. Certificates left behind after a domain is
// deleted must not be warmed into the TLS cache or served by SNI alone.
func (certificates *Certificates) LoadActiveReferenced(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
SELECT
	id,
	hostname,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	provisioning_error,
	created_at,
	updated_at
FROM
	certificates
WHERE
	%s
	AND status = @status
	AND ssl_certificate IS NOT NULL
	AND EXISTS (
		SELECT 1
		FROM custom_domains
		WHERE custom_domains.certificate_id = certificates.id
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"status": string(CertificateStatusActive)}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query active referenced certificates: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Certificate])
	if err != nil {
		return fmt.Errorf("cannot collect certificates: %w", err)
	}

	*certificates = result

	return nil
}

func (certificates *Certificates) ListStaleProvisioning(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
SELECT
	id,
	hostname,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	provisioning_error,
	created_at,
	updated_at
FROM
	certificates
WHERE
	%s
	AND (
		(status IN (@provisioning_status, @renewing_status) AND updated_at < CURRENT_TIMESTAMP - INTERVAL '10 minutes')
		OR
		(ssl_retry_count > 0 AND ssl_last_attempt_at < CURRENT_TIMESTAMP - INTERVAL '24 hours')
	)
	AND status != @failed_status
	AND status != @active_status
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"provisioning_status": string(CertificateStatusProvisioning),
		"renewing_status":     string(CertificateStatusRenewing),
		"failed_status":       string(CertificateStatusFailed),
		"active_status":       string(CertificateStatusActive),
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query stale provisioning certificates: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Certificate])
	if err != nil {
		return fmt.Errorf("cannot collect stale provisioning certificates: %w", err)
	}

	*certificates = result

	return nil
}

func (c *Certificate) LoadNextForProvisioningForUpdateSkipLocked(
	ctx context.Context,
	tx pg.Tx,
) error {
	// PROVISIONING rows with an open order poll every ~30s. Pending/Renewing
	// rows (and provisioning rows without an order) use exponential backoff
	// from ssl_last_attempt_at: 15m * 2^min(retry,5). Ordinary failures only
	// reach retry counts 0–2 before FAILED; higher exponents are unused by the
	// current failure budget but keep the SQL ceiling defensive.
	q := `
SELECT
	id,
	hostname,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	provisioning_error,
	created_at,
	updated_at
FROM
	certificates
WHERE
	status = ANY(@statuses)
	AND (
		ssl_last_attempt_at IS NULL
		OR (
			status = @provisioning_status
			AND http_order_url IS NOT NULL
			AND ssl_last_attempt_at < CURRENT_TIMESTAMP - INTERVAL '30 seconds'
		)
		OR (
			NOT (
				status = @provisioning_status
				AND http_order_url IS NOT NULL
			)
			AND ssl_last_attempt_at < CURRENT_TIMESTAMP - (
				INTERVAL '15 minutes' * (POWER(2, LEAST(ssl_retry_count, 5))::int)
			)
		)
	)
ORDER BY
	updated_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`

	rows, err := tx.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"statuses": []string{
				string(CertificateStatusPending),
				string(CertificateStatusProvisioning),
				string(CertificateStatusRenewing),
			},
			"provisioning_status": string(CertificateStatusProvisioning),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot query certificate provisioning queue: %w", err)
	}

	certificate, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Certificate])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect certificate: %w", err)
	}

	*c = certificate

	return nil
}

func (c *Certificate) LoadNextForRenewalForUpdateSkipLocked(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
SELECT
	id,
	hostname,
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	status,
	ssl_expires_at,
	ssl_retry_count,
	ssl_last_attempt_at,
	provisioning_error,
	created_at,
	updated_at
FROM
	certificates
WHERE
	status = @status
	AND ssl_expires_at <= CURRENT_TIMESTAMP + INTERVAL '30 days'
ORDER BY
	ssl_expires_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`

	rows, err := tx.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{"status": string(CertificateStatusActive)},
	)
	if err != nil {
		return fmt.Errorf("cannot query certificate renewal queue: %w", err)
	}

	certificate, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Certificate])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect certificate: %w", err)
	}

	*c = certificate

	return nil
}
