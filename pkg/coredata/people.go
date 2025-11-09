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

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	People struct {
		ID                       gid.GID    `db:"id"`
		OrganizationID           gid.GID    `db:"organization_id"`
		Kind                     PeopleKind `db:"kind"`
		FullName                 string     `db:"full_name"`
		PrimaryEmailAddress      string     `db:"primary_email_address"`
		AdditionalEmailAddresses []string   `db:"additional_email_addresses"`
		Position                 *string    `db:"position"`
		ContractStartDate        *time.Time `db:"contract_start_date"`
		ContractEndDate          *time.Time `db:"contract_end_date"`
		CreatedAt                time.Time  `db:"created_at"`
		UpdatedAt                time.Time  `db:"updated_at"`
	}

	Peoples []*People

	ErrPeopleNotFound struct {
		Identifier string
	}

	ErrPeopleAlreadyExists struct {
		message string
	}
)

func (e ErrPeopleNotFound) Error() string {
	return fmt.Sprintf("people not found: %s", e.Identifier)
}

func (e ErrPeopleAlreadyExists) Error() string {
	return e.message
}

func (p People) CursorKey(orderBy PeopleOrderField) page.CursorKey {
	switch orderBy {
	case PeopleOrderFieldCreatedAt:
		return page.NewCursorKey(p.ID, p.CreatedAt)
	case PeopleOrderFieldFullName:
		return page.NewCursorKey(p.ID, p.FullName)
	case PeopleOrderFieldKind:
		return page.NewCursorKey(p.ID, p.Kind)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (p *People) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	peopleID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    kind,
    full_name,
    primary_email_address,
    additional_email_addresses,
    position,
    contract_start_date,
    contract_end_date,
    created_at,
    updated_at
FROM
    peoples
WHERE
    %s
    AND id = @people_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"people_id": peopleID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query people: %w", err)
	}

	people, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[People])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrPeopleNotFound{Identifier: peopleID.String()}
		}

		return fmt.Errorf("cannot collect people: %w", err)
	}

	*p = people

	return nil
}

func (p *People) LoadByEmail(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	primaryEmailAddress string,
) error {
	q := `
	SELECT
		id,
		organization_id,
		kind,
		full_name,
		primary_email_address,
		additional_email_addresses,
		position,
		contract_start_date,
		contract_end_date,
		created_at,
		updated_at
	FROM
		peoples
	WHERE
		%s
		AND primary_email_address = @primary_email_address
	LIMIT 1;
	`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"primary_email_address": primaryEmailAddress}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query people: %w", err)
	}

	people, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[People])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrPeopleNotFound{Identifier: primaryEmailAddress}
		}

		return fmt.Errorf("cannot collect people: %w", err)
	}

	*p = people

	return nil
}

func (p *Peoples) LoadByIDs(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	peopleIDs []gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    kind,
    full_name,
    primary_email_address,
    additional_email_addresses,
    position,
    contract_start_date,
    contract_end_date,
    created_at,
    updated_at
FROM
    peoples
WHERE
    %s
    AND id = ANY(@people_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"people_ids": peopleIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query people: %w", err)
	}

	peoples, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[People])
	if err != nil {
		return fmt.Errorf("cannot collect people: %w", err)
	}

	*p = peoples

	return nil
}

func (p People) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    peoples (
        tenant_id,
        id,
        organization_id,
        kind,
        full_name,
        primary_email_address,
        additional_email_addresses,
        position,
        contract_start_date,
        contract_end_date,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @people_id,
    @organization_id,
    @kind,
    @full_name,
    @primary_email_address,
    @additional_email_addresses,
    @position,
    @contract_start_date,
    @contract_end_date,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":                  scope.GetTenantID(),
		"people_id":                  p.ID,
		"organization_id":            p.OrganizationID,
		"kind":                       p.Kind,
		"full_name":                  p.FullName,
		"primary_email_address":      p.PrimaryEmailAddress,
		"additional_email_addresses": p.AdditionalEmailAddresses,
		"position":                   p.Position,
		"contract_start_date":        p.ContractStartDate,
		"contract_end_date":          p.ContractEndDate,
		"created_at":                 p.CreatedAt,
		"updated_at":                 p.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (p People) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM peoples WHERE %s AND id = @people_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"people_id": p.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (p *Peoples) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *PeopleFilter,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    peoples
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
		return 0, fmt.Errorf("cannot count people: %w", err)
	}

	return count, nil
}

func (p *Peoples) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[PeopleOrderField],
	filter *PeopleFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
    kind,
    full_name,
    primary_email_address,
    additional_email_addresses,
    position,
    contract_start_date,
    contract_end_date,
    created_at,
    updated_at
FROM
    peoples
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, cursor.SQLArguments())
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query people: %w", err)
	}

	peoples, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[People])
	if err != nil {
		return fmt.Errorf("cannot collect people: %w", err)
	}

	*p = peoples

	return nil
}

func (p *People) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE peoples SET
	full_name = @full_name,
	primary_email_address = @primary_email_address,
	additional_email_addresses = @additional_email_addresses,
	kind = @kind,
	position = @position,
	contract_start_date = @contract_start_date,
	contract_end_date = @contract_end_date,
	updated_at = @updated_at
WHERE %s
    AND id = @people_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"people_id":                  p.ID,
		"full_name":                  p.FullName,
		"primary_email_address":      p.PrimaryEmailAddress,
		"additional_email_addresses": p.AdditionalEmailAddresses,
		"kind":                       p.Kind,
		"position":                   p.Position,
		"contract_start_date":        p.ContractStartDate,
		"contract_end_date":          p.ContractEndDate,
		"updated_at":                 p.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update people: %w", err)
	}

	return nil
}

func (p *Peoples) LoadAwaitingSigning(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
WITH signatories AS (
	SELECT
		signed_by
	FROM
		document_version_signatures
	WHERE
	    %s
		AND state = 'REQUESTED'
	GROUP BY
		signed_by
)
SELECT
	id,
	organization_id,
	kind,
	full_name,
	primary_email_address,
	additional_email_addresses,
	position,
	contract_start_date,
	contract_end_date,
	created_at,
	updated_at
FROM
	peoples
INNER JOIN signatories ON peoples.id = signatories.signed_by
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err := conn.Query(ctx, q, scope.SQLArguments())
	if err != nil {
		return fmt.Errorf("cannot query people: %w", err)
	}

	peoples, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[People])
	if err != nil {
		return fmt.Errorf("cannot collect people: %w", err)
	}

	*p = peoples

	return nil
}

func (p *Peoples) LoadByMeetingID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	meetingID gid.GID,
) error {
	q := `
WITH people_attendees AS (
	SELECT
		p.id,
		p.organization_id,
		p.kind,
		p.full_name,
		p.primary_email_address,
		p.additional_email_addresses,
		p.position,
		p.contract_start_date,
		p.contract_end_date,
		p.created_at,
		p.updated_at,
		p.tenant_id,
		ma.created_at AS attendee_created_at
	FROM
		peoples p
	INNER JOIN
		meeting_attendees ma ON p.id = ma.attendee_id
	WHERE
		ma.meeting_id = @meeting_id
)
SELECT
	id,
	organization_id,
	kind,
	full_name,
	primary_email_address,
	additional_email_addresses,
	position,
	contract_start_date,
	contract_end_date,
	created_at,
	updated_at
FROM
	people_attendees
WHERE
	%s
ORDER BY
	attendee_created_at ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"meeting_id": meetingID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query people: %w", err)
	}

	peoples, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[People])
	if err != nil {
		return fmt.Errorf("cannot collect people: %w", err)
	}

	*p = peoples

	return nil
}
