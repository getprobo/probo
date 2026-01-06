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
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	SAMLConfiguration struct {
		ID                      gid.GID               `db:"id"`
		OrganizationID          gid.GID               `db:"organization_id"`
		EmailDomain             string                `db:"email_domain"`
		EnforcementPolicy       SAMLEnforcementPolicy `db:"enforcement_policy"`
		IdPEntityID             string                `db:"idp_entity_id"`
		IdPSsoURL               string                `db:"idp_sso_url"`
		IdPCertificate          string                `db:"idp_certificate"`
		IdPMetadataURL          *string               `db:"idp_metadata_url"`
		AttributeEmail          string                `db:"attribute_email"`
		AttributeFirstname      string                `db:"attribute_firstname"`
		AttributeLastname       string                `db:"attribute_lastname"`
		AttributeRole           string                `db:"attribute_role"`
		AutoSignupEnabled       bool                  `db:"auto_signup_enabled"`
		DomainVerificationToken *string               `db:"domain_verification_token"`
		DomainVerifiedAt        *time.Time            `db:"domain_verified_at"`
		CreatedAt               time.Time             `db:"created_at"`
		UpdatedAt               time.Time             `db:"updated_at"`
	}

	SAMLConfigurations []*SAMLConfiguration
)

func (s *SAMLConfiguration) CursorKey(orderBy SAMLConfigurationOrderField) page.CursorKey {
	switch orderBy {
	case SAMLConfigurationOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *SAMLConfiguration) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM iam_saml_configurations WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, s.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query saml configuration authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (s *SAMLConfiguration) GetIdPCertificate() (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(s.IdPCertificate))
	if block == nil {
		return nil, fmt.Errorf("cannot decode PEM block from IdP certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse X.509 certificate: %w", err)
	}

	return cert, nil
}

func (s *SAMLConfiguration) LoadByOrganizationIDAndEmailDomain(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	emailDomain string,
) error {
	q := `
SELECT
    id,
    organization_id,
    email_domain,
    enforcement_policy,
    idp_entity_id,
    idp_sso_url,
    idp_certificate,
    idp_metadata_url,
    attribute_email,
    attribute_firstname,
    attribute_lastname,
    attribute_role,
    auto_signup_enabled,
    domain_verification_token,
    domain_verified_at,
    created_at,
    updated_at
FROM
    iam_saml_configurations
WHERE
    %s
    AND organization_id = @organization_id
    AND email_domain = @email_domain
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"email_domain":    emailDomain,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_saml_configurations: %w", err)
	}

	config, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SAMLConfiguration])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect saml_configuration: %w", err)
	}

	*s = config

	return nil
}

func (s *SAMLConfiguration) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	configID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    email_domain,
    enforcement_policy,
    idp_entity_id,
    idp_sso_url,
    idp_certificate,
    idp_metadata_url,
    attribute_email,
    attribute_firstname,
    attribute_lastname,
    attribute_role,
    auto_signup_enabled,
    domain_verification_token,
    domain_verified_at,
    created_at,
    updated_at
FROM
    iam_saml_configurations
WHERE
    %s
    AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": configID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_saml_configurations: %w", err)
	}

	config, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SAMLConfiguration])
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect saml_configuration: %w", err)
	}

	*s = config

	return nil
}

func (s *SAMLConfiguration) LoadByIDForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Conn,
	configID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    email_domain,
    enforcement_policy,
    idp_entity_id,
    idp_sso_url,
    idp_certificate,
    idp_metadata_url,
    attribute_email,
    attribute_firstname,
    attribute_lastname,
    attribute_role,
    auto_signup_enabled,
    domain_verification_token,
    domain_verified_at,
    created_at,
    updated_at
FROM
    iam_saml_configurations
WHERE
    id = @id
FOR UPDATE SKIP LOCKED;
`

	args := pgx.StrictNamedArgs{"id": configID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_saml_configurations: %w", err)
	}

	config, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SAMLConfiguration])
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect saml_configuration: %w", err)
	}

	*s = config

	return nil
}

func (s *SAMLConfiguration) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO iam_saml_configurations (
    id,
    tenant_id,
    organization_id,
    email_domain,
    enforcement_policy,
    idp_entity_id,
    idp_sso_url,
    idp_certificate,
    idp_metadata_url,
    attribute_email,
    attribute_firstname,
    attribute_lastname,
    attribute_role,
    auto_signup_enabled,
    domain_verification_token,
    domain_verified_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @email_domain,
    @enforcement_policy,
    @idp_entity_id,
    @idp_sso_url,
    @idp_certificate,
    @idp_metadata_url,
    @attribute_email,
    @attribute_firstname,
    @attribute_lastname,
    @attribute_role,
    @auto_signup_enabled,
    @domain_verification_token,
    @domain_verified_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                        s.ID,
		"tenant_id":                 scope.GetTenantID(),
		"organization_id":           s.OrganizationID,
		"email_domain":              s.EmailDomain,
		"enforcement_policy":        s.EnforcementPolicy,
		"idp_entity_id":             s.IdPEntityID,
		"idp_sso_url":               s.IdPSsoURL,
		"idp_certificate":           s.IdPCertificate,
		"idp_metadata_url":          s.IdPMetadataURL,
		"attribute_email":           s.AttributeEmail,
		"attribute_firstname":       s.AttributeFirstname,
		"attribute_lastname":        s.AttributeLastname,
		"attribute_role":            s.AttributeRole,
		"auto_signup_enabled":       s.AutoSignupEnabled,
		"domain_verification_token": s.DomainVerificationToken,
		"domain_verified_at":        s.DomainVerifiedAt,
		"created_at":                s.CreatedAt,
		"updated_at":                s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_saml_config_domain_org_unique" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert saml_configuration: %w", err)
	}

	return nil
}

