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
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

// AsanaWebhookHandshake holds the hook secret Asana sends while a webhook
// is being created. The ciphertext is replaced onto the connector once Asana
// returns the webhook gid.
type AsanaWebhookHandshake struct {
	Nonce            string    `db:"nonce"`
	ConnectorID      gid.GID   `db:"connector_id"`
	SecretCiphertext []byte    `db:"secret_ciphertext"`
	CreatedAt        time.Time `db:"created_at"`
}

func (h *AsanaWebhookHandshake) Insert(ctx context.Context, conn pg.Querier) error {
	q := `
INSERT INTO asana_webhook_handshakes (
    nonce,
    connector_id,
    secret_ciphertext,
    created_at
) VALUES (
    @nonce,
    @connector_id,
    @secret_ciphertext,
    @created_at
)
`

	_, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"nonce":             h.Nonce,
			"connector_id":      h.ConnectorID,
			"secret_ciphertext": h.SecretCiphertext,
			"created_at":        h.CreatedAt,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot store Asana webhook handshake: %w", err)
	}

	return nil
}

func (h *AsanaWebhookHandshake) StoreSecret(ctx context.Context, conn pg.Querier) error {
	q := `
UPDATE asana_webhook_handshakes
SET secret_ciphertext = @secret_ciphertext
WHERE nonce = @nonce
    AND connector_id = @connector_id
    AND octet_length(secret_ciphertext) = 0
RETURNING nonce
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"nonce":             h.Nonce,
			"connector_id":      h.ConnectorID,
			"secret_ciphertext": h.SecretCiphertext,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot store Asana webhook secret: %w", err)
	}

	_, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[string])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot store Asana webhook secret: %w", err)
	}

	return nil
}

func (h *AsanaWebhookHandshake) Load(ctx context.Context, conn pg.Querier, nonce string) error {
	q := `
SELECT
    nonce,
    connector_id,
    secret_ciphertext,
    created_at
FROM
    asana_webhook_handshakes
WHERE
    nonce = @nonce
LIMIT 1
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"nonce": nonce})
	if err != nil {
		return fmt.Errorf("cannot query Asana webhook handshake: %w", err)
	}

	loaded, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AsanaWebhookHandshake])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect Asana webhook handshake: %w", err)
	}

	*h = loaded

	return nil
}

func (h *AsanaWebhookHandshake) Delete(ctx context.Context, conn pg.Querier) error {
	q := `
DELETE FROM asana_webhook_handshakes
WHERE nonce = @nonce
`

	_, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"nonce": h.Nonce})
	if err != nil {
		return fmt.Errorf("cannot delete Asana webhook handshake: %w", err)
	}

	return nil
}
