// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

func (vc *ThirdPartyContact) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM third_party_contacts WHERE id = ANY(@resource_ids::text[])`

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

func (vc *ThirdPartyContacts) LoadByThirdPartyIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyIDs []gid.GID,
) error {
	if len(thirdPartyIDs) == 0 {
		*vc = ThirdPartyContacts{}
		return nil
	}

	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	full_name,
	email,
	phone,
	role,
	created_at,
	updated_at
FROM
	third_party_contacts
WHERE
	%s
	AND third_party_id = ANY(@third_party_ids)
ORDER BY
	third_party_id, full_name ASC
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
