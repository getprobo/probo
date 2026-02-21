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
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	WebhookData struct {
		ID             gid.GID         `db:"id"`
		OrganizationID gid.GID         `db:"organization_id"`
		EventType      WebhookEventType `db:"event_type"`
		Data           json.RawMessage `db:"data"`
		CreatedAt      time.Time       `db:"created_at"`
		ProcessedAt    *time.Time      `db:"processed_at"`
	}

	WebhookDataList []*WebhookData
)

func (w *WebhookData) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO webhook_data (
    id,
    tenant_id,
    organization_id,
    event_type,
    data,
    created_at
)
VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @event_type,
    @data,
    @created_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              w.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": w.OrganizationID,
		"event_type":      w.EventType,
		"data":            w.Data,
		"created_at":      w.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert webhook data: %w", err)
	}

	return nil
}

func (w *WebhookData) LoadNextUnprocessedForUpdate(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
SELECT
    id,
    organization_id,
    event_type,
    data,
    created_at,
    processed_at
FROM webhook_data
WHERE processed_at IS NULL
ORDER BY created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query unprocessed webhook data: %w", err)
	}

	data, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[WebhookData])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect webhook data: %w", err)
	}

	*w = data
	return nil
}

func (w *WebhookData) UpdateProcessedAt(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE webhook_data
SET processed_at = @processed_at
WHERE %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":           w.ID,
		"processed_at": w.ProcessedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update webhook data: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
