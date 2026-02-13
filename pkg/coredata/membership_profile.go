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
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type (
	MembershipProfile struct {
		ID                       gid.GID               `db:"id"`
		IdentityID               gid.GID               `db:"identity_id"`
		OrganizationID           gid.GID               `db:"organization_id"`
		MembershipID             gid.GID               `db:"membership_id"`
		EmailAddress             mail.Addr             `db:"email_address"`
		FullName                 string                `db:"full_name"`
		Kind                     MembershipProfileKind `db:"kind"`
		AdditionalEmailAddresses mail.Addrs            `db:"additional_email_addresses"`
		Position                 *string               `db:"position"`
		ContractStartDate        *time.Time            `db:"contract_start_date"`
		ContractEndDate          *time.Time            `db:"contract_end_date"`
		CreatedAt                time.Time             `db:"created_at"`
		UpdatedAt                time.Time             `db:"updated_at"`
	}

	MembershipProfiles []*MembershipProfile
)

func (p MembershipProfile) CursorKey(orderBy MembershipProfileOrderField) page.CursorKey {
	switch orderBy {
	case MembershipProfileOrderFieldCreatedAt:
		return page.NewCursorKey(p.ID, p.CreatedAt)
	case MembershipProfileOrderFieldFullName:
		return page.NewCursorKey(p.ID, p.FullName)
	case MembershipProfileOrderFieldKind:
		return page.NewCursorKey(p.ID, p.Kind)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (p *MembershipProfile) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT m.organization_id, mp.identity_id FROM iam_membership_profiles mp JOIN iam_memberships m ON mp.membership_id = m.id WHERE mp.id = $1 LIMIT 1;`

	var organizationID gid.GID
	var identityID gid.GID
	if err := conn.QueryRow(ctx, q, p.ID).Scan(&organizationID, &identityID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query membership profile authorization attributes: %w", err)
	}

	return map[string]string{
		"organization_id": organizationID.String(),
		"identity_id":     identityID.String(),
	}, nil
}

func (p *MembershipProfile) LoadByMembershipID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	membershipID gid.GID,
) error {
	q := `
SELECT
    p.id,
    p.identity_id,
    p.organization_id,
    p.membership_id,
    i.email_address,
    p.full_name,
    p.kind,
    p.additional_email_addresses,
    p.position,
    p.contract_start_date,
    p.contract_end_date,
    p.created_at,
    p.updated_at
FROM
    iam_membership_profiles p
INNER JOIN identities i
    ON i.id = p.identity_id
WHERE
    p.%s
    AND p.membership_id = @membership_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"membership_id": membershipID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query profile: %w", err)
	}

	profile, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[MembershipProfile])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect profile: %w", err)
	}

	*p = profile

	return nil
}

func (p *MembershipProfile) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	profileID gid.GID,
) error {
	q := `
SELECT
    p.id,
    p.identity_id,
    p.organization_id,
    p.membership_id,
    i.email_address,
    p.full_name,
    p.kind,
    p.additional_email_addresses,
    p.position,
    p.contract_start_date,
    p.contract_end_date,
    p.created_at,
    p.updated_at
FROM
    iam_membership_profiles p
INNER JOIN identities i
    ON i.id = p.identity_id
WHERE
    p.%s
    AND p.id = @profile_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"profile_id": profileID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query profile: %w", err)
	}

	profile, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[MembershipProfile])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect profile: %w", err)
	}

	*p = profile

	return nil
}

func (p *MembershipProfile) LoadByIdentityIDAndOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	identityID gid.GID,
	organizationID gid.GID,
) error {
	q := `
SELECT
    p.id,
    p.identity_id,
    p.organization_id,
    p.membership_id,
    i.email_address,
    p.full_name,
    p.kind,
    p.additional_email_addresses,
    p.position,
    p.contract_start_date,
    p.contract_end_date,
    p.created_at,
    p.updated_at
FROM
    iam_membership_profiles p
INNER JOIN identities i
    ON i.id = p.identity_id
WHERE
    p.%s
    AND p.identity_id = @identity_id
    AND p.organization_id = @organization_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"identity_id":     identityID,
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query profile: %w", err)
	}

	profile, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[MembershipProfile])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect profile: %w", err)
	}

	*p = profile

	return nil
}

