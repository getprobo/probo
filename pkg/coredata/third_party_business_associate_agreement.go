// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	ThirdPartyBusinessAssociateAgreement struct {
		ID             gid.GID    `db:"id"`
		OrganizationID gid.GID    `db:"organization_id"`
		ThirdPartyID   gid.GID    `db:"third_party_id"`
		ValidFrom      *time.Time `db:"valid_from"`
		ValidUntil     *time.Time `db:"valid_until"`
		FileID         gid.GID    `db:"file_id"`
		CreatedAt      time.Time  `db:"created_at"`
		UpdatedAt      time.Time  `db:"updated_at"`
	}

	ThirdPartyBusinessAssociateAgreements []*ThirdPartyBusinessAssociateAgreement
)

func (v ThirdPartyBusinessAssociateAgreement) CursorKey(orderBy ThirdPartyBusinessAssociateAgreementOrderField) page.CursorKey {
	switch orderBy {
	case ThirdPartyBusinessAssociateAgreementOrderFieldValidFrom:
		return page.NewCursorKey(v.ID, v.ValidFrom)
	case ThirdPartyBusinessAssociateAgreementOrderFieldCreatedAt:
		return page.NewCursorKey(v.ID, v.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (vbaa *ThirdPartyBusinessAssociateAgreement) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM third_party_business_associate_agreements WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, vbaa.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query thirdParty business associate agreement authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (vbaa *ThirdPartyBusinessAssociateAgreement) LoadByThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	valid_from,
	valid_until,
	file_id,
	created_at,
	updated_at
FROM
	third_party_business_associate_agreements
WHERE
	%s
	AND third_party_id = @third_party_id
	AND snapshot_id IS NULL
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"third_party_id": thirdPartyID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParty business associate agreement: %w", err)
	}

	thirdPartyBusinessAssociateAgreement, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ThirdPartyBusinessAssociateAgreement])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParty business associate agreement: %w", err)
	}

	*vbaa = thirdPartyBusinessAssociateAgreement

	return nil
}

func (vbaas *ThirdPartyBusinessAssociateAgreements) LoadByThirdPartyIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyIDs []gid.GID,
) error {
	if len(thirdPartyIDs) == 0 {
		*vbaas = ThirdPartyBusinessAssociateAgreements{}
		return nil
	}

	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	valid_from,
	valid_until,
	file_id,
	created_at,
	updated_at
FROM
	third_party_business_associate_agreements
WHERE
	%s
	AND third_party_id = ANY(@third_party_ids)
	AND snapshot_id IS NULL
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
		return fmt.Errorf("cannot query thirdParty business associate agreements: %w", err)
	}

	agreements, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdPartyBusinessAssociateAgreement])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParty business associate agreements: %w", err)
	}

	*vbaas = agreements

	return nil
}

func (vbaa *ThirdPartyBusinessAssociateAgreement) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyBusinessAssociateAgreementID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	valid_from,
	valid_until,
	file_id,
	created_at,
	updated_at
FROM
	third_party_business_associate_agreements
WHERE
	%s
	AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"id": thirdPartyBusinessAssociateAgreementID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParty business associate agreement: %w", err)
	}

	thirdPartyBusinessAssociateAgreement, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ThirdPartyBusinessAssociateAgreement])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParty business associate agreement: %w", err)
	}

	*vbaa = thirdPartyBusinessAssociateAgreement

	return nil
}

func (vbaa *ThirdPartyBusinessAssociateAgreement) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE
	third_party_business_associate_agreements
SET
	valid_from = @valid_from,
	valid_until = @valid_until,
	file_id = @file_id,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":          vbaa.ID,
		"valid_from":  vbaa.ValidFrom,
		"valid_until": vbaa.ValidUntil,
		"file_id":     vbaa.FileID,
		"updated_at":  vbaa.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update thirdParty business associate agreement: %w", err)
	}

	return nil
}

func (vbaa *ThirdPartyBusinessAssociateAgreement) Upsert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO
	third_party_business_associate_agreements (
		id,
		tenant_id,
		organization_id,
		third_party_id,
		valid_from,
		valid_until,
		file_id,
		created_at,
		updated_at
	)
VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@third_party_id,
	@valid_from,
	@valid_until,
	@file_id,
	@created_at,
	@updated_at
)
ON CONFLICT (organization_id, third_party_id) DO UPDATE SET
	id = EXCLUDED.id,
	valid_from = EXCLUDED.valid_from,
	valid_until = EXCLUDED.valid_until,
	file_id = EXCLUDED.file_id,
	updated_at = EXCLUDED.updated_at
`
	args := pgx.StrictNamedArgs{
		"id":              vbaa.ID,
		"tenant_id":       scope.GetTenantID(),
		"third_party_id":  vbaa.ThirdPartyID,
		"organization_id": vbaa.OrganizationID,
		"valid_from":      vbaa.ValidFrom,
		"valid_until":     vbaa.ValidUntil,
		"file_id":         vbaa.FileID,
		"created_at":      vbaa.CreatedAt,
		"updated_at":      vbaa.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "third_party_business_associate_agreements_source_id_snapshot_id_key" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot upsert thirdParty business associate agreement: %w", err)
	}
	return nil
}

func (vbaa *ThirdPartyBusinessAssociateAgreement) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE
FROM
	third_party_business_associate_agreements
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": vbaa.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (vbaa *ThirdPartyBusinessAssociateAgreement) DeleteByThirdPartyID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	thirdPartyID gid.GID,
) error {
	q := `
DELETE
FROM
	third_party_business_associate_agreements
WHERE
	%s
	AND third_party_id = @third_party_id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"third_party_id": thirdPartyID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}
