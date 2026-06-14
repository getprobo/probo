// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type (
	DocumentVersionSignature struct {
		ID                gid.GID                       `json:"id" db:"id"`
		OrganizationID    gid.GID                       `json:"-" db:"organization_id"`
		DocumentVersionID gid.GID                       `json:"document_version_id" db:"document_version_id"`
		State             DocumentVersionSignatureState `json:"state" db:"state"`
		SignedBy          gid.GID                       `json:"signed_by" db:"signed_by_profile_id"`
		SignedAt          *time.Time                    `json:"signed_at" db:"signed_at"`
		RequestedAt       time.Time                     `json:"requested_at" db:"requested_at"`
		CreatedAt         time.Time                     `json:"created_at" db:"created_at"`
		UpdatedAt         time.Time                     `json:"updated_at" db:"updated_at"`
	}

	DocumentVersionSignatures []*DocumentVersionSignature

	DocumentVersionSignatureWithPeople struct {
		DocumentVersionSignature
		SignedByFullName string `db:"signed_by_full_name"`
	}

	DocumentVersionSignaturesWithPeople []*DocumentVersionSignatureWithPeople
)

func (pvs DocumentVersionSignature) CursorKey(orderBy DocumentVersionSignatureOrderField) page.CursorKey {
	switch orderBy {
	case DocumentVersionSignatureOrderFieldCreatedAt:
		return page.NewCursorKey(pvs.ID, pvs.CreatedAt)
	case DocumentVersionSignatureOrderFieldSignedAt:
		return page.NewCursorKey(pvs.ID, pvs.SignedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (dvs *DocumentVersionSignature) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM document_version_signatures WHERE id = ANY(@resource_ids::text[])`

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

func (pvs *DocumentVersionSignature) LoadByDocumentVersionIDAndSignatory(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
	documentVersionID gid.GID,
	signatory gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	document_version_id,
	state,
	signed_by_profile_id,
	signed_at,
	requested_at,
	created_at,
	updated_at
FROM
	document_version_signatures
WHERE
	%s
	AND document_version_id = @document_version_id
	AND signed_by_profile_id = @signatory
LIMIT 1
`

	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_id": documentVersionID, "signatory": signatory}
	maps.Copy(args, predicate.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signature: %w", err)
	}

	documentVersionSignature, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DocumentVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect document version signature: %w", err)
	}

	*pvs = documentVersionSignature

	return nil
}

// LoadByDocumentMajorAndSignatory loads the signatory's existing signature for
// the whole major that owns documentVersionID, scanning across every minor
// version of that major. A signed signature is preferred over a still pending
// one, then the most recent. It returns ErrResourceNotFound when the signatory
// has no signature anywhere in the major.
func (pvs *DocumentVersionSignature) LoadByDocumentMajorAndSignatory(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
	documentVersionID gid.GID,
	signatory gid.GID,
) error {
	q := `
WITH source_version AS (
	SELECT document_id, major FROM document_versions WHERE id = @document_version_id
),
major_versions AS (
	SELECT dv.id FROM document_versions dv
	INNER JOIN source_version sv ON dv.document_id = sv.document_id AND dv.major = sv.major
),
major_signatures AS (
	SELECT
		dvs.id,
		dvs.organization_id,
		dvs.tenant_id,
		dvs.document_version_id,
		dvs.state,
		dvs.signed_by_profile_id,
		dvs.signed_at,
		dvs.requested_at,
		dvs.created_at,
		dvs.updated_at
	FROM document_version_signatures dvs
	INNER JOIN major_versions mv ON dvs.document_version_id = mv.id
	WHERE dvs.signed_by_profile_id = @signatory
)
SELECT
	id,
	organization_id,
	document_version_id,
	state,
	signed_by_profile_id,
	signed_at,
	requested_at,
	created_at,
	updated_at
FROM
	major_signatures
WHERE
	%s
ORDER BY
	CASE state WHEN 'SIGNED' THEN 0 ELSE 1 END,
	created_at DESC
LIMIT 1
`

	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_id": documentVersionID, "signatory": signatory}
	maps.Copy(args, predicate.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signature by major: %w", err)
	}

	documentVersionSignature, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DocumentVersionSignature])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect document version signature by major: %w", err)
	}

	*pvs = documentVersionSignature

	return nil
}

func (pvs *DocumentVersionSignature) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
	signatureID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	document_version_id,
	state,
	signed_by_profile_id,
	signed_at,
	requested_at,
	created_at,
	updated_at
FROM
	document_version_signatures
WHERE
	id = @document_version_signature_id
	AND %s
