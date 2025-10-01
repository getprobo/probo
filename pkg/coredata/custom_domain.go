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
		ID                     gid.GID               `db:"id"`
		OrganizationID         gid.GID               `db:"organization_id"`
		Domain                 string                `db:"domain"`
		HTTPChallengeToken     *string               `db:"http_challenge_token"`
		HTTPChallengeKeyAuth   *string               `db:"http_challenge_key_auth"`
		HTTPChallengeURL       *string               `db:"http_challenge_url"`
		HTTPOrderURL           *string               `db:"http_order_url"`
		SSLCertificate         *tls.Certificate      `db:"-"`
		SSLCertificatePEM      []byte                `db:"ssl_certificate"`
		SSLPrivateKeyPEM       []byte                `db:"-"`
		EncryptedSSLPrivateKey []byte                `db:"encrypted_ssl_private_key"`
		SSLCertificateChain    *string               `db:"ssl_certificate_chain"`
		SSLStatus              CustomDomainSSLStatus `db:"ssl_status"`
		SSLExpiresAt           *time.Time            `db:"ssl_expires_at"`
		IsActive               bool                  `db:"is_active"`
		CreatedAt              time.Time             `db:"created_at"`
		UpdatedAt              time.Time             `db:"updated_at"`
	}

	CustomDomains []*CustomDomain
)

func NewCustomDomain(orgID gid.GID, domain string) *CustomDomain {
	now := time.Now()
	return &CustomDomain{
		ID:             gid.New(orgID.TenantID(), CustomDomainEntityType),
		OrganizationID: orgID,
		SSLStatus:      CustomDomainSSLStatusPending,
		Domain:         domain,
		IsActive:       false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
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
	is_active,
	created_at,
	updated_at
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

	// Decrypt SSL private key
	if len(cd.EncryptedSSLPrivateKey) > 0 {
		decrypted, err := cipher.Decrypt(cd.EncryptedSSLPrivateKey, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt SSL private key: %w", err)
		}
		cd.SSLPrivateKeyPEM = decrypted
	}

	// Parse certificate and key into tls.Certificate if both are present
	if len(cd.SSLCertificatePEM) > 0 && len(cd.SSLPrivateKeyPEM) > 0 {
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

func (cd *CustomDomain) LoadByIDForUpdate(
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
	is_active,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND id = @domain_id
LIMIT 1
FOR UPDATE
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"domain_id": domainID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query custom domain for update: %w", err)
	}

	customDomain, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CustomDomain])
	if err != nil {
		return fmt.Errorf("cannot collect custom domain: %w", err)
	}

	*cd = customDomain

	// Decrypt SSL private key
	if len(cd.EncryptedSSLPrivateKey) > 0 {
		decrypted, err := cipher.Decrypt(cd.EncryptedSSLPrivateKey, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt SSL private key: %w", err)
		}
		cd.SSLPrivateKeyPEM = decrypted
	}

	// Parse certificate and key into tls.Certificate if both are present
	if len(cd.SSLCertificatePEM) > 0 && len(cd.SSLPrivateKeyPEM) > 0 {
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
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	is_active,
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
		return fmt.Errorf("cannot collect custom domain: %w", err)
	}

	*cd = customDomain

	// Decrypt SSL private key
	if len(cd.EncryptedSSLPrivateKey) > 0 {
		decrypted, err := cipher.Decrypt(cd.EncryptedSSLPrivateKey, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt SSL private key: %w", err)
		}
		cd.SSLPrivateKeyPEM = decrypted
	}

	// Parse certificate and key into tls.Certificate if both are present
	if len(cd.SSLCertificatePEM) > 0 && len(cd.SSLPrivateKeyPEM) > 0 {
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
) error {
	var encryptedKey []byte
	if len(cd.SSLPrivateKeyPEM) > 0 {
		var err error
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
	http_challenge_token,
	http_challenge_key_auth,
	http_challenge_url,
	http_order_url,
	ssl_certificate,
	encrypted_ssl_private_key,
	ssl_certificate_chain,
	ssl_status,
	ssl_expires_at,
	is_active,
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
	@is_active,
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
		"is_active":                 cd.IsActive,
		"created_at":                cd.CreatedAt,
		"updated_at":                cd.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
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
	http_challenge_token = @http_challenge_token,
	http_challenge_key_auth = @http_challenge_key_auth,
	http_challenge_url = @http_challenge_url,
	http_order_url = @http_order_url,
	ssl_certificate = @ssl_certificate,
	encrypted_ssl_private_key = @encrypted_ssl_private_key,
	ssl_certificate_chain = @ssl_certificate_chain,
	ssl_status = @ssl_status,
	ssl_expires_at = @ssl_expires_at,
	is_active = @is_active,
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
		"is_active":                 cd.IsActive,
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

func (domains *CustomDomains) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
	orgID gid.GID,
	cursor *page.Cursor[CustomDomainOrderField],
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
	is_active,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": orgID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query custom domains: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CustomDomain])
	if err != nil {
		return fmt.Errorf("cannot collect custom domains: %w", err)
	}

	for _, cd := range result {
		// Decrypt SSL private key
		if len(cd.EncryptedSSLPrivateKey) > 0 {
			decrypted, err := cipher.Decrypt(cd.EncryptedSSLPrivateKey, encryptionKey)
			if err != nil {
				return fmt.Errorf("cannot decrypt SSL private key: %w", err)
			}
			cd.SSLPrivateKeyPEM = decrypted
		}

		// Parse certificate and key into tls.Certificate if both are present
		if len(cd.SSLCertificatePEM) > 0 && len(cd.SSLPrivateKeyPEM) > 0 {
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
	}

	*domains = result
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
	is_active,
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

	// No need to decrypt anything for challenge validation
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
	is_active,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND ssl_status = 'ACTIVE'
	AND ssl_expires_at IS NOT NULL
	AND ssl_expires_at <= CURRENT_TIMESTAMP + INTERVAL '30 days'
ORDER BY
	ssl_expires_at ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{}
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
	is_active,
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
	is_active,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND ssl_status = 'ACTIVE'
	AND ssl_certificate IS NOT NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query active certificates: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CustomDomain])
	if err != nil {
		return fmt.Errorf("cannot collect custom domains: %w", err)
	}

	for _, cd := range result {
		// Decrypt SSL private key
		if len(cd.EncryptedSSLPrivateKey) > 0 {
			decrypted, err := cipher.Decrypt(cd.EncryptedSSLPrivateKey, encryptionKey)
			if err != nil {
				return fmt.Errorf("cannot decrypt SSL private key: %w", err)
			}
			cd.SSLPrivateKeyPEM = decrypted
		}
	}

	*domains = result
	return nil
}
