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
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	WebhookEvent struct {
		ID                     gid.GID            `db:"id"`
		WebhookDataID          gid.GID            `db:"webhook_data_id"`
		WebhookSubscriptionID gid.GID            `db:"webhook_subscription_id"`
		Status                 WebhookEventStatus `db:"status"`
		Response               json.RawMessage    `db:"response"`
		CreatedAt              time.Time          `db:"created_at"`
	}

	WebhookEvents []*WebhookEvent
)

func (w WebhookEvent) CursorKey(orderBy WebhookEventOrderField) page.CursorKey {
	switch orderBy {
	case WebhookEventOrderFieldCreatedAt:
		return page.NewCursorKey(w.ID, w.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (w *WebhookEvents) LoadBySubscriptionID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	webhookSubscriptionID gid.GID,
	cursor *page.Cursor[WebhookEventOrderField],
) error {
	q := `
SELECT
    id,
    webhook_data_id,
    webhook_subscription_id,
    status,
    response,
    created_at
FROM
    webhook_events
WHERE
    %s
    AND webhook_subscription_id = @webhook_subscription_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"webhook_subscription_id": webhookSubscriptionID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query webhook events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[WebhookEvent])
	if err != nil {
		return fmt.Errorf("cannot collect webhook events: %w", err)
	}

	*w = events
	return nil
}

func (w *WebhookEvents) CountBySubscriptionID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	webhookSubscriptionID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(*)
FROM webhook_events
WHERE %s
    AND webhook_subscription_id = @webhook_subscription_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"webhook_subscription_id": webhookSubscriptionID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count webhook events: %w", err)
	}

	return count, nil
}

func (w *WebhookEvent) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO webhook_events (
    id,
    tenant_id,
    webhook_data_id,
    webhook_subscription_id,
    status,
    response,
    created_at
)
VALUES (
    @id,
    @tenant_id,
    @webhook_data_id,
    @webhook_subscription_id,
    @status,
    @response,
    @created_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                       w.ID,
		"tenant_id":                scope.GetTenantID(),
		"webhook_data_id":          w.WebhookDataID,
		"webhook_subscription_id": w.WebhookSubscriptionID,
		"status":                   w.Status,
		"response":                 w.Response,
		"created_at":               w.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert webhook event: %w", err)
	}

	return nil
}

func (w *WebhookEvent) UpdateStatus(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE webhook_events
SET
    status = @status,
    response = @response
WHERE %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":       w.ID,
		"status":   w.Status,
		"response": w.Response,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update webhook event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