func (p *MembershipProfiles) LoadByIDs(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	profileIDs []gid.GID,
) error {
	q := `
SELECT
    p.id,
    p.identity_id,
    p.organization_id,
    p.membership_id,
    i.email_address,
    p.full_name,
    p.kind,
    p.additional_email_addresses,
    p.position,
    p.contract_start_date,
    p.contract_end_date,
    p.created_at,
    p.updated_at
FROM
    iam_membership_profiles p
INNER JOIN identities i
    ON i.id = p.identity_id
WHERE
    p.%s
    AND p.id = ANY(@profile_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"profile_ids": profileIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query profiles: %w", err)
	}

	profiles, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MembershipProfile])
	if err != nil {
		return fmt.Errorf("cannot collect profiles: %w", err)
	}

	*p = profiles

	return nil
}

func (p *MembershipProfiles) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[MembershipProfileOrderField],
	filter *MembershipProfileFilter,
) error {
	q := `
WITH profiles AS (
    SELECT
        id,
        identity_id,
        organization_id,
        membership_id,
        full_name,
        kind,
        additional_email_addresses,
        position,
        contract_start_date,
        contract_end_date,
        created_at,
        updated_at
    FROM
        iam_membership_profiles
    WHERE
        %s
        AND organization_id = @organization_id
        AND %s
        AND %s
)
SELECT
    p.id,
    p.identity_id,
    p.organization_id,
    p.membership_id,
    i.email_address,
    p.full_name,
    p.kind,
    p.additional_email_addresses,
    p.position,
    p.contract_start_date,
    p.contract_end_date,
    p.created_at,
    p.updated_at
FROM profiles p
INNER JOIN identities i ON i.id = p.identity_id
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query profiles: %w", err)
	}

	profiles, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MembershipProfile])
	if err != nil {
		return fmt.Errorf("cannot collect profiles: %w", err)
	}

	*p = profiles

	return nil
}

func (p *MembershipProfiles) LoadByDocumentID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentID gid.GID,
	cursor *page.Cursor[MembershipProfileOrderField],
) error {
	q := `
WITH profiles AS (
    SELECT
        mp.id,
        mp.identity_id,
        mp.organization_id,
        mp.membership_id,
        mp.full_name,
        mp.kind,
        mp.additional_email_addresses,
        mp.position,
        mp.contract_start_date,
        mp.contract_end_date,
        mp.created_at,
        mp.updated_at
    FROM
        iam_membership_profiles mp
    WHERE
        mp.%s
        AND mp.id IN (
            SELECT approver_profile_id
            FROM document_approvers
            WHERE document_id = @document_id
        )
        AND %s
)
SELECT
    p.id,
    p.identity_id,
    p.organization_id,
    p.membership_id,
    i.email_address,
    p.full_name,
    p.kind,
    p.additional_email_addresses,
    p.position,
    p.contract_start_date,
    p.contract_end_date,
    p.created_at,
    p.updated_at
FROM profiles p
INNER JOIN identities i ON i.id = p.identity_id
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"document_id": documentID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document approver profiles: %w", err)
	}

	profiles, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MembershipProfile])
	if err != nil {
		return fmt.Errorf("cannot collect document approver profiles: %w", err)
	}

	*p = profiles

	return nil
}

func (p *MembershipProfiles) CountByDocumentID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    iam_membership_profiles mp
INNER JOIN document_approvers da ON mp.id = da.approver_profile_id
WHERE
    mp.%s
    AND da.document_id = @document_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_id": documentID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	err := conn.QueryRow(ctx, q, args).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot query document approver profiles count: %w", err)
	}

	return count, nil
}

func (p *MembershipProfiles) LoadByDocumentVersionID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentVersionID gid.GID,
	cursor *page.Cursor[MembershipProfileOrderField],
) error {
	q := `
WITH profiles AS (
    SELECT
        mp.id,
        mp.identity_id,
        mp.organization_id,
        mp.membership_id,
        mp.full_name,
        mp.kind,
        mp.additional_email_addresses,
        mp.position,
        mp.contract_start_date,
        mp.contract_end_date,
        mp.created_at,
        mp.updated_at
    FROM
        iam_membership_profiles mp
    WHERE
        mp.%s
        AND mp.id IN (
            SELECT approver_profile_id
            FROM document_version_approvers
            WHERE document_version_id = @document_version_id
        )
        AND %s
)
SELECT
    p.id,
    p.identity_id,
    p.organization_id,
    p.membership_id,
    i.email_address,
    p.full_name,
    p.kind,
    p.additional_email_addresses,
    p.position,
    p.contract_start_date,
    p.contract_end_date,
    p.created_at,
    p.updated_at
FROM profiles p
INNER JOIN identities i ON i.id = p.identity_id
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"document_version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version approver profiles: %w", err)
	}

	profiles, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MembershipProfile])
	if err != nil {
		return fmt.Errorf("cannot collect document version approver profiles: %w", err)
	}

	*p = profiles

	return nil
}

func (p *MembershipProfiles) CountByDocumentVersionID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentVersionID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    iam_membership_profiles mp
INNER JOIN document_version_approvers dva ON mp.id = dva.approver_profile_id
WHERE
    mp.%s
    AND dva.document_version_id = @document_version_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	err := conn.QueryRow(ctx, q, args).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot query document version approver profiles count: %w", err)
	}

	return count, nil
}

