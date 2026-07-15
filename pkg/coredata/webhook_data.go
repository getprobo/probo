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
		ID             gid.GID          `db:"id"`
		OrganizationID gid.GID          `db:"organization_id"`
		EventType      WebhookEventType `db:"event_type"`
		Data           json.RawMessage  `db:"data"`
		UpdatedFrom    json.RawMessage  `db:"updated_from"`
		CreatedAt      time.Time        `db:"created_at"`
		ProcessedAt    *time.Time       `db:"processed_at"`
	}

	WebhookDataList []*WebhookData
)

func (w *WebhookData) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO webhook_data (
    id,
    tenant_id,
    organization_id,
    event_type,
    data,
    updated_from,
    created_at
)
VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @event_type,
    @data,
    @updated_from,
    @created_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              w.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": w.OrganizationID,
		"event_type":      w.EventType,
		"data":            w.Data,
		"updated_from":    w.UpdatedFrom,
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
	conn pg.Tx,
) error {
	q := `
SELECT
    id,
    organization_id,
    event_type,
    data,
    updated_from,
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
	conn pg.Tx,
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
