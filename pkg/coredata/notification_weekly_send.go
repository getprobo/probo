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

type WeeklyNotificationRecipient struct {
	OrganizationID     gid.GID `db:"organization_id"`
	RecipientProfileID gid.GID `db:"recipient_profile_id"`
}

// ClaimNextWeeklyNotificationRecipient records this week for one membership
// profile that has at least one weekly email preference. An empty digest still
// keeps the row, so the same week is not claimed again.
func ClaimNextWeeklyNotificationRecipient(
	ctx context.Context,
	conn pg.Tx,
	weekStart time.Time,
	now time.Time,
) (WeeklyNotificationRecipient, error) {
	q := `
INSERT INTO notification_weekly_sends (
	organization_id,
	recipient_profile_id,
	week_start,
	created_at
)
SELECT
	organization_id,
	recipient_profile_id,
	@week_start::date,
	@now
FROM
	notification_preferences
WHERE
	email_delivery = @weekly
	AND NOT EXISTS (
		SELECT 1
		FROM notification_weekly_sends existing
		WHERE existing.recipient_profile_id = notification_preferences.recipient_profile_id
			AND existing.week_start = @week_start::date
	)
ORDER BY
	recipient_profile_id
LIMIT 1
ON CONFLICT (recipient_profile_id, week_start) DO NOTHING
RETURNING
	organization_id,
	recipient_profile_id
`

	args := pgx.StrictNamedArgs{
		"week_start": weekStart,
		"now":        now,
		"weekly":     NotificationEmailDeliveryWeekly,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return WeeklyNotificationRecipient{}, fmt.Errorf("cannot claim weekly notification recipient: %w", err)
	}

	recipient, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[WeeklyNotificationRecipient])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WeeklyNotificationRecipient{}, ErrResourceNotFound
		}

		return WeeklyNotificationRecipient{}, fmt.Errorf("cannot collect weekly notification recipient: %w", err)
	}

	return recipient, nil
}