`

	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_signature_id": signatureID}
	maps.Copy(args, predicate.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signature: %w", err)
	}

	documentVersionSignature, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DocumentVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect document version signature: %w", err)
	}

	*pvs = documentVersionSignature

	return nil
}

func (pvs DocumentVersionSignature) Insert(
	ctx context.Context,
	conn pg.Tx,
	predicate Predicater,
) error {
	q := `
INSERT INTO document_version_signatures (
	id,
	tenant_id,
	organization_id,
	document_version_id,
	state,
	signed_by_profile_id,
	signed_at,
	requested_at,
	created_at,
	updated_at
) VALUES (
 	@id,
	@tenant_id,
	@organization_id,
	@document_version_id,
	@state,
	@signed_by_profile_id,
	@signed_at,
	@requested_at,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                   pvs.ID,
		"tenant_id":            predicate.GetTenantID(),
		"organization_id":      pvs.OrganizationID,
		"document_version_id":  pvs.DocumentVersionID,
		"state":                pvs.State,
		"signed_by_profile_id": pvs.SignedBy,
		"signed_at":            pvs.SignedAt,
		"requested_at":         pvs.RequestedAt,
		"created_at":           pvs.CreatedAt,
		"updated_at":           pvs.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "policy_version_signatures_policy_version_id_signed_by_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert document version signature: %w", err)
	}

	return nil
}

func (pvss *DocumentVersionSignatures) LoadByDocumentVersionID(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
	documentVersionID gid.GID,
	cursor *page.Cursor[DocumentVersionSignatureOrderField],
	filter *DocumentVersionSignatureFilter,
) error {
	q := `
WITH source_version AS (
	SELECT document_id, major FROM document_versions WHERE id = @document_version_id
),
major_versions AS (
	SELECT dv.id FROM document_versions dv
	INNER JOIN source_version sv ON dv.document_id = sv.document_id AND dv.major = sv.major
)
SELECT
	document_version_signatures.id,
	document_version_signatures.organization_id,
	document_version_signatures.document_version_id,
	document_version_signatures.state,
	document_version_signatures.signed_by_profile_id,
	document_version_signatures.signed_at,
	document_version_signatures.requested_at,
	document_version_signatures.created_at,
	document_version_signatures.updated_at
FROM
	document_version_signatures
INNER JOIN major_versions mv ON document_version_signatures.document_version_id = mv.id
WHERE
	%s
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, predicate.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_id": documentVersionID}
	maps.Copy(args, predicate.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signatures: %w", err)
	}

	documentVersionSignatures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect document version signatures: %w", err)
	}

	*pvss = documentVersionSignatures

	return nil
}

func (pvs *DocumentVersionSignature) Update(
	ctx context.Context,
	conn pg.Tx,
	predicate Predicater,
) error {
	q := `
UPDATE document_version_signatures
SET
	state = @state,
	signed_by_profile_id = @signed_by_profile_id,
	signed_at = @signed_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                   pvs.ID,
		"state":                pvs.State,
		"signed_by_profile_id": pvs.SignedBy,
		"signed_at":            pvs.SignedAt,
		"updated_at":           pvs.UpdatedAt,
	}

	maps.Copy(args, predicate.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update document version signature: %w", err)
	}

	return nil
}

