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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
)

type (
	File struct {
		ID             gid.GID        `db:"id"`
		OrganizationID gid.GID        `db:"organization_id"`
		BucketName     string         `db:"bucket_name"`
		MimeType       string         `db:"mime_type"`
		FileName       string         `db:"file_name"`
		FileKey        string         `db:"file_key"`
		FileSize       int64          `db:"file_size"`
		Visibility     FileVisibility `db:"visibility"`
		CreatedAt      time.Time      `db:"created_at"`
		UpdatedAt      time.Time      `db:"updated_at"`
		DeletedAt      *time.Time     `db:"deleted_at"`
	}

	Files []*File
)

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (f *File) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM files WHERE id = ANY(@resource_ids::text[])`

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

func (f *File) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	fileID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    bucket_name,
    mime_type,
    file_name,
    file_key,
    file_size,
    visibility,
    created_at,
    updated_at,
    deleted_at
FROM
    files
WHERE
    %s
    AND id = @file_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"file_id": fileID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query file: %w", err)
	}
	defer rows.Close()

	file, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[File])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect file: %w", err)
	}

	*f = file

	return nil
}

// LoadByIDs Loads every given files, whether they are active or not. See
// Files.LoadActiveByIDs for a safer option.
//
// DISCLAIMER: use with caution on user-facing features.
func (f *Files) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	fileIDs []gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    bucket_name,
    mime_type,
    file_name,
    file_key,
    file_size,
    visibility,
    created_at,
    updated_at,
    deleted_at
FROM
    files
WHERE
    %s
    AND id = ANY(@file_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"file_ids": fileIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query files: %w", err)
	}

	files, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[File])
	if err != nil {
		return fmt.Errorf("cannot collect files: %w", err)
	}

	*f = files

	if len(files) != len(gid.NewSet(fileIDs...)) {
		return ErrResourceNotFound
	}

	return nil
}

func (f File) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    files (
        id,
        tenant_id,
        organization_id,
        bucket_name,
        mime_type,
        file_name,
        file_key,
        file_size,
        visibility,
        created_at,
        updated_at,
        deleted_at
    )
VALUES (
    @file_id,
    @tenant_id,
    @organization_id,
    @bucket_name,
    @mime_type,
    @file_name,
    @file_key,
    @file_size,
    @visibility,
    @created_at,
    @updated_at,
    @deleted_at
)
`

	args := pgx.StrictNamedArgs{
		"file_id":         f.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": f.OrganizationID,
		"bucket_name":     f.BucketName,
		"mime_type":       f.MimeType,
		"file_name":       f.FileName,
		"file_key":        f.FileKey,
		"file_size":       f.FileSize,
		"visibility":      f.Visibility,
		"created_at":      f.CreatedAt,
		"updated_at":      f.UpdatedAt,
		"deleted_at":      f.DeletedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "files_file_key_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert file: %w", err)
	}

	return nil
}

func (f *File) LoadActiveByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	fileID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    bucket_name,
    mime_type,
    file_name,
    file_key,
    file_size,
    visibility,
    created_at,
    updated_at,
    deleted_at
FROM
    files
WHERE
    %s
    AND id = @file_id
    AND deleted_at IS NULL
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"file_id": fileID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query file: %w", err)
	}
	defer rows.Close()

	file, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[File])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect file: %w", err)
	}

	*f = file

	return nil
}

func (f *File) LoadPublicByID(
	ctx context.Context,
	conn pg.Querier,
	fileID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    bucket_name,
    mime_type,
    file_name,
    file_key,
    file_size,
    visibility,
    created_at,
    updated_at,
    deleted_at
FROM
    files
WHERE
    id = @file_id
    AND visibility = 'PUBLIC'
    AND deleted_at IS NULL
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"file_id": fileID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query file: %w", err)
	}
	defer rows.Close()

	file, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[File])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect file: %w", err)
	}

	*f = file

	return nil
}

func (f File) SoftDelete(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
UPDATE files
SET deleted_at = NOW()
WHERE %s
    AND id = @file_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"file_id": f.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)

	return err
}
