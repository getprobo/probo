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

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	FrameworkExport struct {
		ID             gid.GID               `db:"id"`
		FrameworkID    gid.GID               `db:"framework_id"`
		RecipientEmail string                `db:"recipient_email"`
		RecipientName  string                `db:"recipient_name"`
		Status         FrameworkExportStatus `db:"status"`
		FileID         *gid.GID              `db:"file_id"`
		CreatedAt      time.Time             `db:"created_at"`
		StartedAt      *time.Time            `db:"started_at"`
		CompletedAt    *time.Time            `db:"completed_at"`
	}

	FrameworkExports []*FrameworkExport
)

var (
	ErrNoFrameworkExportAvailable = errors.New("no framework export available")
)

func (fe FrameworkExport) CursorKey(orderBy FrameworkExportOrderField) page.CursorKey {
	switch orderBy {
	case FrameworkExportOrderFieldCreatedAt:
		return page.NewCursorKey(fe.ID, fe.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (fe *FrameworkExport) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO framework_exports (
	id,
	tenant_id,
	framework_id,
	recipient_email,
	recipient_name,
	status,
	created_at
) VALUES (
	@id,
	@tenant_id,
	@framework_id,
	@recipient_email,
	@recipient_name,
	@status,
	@created_at
)`

	args := pgx.StrictNamedArgs{
		"id":              fe.ID,
		"tenant_id":       scope.GetTenantID(),
		"framework_id":    fe.FrameworkID,
		"recipient_email": fe.RecipientEmail,
		"recipient_name":  fe.RecipientName,
		"status":          fe.Status,
		"created_at":      fe.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (fe *FrameworkExport) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE
	framework_exports
SET
	status = @status,
	file_id = @file_id,
	started_at = @started_at,
	completed_at = @completed_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"status":       fe.Status,
		"file_id":      fe.FileID,
		"started_at":   fe.StartedAt,
		"completed_at": fe.CompletedAt,
		"id":           fe.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (fe *FrameworkExport) LoadNextPendingForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
SELECT
	id,
	framework_id,
	recipient_email,
	recipient_name,
	status,
	file_id,
	created_at,
	started_at,
	completed_at
FROM
	framework_exports
WHERE
	status = @status
ORDER BY
	created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`

	args := pgx.StrictNamedArgs{"status": FrameworkExportStatusPending}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return err
	}

	fe2, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[FrameworkExport])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoFrameworkExportAvailable
		}
		return fmt.Errorf("cannot collect framework export: %w", err)
	}

	*fe = fe2
	return nil
}
