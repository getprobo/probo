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
)

type (
	CompliancePortalDocument struct {
		ID                 gid.GID                    `db:"id"`
		OrganizationID     gid.GID                    `db:"organization_id"`
		CompliancePortalID gid.GID                    `db:"compliance_portal_id"`
		DocumentID         gid.GID                    `db:"document_id"`
		Visibility         CompliancePortalVisibility `db:"visibility"`
		CreatedAt          time.Time                  `db:"created_at"`
		UpdatedAt          time.Time                  `db:"updated_at"`
	}

	CompliancePortalDocuments []*CompliancePortalDocument
)

func (cpd *CompliancePortalDocument) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM cp_documents WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{"resource_ids": resourceIDs}

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

func (cpd *CompliancePortalDocument) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentLinkID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	compliance_portal_id,
	document_id,
	visibility,
	created_at,
	updated_at
FROM
	cp_documents
WHERE
	%s
	AND id = @document_link_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_link_id": documentLinkID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal document: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalDocument])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect compliance portal document: %w", err)
	}

	*cpd = row

	return nil
}

func (cpd *CompliancePortalDocument) LoadByCompliancePortalIDAndDocumentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalID gid.GID,
	documentID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	compliance_portal_id,
	document_id,
	visibility,
	created_at,
	updated_at
FROM
	cp_documents
WHERE
	%s
	AND compliance_portal_id = @compliance_portal_id
	AND document_id = @document_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_id": compliancePortalID,
		"document_id":          documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center document: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalDocument])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect trust center document: %w", err)
	}

	*cpd = row

	return nil
}

func (cpds *CompliancePortalDocuments) LoadByCompliancePortalIDAndDocumentIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalID gid.GID,
	documentIDs []gid.GID,
) error {
	if len(documentIDs) == 0 {
		*cpds = CompliancePortalDocuments{}
		return nil
	}

	q := `
SELECT
	id,
	organization_id,
	compliance_portal_id,
	document_id,
	visibility,
	created_at,
	updated_at
FROM
	cp_documents
WHERE
	%s
	AND compliance_portal_id = @compliance_portal_id
	AND document_id = ANY(@document_ids::text[]);
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_id": compliancePortalID,
		"document_ids":         documentIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center documents: %w", err)
	}

	documents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CompliancePortalDocument])
	if err != nil {
		return fmt.Errorf("cannot collect trust center documents: %w", err)
	}

	*cpds = documents

	return nil
}

func (cpd *CompliancePortalDocument) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO cp_documents (
	id,
	tenant_id,
	organization_id,
	compliance_portal_id,
	document_id,
	visibility,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@compliance_portal_id,
	@document_id,
	@visibility,
	@created_at,
	@updated_at
)
ON CONFLICT (compliance_portal_id, document_id) DO UPDATE
SET
	visibility = EXCLUDED.visibility,
	updated_at = EXCLUDED.updated_at
RETURNING
	id,
	organization_id,
	compliance_portal_id,
	document_id,
	visibility,
	created_at,
	updated_at
`

	args := pgx.StrictNamedArgs{
		"id":                   cpd.ID,
		"tenant_id":            scope.GetTenantID(),
		"organization_id":      cpd.OrganizationID,
		"compliance_portal_id": cpd.CompliancePortalID,
		"document_id":          cpd.DocumentID,
		"visibility":           cpd.Visibility,
		"created_at":           cpd.CreatedAt,
		"updated_at":           cpd.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert trust center document: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalDocument])
	if err != nil {
		return fmt.Errorf("cannot collect trust center document: %w", err)
	}

	*cpd = row

	return nil
}

func DeleteCompliancePortalDocumentByCompliancePortalIDAndDocumentID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	compliancePortalID gid.GID,
	documentID gid.GID,
) error {
	q := `
DELETE FROM cp_documents
WHERE
	%s
	AND compliance_portal_id = @compliance_portal_id
	AND document_id = @document_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_id": compliancePortalID,
		"document_id":          documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center document: %w", err)
	}

	return nil
}

func DeleteCompliancePortalDocumentByID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	id gid.GID,
) error {
	q := `
DELETE FROM cp_documents
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center document: %w", err)
	}

	return nil
}

func DeleteCompliancePortalDocumentsByDocumentIDs(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	documentIDs []gid.GID,
) error {
	q := `
DELETE FROM cp_documents
WHERE
	%s
	AND document_id = ANY(@document_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_ids": documentIDs}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center documents: %w", err)
	}

	return nil
}
