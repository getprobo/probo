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
	CollaborationActor struct {
		DocumentID     gid.GID   `db:"document_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		ActorID        []byte    `db:"actor_id"`
		IdentityID     gid.GID   `db:"identity_id"`
		ProfileID      *gid.GID  `db:"profile_id"`
		CreatedAt      time.Time `db:"created_at"`
	}

	CollaborationActors []*CollaborationActor
)

// Insert records the first authenticated identity that introduced an
// Automerge actor on this collaboration document. Later inserts for the same
// actor are ignored so a peer replaying someone else's changes cannot steal
// authorship.
func (a CollaborationActor) Insert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO collaboration_actors (
	tenant_id,
	document_id,
	organization_id,
	actor_id,
	identity_id,
	profile_id,
	created_at
)
VALUES (
	@tenant_id,
	@document_id,
	@organization_id,
	@actor_id,
	@identity_id,
	@profile_id,
	@created_at
)
ON CONFLICT (document_id, actor_id) DO NOTHING
`
	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"document_id":     a.DocumentID,
		"organization_id": a.OrganizationID,
		"actor_id":        a.ActorID,
		"identity_id":     a.IdentityID,
		"profile_id":      a.ProfileID,
		"created_at":      a.CreatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert collaboration actor: %w", err)
	}

	return nil
}

// LoadByDocumentIDAndIdentityID returns the oldest actor bound to identityID
// on this document.
func (a *CollaborationActor) LoadByDocumentIDAndIdentityID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
	identityID gid.GID,
) error {
	q := `
SELECT
	document_id,
	organization_id,
	actor_id,
	identity_id,
	profile_id,
	created_at
FROM
	collaboration_actors
WHERE
	%s
	AND document_id = @document_id
	AND identity_id = @identity_id
ORDER BY created_at ASC, actor_id ASC
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"document_id": documentID,
		"identity_id": identityID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query collaboration actor: %w", err)
	}

	actor, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[CollaborationActor],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect collaboration actor: %w", err)
	}

	*a = actor

	return nil
}

// LoadByDocumentIDAndActorIDs loads the actors named by actorIDs. The caller
// bounds the set.
func (as *CollaborationActors) LoadByDocumentIDAndActorIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
	actorIDs [][]byte,
) error {
	if len(actorIDs) == 0 {
		*as = nil

		return nil
	}

	q := `
SELECT
	document_id,
	organization_id,
	actor_id,
	identity_id,
	profile_id,
	created_at
FROM
	collaboration_actors
WHERE
	%s
	AND document_id = @document_id
	AND actor_id = ANY(@actor_ids)
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"document_id": documentID,
		"actor_ids":   actorIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query collaboration actors by id: %w", err)
	}

	actors, err := pgx.CollectRows(
		rows,
		pgx.RowToAddrOfStructByName[CollaborationActor],
	)
	if err != nil {
		return fmt.Errorf("cannot collect collaboration actors by id: %w", err)
	}

	*as = actors

	return nil
}
