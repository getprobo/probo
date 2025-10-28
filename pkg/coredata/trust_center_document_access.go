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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	TrustCenterDocumentAccess struct {
		ID                  gid.GID   `db:"id"`
		TrustCenterAccessID gid.GID   `db:"trust_center_access_id"`
		DocumentID          *gid.GID  `db:"document_id"`
		ReportID            *gid.GID  `db:"report_id"`
		TrustCenterFileID   *gid.GID  `db:"trust_center_file_id"`
		Active              bool      `db:"active"`
		Requested           bool      `db:"requested"`
		CreatedAt           time.Time `db:"created_at"`
		UpdatedAt           time.Time `db:"updated_at"`
	}

	TrustCenterDocumentAccesses []*TrustCenterDocumentAccess

	ErrTrustCenterDocumentAccessNotFound struct {
		Identifier string
	}
)

func (e ErrTrustCenterDocumentAccessNotFound) Error() string {
	return fmt.Sprintf("trust center document access not found: %s", e.Identifier)
}

func (tcda *TrustCenterDocumentAccess) CursorKey(orderBy TrustCenterDocumentAccessOrderField) page.CursorKey {
	switch orderBy {
	case TrustCenterDocumentAccessOrderFieldCreatedAt:
		return page.NewCursorKey(tcda.ID, tcda.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (tcda *TrustCenterDocumentAccess) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	accessID gid.GID,
) error {
	q := `
SELECT
	id,
	trust_center_access_id,
	document_id,
	report_id,
	trust_center_file_id,
	active,
	requested,
	created_at,
	updated_at
FROM
	trust_center_document_accesses
WHERE
	%s
	AND id = @access_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"access_id": accessID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center document access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterDocumentAccess])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrTrustCenterDocumentAccessNotFound{Identifier: accessID.String()}
		}
		return fmt.Errorf("cannot collect trust center document access: %w", err)
	}

	*tcda = access

	return nil
}

func (tcda *TrustCenterDocumentAccess) LoadByTrustCenterAccessIDAndDocumentID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	documentID gid.GID,
) error {
	q := `
SELECT
	id,
	trust_center_access_id,
	document_id,
	report_id,
	trust_center_file_id,
	active,
	requested,
	created_at,
	updated_at
FROM
	trust_center_document_accesses
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
	AND document_id = @document_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
		"document_id":            documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center document access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterDocumentAccess])
	if err != nil {
		return fmt.Errorf("cannot collect trust center document access: %w", err)
	}

	*tcda = access

	return nil
}

