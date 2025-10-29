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
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type SAMLConfiguration struct {
	ID                      gid.GID               `db:"id"`
	OrganizationID          gid.GID               `db:"organization_id"`
	EmailDomain             string                `db:"email_domain"`
	Enabled                 bool                  `db:"enabled"`
	EnforcementPolicy       SAMLEnforcementPolicy `db:"enforcement_policy"`
	IdPEntityID             string                `db:"idp_entity_id"`
	IdPSsoURL               string                `db:"idp_sso_url"`
	IdPCertificate          string                `db:"idp_certificate"`
	IdPMetadataURL          *string               `db:"idp_metadata_url"`
	AttributeEmail          string                `db:"attribute_email"`
	AttributeFirstname      string                `db:"attribute_firstname"`
	AttributeLastname       string                `db:"attribute_lastname"`
	AttributeRole           string                `db:"attribute_role"`
	DefaultRole             string                `db:"default_role"`
	AutoSignupEnabled       bool                  `db:"auto_signup_enabled"`
	DomainVerified          bool                  `db:"domain_verified"`
	DomainVerificationToken *string               `db:"domain_verification_token"`
	DomainVerifiedAt        *time.Time            `db:"domain_verified_at"`
	CreatedAt               time.Time             `db:"created_at"`
	UpdatedAt               time.Time             `db:"updated_at"`
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
    enabled,
    enforcement_policy,
    idp_entity_id,
    idp_sso_url,
    idp_certificate,
    idp_metadata_url,
    attribute_email,
    attribute_firstname,
    attribute_lastname,
    attribute_role,
    default_role,
    auto_signup_enabled,
    domain_verified,
    domain_verification_token,
    domain_verified_at,
    created_at,
    updated_at
FROM
    auth_saml_configurations
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
		return fmt.Errorf("cannot query auth_saml_configurations: %w", err)
	}

	config, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SAMLConfiguration])
	if err != nil {
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
    enabled,
    enforcement_policy,
    idp_entity_id,
    idp_sso_url,
    idp_certificate,
    idp_metadata_url,
    attribute_email,
    attribute_firstname,
    attribute_lastname,
    attribute_role,
    default_role,
    auto_signup_enabled,
    domain_verified,
    domain_verification_token,
    domain_verified_at,
    created_at,
    updated_at
FROM
    auth_saml_configurations
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
		return fmt.Errorf("cannot query auth_saml_configurations: %w", err)
	}

	config, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SAMLConfiguration])
	if err != nil {
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
INSERT INTO auth_saml_configurations (
    id,
    tenant_id,
    organization_id,
    email_domain,
    enabled,
    enforcement_policy,
    idp_entity_id,
    idp_sso_url,
    idp_certificate,
    idp_metadata_url,
    attribute_email,
    attribute_firstname,
    attribute_lastname,
    attribute_role,
    default_role,
    auto_signup_enabled,
    domain_verified,
    domain_verification_token,
    domain_verified_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @email_domain,
    @enabled,
    @enforcement_policy,
    @idp_entity_id,
    @idp_sso_url,
    @idp_certificate,
    @idp_metadata_url,
    @attribute_email,
    @attribute_firstname,
    @attribute_lastname,
    @attribute_role,
    @default_role,
    @auto_signup_enabled,
    @domain_verified,
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
		"enabled":                   s.Enabled,
		"enforcement_policy":        s.EnforcementPolicy,
		"idp_entity_id":             s.IdPEntityID,
		"idp_sso_url":               s.IdPSsoURL,
		"idp_certificate":           s.IdPCertificate,
		"idp_metadata_url":          s.IdPMetadataURL,
		"attribute_email":           s.AttributeEmail,
		"attribute_firstname":       s.AttributeFirstname,
		"attribute_lastname":        s.AttributeLastname,
		"attribute_role":            s.AttributeRole,
		"default_role":              s.DefaultRole,
		"auto_signup_enabled":       s.AutoSignupEnabled,
		"domain_verified":           s.DomainVerified,
		"domain_verification_token": s.DomainVerificationToken,
		"domain_verified_at":        s.DomainVerifiedAt,
		"created_at":                s.CreatedAt,
		"updated_at":                s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
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
UPDATE auth_saml_configurations
SET
    enabled = @enabled,
    enforcement_policy = @enforcement_policy,
    idp_entity_id = @idp_entity_id,
    idp_sso_url = @idp_sso_url,
    idp_certificate = @idp_certificate,
    idp_metadata_url = @idp_metadata_url,
    attribute_email = @attribute_email,
    attribute_firstname = @attribute_firstname,
    attribute_lastname = @attribute_lastname,
    attribute_role = @attribute_role,
    default_role = @default_role,
    auto_signup_enabled = @auto_signup_enabled,
    domain_verified = @domain_verified,
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
		"enabled":                   s.Enabled,
		"enforcement_policy":        s.EnforcementPolicy,
		"idp_entity_id":             s.IdPEntityID,
		"idp_sso_url":               s.IdPSsoURL,
		"idp_certificate":           s.IdPCertificate,
		"idp_metadata_url":          s.IdPMetadataURL,
		"attribute_email":           s.AttributeEmail,
		"attribute_firstname":       s.AttributeFirstname,
		"attribute_lastname":        s.AttributeLastname,
		"attribute_role":            s.AttributeRole,
		"default_role":              s.DefaultRole,
		"auto_signup_enabled":       s.AutoSignupEnabled,
		"domain_verified":           s.DomainVerified,
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
DELETE FROM auth_saml_configurations
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

func LoadSAMLConfigurationsByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) ([]*SAMLConfiguration, error) {
	q := `
SELECT
    id,
    organization_id,
    email_domain,
    enabled,
    enforcement_policy,
    idp_entity_id,
    idp_sso_url,
    idp_certificate,
    idp_metadata_url,
    attribute_email,
    attribute_firstname,
    attribute_lastname,
    attribute_role,
    default_role,
    auto_signup_enabled,
    domain_verified,
    domain_verification_token,
    domain_verified_at,
    created_at,
    updated_at
FROM
    auth_saml_configurations
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
		return nil, fmt.Errorf("cannot query auth_saml_configurations: %w", err)
	}

	configs, err := pgx.CollectRows(rows, pgx.RowToStructByName[SAMLConfiguration])
	if err != nil {
		return nil, fmt.Errorf("cannot collect saml_configurations: %w", err)
	}

	result := make([]*SAMLConfiguration, len(configs))
	for i := range configs {
		result[i] = &configs[i]
	}

	return result, nil
}

// LoadAllEnabledSAMLConfigurationsByEmailDomain loads all enabled SAML configurations for a given email domain
// This is used for SSO login detection when multiple organizations may have SAML configured for the same domain
func LoadAllEnabledSAMLConfigurationsByEmailDomain(
	ctx context.Context,
	conn pg.Conn,
	emailDomain string,
) ([]*SAMLConfiguration, error) {
	q := `
SELECT
    id,
    organization_id,
    email_domain,
    enabled,
    enforcement_policy,
    idp_entity_id,
    idp_sso_url,
    idp_certificate,
    idp_metadata_url,
    attribute_email,
    attribute_firstname,
    attribute_lastname,
    attribute_role,
    default_role,
    auto_signup_enabled,
    domain_verified,
    domain_verification_token,
    domain_verified_at,
    created_at,
    updated_at
FROM
    auth_saml_configurations
WHERE
    email_domain = $1
    AND enabled = true
    AND domain_verified = true
ORDER BY created_at ASC;
`

	rows, err := conn.Query(ctx, q, emailDomain)
	if err != nil {
		return nil, fmt.Errorf("cannot query auth_saml_configurations: %w", err)
	}

	configs, err := pgx.CollectRows(rows, pgx.RowToStructByName[SAMLConfiguration])
	if err != nil {
		return nil, fmt.Errorf("cannot collect saml_configurations: %w", err)
	}

	result := make([]*SAMLConfiguration, len(configs))
	for i := range configs {
		result[i] = &configs[i]
	}

	return result, nil
}

