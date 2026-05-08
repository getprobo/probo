// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	// CloudAccount is the tenant-scoped, polymorphic record holding
	// the (provider, credential_kind, scope, encrypted_credentials,
	// status) tuple for a customer's cloud account connection.
	//
	// DecryptedCredentials holds the cleartext JSON envelope after
	// LoadByID decrypts EncryptedCredentials with the configured
	// EncryptionKey. It is never persisted directly; callers must
	// not log or otherwise expose it.
	CloudAccount struct {
		ID                       gid.GID                    `db:"id"`
		OrganizationID           gid.GID                    `db:"organization_id"`
		Label                    string                     `db:"label"`
		Provider                 CloudAccountProvider       `db:"provider"`
		CredentialKind           CloudAccountCredentialKind `db:"credential_kind"`
		ScopeKind                CloudAccountScopeKind      `db:"scope_kind"`
		ScopeIdentifier          string                     `db:"scope_identifier"`
		EnabledAuditModules      []CloudAccountAuditModule  `db:"enabled_audit_modules"`
		Status                   CloudAccountStatus         `db:"status"`
		ConsecutiveProbeFailures int                        `db:"consecutive_probe_failures"`
		FirstProbeFailureAt      *time.Time                 `db:"first_probe_failure_at"`
		EncryptedCredentials     []byte                     `db:"encrypted_credentials"`
		DecryptedCredentials     []byte                     `db:"-"`
		ExternalID               *string                    `db:"external_id"`
		TemplateSHA256           *string                    `db:"template_sha256"`
		LastProbeAt              *time.Time                 `db:"last_probe_at"`
		LastProbeError           *string                    `db:"last_probe_error"`
		LastVerifiedAt           *time.Time                 `db:"last_verified_at"`
		CreatedAt                time.Time                  `db:"created_at"`
		UpdatedAt                time.Time                  `db:"updated_at"`
	}

	CloudAccounts []*CloudAccount
)

func (c *CloudAccount) CursorKey(orderBy CloudAccountOrderField) page.CursorKey {
	switch orderBy {
	case CloudAccountOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	case CloudAccountOrderFieldStatus:
		return page.NewCursorKey(c.ID, c.Status.String())
	case CloudAccountOrderFieldProvider:
		return page.NewCursorKey(c.ID, c.Provider.String())
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (c *CloudAccount) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM cloud_accounts WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, c.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot query cloud account authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

// LoadByID loads a cloud account by ID and decrypts its credentials
// envelope into DecryptedCredentials. Returns ErrResourceNotFound when
// no row matches in scope.
func (c *CloudAccount) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cloudAccountID gid.GID,
	encryptionKey cipher.EncryptionKey,
) error {
	if err := c.loadMetadataByID(ctx, conn, scope, cloudAccountID); err != nil {
		return err
	}

	if len(c.EncryptedCredentials) > 0 {
		decrypted, err := cipher.Decrypt(c.EncryptedCredentials, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt cloud account credentials: %w", err)
		}

		c.DecryptedCredentials = decrypted
	}

	return nil
}

// LoadMetadataByID loads cloud account metadata without decrypting
// the credentials envelope. Use when only the row metadata
// (provider, scope, status, timestamps) is required.
func (c *CloudAccount) LoadMetadataByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cloudAccountID gid.GID,
) error {
	return c.loadMetadataByID(ctx, conn, scope, cloudAccountID)
}

func (c *CloudAccount) loadMetadataByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cloudAccountID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    label,
    provider,
    credential_kind,
    scope_kind,
    scope_identifier,
    enabled_audit_modules,
    status,
    consecutive_probe_failures,
    first_probe_failure_at,
    encrypted_credentials,
    external_id,
    template_sha256,
    last_probe_at,
    last_probe_error,
    last_verified_at,
    created_at,
    updated_at
FROM
    cloud_accounts
WHERE
    %s
    AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": cloudAccountID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cloud accounts: %w", err)
	}

	loaded, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CloudAccount])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect cloud account row: %w", err)
	}

	*c = loaded

	return nil
}

// LoadByOrganizationID loads paginated, filtered cloud accounts in scope.
// Credentials are NOT decrypted; callers that need the cleartext
// envelope must call LoadByID for each row they care about.
func (c *CloudAccounts) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[CloudAccountOrderField],
	filter *CloudAccountFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
    label,
    provider,
    credential_kind,
    scope_kind,
    scope_identifier,
    enabled_audit_modules,
    status,
    consecutive_probe_failures,
    first_probe_failure_at,
    encrypted_credentials,
    external_id,
    template_sha256,
    last_probe_at,
    last_probe_error,
    last_verified_at,
    created_at,
    updated_at
FROM
    cloud_accounts
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cloud accounts: %w", err)
	}

	loaded, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CloudAccount])
	if err != nil {
		return fmt.Errorf("cannot collect cloud accounts: %w", err)
	}

	*c = loaded

	return nil
}

