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

const documentVersionAutomergeActorMaxLoad = 500

type (
	DocumentVersionAutomergeActor struct {
		DocumentVersionID gid.GID   `db:"document_version_id"`
		OrganizationID    gid.GID   `db:"organization_id"`
		ActorID           []byte    `db:"actor_id"`
		IdentityID        gid.GID   `db:"identity_id"`
		ProfileID         *gid.GID  `db:"profile_id"`
		CreatedAt         time.Time `db:"created_at"`
	}

	DocumentVersionAutomergeActors []*DocumentVersionAutomergeActor
)

// InsertIfAbsent records the first authenticated identity that introduced an
// Automerge actor on this document version. Later inserts for the same actor
// are ignored so a peer replaying someone else's changes cannot steal authorship.
func (a DocumentVersionAutomergeActor) InsertIfAbsent(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO document_version_automerge_actors (
	tenant_id,
	document_version_id,
	organization_id,
	actor_id,
	identity_id,
	profile_id,
	created_at
)
VALUES (
	@tenant_id,
	@document_version_id,
	@organization_id,
	@actor_id,
	@identity_id,
	@profile_id,
	@created_at
)
ON CONFLICT (document_version_id, actor_id) DO NOTHING
`
	args := pgx.StrictNamedArgs{
		"tenant_id":           scope.GetTenantID(),
		"document_version_id": a.DocumentVersionID,
		"organization_id":     a.OrganizationID,
		"actor_id":            a.ActorID,
		"identity_id":         a.IdentityID,
		"profile_id":          a.ProfileID,
		"created_at":          a.CreatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert document version Automerge actor: %w", err)
	}

	return nil
}

// LoadByDocumentVersionID loads Automerge actor bindings for a document version.
// The result is hard-capped; a document with more concurrent writers than the
// cap still returns a usable prefix for history attribution.
func (as *DocumentVersionAutomergeActors) LoadByDocumentVersionID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
	limit int,
) error {
	if limit <= 0 || limit > documentVersionAutomergeActorMaxLoad {
		limit = documentVersionAutomergeActorMaxLoad
	}

	q := `
SELECT
	document_version_id,
	organization_id,
	actor_id,
	identity_id,
	profile_id,
	created_at
FROM
	document_version_automerge_actors
WHERE
	%s
	AND document_version_id = @document_version_id
ORDER BY created_at ASC, actor_id ASC
LIMIT @limit
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"document_version_id": documentVersionID,
		"limit":               limit,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version Automerge actors: %w", err)
	}

	actors, err := pgx.CollectRows(
		rows,
		pgx.RowToAddrOfStructByName[DocumentVersionAutomergeActor],
	)
	if err != nil {
		return fmt.Errorf("cannot collect document version Automerge actors: %w", err)
	}

	*as = actors

	return nil
}