func (tcda *TrustCenterDocumentAccess) LoadByTrustCenterAccessIDAndReportID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	reportID gid.GID,
) error {
	q := `
SELECT
	id,
	trust_center_access_id,
	document_id,
	report_id,
	trust_center_file_id,
	active,
	requested,
	created_at,
	updated_at
FROM
	trust_center_document_accesses
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
	AND report_id = @report_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
		"report_id":              reportID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center document access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterDocumentAccess])
	if err != nil {
		return fmt.Errorf("cannot collect trust center document access: %w", err)
	}

	*tcda = access

	return nil
}

func (tcda *TrustCenterDocumentAccess) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO trust_center_document_accesses (
	id,
	tenant_id,
	trust_center_access_id,
	document_id,
	report_id,
	trust_center_file_id,
	active,
	requested,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@trust_center_access_id,
	@document_id,
	@report_id,
	@trust_center_file_id,
	@active,
	@requested,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                     tcda.ID,
		"tenant_id":              scope.GetTenantID(),
		"trust_center_access_id": tcda.TrustCenterAccessID,
		"document_id":            tcda.DocumentID,
		"report_id":              tcda.ReportID,
		"trust_center_file_id":   tcda.TrustCenterFileID,
		"active":                 tcda.Active,
		"requested":              tcda.Requested,
		"created_at":             tcda.CreatedAt,
		"updated_at":             tcda.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert trust center document access: %w", err)
	}

	return nil
}

func (tcda *TrustCenterDocumentAccess) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE trust_center_document_accesses SET
	active = @active,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         tcda.ID,
		"active":     tcda.Active,
		"updated_at": tcda.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update trust center document access: %w", err)
	}

	return nil
}

func (tcda *TrustCenterDocumentAccess) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM trust_center_document_accesses
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id": tcda.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center document access: %w", err)
	}

	return nil
}

func (tcdas *TrustCenterDocumentAccesses) CountByTrustCenterAccessID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	trust_center_document_accesses
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
	}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (tcdas *TrustCenterDocumentAccesses) CountPendingRequestByTrustCenterAccessID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	trust_center_document_accesses
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
	AND requested = true
	AND active = false
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
	}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (tcdas *TrustCenterDocumentAccesses) CountActiveByTrustCenterAccessID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	trust_center_document_accesses
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
	AND active = true
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
	}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (tcdas *TrustCenterDocumentAccesses) LoadAvailableByTrustCenterAccessID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	cursor *page.Cursor[TrustCenterDocumentAccessOrderField],
) error {
	q := `
WITH organization AS (
	SELECT tc.organization_id
	FROM trust_center_accesses tca
	INNER JOIN trust_centers tc ON tca.trust_center_id = tc.id
	WHERE tca.tenant_id = @tenant_id
		AND tca.id = @trust_center_access_id
),
tenant_organization AS (
	SELECT o.id AS organization_id
	FROM organizations o
	WHERE %s
),
all_items AS (
	SELECT
		d.id AS item_id,
		d.id AS document_id,
		NULL::text AS report_id,
		NULL::text AS trust_center_file_id,
		d.created_at AS item_created_at,
		d.updated_at AS item_updated_at
	FROM documents d, tenant_organization o
	WHERE d.organization_id = o.organization_id
		AND d.deleted_at IS NULL
		AND d.trust_center_visibility = 'PRIVATE'::trust_center_visibility

	UNION ALL

	SELECT
		r.id AS item_id,
		NULL::text AS document_id,
		r.id AS report_id,
		NULL::text AS trust_center_file_id,
		r.created_at AS item_created_at,
		r.updated_at AS item_updated_at
	FROM audits r, tenant_organization o
	WHERE r.organization_id = o.organization_id
		AND r.trust_center_visibility = 'PRIVATE'::trust_center_visibility

	UNION ALL

	SELECT
		tcf.id AS item_id,
		NULL::text AS document_id,
		NULL::text AS report_id,
		tcf.id AS trust_center_file_id,
		tcf.created_at AS item_created_at,
		tcf.updated_at AS item_updated_at
	FROM trust_center_files tcf, tenant_organization o
	WHERE tcf.organization_id = o.organization_id
		AND tcf.trust_center_visibility = 'PRIVATE'::trust_center_visibility
),
final_items AS (
	SELECT
		COALESCE(tcda.id, ai.item_id) AS id,
		tcda.tenant_id,
		@trust_center_access_id AS trust_center_access_id,
		ai.document_id,
		ai.report_id,
		ai.trust_center_file_id,
		COALESCE(tcda.active, false) AS active,
		COALESCE(tcda.requested, false) AS requested,
		COALESCE(tcda.created_at, ai.item_created_at) AS created_at,
		COALESCE(tcda.updated_at, ai.item_updated_at) AS updated_at
	FROM all_items ai
	LEFT JOIN trust_center_document_accesses tcda ON (
		tcda.trust_center_access_id = @trust_center_access_id
		AND (
			(tcda.document_id = ai.document_id AND ai.document_id IS NOT NULL)
			OR (tcda.report_id = ai.report_id AND ai.report_id IS NOT NULL)
			OR (tcda.trust_center_file_id = ai.trust_center_file_id AND ai.trust_center_file_id IS NOT NULL)
		)
	)
)
SELECT
	id,
	trust_center_access_id,
	document_id,
	report_id,
	trust_center_file_id,
	active,
	requested,
	created_at,
	updated_at
FROM final_items
WHERE %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center document accesses: %w", err)
	}

	accesses, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrustCenterDocumentAccess])
	if err != nil {
		return fmt.Errorf("cannot collect trust center document accesses: %w", err)
	}

	*tcdas = accesses

	return nil
}

func (tcdas *TrustCenterDocumentAccesses) LoadAllByTrustCenterAccessID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
) error {
	q := `
SELECT
	id,
	trust_center_access_id,
	document_id,
	report_id,
	trust_center_file_id,
	active,
	requested,
	created_at,
	updated_at
FROM
	trust_center_document_accesses
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center document accesses: %w", err)
	}

	accesses, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrustCenterDocumentAccess])
	if err != nil {
		return fmt.Errorf("cannot collect trust center document accesses: %w", err)
	}

	*tcdas = accesses

	return nil
}

func DeactivateByTrustCenterAccessID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	updatedAt time.Time,
) error {
	q := `
UPDATE trust_center_document_accesses
SET active = false, updated_at = @updated_at
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
		"updated_at":             updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot deactivate trust center document accesses: %w", err)
	}

	return nil
}

func ActivateByDocumentIDs(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	documentIDs []gid.GID,
	updatedAt time.Time,
) error {
	q := `
UPDATE trust_center_document_accesses
SET active = true, updated_at = @updated_at
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
	AND document_id = ANY(@document_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
		"document_ids":           documentIDs,
		"updated_at":             updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot activate trust center document accesses by document IDs: %w", err)
	}

	return nil
}

func ActivateByReportIDs(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	reportIDs []gid.GID,
	updatedAt time.Time,
) error {
	q := `
UPDATE trust_center_document_accesses
SET active = true, updated_at = @updated_at
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
	AND report_id = ANY(@report_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
		"report_ids":             reportIDs,
		"updated_at":             updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot activate trust center document accesses by report IDs: %w", err)
	}

	return nil
}

