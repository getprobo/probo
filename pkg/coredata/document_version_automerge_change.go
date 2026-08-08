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
	DocumentVersionAutomergeChange struct {
		DocumentVersionID gid.GID   `db:"document_version_id"`
		OrganizationID    gid.GID   `db:"organization_id"`
		Revision          int64     `db:"revision"`
		ChangeHash        []byte    `db:"change_hash"`
		ChangeBytes       []byte    `db:"change_bytes"`
		CreatedAt         time.Time `db:"created_at"`
	}

	DocumentVersionAutomergeChanges []*DocumentVersionAutomergeChange
)

func (c DocumentVersionAutomergeChange) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO document_version_automerge_changes (
	tenant_id,
	document_version_id,
	organization_id,
	revision,
	change_hash,
	change_bytes,
	created_at
)
VALUES (
	@tenant_id,
	@document_version_id,
	@organization_id,
	@revision,
	@change_hash,
	@change_bytes,
	@created_at
)
ON CONFLICT (document_version_id, change_hash) DO NOTHING
`
	args := pgx.StrictNamedArgs{
		"tenant_id":           scope.GetTenantID(),
		"document_version_id": c.DocumentVersionID,
		"organization_id":     c.OrganizationID,
		"revision":            c.Revision,
		"change_hash":         c.ChangeHash,
		"change_bytes":        c.ChangeBytes,
		"created_at":          c.CreatedAt,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert document version Automerge change: %w", err)
	}

	return nil
}

func (cs *DocumentVersionAutomergeChanges) LoadAfterRevision(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
	afterRevision int64,
	limit int,
) error {
	q := `
SELECT
	document_version_id,
	organization_id,
	revision,
	change_hash,
	change_bytes,
	created_at
FROM
	document_version_automerge_changes
WHERE
	%s
	AND document_version_id = @document_version_id
	AND revision > @after_revision
ORDER BY revision ASC
LIMIT @limit
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"document_version_id": documentVersionID,
		"after_revision":      afterRevision,
		"limit":               limit,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version Automerge changes: %w", err)
	}

	changes, err := pgx.CollectRows(
		rows,
		pgx.RowToAddrOfStructByName[DocumentVersionAutomergeChange],
	)
	if err != nil {
		return fmt.Errorf("cannot collect document version Automerge changes: %w", err)
	}

	*cs = changes

	return nil
}

func DeleteDocumentVersionAutomergeChangesThroughRevision(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	documentVersionID gid.GID,
	revision int64,
) error {
	q := `
DELETE FROM document_version_automerge_changes
WHERE
	%s
	AND document_version_id = @document_version_id
	AND revision <= @revision
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"document_version_id": documentVersionID,
		"revision":            revision,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete compacted document version Automerge changes: %w", err)
	}

	return nil
}
