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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	ElectronicSignatureEvent struct {
		ID                    gid.GID                        `db:"id"`
		TenantID              gid.TenantID                   `db:"tenant_id"`
		ElectronicSignatureID gid.GID                        `db:"electronic_signature_id"`
		EventType             ElectronicSignatureEventType   `db:"event_type"`
		EventSource           ElectronicSignatureEventSource `db:"event_source"`
		ActorEmail            string                         `db:"actor_email"`
		ActorIPAddress        string                         `db:"actor_ip_address"`
		ActorUserAgent        string                         `db:"actor_user_agent"`
		OccurredAt            time.Time                      `db:"occurred_at"`
		CreatedAt             time.Time                      `db:"created_at"`
	}

	ElectronicSignatureEvents []*ElectronicSignatureEvent
)

func (e *ElectronicSignatureEvent) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO electronic_signature_events (
	id, tenant_id, electronic_signature_id, event_type, event_source,
	actor_email, actor_ip_address, actor_user_agent,
	occurred_at, created_at
) VALUES (
	@id, @tenant_id, @electronic_signature_id, @event_type, @event_source,
	@actor_email, @actor_ip_address, @actor_user_agent,
	@occurred_at, @created_at
)
`
	args := pgx.StrictNamedArgs{
		"id":                      e.ID,
		"tenant_id":               scope.GetTenantID(),
		"electronic_signature_id": e.ElectronicSignatureID,
		"event_type":              e.EventType,
		"event_source":            e.EventSource,
		"actor_email":             e.ActorEmail,
		"actor_ip_address":        e.ActorIPAddress,
		"actor_user_agent":        e.ActorUserAgent,
		"occurred_at":             e.OccurredAt,
		"created_at":              e.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert electronic signature event: %w", err)
	}

	return nil
}

func (es *ElectronicSignatureEvents) LoadBySignatureID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	sigID gid.GID,
) error {
	q := `
SELECT
	id, tenant_id, electronic_signature_id, event_type, event_source,
	actor_email, actor_ip_address, actor_user_agent,
	occurred_at, created_at
FROM electronic_signature_events
WHERE %s AND electronic_signature_id = @electronic_signature_id
ORDER BY occurred_at ASC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"electronic_signature_id": sigID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query electronic signature events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ElectronicSignatureEvent])
	if err != nil {
		return fmt.Errorf("cannot collect electronic signature events: %w", err)
	}

	*es = events

	return nil
}