func (pvs *DocumentVersionSignature) Delete(
	ctx context.Context,
	conn pg.Tx,
	predicate Predicater,
	documentVersionSignatureID gid.GID,
) error {
	q := `
DELETE FROM document_version_signatures
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{"id": documentVersionSignatureID}
	maps.Copy(args, predicate.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete document version signature: %w", err)
	}

	return nil
}

func (pvss *DocumentVersionSignatures) DeleteRequestedBySignatory(
	ctx context.Context,
	conn pg.Tx,
	predicate Predicater,
	signatoryID gid.GID,
) error {
	q := `
DELETE FROM document_version_signatures
WHERE
	%s
	AND signed_by_profile_id = @signatory_id
	AND state = @state
`

	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{
		"signatory_id": signatoryID,
		"state":        DocumentVersionSignatureStateRequested,
	}
	maps.Copy(args, predicate.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete requested document version signatures: %w", err)
	}

	return nil
}

func (pvss *DocumentVersionSignatures) DeleteRequestedByDocumentIDBelowMajor(
	ctx context.Context,
	conn pg.Tx,
	predicate Predicater,
	documentID gid.GID,
	major int,
) error {
	q := `
DELETE FROM document_version_signatures
WHERE
	%s
	AND state = @state
	AND document_version_id IN (
		SELECT id
		FROM document_versions
		WHERE document_id = @document_id
			AND major < @major
	)
`

	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id": documentID,
		"major":       major,
		"state":       DocumentVersionSignatureStateRequested,
	}
	maps.Copy(args, predicate.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete requested document version signatures from previous major versions: %w", err)
	}

	return nil
}

func (pvss *DocumentVersionSignaturesWithPeople) LoadByDocumentVersionIDWithPeople(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
	documentVersionID gid.GID,
	limit int,
) error {
	q := `
WITH source_version AS (
	SELECT document_id, major FROM document_versions WHERE id = @document_version_id
),
major_versions AS (
	SELECT dv.id FROM document_versions dv
	INNER JOIN source_version sv ON dv.document_id = sv.document_id AND dv.major = sv.major
),
signatures_with_people AS (
	SELECT
		dvs.id,
		dvs.organization_id,
		dvs.tenant_id,
		dvs.document_version_id,
		dvs.state,
		dvs.signed_by_profile_id,
		dvs.signed_at,
		dvs.requested_at,
		dvs.created_at,
		dvs.updated_at,
		p.full_name AS signed_by_full_name
	FROM document_version_signatures dvs
	INNER JOIN major_versions mv ON dvs.document_version_id = mv.id
	INNER JOIN iam_membership_profiles p ON dvs.signed_by_profile_id = p.id
	WHERE
		dvs.state = 'SIGNED'
		OR (
			p.state = 'ACTIVE'
			AND (p.contract_end_date IS NULL OR p.contract_end_date >= CURRENT_DATE)
			AND EXISTS (
				SELECT 1
				FROM iam_memberships m
				WHERE m.identity_id = p.identity_id
					AND m.organization_id = p.organization_id
			)
		)
)
SELECT
	id,
	organization_id,
	document_version_id,
	state,
	signed_by_profile_id,
	signed_at,
	requested_at,
	created_at,
	updated_at,
	signed_by_full_name
FROM
	signatures_with_people
WHERE
	%s
ORDER BY
	signed_by_full_name ASC
LIMIT @limit
`

	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_version_id": documentVersionID,
		"limit":               limit,
	}
	maps.Copy(args, predicate.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signatures with people: %w", err)
	}

	signatures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionSignatureWithPeople])
	if err != nil {
		return fmt.Errorf("cannot collect document version signatures with people: %w", err)
	}

	*pvss = signatures

	return nil
}

func (pvs *DocumentVersionSignature) IsSignedByUserEmail(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
	documentVersionID gid.GID,
	userEmail mail.Addr,
) (bool, error) {
	q := `
WITH source_version AS (
	SELECT document_id, major FROM document_versions WHERE id = @document_version_id
),
major_versions AS (
	SELECT dv.id FROM document_versions dv
	INNER JOIN source_version sv ON dv.document_id = sv.document_id AND dv.major = sv.major
),
signed_emails AS (
	SELECT dvs.id, dvs.tenant_id
	FROM document_version_signatures dvs
	INNER JOIN major_versions mv ON dvs.document_version_id = mv.id
	INNER JOIN iam_membership_profiles p ON dvs.signed_by_profile_id = p.id
	INNER JOIN identities i ON p.identity_id = i.id
	WHERE i.email_address = @user_email::CITEXT
		AND dvs.state = 'SIGNED'
)
SELECT EXISTS (
	SELECT 1
	FROM signed_emails
	WHERE %s
) AS signed
`

	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_version_id": documentVersionID,
		"user_email":          userEmail,
	}
	maps.Copy(args, predicate.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot query document version signature: %w", err)
	}

	signed, err := pgx.CollectOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return false, fmt.Errorf("cannot collect signed status: %w", err)
	}

	return signed, nil
}

func (dvs *DocumentVersionSignatures) CountByDocumentVersionID(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
	documentVersionID gid.GID,
	filter *DocumentVersionSignatureFilter,
) (int, error) {
	q := `
WITH source_version AS (
	SELECT document_id, major FROM document_versions WHERE id = @document_version_id
),
major_versions AS (
	SELECT dv.id FROM document_versions dv
	INNER JOIN source_version sv ON dv.document_id = sv.document_id AND dv.major = sv.major
)
SELECT
	COUNT(document_version_signatures.id)
FROM
	document_version_signatures
INNER JOIN major_versions mv ON document_version_signatures.document_version_id = mv.id
WHERE
	%s
	AND %s
`

	q = fmt.Sprintf(q, predicate.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"document_version_id": documentVersionID}
	maps.Copy(args, predicate.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}
