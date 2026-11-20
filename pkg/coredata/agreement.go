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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	Agreement struct {
		ID               gid.GID       `db:"id"`
		OrganizationID   gid.GID       `db:"organization_id"`
		SiteID           gid.GID       `db:"site_id"`
		Kind             AgreementKind `db:"kind"`
		DocumentID       gid.GID       `db:"document_id"`
		CurrentVersionID *gid.GID      `db:"current_version_id"`
		CreatedAt        time.Time     `db:"created_at"`
		UpdatedAt        time.Time     `db:"updated_at"`
	}

	Agreements []*Agreement
)

func (a *Agreement) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM agreements WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query agreement authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID
		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan agreement authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate agreement authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (a *Agreement) CursorKey(orderBy AgreementOrderField) page.CursorKey {
	switch orderBy {
	case AgreementOrderFieldCreatedAt:
		return page.NewCursorKey(a.ID, a.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (a *Agreement) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO agreements (
	id, tenant_id, organization_id, site_id, kind, document_id, current_version_id,
	created_at, updated_at
) VALUES (
	@id, @tenant_id, @organization_id, @site_id, @kind, @document_id, @current_version_id,
	@created_at, @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                 a.ID,
		"tenant_id":          scope.GetTenantID(),
		"organization_id":    a.OrganizationID,
		"site_id":            a.SiteID,
		"kind":               a.Kind,
		"document_id":        a.DocumentID,
		"current_version_id": a.CurrentVersionID,
		"created_at":         a.CreatedAt,
		"updated_at":         a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "agreements_document_id_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert agreement: %w", err)
	}

	return nil
}

func (a *Agreement) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE agreements SET
	current_version_id = @current_version_id,
	updated_at = @updated_at
WHERE %s AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                 a.ID,
		"current_version_id": a.CurrentVersionID,
		"updated_at":         a.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update agreement: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (a *Agreement) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id, organization_id, site_id, kind, document_id, current_version_id,
	created_at, updated_at
FROM agreements
WHERE %s AND id = @id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agreement: %w", err)
	}

	a2, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Agreement])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect agreement: %w", err)
	}

	*a = a2

	return nil
}

func (aa *Agreements) LoadBySiteID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	siteID gid.GID,
	cursor *page.Cursor[AgreementOrderField],
) error {
	q := `
SELECT
	id, organization_id, site_id, kind, document_id, current_version_id,
	created_at, updated_at
FROM agreements
WHERE
	%s
	AND site_id = @site_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"site_id": siteID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agreements: %w", err)
	}

	agreements, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Agreement])
	if err != nil {
		return fmt.Errorf("cannot collect agreements: %w", err)
	}

	*aa = agreements

	return nil
}
