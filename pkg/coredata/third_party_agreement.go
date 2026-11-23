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
	ThirdPartyAgreement struct {
		ID             gid.GID                 `db:"id"`
		OrganizationID gid.GID                 `db:"organization_id"`
		ThirdPartyID   gid.GID                 `db:"third_party_id"`
		Title          string                  `db:"title"`
		Type           ThirdPartyAgreementType `db:"type"`
		ValidFrom      *time.Time              `db:"valid_from"`
		ValidUntil     *time.Time              `db:"valid_until"`
		FileID         gid.GID                 `db:"file_id"`
		CreatedAt      time.Time               `db:"created_at"`
		UpdatedAt      time.Time               `db:"updated_at"`
	}

	ThirdPartyAgreements []*ThirdPartyAgreement
)

func (a ThirdPartyAgreement) CursorKey(orderBy ThirdPartyAgreementOrderField) page.CursorKey {
	switch orderBy {
	case ThirdPartyAgreementOrderFieldCreatedAt:
		return page.NewCursorKey(a.ID, a.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (a *ThirdPartyAgreement) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM third_party_agreements WHERE id = ANY(@resource_ids::text[])`

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

func (as *ThirdPartyAgreements) LoadForThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyID gid.GID,
	cursor *page.Cursor[ThirdPartyAgreementOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	title,
	type,
	valid_from,
	valid_until,
	file_id,
	created_at,
	updated_at
FROM
	third_party_agreements
WHERE
	%s
	AND third_party_id = @third_party_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"third_party_id": thirdPartyID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query third party agreements: %w", err)
	}

	agreements, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdPartyAgreement])
	if err != nil {
		return fmt.Errorf("cannot collect third party agreements: %w", err)
	}

	*as = agreements

	return nil
}

func (as *ThirdPartyAgreements) LoadByThirdPartyIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyIDs []gid.GID,
) error {
	if len(thirdPartyIDs) == 0 {
		*as = ThirdPartyAgreements{}
		return nil
	}

	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	title,
	type,
	valid_from,
	valid_until,
	file_id,
	created_at,
	updated_at
FROM
	third_party_agreements
WHERE
	%s
	AND third_party_id = ANY(@third_party_ids)
ORDER BY
	third_party_id, created_at DESC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	ids := make([]string, len(thirdPartyIDs))
	for i, id := range thirdPartyIDs {
		ids[i] = id.String()
	}

	args := pgx.NamedArgs{"third_party_ids": ids}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query third party agreements: %w", err)
	}

	agreements, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdPartyAgreement])
	if err != nil {
		return fmt.Errorf("cannot collect third party agreements: %w", err)
	}

	*as = agreements

	return nil
}

func (a *ThirdPartyAgreement) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	agreementID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	title,
	type,
	valid_from,
	valid_until,
	file_id,
	created_at,
	updated_at
FROM
	third_party_agreements
WHERE
	%s
	AND id = @id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"id": agreementID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query third party agreement: %w", err)
	}

	agreement, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ThirdPartyAgreement])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect third party agreement: %w", err)
	}

	*a = agreement

	return nil
}

func (a *ThirdPartyAgreement) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
	third_party_agreements (
		id,
		organization_id,
		tenant_id,
		third_party_id,
		title,
		type,
		valid_from,
		valid_until,
		file_id,
		created_at,
		updated_at
	)
VALUES (
	@id,
	@organization_id,
	@tenant_id,
	@third_party_id,
	@title,
	@type,
	@valid_from,
	@valid_until,
	@file_id,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              a.ID,
		"organization_id": a.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"third_party_id":  a.ThirdPartyID,
		"title":           a.Title,
		"type":            a.Type,
		"valid_from":      a.ValidFrom,
		"valid_until":     a.ValidUntil,
		"file_id":         a.FileID,
		"created_at":      a.CreatedAt,
		"updated_at":      a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert third party agreement: %w", err)
	}

	return nil
}

func (a *ThirdPartyAgreement) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE
	third_party_agreements
SET
	title = @title,
	type = @type,
	valid_from = @valid_from,
	valid_until = @valid_until,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":          a.ID,
		"title":       a.Title,
		"type":        a.Type,
		"valid_from":  a.ValidFrom,
		"valid_until": a.ValidUntil,
		"updated_at":  a.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update third party agreement: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (a *ThirdPartyAgreement) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM
	third_party_agreements
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": a.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete third party agreement: %w", err)
	}

	return nil
}
