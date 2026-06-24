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
	conn pg.Tx,
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
	conn pg.Querier,
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
