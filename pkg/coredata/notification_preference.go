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
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	NotificationPreference struct {
		OrganizationID     gid.GID                   `db:"organization_id"`
		RecipientProfileID gid.GID                   `db:"recipient_profile_id"`
		Type               NotificationType          `db:"notification_type"`
		EmailDelivery      NotificationEmailDelivery `db:"email_delivery"`
		CreatedAt          time.Time                 `db:"created_at"`
		UpdatedAt          time.Time                 `db:"updated_at"`
	}

	NotificationPreferences []*NotificationPreference
)

func (p *NotificationPreference) Upsert(ctx context.Context, conn pg.Tx) error {
	q := `
INSERT INTO notification_preferences (
	organization_id,
	recipient_profile_id,
	notification_type,
	email_delivery,
	created_at,
	updated_at
) VALUES (
	@organization_id,
	@recipient_profile_id,
	@notification_type,
	@email_delivery,
	@created_at,
	@updated_at
)
ON CONFLICT (recipient_profile_id, notification_type) DO UPDATE
SET
	email_delivery = EXCLUDED.email_delivery,
	updated_at = EXCLUDED.updated_at
RETURNING
	organization_id,
	recipient_profile_id,
	notification_type,
	email_delivery,
	created_at,
	updated_at
`

	args := pgx.StrictNamedArgs{
		"organization_id":      p.OrganizationID,
		"recipient_profile_id": p.RecipientProfileID,
		"notification_type":    p.Type,
		"email_delivery":       p.EmailDelivery,
		"created_at":           p.CreatedAt,
		"updated_at":           p.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert notification preference: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[NotificationPreference])
	if err != nil {
		return fmt.Errorf("cannot collect notification preference: %w", err)
	}

	*p = row

	return nil
}

func (p *NotificationPreferences) LoadByProfileID(
	ctx context.Context,
	conn pg.Querier,
	recipientProfileID gid.GID,
) error {
	q := `
SELECT
	organization_id,
	recipient_profile_id,
	notification_type,
	email_delivery,
	created_at,
	updated_at
FROM
	notification_preferences
WHERE
	recipient_profile_id = @recipient_profile_id
ORDER BY
	notification_type
`

	args := pgx.StrictNamedArgs{"recipient_profile_id": recipientProfileID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query notification preferences: %w", err)
	}

	preferences, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[NotificationPreference])
	if err != nil {
		return fmt.Errorf("cannot collect notification preferences: %w", err)
	}

	*p = preferences

	return nil
}
