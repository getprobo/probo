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
	ServiceAccount struct {
		ID             gid.GID      `db:"id"`
		OrganizationID gid.GID      `db:"organization_id"`
		Name           string       `db:"name"`
		Description    *string      `db:"description"`
		Scopes         OAuth2Scopes `db:"scopes"`
		DisabledAt     *time.Time   `db:"disabled_at"`
		DeletedAt      *time.Time   `db:"deleted_at"`
		CreatedAt      time.Time    `db:"created_at"`
		UpdatedAt      time.Time    `db:"updated_at"`
	}

	ServiceAccounts []*ServiceAccount
)

func (s *ServiceAccount) CursorKey(orderBy ServiceAccountOrderField) page.CursorKey {
	switch orderBy {
	case ServiceAccountOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	case ServiceAccountOrderFieldName:
		return page.NewCursorKey(s.ID, s.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *ServiceAccount) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT id, organization_id, disabled_at IS NOT NULL
FROM iam_service_accounts
WHERE id = ANY(@resource_ids::text[])
    AND deleted_at IS NULL
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"resource_ids": resourceIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot query service account authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID, len(resourceIDs))
	for rows.Next() {
		var (
			id             gid.GID
			organizationID gid.GID
			disabled       bool
		)
		if err := rows.Scan(&id, &organizationID, &disabled); err != nil {
			return nil, fmt.Errorf("cannot scan service account authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"disabled":        fmt.Sprintf("%t", disabled),
			"organization_id": organizationID.String(),
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate service account authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (s *ServiceAccount) LoadByID(ctx context.Context, conn pg.Querier, scope Scoper, id gid.GID) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    description,
    scopes,
    disabled_at,
    deleted_at,
    created_at,
    updated_at
FROM iam_service_accounts
WHERE %s
    AND id = @id
    AND deleted_at IS NULL
LIMIT 1;
`

	return s.loadByID(ctx, conn, scope, id, q)
}

func (s *ServiceAccount) LoadByIDForUpdate(ctx context.Context, conn pg.Tx, scope Scoper, id gid.GID) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    description,
    scopes,
    disabled_at,
    deleted_at,
    created_at,
    updated_at
FROM iam_service_accounts
WHERE %s
    AND id = @id
    AND deleted_at IS NULL
LIMIT 1
FOR UPDATE;
`

	return s.loadByID(ctx, conn, scope, id, q)
}

func (s *ServiceAccount) loadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
	q string,
) error {
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query service account: %w", err)
	}

	account, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ServiceAccount])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect service account: %w", err)
	}

	*s = account

	return nil
}

func (ss *ServiceAccounts) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ServiceAccountOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    description,
    scopes,
    disabled_at,
    deleted_at,
    created_at,
    updated_at
FROM iam_service_accounts
WHERE %s
    AND organization_id = @organization_id
    AND deleted_at IS NULL
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query service accounts: %w", err)
	}

	accounts, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ServiceAccount])
	if err != nil {
		return fmt.Errorf("cannot collect service accounts: %w", err)
	}

	*ss = accounts

	return nil
}

func (ss *ServiceAccounts) LoadByIDs(ctx context.Context, conn pg.Querier, ids []gid.GID) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    description,
    scopes,
    disabled_at,
    deleted_at,
    created_at,
    updated_at
FROM iam_service_accounts
WHERE id = ANY(@ids)
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"ids": ids})
	if err != nil {
		return fmt.Errorf("cannot query service accounts by IDs: %w", err)
	}

	accounts, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ServiceAccount])
	if err != nil {
		return fmt.Errorf("cannot collect service accounts by IDs: %w", err)
	}

	*ss = accounts

	return nil
}

func (ss *ServiceAccounts) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM iam_service_accounts
WHERE %s
    AND organization_id = @organization_id
    AND deleted_at IS NULL
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count service accounts: %w", err)
	}

	return count, nil
}

func (s *ServiceAccount) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO iam_service_accounts (
    id,
    tenant_id,
    organization_id,
    name,
    description,
    scopes,
    disabled_at,
    deleted_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @name,
    @description,
    @scopes,
    @disabled_at,
    @deleted_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              s.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": s.OrganizationID,
		"name":            s.Name,
		"description":     s.Description,
		"scopes":          s.Scopes,
		"disabled_at":     s.DisabledAt,
		"deleted_at":      s.DeletedAt,
		"created_at":      s.CreatedAt,
		"updated_at":      s.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert service account: %w", err)
	}

	return nil
}

func (s *ServiceAccount) Update(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
UPDATE iam_service_accounts
SET
    name = @name,
    description = @description,
    scopes = @scopes,
    disabled_at = @disabled_at,
    deleted_at = @deleted_at,
    updated_at = @updated_at
WHERE %s
    AND id = @id
    AND deleted_at IS NULL
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":          s.ID,
		"name":        s.Name,
		"description": s.Description,
		"scopes":      s.Scopes,
		"disabled_at": s.DisabledAt,
		"deleted_at":  s.DeletedAt,
		"updated_at":  s.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update service account: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *ServiceAccount) SoftDelete(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
UPDATE iam_service_accounts
SET
    disabled_at = @disabled_at,
    deleted_at = @deleted_at,
    updated_at = @updated_at
WHERE %s
    AND id = @id
    AND deleted_at IS NULL
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":          s.ID,
		"disabled_at": s.DisabledAt,
		"deleted_at":  s.DeletedAt,
		"updated_at":  s.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot soft delete service account: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
