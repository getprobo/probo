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
)

type (
	// SCIMProfilePhoto records the directory photo version that was last
	// applied to a SCIM-managed membership profile. A profile without a row
	// has no synced photo.
	SCIMProfilePhoto struct {
		MembershipProfileID gid.GID   `db:"membership_profile_id"`
		OrganizationID      gid.GID   `db:"organization_id"`
		ETag                string    `db:"etag"`
		CreatedAt           time.Time `db:"created_at"`
		UpdatedAt           time.Time `db:"updated_at"`
	}

	SCIMProfilePhotos []*SCIMProfilePhoto
)

func (p *SCIMProfilePhoto) LoadByMembershipProfileID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	profileID gid.GID,
) error {
	q := `
SELECT
    membership_profile_id,
    organization_id,
    etag,
    created_at,
    updated_at
FROM
    scim_profile_photos
WHERE
    %s
    AND membership_profile_id = @membership_profile_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"membership_profile_id": profileID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query scim profile photo: %w", err)
	}

	photo, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SCIMProfilePhoto])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect scim profile photo: %w", err)
	}

	*p = photo

	return nil
}

func (p *SCIMProfilePhotos) LoadByMembershipProfileIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	profileIDs []gid.GID,
) error {
	q := `
SELECT
    membership_profile_id,
    organization_id,
    etag,
    created_at,
    updated_at
FROM
    scim_profile_photos
WHERE
    %s
    AND membership_profile_id = ANY(@membership_profile_ids::text[]);
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"membership_profile_ids": profileIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query scim profile photos: %w", err)
	}

	photos, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SCIMProfilePhoto])
	if err != nil {
		return fmt.Errorf("cannot collect scim profile photos: %w", err)
	}

	*p = photos

	return nil
}

func (p *SCIMProfilePhoto) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    scim_profile_photos (
        membership_profile_id,
        organization_id,
        tenant_id,
        etag,
        created_at,
        updated_at
    )
VALUES (
    @membership_profile_id,
    @organization_id,
    @tenant_id,
    @etag,
    @created_at,
    @updated_at
)
ON CONFLICT (membership_profile_id) DO UPDATE
SET
    etag = EXCLUDED.etag,
    updated_at = EXCLUDED.updated_at;
`

	args := pgx.StrictNamedArgs{
		"membership_profile_id": p.MembershipProfileID,
		"organization_id":       p.OrganizationID,
		"tenant_id":             scope.GetTenantID(),
		"etag":                  p.ETag,
		"created_at":            p.CreatedAt,
		"updated_at":            p.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert scim profile photo: %w", err)
	}

	return nil
}

func (p *SCIMProfilePhoto) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	profileID gid.GID,
) error {
	q := `
DELETE FROM
    scim_profile_photos
WHERE
    %s
    AND membership_profile_id = @membership_profile_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"membership_profile_id": profileID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete scim profile photo: %w", err)
	}

	return nil
}
