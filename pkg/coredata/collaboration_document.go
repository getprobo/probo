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
	CollaborationDocument struct {
		DocumentID       gid.GID    `db:"document_id"`
		OrganizationID   gid.GID    `db:"organization_id"`
		Snapshot         []byte     `db:"snapshot"`
		Heads            []byte     `db:"heads"`
		Revision         int64      `db:"revision"`
		SnapshotRevision int64      `db:"snapshot_revision"`
		ChangeRevision   int64      `db:"change_revision"`
		Seeded           bool       `db:"seeded"`
		SeedClaimedAt    *time.Time `db:"seed_claimed_at"`
		CreatedAt        time.Time  `db:"created_at"`
		UpdatedAt        time.Time  `db:"updated_at"`
	}
)

func (d *CollaborationDocument) LoadByDocumentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
SELECT
	document_id,
	organization_id,
	snapshot,
	heads,
	revision,
	snapshot_revision,
	change_revision,
	seeded,
	seed_claimed_at,
	created_at,
	updated_at
FROM
	collaboration_documents
WHERE
	%s
	AND document_id = @document_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id": documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query collaboration document: %w", err)
	}

	document, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[CollaborationDocument],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect collaboration document: %w", err)
	}

	*d = document

	return nil
}

func (d *CollaborationDocument) LoadByDocumentIDForUpdate(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
SELECT
	document_id,
	organization_id,
	snapshot,
	heads,
	revision,
	snapshot_revision,
	change_revision,
	seeded,
	seed_claimed_at,
	created_at,
	updated_at
FROM
	collaboration_documents
WHERE
	%s
	AND document_id = @document_id
FOR UPDATE
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id": documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := tx.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query collaboration document: %w", err)
	}

	document, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[CollaborationDocument],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect collaboration document: %w", err)
	}

	*d = document

	return nil
}

// InsertIfAbsent creates the collaboration document for a document the first
// time it is opened for editing. It reports whether the row was created.
func (d CollaborationDocument) InsertIfAbsent(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO collaboration_documents (
	tenant_id,
	document_id,
	organization_id,
	snapshot,
	heads,
	revision,
	snapshot_revision,
	change_revision,
	seeded,
	seed_claimed_at,
	created_at,
	updated_at
)
VALUES (
	@tenant_id,
	@document_id,
	@organization_id,
	@snapshot,
	@heads,
	@revision,
	@snapshot_revision,
	@change_revision,
	@seeded,
	@seed_claimed_at,
	@created_at,
	@updated_at
)
ON CONFLICT (document_id) DO NOTHING
`
	args := pgx.StrictNamedArgs{
		"tenant_id":         scope.GetTenantID(),
		"document_id":       d.DocumentID,
		"organization_id":   d.OrganizationID,
		"snapshot":          d.Snapshot,
		"heads":             d.Heads,
		"revision":          d.Revision,
		"snapshot_revision": d.SnapshotRevision,
		"change_revision":   d.ChangeRevision,
		"seeded":            d.Seeded,
		"seed_claimed_at":   d.SeedClaimedAt,
		"created_at":        d.CreatedAt,
		"updated_at":        d.UpdatedAt,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot insert collaboration document if absent: %w", err)
	}

	return result.RowsAffected() == 1, nil
}

func (d CollaborationDocument) Update(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE collaboration_documents SET
	snapshot = @snapshot,
	heads = @heads,
	revision = @revision,
	snapshot_revision = @snapshot_revision,
	change_revision = @change_revision,
	seeded = @seeded,
	seed_claimed_at = @seed_claimed_at,
	updated_at = @updated_at
WHERE
	%s
	AND document_id = @document_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id":       d.DocumentID,
		"snapshot":          d.Snapshot,
		"heads":             d.Heads,
		"revision":          d.Revision,
		"snapshot_revision": d.SnapshotRevision,
		"change_revision":   d.ChangeRevision,
		"seeded":            d.Seeded,
		"seed_claimed_at":   d.SeedClaimedAt,
		"updated_at":        d.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update collaboration document: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// ReleaseSeedClaim clears seed_claimed_at only when it is still the claim
// acquired by this opener. A newer claim is left in place.
func ReleaseCollaborationSeedClaim(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	documentID gid.GID,
	claimedAt time.Time,
) error {
	q := `
UPDATE collaboration_documents SET
	seed_claimed_at = NULL,
	revision = revision + 1,
	updated_at = @updated_at
WHERE
	%s
	AND document_id = @document_id
	AND seeded = FALSE
	AND seed_claimed_at = @seed_claimed_at
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id":     documentID,
		"seed_claimed_at": claimedAt,
		"updated_at":      time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot release collaboration seed claim: %w", err)
	}

	return nil
}
