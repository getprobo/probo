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
	BusinessFunction struct {
		ID              gid.GID                        `db:"id"`
		OrganizationID  gid.GID                        `db:"organization_id"`
		ReferenceID     string                         `db:"reference_id"`
		Name            string                         `db:"name"`
		Classification  BusinessFunctionClassification `db:"classification"`
		MTDMinutes      int                            `db:"mtd_minutes"`
		RTOMinutes      int                            `db:"rto_minutes"`
		RPOMinutes      int                            `db:"rpo_minutes"`
		ImpactTolerance *string                        `db:"impact_tolerance"`
		Notes           *string                        `db:"notes"`
		OwnerID         *gid.GID                       `db:"owner_id"`
		CreatedAt       time.Time                      `db:"created_at"`
		UpdatedAt       time.Time                      `db:"updated_at"`
	}

	BusinessFunctions []*BusinessFunction
)

func (bf *BusinessFunction) CursorKey(field BusinessFunctionOrderField) page.CursorKey {
	switch field {
	case BusinessFunctionOrderFieldCreatedAt:
		return page.NewCursorKey(bf.ID, bf.CreatedAt)
	case BusinessFunctionOrderFieldReferenceID:
		return page.NewCursorKey(bf.ID, bf.ReferenceID)
	case BusinessFunctionOrderFieldName:
		return page.NewCursorKey(bf.ID, bf.Name)
	case BusinessFunctionOrderFieldClassification:
		return page.NewCursorKey(bf.ID, bf.Classification)
	case BusinessFunctionOrderFieldMTDMinutes:
		return page.NewCursorKey(bf.ID, bf.MTDMinutes)
	case BusinessFunctionOrderFieldRTOMinutes:
		return page.NewCursorKey(bf.ID, bf.RTOMinutes)
	case BusinessFunctionOrderFieldRPOMinutes:
		return page.NewCursorKey(bf.ID, bf.RPOMinutes)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (bf *BusinessFunction) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM business_functions WHERE id = ANY(@resource_ids::text[])`

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

func (bf *BusinessFunction) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	businessFunctionID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	reference_id,
	name,
	classification,
	mtd_minutes,
	rto_minutes,
	rpo_minutes,
	impact_tolerance,
	notes,
	owner_id,
	created_at,
	updated_at
FROM
	business_functions
WHERE
	%s
	AND id = @business_function_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"business_function_id": businessFunctionID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query business function: %w", err)
	}

	businessFunction, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[BusinessFunction])
	if err != nil {
		return fmt.Errorf("cannot collect business function: %w", err)
	}

	*bf = businessFunction

	return nil
}

func (bfs *BusinessFunctions) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	filter *BusinessFunctionFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	business_functions
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int

	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count business functions: %w", err)
	}

	return count, nil
}

func (bfs *BusinessFunctions) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[BusinessFunctionOrderField],
	filter *BusinessFunctionFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	reference_id,
	name,
	classification,
	mtd_minutes,
	rto_minutes,
	rpo_minutes,
	impact_tolerance,
	notes,
	owner_id,
	created_at,
	updated_at
FROM
	business_functions
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
		return fmt.Errorf("cannot query business functions: %w", err)
	}

	businessFunctions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[BusinessFunction])
	if err != nil {
		return fmt.Errorf("cannot collect business functions: %w", err)
	}

	*bfs = businessFunctions

	return nil
}

func (bf *BusinessFunction) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO business_functions (
	id,
	tenant_id,
	organization_id,
	reference_id,
	name,
	classification,
	mtd_minutes,
	rto_minutes,
	rpo_minutes,
	impact_tolerance,
	notes,
	owner_id,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@reference_id,
	@name,
	@classification,
	@mtd_minutes,
	@rto_minutes,
	@rpo_minutes,
	@impact_tolerance,
	@notes,
	@owner_id,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":               bf.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  bf.OrganizationID,
		"reference_id":     bf.ReferenceID,
		"name":             bf.Name,
		"classification":   bf.Classification,
		"mtd_minutes":      bf.MTDMinutes,
		"rto_minutes":      bf.RTOMinutes,
		"rpo_minutes":      bf.RPOMinutes,
		"impact_tolerance": bf.ImpactTolerance,
		"notes":            bf.Notes,
		"owner_id":         bf.OwnerID,
		"created_at":       bf.CreatedAt,
		"updated_at":       bf.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "business_functions_organization_id_reference_id_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert business function: %w", err)
	}

	return nil
}

func (bf *BusinessFunction) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE business_functions
SET
	reference_id = @reference_id,
	name = @name,
	classification = @classification,
	mtd_minutes = @mtd_minutes,
	rto_minutes = @rto_minutes,
	rpo_minutes = @rpo_minutes,
	impact_tolerance = @impact_tolerance,
	notes = @notes,
	owner_id = @owner_id,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":               bf.ID,
		"reference_id":     bf.ReferenceID,
		"name":             bf.Name,
		"classification":   bf.Classification,
		"mtd_minutes":      bf.MTDMinutes,
		"rto_minutes":      bf.RTOMinutes,
		"rpo_minutes":      bf.RPOMinutes,
		"impact_tolerance": bf.ImpactTolerance,
		"notes":            bf.Notes,
		"owner_id":         bf.OwnerID,
		"updated_at":       bf.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "business_functions_organization_id_reference_id_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot update business function: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (bf *BusinessFunction) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM business_functions
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": bf.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete business function: %w", err)
	}

	return nil
}

func (bf BusinessFunction) GetGeneratedDocumentID(
	ctx context.Context,
	conn pg.Querier,
	organizationID gid.GID,
) (*gid.GID, error) {
	var documentID *gid.GID

	err := conn.QueryRow(
		ctx,
		`
SELECT
	business_functions_document_id
FROM
	generated_documents
WHERE
	organization_id = @organization_id
`,
		pgx.NamedArgs{"organization_id": organizationID},
	).Scan(&documentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("cannot get business function list document ID: %w", err)
	}

	return documentID, nil
}

func (bf BusinessFunction) UpsertGeneratedDocumentID(
	ctx context.Context,
	conn pg.Tx,
	organizationID gid.GID,
	tenantID gid.TenantID,
	documentID gid.GID,
) error {
	now := time.Now()

	_, err := conn.Exec(
		ctx,
		`
INSERT INTO generated_documents (
	organization_id,
	tenant_id,
	business_functions_document_id,
	created_at,
	updated_at
) VALUES (
	@organization_id,
	@tenant_id,
	@business_functions_document_id,
	@created_at,
	@updated_at
)
ON CONFLICT (organization_id) DO UPDATE
SET
	business_functions_document_id = @business_functions_document_id,
	updated_at = @updated_at
`,
		pgx.NamedArgs{
			"organization_id":                organizationID,
			"tenant_id":                      tenantID,
			"business_functions_document_id": documentID,
			"created_at":                     now,
			"updated_at":                     now,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot upsert business function list document ID: %w", err)
	}

	return nil
}

func (bf BusinessFunction) ClearGeneratedDocumentID(
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
	generated_documents
SET
	business_functions_document_id = NULL,
	updated_at = @now
WHERE
	business_functions_document_id = ANY(@ids)
`,
		pgx.NamedArgs{
			"ids": ids,
			"now": time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot clear business function list document references: %w", err)
	}

	return nil
}
