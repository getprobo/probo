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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	DocumentVersionAutomergeState struct {
		DocumentVersionID gid.GID    `db:"document_version_id"`
		OrganizationID    gid.GID    `db:"organization_id"`
		Snapshot          []byte     `db:"snapshot"`
		Revision          int64      `db:"revision"`
		Seeded            bool       `db:"seeded"`
		SeedClaimedAt     *time.Time `db:"seed_claimed_at"`
		CreatedAt         time.Time  `db:"created_at"`
		UpdatedAt         time.Time  `db:"updated_at"`
	}
)

func (s *DocumentVersionAutomergeState) LoadByDocumentVersionID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
) error {
	return s.loadByDocumentVersionID(ctx, conn, scope, documentVersionID, false)
}

func (s *DocumentVersionAutomergeState) LoadByDocumentVersionIDForUpdate(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	documentVersionID gid.GID,
) error {
	return s.loadByDocumentVersionID(ctx, tx, scope, documentVersionID, true)
}

func (s *DocumentVersionAutomergeState) loadByDocumentVersionID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
	forUpdate bool,
) error {
	q := `
SELECT
	document_version_id,
	organization_id,
	snapshot,
	revision,
	seeded,
	seed_claimed_at,
	created_at,
	updated_at
FROM
	document_version_automerge_states
WHERE
	%s
	AND document_version_id = @document_version_id
`
	if forUpdate {
		q += "FOR UPDATE\n"
	}

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_version_id": documentVersionID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version Automerge state: %w", err)
	}

	state, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[DocumentVersionAutomergeState],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect document version Automerge state: %w", err)
	}

	*s = state

	return nil
}

func (s DocumentVersionAutomergeState) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO document_version_automerge_states (
	tenant_id,
	document_version_id,
	organization_id,
	snapshot,
	revision,
	seeded,
	seed_claimed_at,
	created_at,
	updated_at
)
VALUES (
	@tenant_id,
	@document_version_id,
	@organization_id,
	@snapshot,
	@revision,
	@seeded,
	@seed_claimed_at,
	@created_at,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"tenant_id":           scope.GetTenantID(),
		"document_version_id": s.DocumentVersionID,
		"organization_id":     s.OrganizationID,
		"snapshot":            s.Snapshot,
		"revision":            s.Revision,
		"seeded":              s.Seeded,
		"seed_claimed_at":     s.SeedClaimedAt,
		"created_at":          s.CreatedAt,
		"updated_at":          s.UpdatedAt,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "document_version_automerge_states_pkey" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert document version Automerge state: %w", err)
	}

	return nil
}

func (s DocumentVersionAutomergeState) InsertIfAbsent(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO document_version_automerge_states (
	tenant_id,
	document_version_id,
	organization_id,
	snapshot,
	revision,
	seeded,
	seed_claimed_at,
	created_at,
	updated_at
)
VALUES (
	@tenant_id,
	@document_version_id,
	@organization_id,
	@snapshot,
	@revision,
	@seeded,
	@seed_claimed_at,
	@created_at,
	@updated_at
)
ON CONFLICT (document_version_id) DO NOTHING
`
	args := pgx.StrictNamedArgs{
		"tenant_id":           scope.GetTenantID(),
		"document_version_id": s.DocumentVersionID,
		"organization_id":     s.OrganizationID,
		"snapshot":            s.Snapshot,
		"revision":            s.Revision,
		"seeded":              s.Seeded,
		"seed_claimed_at":     s.SeedClaimedAt,
		"created_at":          s.CreatedAt,
		"updated_at":          s.UpdatedAt,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot insert document version Automerge state if absent: %w", err)
	}

	return result.RowsAffected() == 1, nil
}

func (s DocumentVersionAutomergeState) Update(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE document_version_automerge_states SET
	snapshot = @snapshot,
	revision = @revision,
	seeded = @seeded,
	seed_claimed_at = @seed_claimed_at,
	updated_at = @updated_at
WHERE
	%s
	AND document_version_id = @document_version_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_version_id": s.DocumentVersionID,
		"snapshot":            s.Snapshot,
		"revision":            s.Revision,
		"seeded":              s.Seeded,
		"seed_claimed_at":     s.SeedClaimedAt,
		"updated_at":          s.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update document version Automerge state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
