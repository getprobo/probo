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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	ThirdPartyService struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		ThirdPartyID   gid.GID   `db:"third_party_id"`
		Name           string    `db:"name"`
		Description    *string   `db:"description"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	ThirdPartyServices []*ThirdPartyService
)

func (vs ThirdPartyService) CursorKey(orderBy ThirdPartyServiceOrderField) page.CursorKey {
	switch orderBy {
	case ThirdPartyServiceOrderFieldCreatedAt:
		return page.CursorKey{ID: vs.ID, Value: vs.CreatedAt}
	case ThirdPartyServiceOrderFieldName:
		return page.CursorKey{ID: vs.ID, Value: vs.Name}
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (vs *ThirdPartyService) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM third_party_services WHERE id = ANY(@resource_ids::text[])`

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

func (vs *ThirdPartyService) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyServiceID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	name,
	description,
	created_at,
	updated_at
FROM
	third_party_services
WHERE
	%s
	AND id = @third_party_service_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"third_party_service_id": thirdPartyServiceID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParty service: %w", err)
	}
	defer rows.Close()

	thirdPartyService, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ThirdPartyService])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect thirdParty service: %w", err)
	}

	*vs = thirdPartyService

	return nil
}

func (vs *ThirdPartyServices) LoadByThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyID gid.GID,
	cursor *page.Cursor[ThirdPartyServiceOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	name,
	description,
	created_at,
	updated_at
FROM
	third_party_services
WHERE
	%s
	AND third_party_id = @third_party_id
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"third_party_id": thirdPartyID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParty services: %w", err)
	}
	defer rows.Close()

	thirdPartyServices, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdPartyService])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParty services: %w", err)
	}

	*vs = thirdPartyServices

	return nil
}

func (vs *ThirdPartyServices) LoadByThirdPartyIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyIDs []gid.GID,
) error {
	if len(thirdPartyIDs) == 0 {
		*vs = ThirdPartyServices{}
		return nil
	}

	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	name,
	description,
	created_at,
	updated_at
FROM
	third_party_services
WHERE
	%s
	AND third_party_id = ANY(@third_party_ids)
ORDER BY
	third_party_id, name ASC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	ids := make([]string, len(thirdPartyIDs))
	for i, id := range thirdPartyIDs {
		ids[i] = id.String()
	}

	args := pgx.StrictNamedArgs{"third_party_ids": ids}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParty services: %w", err)
	}
	defer rows.Close()

	thirdPartyServices, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdPartyService])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParty services: %w", err)
	}

	*vs = thirdPartyServices

	return nil
}

func (vs ThirdPartyService) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
	third_party_services (
		tenant_id,
		id,
		organization_id,
		third_party_id,
		name,
		description,
		created_at,
		updated_at
	)
VALUES (
	@tenant_id,
	@third_party_service_id,
	@organization_id,
	@third_party_id,
	@name,
	@description,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":              scope.GetTenantID(),
		"third_party_service_id": vs.ID,
		"organization_id":        vs.OrganizationID,
		"third_party_id":         vs.ThirdPartyID,
		"name":                   vs.Name,
		"description":            vs.Description,
		"created_at":             vs.CreatedAt,
		"updated_at":             vs.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert thirdParty service: %w", err)
	}

	return nil
}

func (vs ThirdPartyService) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE
	third_party_services
SET
	name = @name,
	description = @description,
	updated_at = @updated_at
WHERE
	%s
	AND id = @third_party_service_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"third_party_service_id": vs.ID,
		"name":                   vs.Name,
		"description":            vs.Description,
		"updated_at":             vs.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update thirdParty service: %w", err)
	}

	return nil
}

func (vs ThirdPartyService) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM
	third_party_services
WHERE
	%s
	AND id = @third_party_service_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"third_party_service_id": vs.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete thirdParty service: %w", err)
	}

	return nil
}
