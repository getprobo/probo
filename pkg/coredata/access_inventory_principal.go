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
	AccessInventoryPrincipal struct {
		ID                   gid.GID                      `db:"id"`
		OrganizationID       gid.GID                      `db:"organization_id"`
		AccessReviewSourceID gid.GID                      `db:"access_review_source_id"`
		PrincipalKey         string                       `db:"principal_key"`
		Email                string                       `db:"email"`
		FullName             string                       `db:"full_name"`
		ExternalID           string                       `db:"external_id"`
		JobTitle             string                       `db:"job_title"`
		IsAdmin              *bool                        `db:"is_admin"`
		MFAStatus            MFAStatus                    `db:"mfa_status"`
		AuthMethod           AccessReviewEntryAuthMethod  `db:"auth_method"`
		AccountType          AccessReviewEntryAccountType `db:"account_type"`
		Active               *bool                        `db:"active"`
		LastLogin            *time.Time                   `db:"last_login"`
		AccountCreatedAt     *time.Time                   `db:"account_created_at"`
		FirstSeenAt          time.Time                    `db:"first_seen_at"`
		LastSeenAt           time.Time                    `db:"last_seen_at"`
		CreatedAt            time.Time                    `db:"created_at"`
		UpdatedAt            time.Time                    `db:"updated_at"`
	}

	AccessInventoryPrincipals []*AccessInventoryPrincipal
)

func (p AccessInventoryPrincipal) CursorKey(orderBy AccessInventoryPrincipalOrderField) page.CursorKey {
	switch orderBy {
	case AccessInventoryPrincipalOrderFieldCreatedAt:
		return page.NewCursorKey(p.ID, p.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (p *AccessInventoryPrincipal) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM access_inventory_principals WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (p *AccessInventoryPrincipal) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    access_review_source_id,
    principal_key,
    email,
    full_name,
    external_id,
    job_title,
    is_admin,
    mfa_status,
    auth_method,
    account_type,
    active,
    last_login,
    account_created_at,
    first_seen_at,
    last_seen_at,
    created_at,
    updated_at
FROM
    access_inventory_principals
WHERE
    %s
    AND id = @id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_inventory_principals: %w", err)
	}

	principal, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessInventoryPrincipal])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access inventory principal: %w", err)
	}

	*p = principal

	return nil
}

func (p *AccessInventoryPrincipal) LoadBySourceIDAndPrincipalKey(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sourceID gid.GID,
	principalKey string,
) error {
	q := `
SELECT
    id,
    organization_id,
    access_review_source_id,
    principal_key,
    email,
    full_name,
    external_id,
    job_title,
    is_admin,
    mfa_status,
    auth_method,
    account_type,
    active,
    last_login,
    account_created_at,
    first_seen_at,
    last_seen_at,
    created_at,
    updated_at
FROM
    access_inventory_principals
WHERE
    %s
    AND access_review_source_id = @access_review_source_id
    AND principal_key = @principal_key
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"access_review_source_id": sourceID,
		"principal_key":           principalKey,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_inventory_principals: %w", err)
	}

	principal, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessInventoryPrincipal])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access inventory principal: %w", err)
	}

	*p = principal

	return nil
}

func (ps *AccessInventoryPrincipals) LoadBySourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sourceID gid.GID,
	cursor *page.Cursor[AccessInventoryPrincipalOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    access_review_source_id,
    principal_key,
    email,
    full_name,
    external_id,
    job_title,
    is_admin,
    mfa_status,
    auth_method,
    account_type,
    active,
    last_login,
    account_created_at,
    first_seen_at,
    last_seen_at,
    created_at,
    updated_at
FROM
    access_inventory_principals
WHERE
    %s
    AND access_review_source_id = @access_review_source_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"access_review_source_id": sourceID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_inventory_principals: %w", err)
	}

	principals, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessInventoryPrincipal])
	if err != nil {
		return fmt.Errorf("cannot collect access_inventory_principals: %w", err)
	}

	*ps = principals

	return nil
}

func (p *AccessInventoryPrincipal) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) (inserted bool, err error) {
	q := `
INSERT INTO access_inventory_principals (
    id,
    tenant_id,
    organization_id,
    access_review_source_id,
    principal_key,
    email,
    full_name,
    external_id,
    job_title,
    is_admin,
    mfa_status,
    auth_method,
    account_type,
    active,
    last_login,
    account_created_at,
    first_seen_at,
    last_seen_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @access_review_source_id,
    @principal_key,
    @email,
    @full_name,
    @external_id,
    @job_title,
    @is_admin,
    @mfa_status,
    @auth_method,
    @account_type,
    @active,
    @last_login,
    @account_created_at,
    @first_seen_at,
    @last_seen_at,
    @created_at,
    @updated_at
)
ON CONFLICT (access_review_source_id, principal_key) DO UPDATE SET
    email              = EXCLUDED.email,
    full_name          = EXCLUDED.full_name,
    external_id        = EXCLUDED.external_id,
    job_title          = EXCLUDED.job_title,
    is_admin           = EXCLUDED.is_admin,
    mfa_status         = EXCLUDED.mfa_status,
    auth_method        = EXCLUDED.auth_method,
    account_type       = EXCLUDED.account_type,
    active             = EXCLUDED.active,
    last_login         = EXCLUDED.last_login,
    account_created_at = EXCLUDED.account_created_at,
    last_seen_at       = EXCLUDED.last_seen_at,
    updated_at         = EXCLUDED.updated_at
RETURNING
    id,
    organization_id,
    access_review_source_id,
    principal_key,
    email,
    full_name,
    external_id,
    job_title,
    is_admin,
    mfa_status,
    auth_method,
    account_type,
    active,
    last_login,
    account_created_at,
    first_seen_at,
    last_seen_at,
    created_at,
    updated_at
`
	originalID := p.ID

	args := pgx.StrictNamedArgs{
		"id":                      p.ID,
		"tenant_id":               scope.GetTenantID(),
		"organization_id":         p.OrganizationID,
		"access_review_source_id": p.AccessReviewSourceID,
		"principal_key":           p.PrincipalKey,
		"email":                   p.Email,
		"full_name":               p.FullName,
		"external_id":             p.ExternalID,
		"job_title":               p.JobTitle,
		"is_admin":                p.IsAdmin,
		"mfa_status":              p.MFAStatus,
		"auth_method":             p.AuthMethod,
		"account_type":            p.AccountType,
		"active":                  p.Active,
		"last_login":              p.LastLogin,
		"account_created_at":      p.AccountCreatedAt,
		"first_seen_at":           p.FirstSeenAt,
		"last_seen_at":            p.LastSeenAt,
		"created_at":              p.CreatedAt,
		"updated_at":              p.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert access inventory principal: %w", err)
	}

	principal, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessInventoryPrincipal])
	if err != nil {
		return false, fmt.Errorf("cannot collect access inventory principal upsert result: %w", err)
	}

	*p = principal

	return originalID == p.ID, nil
}

func (p *AccessInventoryPrincipal) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM access_inventory_principals
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": p.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete access inventory principal: %w", err)
	}

	return nil
}
