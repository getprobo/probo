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
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type (
	ThirdPartyContact struct {
		ID             gid.GID    `db:"id"`
		OrganizationID gid.GID    `db:"organization_id"`
		ThirdPartyID   gid.GID    `db:"third_party_id"`
		FullName       *string    `db:"full_name"`
		Email          *mail.Addr `db:"email"`
		Phone          *string    `db:"phone"`
		Role           *string    `db:"role"`
		SnapshotID     *gid.GID   `db:"snapshot_id"`
		SourceID       *gid.GID   `db:"source_id"`
		CreatedAt      time.Time  `db:"created_at"`
		UpdatedAt      time.Time  `db:"updated_at"`
	}

	ThirdPartyContacts []*ThirdPartyContact
)

func (vc ThirdPartyContact) CursorKey(orderBy ThirdPartyContactOrderField) page.CursorKey {
	switch orderBy {
	case ThirdPartyContactOrderFieldCreatedAt:
		return page.CursorKey{ID: vc.ID, Value: vc.CreatedAt}
	case ThirdPartyContactOrderFieldFullName:
		return page.CursorKey{ID: vc.ID, Value: vc.FullName}
	case ThirdPartyContactOrderFieldEmail:
		return page.CursorKey{ID: vc.ID, Value: vc.Email}
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (vc *ThirdPartyContact) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM third_party_contacts WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, vc.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query thirdParty contact authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (vc *ThirdPartyContact) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyContactID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	full_name,
	email,
	phone,
	role,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	third_party_contacts
WHERE
	%s
	AND id = @third_party_contact_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"third_party_contact_id": thirdPartyContactID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParty contact: %w", err)
	}
	defer rows.Close()

	thirdPartyContact, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ThirdPartyContact])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect thirdParty contact: %w", err)
	}

	*vc = thirdPartyContact

	return nil
}

func (vc *ThirdPartyContacts) LoadByThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyID gid.GID,
	cursor *page.Cursor[ThirdPartyContactOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	full_name,
	email,
	phone,
	role,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	third_party_contacts
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
		return fmt.Errorf("cannot query thirdParty contacts: %w", err)
	}
	defer rows.Close()

	thirdPartyContacts, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdPartyContact])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParty contacts: %w", err)
	}

	*vc = thirdPartyContacts

	return nil
}

func (vc ThirdPartyContact) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
	third_party_contacts (
		tenant_id,
		id,
		organization_id,
		third_party_id,
		full_name,
		email,
		phone,
		role,
		created_at,
		updated_at
	)
VALUES (
	@tenant_id,
	@third_party_contact_id,
	@organization_id,
	@third_party_id,
	@full_name,
	@email,
	@phone,
	@role,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":              scope.GetTenantID(),
		"third_party_contact_id": vc.ID,
		"organization_id":        vc.OrganizationID,
		"third_party_id":         vc.ThirdPartyID,
		"full_name":              vc.FullName,
		"email":                  vc.Email,
		"phone":                  vc.Phone,
		"role":                   vc.Role,
		"created_at":             vc.CreatedAt,
		"updated_at":             vc.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert thirdParty contact: %w", err)
	}

	return nil
}

func (vc ThirdPartyContact) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE
	third_party_contacts
SET
	full_name = @full_name,
	email = @email,
	phone = @phone,
	role = @role,
	updated_at = @updated_at
WHERE
	%s
	AND id = @third_party_contact_id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"third_party_contact_id": vc.ID,
		"full_name":              vc.FullName,
		"email":                  vc.Email,
		"phone":                  vc.Phone,
		"role":                   vc.Role,
		"updated_at":             vc.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update thirdParty contact: %w", err)
	}

	return nil
}

func (vc ThirdPartyContact) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM
	third_party_contacts
WHERE
	%s
	AND id = @third_party_contact_id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"third_party_contact_id": vc.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete thirdParty contact: %w", err)
	}

	return nil
}

func (vc ThirdPartyContacts) InsertThirdPartySnapshots(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
WITH
	snapshot_third_parties AS (
		SELECT id, source_id
		FROM third_parties
		WHERE organization_id = @organization_id AND snapshot_id = @snapshot_id
	)
INSERT INTO third_party_contacts (
	tenant_id,
	id,
	organization_id,
	snapshot_id,
	source_id,
	third_party_id,
	full_name,
	email,
	phone,
	role,
	created_at,
	updated_at
)
SELECT
	@tenant_id,
	generate_gid(decode_base64_unpadded(@tenant_id), @third_party_contact_entity_type),
	@organization_id,
	@snapshot_id,
	vc.id,
	sv.id,
	vc.full_name,
	vc.email,
	vc.phone,
	vc.role,
	vc.created_at,
	vc.updated_at
FROM third_party_contacts vc
INNER JOIN snapshot_third_parties sv ON sv.source_id = vc.third_party_id
WHERE %s AND vc.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":                       scope.GetTenantID(),
		"snapshot_id":                     snapshotID,
		"organization_id":                 organizationID,
		"third_party_contact_entity_type": ThirdPartyContactEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert thirdParty contact snapshots: %w", err)
	}

	return nil
}
