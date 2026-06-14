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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	Invitation struct {
		ID             gid.GID          `db:"id"`
		OrganizationID gid.GID          `db:"organization_id"`
		UserID         gid.GID          `db:"user_id"`
		Status         InvitationStatus `db:"status"`
		ExpiresAt      time.Time        `db:"expires_at"`
		AcceptedAt     *time.Time       `db:"accepted_at"`
		CreatedAt      time.Time        `db:"created_at"`
	}

	Invitations []*Invitation
)

func (i Invitation) CursorKey(orderBy InvitationOrderField) page.CursorKey {
	switch orderBy {
	case InvitationOrderFieldCreatedAt:
		return page.NewCursorKey(i.ID, i.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (i *Invitation) Insert(ctx context.Context, conn pg.Tx, predicate Predicater) error {
	query := `
INSERT INTO
    iam_invitations (
        tenant_id,
		organization_id,
		user_id,
        id,
        expires_at,
        created_at
    )
VALUES (
    @tenant_id,
	@organization_id,
	@user_id,
    @id,
    @expires_at,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       predicate.GetTenantID(),
		"organization_id": i.OrganizationID,
		"id":              i.ID,
		"user_id":         i.UserID,
		"expires_at":      i.ExpiresAt,
		"created_at":      i.CreatedAt,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot create invitation: %w", err)
	}

	return nil
}

func (i *Invitation) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
	id gid.GID,
) error {
	query := `
SELECT
    id,
	organization_id,
	user_id,
    CASE
        WHEN accepted_at IS NOT NULL THEN 'ACCEPTED'
        WHEN expires_at < NOW() THEN 'EXPIRED'
        ELSE 'PENDING'
    END as status,
    expires_at,
    accepted_at,
    created_at
FROM
    iam_invitations
WHERE
    id = @id
    AND %s
`

	query = fmt.Sprintf(query, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id": id,
	}
	maps.Copy(args, predicate.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query invitation: %w", err)
	}

	invitation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Invitation])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect invitation: %w", err)
	}

	*i = invitation

	return nil
}

// AuthorizationAttributes loads the minimal authorization attributes for policy condition evaluation.
// It is intentionally lightweight and does not populate the Invitation struct.
func (i *Invitation) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
    id,
    email,
    organization_id
FROM
    iam_invitations
WHERE
    id = ANY(@resource_ids::text[])
`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query invitation authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID, len(resourceIDs))

	for rows.Next() {
		var (
			id             gid.GID
			email          string
			organizationID gid.GID
		)

		err = rows.Scan(&id, &email, &organizationID)
		if err != nil {
			return nil, fmt.Errorf("cannot scan invitation authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"email":           email,
			"organization_id": organizationID.String(),
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate invitation authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (i *Invitation) Update(ctx context.Context, conn pg.Tx, predicate Predicater) error {
	query := `
UPDATE
    iam_invitations
SET
    accepted_at = @accepted_at
WHERE
    id = @id
    AND %s
`

	query = fmt.Sprintf(query, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":          i.ID,
		"accepted_at": i.AcceptedAt,
	}
	maps.Copy(args, predicate.SQLArguments())

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot update invitation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (i *Invitation) Delete(ctx context.Context, conn pg.Tx, predicate Predicater, invitationID gid.GID) error {
	query := `
DELETE FROM
    iam_invitations
WHERE
    %s
    AND id = @invitation_id
`

	query = fmt.Sprintf(query, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{
		"invitation_id": invitationID,
	}
	maps.Copy(args, predicate.SQLArguments())

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot delete invitation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (i *Invitations) LoadByUserID(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
	userID gid.GID,
	cursor *page.Cursor[InvitationOrderField],
	filter *InvitationFilter,
) error {
	query := `
SELECT
    id,
    organization_id,
	user_id,
    CASE
        WHEN accepted_at IS NOT NULL THEN 'ACCEPTED'
        WHEN expires_at < NOW() THEN 'EXPIRED'
        ELSE 'PENDING'
    END as status,
    expires_at,
    accepted_at,
    created_at
FROM
    iam_invitations
WHERE
    user_id = @user_id
    AND %s
    AND %s
    AND %s
`

	query = fmt.Sprintf(query, predicate.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"user_id": userID,
	}
	maps.Copy(args, predicate.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query invitations: %w", err)
	}

	invitations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Invitation])
	if err != nil {
		return fmt.Errorf("cannot collect invitations: %w", err)
	}

	*i = invitations

	return nil
}

func (i *Invitations) ExpireByUserID(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
	userID gid.GID,
	filter *InvitationFilter,
) error {
	q := `
	UPDATE
	    iam_invitations
	SET
	    expires_at = NOW()
	WHERE
	    user_id = @user_id
	    AND %s
	    AND %s
	`

	q = fmt.Sprintf(q, predicate.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{
		"user_id": userID,
	}
	maps.Copy(args, predicate.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot expire invitations: %w", err)
	}

	return nil
}
