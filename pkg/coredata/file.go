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
	"fmt"
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	File struct {
		ID         gid.GID    `db:"id"`
		BucketName string     `db:"bucket_name"`
		MimeType   string     `db:"mime_type"`
		FileName   string     `db:"file_name"`
		FileKey    string     `db:"file_key"`
		FileSize   int64      `db:"file_size"`
		CreatedAt  time.Time  `db:"created_at"`
		UpdatedAt  time.Time  `db:"updated_at"`
		DeletedAt  *time.Time `db:"deleted_at"`
	}

	Files []*File
)

func (f *File) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	fileID gid.GID,
) error {
	q := `
SELECT
    id,
    bucket_name,
    mime_type,
    file_name,
    file_key,
    file_size,
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
		return fmt.Errorf("cannot collect file: %w", err)
	}

	*f = file

	return nil
}

func (f File) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    files (
        id,
        tenant_id,
        bucket_name,
        mime_type,
        file_name,
        file_key,
        file_size,
        created_at,
        updated_at,
        deleted_at
    )
VALUES (
    @file_id,
    @tenant_id,
    @bucket_name,
    @mime_type,
    @file_name,
    @file_key,
    @file_size,
    @created_at,
    @updated_at,
    @deleted_at
)
`

	args := pgx.StrictNamedArgs{
		"file_id":     f.ID,
		"tenant_id":   scope.GetTenantID(),
		"bucket_name": f.BucketName,
		"mime_type":   f.MimeType,
		"file_name":   f.FileName,
		"file_key":    f.FileKey,
		"file_size":   f.FileSize,
		"created_at":  f.CreatedAt,
		"updated_at":  f.UpdatedAt,
		"deleted_at":  f.DeletedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (f File) SoftDelete(ctx context.Context, conn pg.Conn, scope Scoper) error {
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