func (s *SAMLConfiguration) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE iam_saml_configurations
SET
    enforcement_policy = @enforcement_policy,
    idp_entity_id = @idp_entity_id,
    idp_sso_url = @idp_sso_url,
    idp_certificate = @idp_certificate,
    idp_metadata_url = @idp_metadata_url,
    attribute_email = @attribute_email,
    attribute_firstname = @attribute_firstname,
    attribute_lastname = @attribute_lastname,
    attribute_role = @attribute_role,
    auto_signup_enabled = @auto_signup_enabled,
    domain_verification_token = @domain_verification_token,
    domain_verified_at = @domain_verified_at,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                        s.ID,
		"enforcement_policy":        s.EnforcementPolicy,
		"idp_entity_id":             s.IdPEntityID,
		"idp_sso_url":               s.IdPSsoURL,
		"idp_certificate":           s.IdPCertificate,
		"idp_metadata_url":          s.IdPMetadataURL,
		"attribute_email":           s.AttributeEmail,
		"attribute_firstname":       s.AttributeFirstname,
		"attribute_lastname":        s.AttributeLastname,
		"attribute_role":            s.AttributeRole,
		"auto_signup_enabled":       s.AutoSignupEnabled,
		"domain_verification_token": s.DomainVerificationToken,
		"domain_verified_at":        s.DomainVerifiedAt,
		"updated_at":                s.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update saml_configuration: %w", err)
	}

	return nil
}

func (s *SAMLConfiguration) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM iam_saml_configurations
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": s.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete saml_configuration: %w", err)
	}

	return nil
}

func (s *SAMLConfigurations) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    email_domain,
    enforcement_policy,
    idp_entity_id,
    idp_sso_url,
    idp_certificate,
    idp_metadata_url,
    attribute_email,
    attribute_firstname,
    attribute_lastname,
    attribute_role,
    auto_signup_enabled,
    domain_verification_token,
    domain_verified_at,
    created_at,
    updated_at
FROM
    iam_saml_configurations
WHERE
    %s
    AND organization_id = @organization_id
ORDER BY email_domain ASC;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_saml_configurations: %w", err)
	}

	samlConfigurations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SAMLConfiguration])
	if err != nil {
		return fmt.Errorf("cannot collect saml_configurations: %w", err)
	}

	*s = samlConfigurations

	return nil
}

func (s *SAMLConfigurations) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    iam_saml_configurations
WHERE
    %s
    AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot query iam_saml_configurations: %w", err)
	}

	var count int
	err = rows.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot collect count: %w", err)
	}

	return count, nil
}

func (s *SAMLConfigurations) LoadVerifiedByEmailDomain(ctx context.Context, conn pg.Conn, emailDomain string) error {
	q := `
SELECT
    id,
    organization_id,
    email_domain,
    enforcement_policy,
    idp_entity_id,
    idp_sso_url,
    idp_certificate,
    idp_metadata_url,
    attribute_email,
    attribute_firstname,
    attribute_lastname,
    attribute_role,
    auto_signup_enabled,
    domain_verification_token,
    domain_verified_at,
    created_at,
    updated_at
FROM
    iam_saml_configurations
WHERE
    email_domain = @email_domain
    AND domain_verified_at IS NOT NULL
ORDER BY email_domain ASC;
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"email_domain": emailDomain})
	if err != nil {
		return fmt.Errorf("cannot query iam_saml_configurations: %w", err)
	}

	samlConfigurations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SAMLConfiguration])
	if err != nil {
		return fmt.Errorf("cannot collect saml_configurations: %w", err)
	}

	*s = samlConfigurations

	return nil
}

func (s *SAMLConfigurations) CountVerifiedByEmailDomain(
	ctx context.Context,
	conn pg.Conn,
	emailDomain string,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    iam_saml_configurations
WHERE
    email_domain = @email_domain
    AND domain_verified_at IS NOT NULL
`

	row := conn.QueryRow(ctx, q, pgx.StrictNamedArgs{"email_domain": emailDomain})
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count SAML configurations: %w", err)
	}

	return count, nil
}

func (s *SAMLConfigurations) LoadUnverified(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
SELECT
    id,
    organization_id,
    email_domain,
    enforcement_policy,
    idp_entity_id,
    idp_sso_url,
    idp_certificate,
    idp_metadata_url,
    attribute_email,
    attribute_firstname,
    attribute_lastname,
    attribute_role,
    auto_signup_enabled,
    domain_verification_token,
    domain_verified_at,
    created_at,
    updated_at
FROM
    iam_saml_configurations
WHERE
    domain_verified_at IS NULL
    AND domain_verification_token IS NOT NULL
ORDER BY created_at ASC
LIMIT 100;
`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query unverified iam_saml_configurations: %w", err)
	}

	configs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SAMLConfiguration])
	if err != nil {
		return fmt.Errorf("cannot collect unverified saml_configurations: %w", err)
	}

	*s = configs

	return nil
}
