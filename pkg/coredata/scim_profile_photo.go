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
	SCIMProfilePhoto struct {
		MembershipProfileID gid.GID   `db:"membership_profile_id"`
		OrganizationID      gid.GID   `db:"organization_id"`
		ETag                string    `db:"etag"`
		AvatarFileID        *gid.GID  `db:"avatar_file_id"`
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
    avatar_file_id,
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

func (p *SCIMProfilePhotos) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
    membership_profile_id,
    organization_id,
    etag,
    avatar_file_id,
    created_at,
    updated_at
FROM
    scim_profile_photos
WHERE
    %s
    AND organization_id = @organization_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
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
        avatar_file_id,
        created_at,
        updated_at
    )
VALUES (
    @membership_profile_id,
    @organization_id,
    @tenant_id,
    @etag,
    @avatar_file_id,
    @created_at,
    @updated_at
)
ON CONFLICT (membership_profile_id) DO UPDATE
SET
    etag = EXCLUDED.etag,
    avatar_file_id = EXCLUDED.avatar_file_id,
    updated_at = EXCLUDED.updated_at;
`

	args := pgx.StrictNamedArgs{
		"membership_profile_id": p.MembershipProfileID,
		"organization_id":       p.OrganizationID,
		"tenant_id":             scope.GetTenantID(),
		"etag":                  p.ETag,
		"avatar_file_id":        p.AvatarFileID,
		"created_at":            p.CreatedAt,
		"updated_at":            p.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert scim profile photo: %w", err)
	}

	return nil
}

// DeleteByIdentityID forgets the photo state of every profile of the identity.
// Identities have no tenant, so rows are matched through the profiles.
func (p *SCIMProfilePhoto) DeleteByIdentityID(
	ctx context.Context,
	conn pg.Tx,
	identityID gid.GID,
) error {
	q := `
DELETE FROM
    scim_profile_photos
WHERE
    membership_profile_id IN (
        SELECT id
        FROM iam_membership_profiles
        WHERE identity_id = @identity_id
    );
`

	args := pgx.StrictNamedArgs{"identity_id": identityID}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete scim profile photos: %w", err)
	}

	return nil
}
