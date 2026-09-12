// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	ServiceAccountCredential struct {
		ID               gid.GID      `db:"id"`
		OrganizationID   gid.GID      `db:"organization_id"`
		ServiceAccountID gid.GID      `db:"service_account_id"`
		Name             string       `db:"name"`
		HashedToken      []byte       `db:"hashed_token"`
		Scopes           OAuth2Scopes `db:"scopes"`
		ExpiresAt        time.Time    `db:"expires_at"`
		LastUsedAt       *time.Time   `db:"last_used_at"`
		RevokedAt        *time.Time   `db:"revoked_at"`
		CreatedAt        time.Time    `db:"created_at"`
		UpdatedAt        time.Time    `db:"updated_at"`
	}

	ServiceAccountCredentials []*ServiceAccountCredential
)

func (c *ServiceAccountCredential) CursorKey(orderBy ServiceAccountCredentialOrderField) page.CursorKey {
	switch orderBy {
	case ServiceAccountCredentialOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	case ServiceAccountCredentialOrderFieldName:
		return page.NewCursorKey(c.ID, c.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (c *ServiceAccountCredential) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT id, organization_id
FROM iam_service_account_credentials
WHERE id = ANY(@resource_ids::text[])
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"resource_ids": resourceIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot query service account credential authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID, len(resourceIDs))
	for rows.Next() {
		var id, organizationID gid.GID
		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan service account credential authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{"organization_id": organizationID.String()}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate service account credential authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (c *ServiceAccountCredential) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    service_account_id,
    name,
    hashed_token,
    scopes,
    expires_at,
    last_used_at,
    revoked_at,
    created_at,
    updated_at
FROM iam_service_account_credentials
WHERE %s
    AND id = @id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	return c.load(ctx, conn, q, args)
}

func (c *ServiceAccountCredential) LoadByHashedToken(
	ctx context.Context,
	conn pg.Querier,
	hashedToken []byte,
) error {
	q := `
SELECT
    id,
    organization_id,
    service_account_id,
    name,
    hashed_token,
    scopes,
    expires_at,
    last_used_at,
    revoked_at,
    created_at,
    updated_at
FROM iam_service_account_credentials
WHERE hashed_token = @hashed_token
LIMIT 1;
`

	return c.load(ctx, conn, q, pgx.StrictNamedArgs{"hashed_token": hashedToken})
}

func (c *ServiceAccountCredential) LoadByHashedTokenForUpdate(
	ctx context.Context,
	conn pg.Tx,
	hashedToken []byte,
) error {
	q := `
SELECT
    id,
    organization_id,
    service_account_id,
    name,
    hashed_token,
    scopes,
    expires_at,
    last_used_at,
    revoked_at,
    created_at,
    updated_at
FROM iam_service_account_credentials
WHERE hashed_token = @hashed_token
LIMIT 1
FOR UPDATE;
`

	return c.load(ctx, conn, q, pgx.StrictNamedArgs{"hashed_token": hashedToken})
}

func (c *ServiceAccountCredential) load(
	ctx context.Context,
	conn pg.Querier,
	q string,
	args pgx.StrictNamedArgs,
) error {
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query service account credential: %w", err)
	}

	credential, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ServiceAccountCredential])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect service account credential: %w", err)
	}

	*c = credential

	return nil
}

func (cs *ServiceAccountCredentials) LoadByServiceAccountID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	serviceAccountID gid.GID,
	cursor *page.Cursor[ServiceAccountCredentialOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    service_account_id,
    name,
    hashed_token,
    scopes,
    expires_at,
    last_used_at,
    revoked_at,
    created_at,
    updated_at
FROM iam_service_account_credentials
WHERE %s
    AND service_account_id = @service_account_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"service_account_id": serviceAccountID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query service account credentials: %w", err)
	}

	credentials, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ServiceAccountCredential])
	if err != nil {
		return fmt.Errorf("cannot collect service account credentials: %w", err)
	}

	*cs = credentials

	return nil
}

func (c *ServiceAccountCredential) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO iam_service_account_credentials (
    id,
    tenant_id,
    organization_id,
    service_account_id,
    name,
    hashed_token,
    scopes,
    expires_at,
    last_used_at,
    revoked_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @service_account_id,
    @name,
    @hashed_token,
    @scopes,
    @expires_at,
    @last_used_at,
    @revoked_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                 c.ID,
		"tenant_id":          scope.GetTenantID(),
		"organization_id":    c.OrganizationID,
		"service_account_id": c.ServiceAccountID,
		"name":               c.Name,
		"hashed_token":       c.HashedToken,
		"scopes":             c.Scopes,
		"expires_at":         c.ExpiresAt,
		"last_used_at":       c.LastUsedAt,
		"revoked_at":         c.RevokedAt,
		"created_at":         c.CreatedAt,
		"updated_at":         c.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert service account credential: %w", err)
	}

	return nil
}

func (c *ServiceAccountCredential) Update(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
UPDATE iam_service_account_credentials
SET
    last_used_at = @last_used_at,
    revoked_at = @revoked_at,
    updated_at = @updated_at
WHERE %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":           c.ID,
		"last_used_at": c.LastUsedAt,
		"revoked_at":   c.RevokedAt,
		"updated_at":   c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update service account credential: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (cs *ServiceAccountCredentials) RevokeByServiceAccountID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	serviceAccountID gid.GID,
	now time.Time,
) error {
	q := `
UPDATE iam_service_account_credentials
SET
    revoked_at = @now,
    updated_at = @now
WHERE %s
    AND service_account_id = @service_account_id
    AND revoked_at IS NULL
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"service_account_id": serviceAccountID,
		"now":                now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot revoke service account credentials: %w", err)
	}

	return nil
}

func (cs *ServiceAccountCredentials) RevokeOutsideScopesByServiceAccountID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	serviceAccountID gid.GID,
	scopes OAuth2Scopes,
	now time.Time,
) error {
	q := `
UPDATE iam_service_account_credentials
SET
    revoked_at = @now,
    updated_at = @now
WHERE %s
    AND service_account_id = @service_account_id
    AND revoked_at IS NULL
    AND NOT (scopes <@ @scopes::text[])
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"service_account_id": serviceAccountID,
		"scopes":             scopes,
		"now":                now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot revoke service account credentials outside scopes: %w", err)
	}

	return nil
}