// LoadNextStaleForUpdateSkipLocked claims the next cloud account due
// for re-probing using FOR UPDATE SKIP LOCKED so concurrent worker
// shards never claim the same row. Returns ErrResourceNotFound when
// no row is currently due. Callers convert that sentinel to
// worker.ErrNoTask at the worker boundary (do NOT return worker
// sentinels from coredata).
func (c *CloudAccount) LoadNextStaleForUpdateSkipLocked(
	ctx context.Context,
	tx pg.Tx,
	staleAfter time.Duration,
) error {
	q := `
SELECT
    id,
    organization_id,
    label,
    provider,
    credential_kind,
    scope_kind,
    scope_identifier,
    enabled_audit_modules,
    status,
    consecutive_probe_failures,
    first_probe_failure_at,
    encrypted_credentials,
    external_id,
    template_sha256,
    last_probe_at,
    last_probe_error,
    last_verified_at,
    created_at,
    updated_at
FROM
    cloud_accounts
WHERE
    last_probe_at IS NULL
    OR last_probe_at < (now() - @stale_after::interval)
ORDER BY
    last_probe_at NULLS FIRST
LIMIT 1
FOR UPDATE SKIP LOCKED;
`

	args := pgx.StrictNamedArgs{
		"stale_after": fmt.Sprintf("%d milliseconds", staleAfter.Milliseconds()),
	}

	rows, err := tx.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query next stale cloud account: %w", err)
	}

	loaded, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CloudAccount])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect cloud account row: %w", err)
	}

	*c = loaded

	return nil
}

// Insert encrypts c.DecryptedCredentials with the configured key,
// persists the row in a single statement, and updates
// c.EncryptedCredentials in-place so callers can re-read the
// envelope without decrypting again.
func (c *CloudAccount) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
) error {
	q := `
INSERT INTO cloud_accounts (
    id,
    tenant_id,
    organization_id,
    label,
    provider,
    credential_kind,
    scope_kind,
    scope_identifier,
    enabled_audit_modules,
    status,
    consecutive_probe_failures,
    first_probe_failure_at,
    encrypted_credentials,
    external_id,
    template_sha256,
    last_probe_at,
    last_probe_error,
    last_verified_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @label,
    @provider,
    @credential_kind,
    @scope_kind,
    @scope_identifier,
    @enabled_audit_modules,
    @status,
    @consecutive_probe_failures,
    @first_probe_failure_at,
    @encrypted_credentials,
    @external_id,
    @template_sha256,
    @last_probe_at,
    @last_probe_error,
    @last_verified_at,
    @created_at,
    @updated_at
)
`

	if len(c.DecryptedCredentials) == 0 {
		return fmt.Errorf("cannot insert cloud account: decrypted credentials are empty")
	}

	encrypted, err := cipher.Encrypt(c.DecryptedCredentials, encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt cloud account credentials: %w", err)
	}

	args := pgx.StrictNamedArgs{
		"id":                         c.ID,
		"tenant_id":                  scope.GetTenantID(),
		"organization_id":            c.OrganizationID,
		"label":                      c.Label,
		"provider":                   c.Provider,
		"credential_kind":            c.CredentialKind,
		"scope_kind":                 c.ScopeKind,
		"scope_identifier":           c.ScopeIdentifier,
		"enabled_audit_modules":      c.EnabledAuditModules,
		"status":                     c.Status,
		"consecutive_probe_failures": c.ConsecutiveProbeFailures,
		"first_probe_failure_at":     c.FirstProbeFailureAt,
		"encrypted_credentials":      encrypted,
		"external_id":                c.ExternalID,
		"template_sha256":            c.TemplateSHA256,
		"last_probe_at":              c.LastProbeAt,
		"last_probe_error":           c.LastProbeError,
		"last_verified_at":           c.LastVerifiedAt,
		"created_at":                 c.CreatedAt,
		"updated_at":                 c.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert cloud account: %w", err)
	}

	c.EncryptedCredentials = encrypted

	return nil
}

// Update writes back the mutable columns. When DecryptedCredentials
// is non-empty the cleartext is re-encrypted; otherwise the existing
// EncryptedCredentials value is preserved untouched. Returns
// ErrResourceNotFound when no row matches in scope.
func (c *CloudAccount) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
) error {
	q := `
UPDATE cloud_accounts
SET
    label = @label,
    credential_kind = @credential_kind,
    scope_kind = @scope_kind,
    scope_identifier = @scope_identifier,
    enabled_audit_modules = @enabled_audit_modules,
    status = @status,
    consecutive_probe_failures = @consecutive_probe_failures,
    first_probe_failure_at = @first_probe_failure_at,
    encrypted_credentials = @encrypted_credentials,
    external_id = @external_id,
    template_sha256 = @template_sha256,
    last_probe_at = @last_probe_at,
    last_probe_error = @last_probe_error,
    last_verified_at = @last_verified_at,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	encrypted := c.EncryptedCredentials
	if len(c.DecryptedCredentials) > 0 {
		next, err := cipher.Encrypt(c.DecryptedCredentials, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot encrypt cloud account credentials: %w", err)
		}

		encrypted = next
	}

	args := pgx.StrictNamedArgs{
		"id":                         c.ID,
		"label":                      c.Label,
		"credential_kind":            c.CredentialKind,
		"scope_kind":                 c.ScopeKind,
		"scope_identifier":           c.ScopeIdentifier,
		"enabled_audit_modules":      c.EnabledAuditModules,
		"status":                     c.Status,
		"consecutive_probe_failures": c.ConsecutiveProbeFailures,
		"first_probe_failure_at":     c.FirstProbeFailureAt,
		"encrypted_credentials":      encrypted,
		"external_id":                c.ExternalID,
		"template_sha256":            c.TemplateSHA256,
		"last_probe_at":              c.LastProbeAt,
		"last_probe_error":           c.LastProbeError,
		"last_verified_at":           c.LastVerifiedAt,
		"updated_at":                 c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update cloud account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	c.EncryptedCredentials = encrypted

	return nil
}

// Delete removes the cloud account row in scope. Returns
// ErrResourceNotFound when no row matches.
func (c *CloudAccount) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM cloud_accounts
WHERE %s AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete cloud account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
