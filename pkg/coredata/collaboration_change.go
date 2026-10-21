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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	CollaborationChange struct {
		DocumentID     gid.GID   `db:"document_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		Revision       int64     `db:"revision"`
		ChangeHash     []byte    `db:"change_hash"`
		ChangeBytes    []byte    `db:"change_bytes"`
		CreatedAt      time.Time `db:"created_at"`
	}

	CollaborationChanges []*CollaborationChange
)

func (c CollaborationChange) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO collaboration_changes (
	tenant_id,
	document_id,
	organization_id,
	revision,
	change_hash,
	change_bytes,
	created_at
)
VALUES (
	@tenant_id,
	@document_id,
	@organization_id,
	@revision,
	@change_hash,
	@change_bytes,
	@created_at
)
ON CONFLICT (document_id, change_hash) DO NOTHING
`
	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"document_id":     c.DocumentID,
		"organization_id": c.OrganizationID,
		"revision":        c.Revision,
		"change_hash":     c.ChangeHash,
		"change_bytes":    c.ChangeBytes,
		"created_at":      c.CreatedAt,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert collaboration change: %w", err)
	}

	return nil
}

func (cs *CollaborationChanges) LoadAfterRevision(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
	afterRevision int64,
	limit int,
) error {
	q := `
SELECT
	document_id,
	organization_id,
	revision,
	change_hash,
	change_bytes,
	created_at
FROM
	collaboration_changes
WHERE
	%s
	AND document_id = @document_id
	AND revision > @after_revision
ORDER BY revision ASC
LIMIT @limit
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"document_id":    documentID,
		"after_revision": afterRevision,
		"limit":          limit,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query collaboration changes: %w", err)
	}

	changes, err := pgx.CollectRows(
		rows,
		pgx.RowToAddrOfStructByName[CollaborationChange],
	)
	if err != nil {
		return fmt.Errorf("cannot collect collaboration changes: %w", err)
	}

	*cs = changes

	return nil
}

func DeleteCollaborationChangesThroughRevision(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	documentID gid.GID,
	revision int64,
) error {
	q := `
DELETE FROM collaboration_changes
WHERE
	%s
	AND document_id = @document_id
	AND revision <= @revision
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"document_id": documentID,
		"revision":    revision,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete compacted collaboration changes: %w", err)
	}

	return nil
}
