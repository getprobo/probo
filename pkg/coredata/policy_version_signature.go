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
	"fmt"
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	PolicyVersionSignature struct {
		ID              gid.GID                     `json:"id"`
		PolicyVersionID gid.GID                     `json:"policy_version_id"`
		State           PolicyVersionSignatureState `json:"state"`
		SignedBy        gid.GID                     `json:"signed_by"`
		SignedAt        *time.Time                  `json:"signed_at"`
		RequestedAt     time.Time                   `json:"requested_at"`
		RequestedBy     gid.GID                     `json:"requested_by"`
		CreatedAt       time.Time                   `json:"created_at"`
		UpdatedAt       time.Time                   `json:"updated_at"`
	}

	PolicyVersionSignatures []*PolicyVersionSignature
)

func (pvs PolicyVersionSignature) CursorKey(orderBy PolicyVersionSignatureOrderField) page.CursorKey {
	switch orderBy {
	case PolicyVersionSignatureOrderFieldCreatedAt:
		return page.NewCursorKey(pvs.ID, pvs.CreatedAt)
	case PolicyVersionSignatureOrderFieldSignedAt:
		return page.NewCursorKey(pvs.ID, pvs.SignedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (pvs *PolicyVersionSignature) LoadByPolicyVersionIDAndSignatory(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	policyVersionID gid.GID,
	signatory gid.GID,
) error {
	q := `
SELECT
	id,
	policy_version_id,
	state,
	signed_by,
	signed_at,
	requested_at,
	requested_by,
	created_at,
	updated_at
FROM
	policy_version_signatures
WHERE
	%s
	AND policy_version_id = @policy_version_id
	AND signed_by = @signatory
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"policy_version_id": policyVersionID, "signatory": signatory}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query policy version signature: %w", err)
	}

	policyVersionSignature, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[PolicyVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect policy version signature: %w", err)
	}

	*pvs = policyVersionSignature

	return nil
}

func (pvs *PolicyVersionSignature) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	signatureID gid.GID,
) error {
	q := `
SELECT
	id,
	policy_version_id,
	state,
	signed_by,
	signed_at,
	requested_at,
	requested_by,
	created_at,
	updated_at
FROM
	policy_version_signatures
WHERE
	id = @policy_version_signature_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"policy_version_signature_id": signatureID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query policy version signature: %w", err)
	}

	policyVersionSignature, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[PolicyVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect policy version signature: %w", err)
	}

	*pvs = policyVersionSignature

	return nil
}

func (pvs PolicyVersionSignature) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO policy_version_signatures (
	id,
	tenant_id,
	policy_version_id,
	state,
	signed_by,
	signed_at,
	requested_at,
	requested_by,
	created_at,
	updated_at
) VALUES (
 	@id,
	@tenant_id, 
	@policy_version_id,
	@state,
	@signed_by,
	@signed_at,
	@requested_at,
	@requested_by,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                pvs.ID,
		"tenant_id":         scope.GetTenantID(),
		"policy_version_id": pvs.PolicyVersionID,
		"state":             pvs.State,
		"signed_by":         pvs.SignedBy,
		"signed_at":         pvs.SignedAt,
		"requested_at":      pvs.RequestedAt,
		"requested_by":      pvs.RequestedBy,
		"created_at":        pvs.CreatedAt,
		"updated_at":        pvs.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert policy version signature: %w", err)
	}

	return nil
}

func (pvss *PolicyVersionSignatures) LoadByPolicyVersionID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	policyVersionID gid.GID,
	cursor *page.Cursor[PolicyVersionSignatureOrderField],
) error {
	q := `
SELECT
	id,
	policy_version_id,
	state,
	signed_by,
	signed_at,
	requested_at,
	requested_by,
	created_at,
	updated_at
FROM
	policy_version_signatures
WHERE
	%s
	AND policy_version_id = @policy_version_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"policy_version_id": policyVersionID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query policy version signatures: %w", err)
	}

	policyVersionSignatures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[PolicyVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect policy version signatures: %w", err)
	}

	*pvss = policyVersionSignatures

	return nil
}

func (pvs *PolicyVersionSignature) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE policy_version_signatures
SET
	state = @state,
	signed_by = @signed_by,
	signed_at = @signed_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         pvs.ID,
		"state":      pvs.State,
		"signed_by":  pvs.SignedBy,
		"signed_at":  pvs.SignedAt,
		"updated_at": pvs.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update policy version signature: %w", err)
	}

	return nil
}
