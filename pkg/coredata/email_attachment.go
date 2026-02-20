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
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	EmailAttachment struct {
		ID        gid.GID   `db:"id"`
		EmailID   gid.GID   `db:"email_id"`
		FileID    gid.GID   `db:"file_id"`
		Filename  string    `db:"filename"`
		CreatedAt time.Time `db:"created_at"`
	}

	EmailAttachments []*EmailAttachment
)

func NewEmailAttachment(emailID, fileID gid.GID, filename string) *EmailAttachment {
	return &EmailAttachment{
		ID:        gid.New(gid.NilTenant, EmailAttachmentEntityType),
		EmailID:   emailID,
		FileID:    fileID,
		Filename:  filename,
		CreatedAt: time.Now(),
	}
}

func (a *EmailAttachment) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO email_attachments (id, email_id, file_id, filename, created_at)
VALUES (@id, @email_id, @file_id, @filename, @created_at)
`
	args := pgx.StrictNamedArgs{
		"id":         a.ID,
		"email_id":   a.EmailID,
		"file_id":    a.FileID,
		"filename":   a.Filename,
		"created_at": a.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert email attachment: %w", err)
	}

	return nil
}

func (a *EmailAttachments) LoadByEmailID(
	ctx context.Context,
	conn pg.Conn,
	emailID gid.GID,
) error {
	q := `
SELECT id, email_id, file_id, filename, created_at
FROM email_attachments
WHERE email_id = @email_id
ORDER BY created_at ASC
`
	args := pgx.StrictNamedArgs{
		"email_id": emailID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query email attachments: %w", err)
	}

	attachments, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[EmailAttachment])
	if err != nil {
		return fmt.Errorf("cannot collect email attachments: %w", err)
	}

	*a = attachments

	return nil
}
