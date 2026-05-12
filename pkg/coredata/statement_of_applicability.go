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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	StatementOfApplicability struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		Name           string    `db:"name"`
		DocumentID     *gid.GID  `db:"document_id"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	StatementsOfApplicability []*StatementOfApplicability
)

func (s StatementOfApplicability) CursorKey(orderBy StatementOfApplicabilityOrderField) page.CursorKey {
	switch orderBy {
	case StatementOfApplicabilityOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	case StatementOfApplicabilityOrderFieldName:
		return page.NewCursorKey(s.ID, s.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *StatementOfApplicability) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM statements_of_applicability WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, s.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query statement of applicability authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (s *StatementOfApplicability) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	statementOfApplicabilityID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    document_id,
    created_at,
    updated_at
FROM
    statements_of_applicability
WHERE
    %s
    AND id = @statement_of_applicability_id
    AND snapshot_id IS NULL
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"statement_of_applicability_id": statementOfApplicabilityID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query statements_of_applicability: %w", err)
	}

	statementOfApplicability, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[StatementOfApplicability])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect statement_of_applicability: %w", err)
	}

	*s = statementOfApplicability
	return nil
}

func (s *StatementsOfApplicability) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[StatementOfApplicabilityOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    document_id,
    created_at,
    updated_at
FROM
    statements_of_applicability
WHERE
    %s
    AND organization_id = @organization_id
    AND snapshot_id IS NULL
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query statements_of_applicability: %w", err)
	}

	statementsOfApplicability, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[StatementOfApplicability])
	if err != nil {
		return fmt.Errorf("cannot collect statements_of_applicability: %w", err)
	}

	*s = statementsOfApplicability
	return nil
}

func (s *StatementsOfApplicability) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    statements_of_applicability
WHERE
    %s
    AND organization_id = @organization_id
    AND snapshot_id IS NULL
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count statements_of_applicability: %w", err)
	}

	return count, nil
}

func (s *StatementOfApplicability) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    statements_of_applicability (
        tenant_id,
        id,
        organization_id,
        name,
        document_id,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @statement_of_applicability_id,
    @organization_id,
    @name,
    @document_id,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":                     scope.GetTenantID(),
		"statement_of_applicability_id": s.ID,
		"organization_id":               s.OrganizationID,
		"name":                          s.Name,
		"document_id":                   s.DocumentID,
		"created_at":                    s.CreatedAt,
		"updated_at":                    s.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert statement_of_applicability: %w", err)
	}

	return nil
}

func (s *StatementOfApplicability) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE statements_of_applicability
SET
    name = @name,
    document_id = @document_id,
    updated_at = @updated_at
WHERE
    %s
    AND id = @statement_of_applicability_id
    AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"statement_of_applicability_id": s.ID,
		"name":                          s.Name,
		"document_id":                   s.DocumentID,
		"updated_at":                    s.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot update statement_of_applicability: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *StatementOfApplicability) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM statements_of_applicability
WHERE
    %s
    AND id = @statement_of_applicability_id
    AND snapshot_id IS NULL
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"statement_of_applicability_id": s.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete statement_of_applicability: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (s StatementOfApplicability) ClearDocumentIDByDocumentIDs(
	ctx context.Context,
	conn pg.Tx,
	documentIDs []gid.GID,
) error {
	ids := make([]string, len(documentIDs))
	for i, id := range documentIDs {
		ids[i] = id.String()
	}

	_, err := conn.Exec(
		ctx,
		`
UPDATE
	statements_of_applicability
SET
	document_id = NULL,
	updated_at = @now
WHERE
	document_id = ANY(@ids)
`,
		pgx.NamedArgs{
			"ids": ids,
			"now": time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot clear statement of applicability document references: %w", err)
	}

	return nil
}
