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
	conn pg.Tx,
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
	conn pg.Querier,
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