func (tcdas TrustCenterDocumentAccesses) BulkInsertDocumentAccesses(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	documentIDs []gid.GID,
	requested bool,
	createdAt time.Time,
) error {
	if len(documentIDs) == 0 {
		return nil
	}

	q := `
WITH document_access_data AS (
	SELECT
		generate_gid(decode_base64_unpadded(@tenant_id), @trust_center_document_access_entity_type) AS id,
		@tenant_id AS tenant_id,
		@trust_center_access_id AS trust_center_access_id,
		unnest(@document_ids::text[]) AS document_id,
		null::text AS report_id,
		null::text AS trust_center_file_id,
		false AS active,
		@requested::boolean AS requested,
		@created_at::timestamptz AS created_at,
		@updated_at::timestamptz AS updated_at
)
INSERT INTO trust_center_document_accesses (
	id, tenant_id, trust_center_access_id, document_id, report_id, trust_center_file_id, active, requested, created_at, updated_at
)
SELECT * FROM document_access_data
ON CONFLICT DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"tenant_id":              scope.GetTenantID(),
		"trust_center_access_id": trustCenterAccessID,
		"document_ids":           documentIDs,
		"trust_center_document_access_entity_type": TrustCenterDocumentAccessEntityType,
		"requested":  requested,
		"created_at": createdAt,
		"updated_at": createdAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot bulk insert trust center document accesses: %w", err)
	}

	return nil
}

func (tcdas TrustCenterDocumentAccesses) BulkInsertReportAccesses(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	reportIDs []gid.GID,
	requested bool,
	createdAt time.Time,
) error {
	if len(reportIDs) == 0 {
		return nil
	}

	q := `
WITH report_access_data AS (
	SELECT
		generate_gid(decode_base64_unpadded(@tenant_id), @trust_center_document_access_entity_type) AS id,
		@tenant_id AS tenant_id,
		@trust_center_access_id AS trust_center_access_id,
		null::text AS document_id,
		unnest(@report_ids::text[]) AS report_id,
		null::text AS trust_center_file_id,
		false AS active,
		@requested::boolean AS requested,
		@created_at::timestamptz AS created_at,
		@updated_at::timestamptz AS updated_at
)
INSERT INTO trust_center_document_accesses (
	id, tenant_id, trust_center_access_id, document_id, report_id, trust_center_file_id, active, requested, created_at, updated_at
)
SELECT * FROM report_access_data
ON CONFLICT DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"tenant_id": scope.GetTenantID(),
		"trust_center_document_access_entity_type": TrustCenterDocumentAccessEntityType,
		"trust_center_access_id":                   trustCenterAccessID,
		"report_ids":                               reportIDs,
		"created_at":                               createdAt,
		"updated_at":                               createdAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot bulk insert trust center report accesses: %w", err)
	}

	return nil
}

func (tcda *TrustCenterDocumentAccess) LoadByTrustCenterAccessIDAndTrustCenterFileID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	trustCenterFileID gid.GID,
) error {
	q := `
SELECT
	id,
	trust_center_access_id,
	document_id,
	report_id,
	trust_center_file_id,
	active,
	requested,
	created_at,
	updated_at
FROM
	trust_center_document_accesses
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
	AND trust_center_file_id = @trust_center_file_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
		"trust_center_file_id":   trustCenterFileID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center document access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterDocumentAccess])
	if err != nil {
		return fmt.Errorf("cannot collect trust center document access: %w", err)
	}

	*tcda = access

	return nil
}

func ActivateByTrustCenterFileIDs(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	trustCenterFileIDs []gid.GID,
	updatedAt time.Time,
) error {
	q := `
UPDATE trust_center_document_accesses
SET active = true, updated_at = @updated_at
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
	AND trust_center_file_id = ANY(@trust_center_file_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
		"trust_center_file_ids":  trustCenterFileIDs,
		"updated_at":             updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot activate trust center document accesses by trust center file IDs: %w", err)
	}

	return nil
}

func (tcdas TrustCenterDocumentAccesses) BulkInsertTrustCenterFileAccesses(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	trustCenterFileIDs []gid.GID,
	requested bool,
	createdAt time.Time,
) error {
	q := `
WITH trust_center_file_access_data AS (
	SELECT
		generate_gid(decode_base64_unpadded(@tenant_id), @trust_center_document_access_entity_type) AS id,
		@tenant_id AS tenant_id,
		@trust_center_access_id AS trust_center_access_id,
		null::text AS document_id,
		null::text AS report_id,
		unnest(@trust_center_file_ids::text[]) AS trust_center_file_id,
		false AS active,
		@requested::boolean AS requested,
		@created_at::timestamptz AS created_at,
		@updated_at::timestamptz AS updated_at
)
INSERT INTO trust_center_document_accesses (
	id, tenant_id, trust_center_access_id, document_id, report_id, trust_center_file_id, active, requested, created_at, updated_at
)
SELECT * FROM trust_center_file_access_data
ON CONFLICT DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"tenant_id": scope.GetTenantID(),
		"trust_center_document_access_entity_type": TrustCenterDocumentAccessEntityType,
		"trust_center_access_id":                   trustCenterAccessID,
		"trust_center_file_ids":                    trustCenterFileIDs,
		"requested":                                requested,
		"created_at":                               createdAt,
		"updated_at":                               createdAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot bulk insert trust center file accesses: %w", err)
	}

	return nil
}
