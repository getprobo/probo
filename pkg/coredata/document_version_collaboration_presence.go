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
	DocumentVersionCollaborationPresence struct {
		ConnectionID      string    `db:"connection_id"`
		DocumentVersionID gid.GID   `db:"document_version_id"`
		OrganizationID    gid.GID   `db:"organization_id"`
		IdentityID        gid.GID   `db:"identity_id"`
		AnchorPosition    int       `db:"anchor_position"`
		HeadPosition      int       `db:"head_position"`
		UpdatedAt         time.Time `db:"updated_at"`
	}

	DocumentVersionCollaborationPresences []*DocumentVersionCollaborationPresence
)

func (p DocumentVersionCollaborationPresence) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO document_version_collaboration_presences (
	tenant_id,
	connection_id,
	document_version_id,
	organization_id,
	identity_id,
	anchor_position,
	head_position,
	updated_at
)
VALUES (
	@tenant_id,
	@connection_id,
	@document_version_id,
	@organization_id,
	@identity_id,
	@anchor_position,
	@head_position,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"tenant_id":           scope.GetTenantID(),
		"connection_id":       p.ConnectionID,
		"document_version_id": p.DocumentVersionID,
		"organization_id":     p.OrganizationID,
		"identity_id":         p.IdentityID,
		"anchor_position":     p.AnchorPosition,
		"head_position":       p.HeadPosition,
		"updated_at":          p.UpdatedAt,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert document collaboration presence: %w", err)
	}

	return nil
}

func (p DocumentVersionCollaborationPresence) Update(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE document_version_collaboration_presences SET
	anchor_position = @anchor_position,
	head_position = @head_position,
	updated_at = @updated_at
WHERE
	%s
	AND connection_id = @connection_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"connection_id":   p.ConnectionID,
		"anchor_position": p.AnchorPosition,
		"head_position":   p.HeadPosition,
		"updated_at":      p.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update document collaboration presence: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (ps *DocumentVersionCollaborationPresences) Load(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
	excludeConnectionID string,
	updatedAfter time.Time,
	limit int,
) error {
	q := `
SELECT
	connection_id,
	document_version_id,
	organization_id,
	identity_id,
	anchor_position,
	head_position,
	updated_at
FROM
	document_version_collaboration_presences
WHERE
	%s
	AND document_version_id = @document_version_id
	AND connection_id <> @exclude_connection_id
	AND updated_at > @updated_after
ORDER BY updated_at DESC
LIMIT @limit
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"document_version_id":   documentVersionID,
		"exclude_connection_id": excludeConnectionID,
		"updated_after":         updatedAfter,
		"limit":                 limit,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document collaboration presences: %w", err)
	}

	presences, err := pgx.CollectRows(
		rows,
		pgx.RowToAddrOfStructByName[DocumentVersionCollaborationPresence],
	)
	if err != nil {
		return fmt.Errorf("cannot collect document collaboration presences: %w", err)
	}

	*ps = presences

	return nil
}

func (p DocumentVersionCollaborationPresence) Delete(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM document_version_collaboration_presences
WHERE
	%s
	AND connection_id = @connection_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"connection_id": p.ConnectionID,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete document collaboration presence: %w", err)
	}

	return nil
}

func DeleteExpiredDocumentVersionCollaborationPresences(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	documentVersionID gid.GID,
	updatedBefore time.Time,
) error {
	q := `
DELETE FROM document_version_collaboration_presences
WHERE
	%s
	AND document_version_id = @document_version_id
	AND updated_at < @updated_before
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"document_version_id": documentVersionID,
		"updated_before":      updatedBefore,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete expired document collaboration presences: %w", err)
	}

	return nil
}
