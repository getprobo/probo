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
)

type (
	IdentityProfile struct {
		ID           gid.GID   `db:"id"`
		IdentityID   gid.GID   `db:"identity_id"`
		MembershipID *gid.GID  `db:"membership_id"`
		FullName     string    `db:"full_name"`
		CreatedAt    time.Time `db:"created_at"`
		UpdatedAt    time.Time `db:"updated_at"`
	}

	IdentityProfiles []*IdentityProfile
)

func (p *IdentityProfile) IsDefault() bool {
	return p.MembershipID == nil
}

// LoadDefaultByIdentityID loads the default profile for an identity (where membership_id is NULL)
func (p *IdentityProfile) LoadDefaultByIdentityID(
	ctx context.Context,
	conn pg.Conn,
	identityID gid.GID,
) error {
	q := `
SELECT
    id,
    identity_id,
    membership_id,
    full_name,
    created_at,
    updated_at
FROM
    iam_identity_profiles
WHERE
	tenant_id IS NULL
    AND identity_id = @identity_id
    AND membership_id IS NULL
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"identity_id": identityID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query default identity profile: %w", err)
	}

	profile, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[IdentityProfile])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect default identity profile: %w", err)
	}

	*p = profile

	return nil
}

// LoadByMembershipID loads the profile for a specific membership
func (p *IdentityProfile) LoadByMembershipID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	membershipID gid.GID,
) error {
	q := `
SELECT
    id,
    identity_id,
    membership_id,
    full_name,
    created_at,
    updated_at
FROM
    iam_identity_profiles
WHERE
	%s
    AND membership_id = @membership_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"membership_id": membershipID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query identity profile: %w", err)
	}

	profile, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[IdentityProfile])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect identity profile: %w", err)
	}

	*p = profile

	return nil
}

func (p *IdentityProfile) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	profileID gid.GID,
) error {
	q := `
SELECT
    id,
    identity_id,
    membership_id,
    full_name,
    created_at,
    updated_at
FROM
    iam_identity_profiles
WHERE
	%s
    AND id = @profile_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"profile_id": profileID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query identity profile: %w", err)
	}

	profile, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[IdentityProfile])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect identity profile: %w", err)
	}

	*p = profile

	return nil
}

func (p *IdentityProfile) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO
    iam_identity_profiles (
        tenant_id,
        id,
        identity_id,
        membership_id,
        full_name,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @id,
    @identity_id,
    @membership_id,
    @full_name,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":     p.ID.TenantID().String(),
		"id":            p.ID,
		"identity_id":   p.IdentityID,
		"membership_id": p.MembershipID,
		"full_name":     p.FullName,
		"created_at":    p.CreatedAt,
		"updated_at":    p.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert identity profile: %w", err)
	}

	return nil
}

func (p *IdentityProfile) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE
    iam_identity_profiles
SET
    full_name = @full_name,
    updated_at = @updated_at
WHERE
    id = @id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         p.ID,
		"full_name":  p.FullName,
		"updated_at": p.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update identity profile: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (p *IdentityProfile) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	profileID gid.GID,
) error {
	q := `
DELETE FROM
    iam_identity_profiles
WHERE
    id = @profile_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"profile_id": profileID}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete identity profile: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