func (p *MembershipProfiles) LoadByMeetingID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	meetingID gid.GID,
) error {
	q := `
WITH attendees AS (
    SELECT
        p.id,
        p.tenant_id,
        p.identity_id,
        p.organization_id,
        p.membership_id,
        i.email_address,
        p.full_name,
        p.kind,
        p.additional_email_addresses,
        p.position,
        p.contract_start_date,
        p.contract_end_date,
        p.created_at,
        p.updated_at,
        ma.created_at AS attendee_created_at
    FROM
        iam_membership_profiles p
    INNER JOIN identities i
        ON i.id = p.identity_id
    INNER JOIN
        meeting_attendees ma ON p.id = ma.attendee_profile_id
    WHERE
        ma.meeting_id = @meeting_id
)
SELECT
    id,
    identity_id,
    organization_id,
    membership_id,
    kind,
    email_address,
    full_name,
    additional_email_addresses,
    position,
    contract_start_date,
    contract_end_date,
    created_at,
    updated_at
FROM
    attendees
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
		return fmt.Errorf("cannot query profiles: %w", err)
	}

	profiles, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MembershipProfile])
	if err != nil {
		return fmt.Errorf("cannot collect profiles: %w", err)
	}

	*p = profiles

	return nil
}

func (p *MembershipProfiles) LoadAwaitingSigning(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
WITH signatories AS (
    SELECT
        signed_by_profile_id
    FROM
        document_version_signatures
    WHERE
        %s
        AND state = 'REQUESTED'
    GROUP BY
        signed_by_profile_id
)
SELECT
    p.id,
    p.identity_id,
    p.organization_id,
    p.membership_id,
    p.kind,
    p.full_name,
    i.email_address,
    p.additional_email_addresses,
    p.position,
    p.contract_start_date,
    p.contract_end_date,
    p.created_at,
    p.updated_at
FROM
    iam_membership_profiles p
INNER JOIN identities i
    ON i.id = p.identity_id
INNER JOIN signatories ON p.id = signatories.signed_by_profile_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err := conn.Query(ctx, q, scope.SQLArguments())
	if err != nil {
		return fmt.Errorf("cannot query profiles: %w", err)
	}

	profiles, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MembershipProfile])
	if err != nil {
		return fmt.Errorf("cannot collect profiles: %w", err)
	}

	*p = profiles

	return nil
}

func (p *MembershipProfiles) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *MembershipProfileFilter,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    iam_membership_profiles
WHERE
    %s
	AND %s
    AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot query iam_membership_profiles: %w", err)
	}

	var count int
	err = rows.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot collect count: %w", err)
	}

	return count, nil
}

func (p *MembershipProfile) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO
    iam_membership_profiles (
        tenant_id,
        id,
        identity_id,
        organization_id,
        membership_id,
        full_name,
        kind,
        additional_email_addresses,
        position,
        contract_start_date,
        contract_end_date,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @id,
    @identity_id,
    @organization_id,
    @membership_id,
    @full_name,
    @kind,
    COALESCE(@additional_email_addresses, '{}'::CITEXT[]),
    @position,
    @contract_start_date,
    @contract_end_date,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":                  p.ID.TenantID().String(),
		"id":                         p.ID,
		"identity_id":                p.IdentityID,
		"organization_id":            p.OrganizationID,
		"membership_id":              p.MembershipID,
		"full_name":                  p.FullName,
		"kind":                       p.Kind,
		"additional_email_addresses": p.AdditionalEmailAddresses,
		"position":                   p.Position,
		"contract_start_date":        p.ContractStartDate,
		"contract_end_date":          p.ContractEndDate,
		"created_at":                 p.CreatedAt,
		"updated_at":                 p.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert profile: %w", err)
	}

	return nil
}

func (p *MembershipProfile) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE
    iam_membership_profiles
SET
    full_name = @full_name,
    kind = @kind,
    additional_email_addresses = @additional_email_addresses,
    position = @position,
    contract_start_date = @contract_start_date,
    contract_end_date = @contract_end_date,
    updated_at = @updated_at
WHERE
    id = @id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                         p.ID,
		"full_name":                  p.FullName,
		"kind":                       p.Kind,
		"additional_email_addresses": p.AdditionalEmailAddresses,
		"position":                   p.Position,
		"contract_start_date":        p.ContractStartDate,
		"contract_end_date":          p.ContractEndDate,
		"updated_at":                 p.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update profile: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (p *MembershipProfile) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	profileID gid.GID,
) error {
	q := `
DELETE FROM
    iam_membership_profiles
WHERE
    id = @profile_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"profile_id": profileID}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete profile: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
