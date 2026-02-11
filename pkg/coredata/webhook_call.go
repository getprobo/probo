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
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	WebhookCall struct {
		ID                     gid.GID           `db:"id"`
		WebhookEventID         gid.GID           `db:"webhook_event_id"`
		WebhookConfigurationID gid.GID           `db:"webhook_configuration_id"`
		EndpointURL            string            `db:"endpoint_url"`
		Status                 WebhookCallStatus `db:"status"`
		Response               json.RawMessage   `db:"response"`
		CreatedAt              time.Time         `db:"created_at"`
	}

	WebhookCalls []*WebhookCall
)

func (w *WebhookCall) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO webhook_calls (
    id,
    tenant_id,
    webhook_event_id,
    webhook_configuration_id,
    endpoint_url,
    status,
    response,
    created_at
)
VALUES (
    @id,
    @tenant_id,
    @webhook_event_id,
    @webhook_configuration_id,
    @endpoint_url,
    @status,
    @response,
    @created_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                       w.ID,
		"tenant_id":                scope.GetTenantID(),
		"webhook_event_id":         w.WebhookEventID,
		"webhook_configuration_id": w.WebhookConfigurationID,
		"endpoint_url":             w.EndpointURL,
		"status":                   w.Status,
		"response":                 w.Response,
		"created_at":               w.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert webhook call: %w", err)
	}

	return nil
}
