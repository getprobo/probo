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

type NotificationEmailGroup struct {
	OrganizationID     gid.GID          `db:"organization_id"`
	RecipientProfileID gid.GID          `db:"recipient_profile_id"`
	Type               NotificationType `db:"type"`
	CreatedAt          time.Time        `db:"created_at"`
}

func (g *NotificationEmailGroup) LoadNextEachDue(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
	debounceBefore time.Time,
) error {
	q := `
SELECT
    organization_id,
    recipient_profile_id,
    type,
    created_at
FROM
    notifications
WHERE
    status = @pending
    AND email_count = 0
    AND created_at < @debounce_before
    AND COALESCE(
        (
            SELECT
                notification_preferences.email_delivery
            FROM
                notification_preferences
            WHERE
                notification_preferences.recipient_profile_id = notifications.recipient_profile_id
                AND notification_preferences.notification_type = notifications.type
        ),
        CASE notifications.type
            WHEN @task_assigned::notification_type THEN @none::notification_email_delivery
            WHEN @task_mentioned::notification_type THEN @none::notification_email_delivery
            ELSE @reminders::notification_email_delivery
        END
    ) = @each::notification_email_delivery
    AND EXISTS (
        SELECT
            1
        FROM
            iam_membership_profiles
        WHERE
            iam_membership_profiles.id = notifications.recipient_profile_id
            AND iam_membership_profiles.state = @recipient_state
            AND (
                iam_membership_profiles.contract_end_date IS NULL
                OR iam_membership_profiles.contract_end_date >= @now::date
            )
    )
ORDER BY
    created_at
LIMIT 1
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"pending":         NotificationStatusPending,
			"debounce_before": debounceBefore,
			"task_assigned":   NotificationTypeTaskAssigned,
			"task_mentioned":  NotificationTypeTaskMentioned,
			"none":            NotificationEmailDeliveryNone,
			"each":            NotificationEmailDeliveryEach,
			"reminders":       NotificationEmailDeliveryReminders,
			"recipient_state": ProfileStateActive,
			"now":             now,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot query next each notification email: %w", err)
	}

	group, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[NotificationEmailGroup])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect next each notification email: %w", err)
	}

	*g = group

	return nil
}

func (g *NotificationEmailGroup) LoadNextRemindersDue(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
	debounceBefore time.Time,
	reminderInterval time.Duration,
) error {
	q := `
SELECT
    organization_id,
    recipient_profile_id,
    type,
    created_at
FROM
    notifications
WHERE
    status = @pending
    AND email_count < @max_notifications
    AND (
        (
            email_count = 0
            AND created_at < @debounce_before
        )
        OR (
            email_count > 0
            AND EXTRACT(ISODOW FROM @now AT TIME ZONE 'UTC') BETWEEN 1 AND 5
            AND @now > (
                last_emailed_at
                    + make_interval(secs => @reminder_interval_seconds * email_count)
                    + CASE EXTRACT(ISODOW FROM (last_emailed_at + make_interval(secs => @reminder_interval_seconds * email_count)) AT TIME ZONE 'UTC')
                        WHEN 6 THEN INTERVAL '2 days'
                        WHEN 7 THEN INTERVAL '1 day'
                        ELSE INTERVAL '0'
                    END
            )
        )
    )
    AND COALESCE(
        (
            SELECT
                notification_preferences.email_delivery
            FROM
                notification_preferences
            WHERE
                notification_preferences.recipient_profile_id = notifications.recipient_profile_id
                AND notification_preferences.notification_type = notifications.type
        ),
        CASE notifications.type
            WHEN @task_assigned::notification_type THEN @none::notification_email_delivery
            WHEN @task_mentioned::notification_type THEN @none::notification_email_delivery
            ELSE @reminders::notification_email_delivery
        END
    ) = @reminders::notification_email_delivery
    AND EXISTS (
        SELECT
            1
        FROM
            iam_membership_profiles
        WHERE
            iam_membership_profiles.id = notifications.recipient_profile_id
            AND iam_membership_profiles.state = @recipient_state
            AND (
                iam_membership_profiles.contract_end_date IS NULL
                OR iam_membership_profiles.contract_end_date >= @now::date
            )
    )
ORDER BY
    created_at
LIMIT 1
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"pending":                   NotificationStatusPending,
			"max_notifications":         NotificationEmailMaxCount,
			"debounce_before":           debounceBefore,
			"now":                       now,
			"reminder_interval_seconds": reminderInterval.Seconds(),
			"task_assigned":             NotificationTypeTaskAssigned,
			"task_mentioned":            NotificationTypeTaskMentioned,
			"none":                      NotificationEmailDeliveryNone,
			"reminders":                 NotificationEmailDeliveryReminders,
			"recipient_state":           ProfileStateActive,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot query next reminders notification email: %w", err)
	}

	group, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[NotificationEmailGroup])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect next reminders notification email: %w", err)
	}

	*g = group

	return nil
}

func (ns *Notifications) UpdateDueEach(
	ctx context.Context,
	conn pg.Tx,
	group NotificationEmailGroup,
	now time.Time,
	debounceBefore time.Time,
) error {
	q := `
UPDATE
    notifications
SET
    email_count = email_count + 1,
    last_emailed_at = @now,
    updated_at = @now
WHERE
    organization_id = @organization_id
    AND recipient_profile_id = @recipient_profile_id
    AND type = @notification_type
    AND status = @pending
    AND email_count = 0
    AND created_at < @debounce_before
    AND COALESCE(
        (
            SELECT
                notification_preferences.email_delivery
            FROM
                notification_preferences
            WHERE
                notification_preferences.recipient_profile_id = notifications.recipient_profile_id
                AND notification_preferences.notification_type = notifications.type
        ),
        CASE notifications.type
            WHEN @task_assigned::notification_type THEN @none::notification_email_delivery
            WHEN @task_mentioned::notification_type THEN @none::notification_email_delivery
            ELSE @reminders::notification_email_delivery
        END
    ) = @each::notification_email_delivery
RETURNING
    id,
    organization_id,
    recipient_profile_id,
    type,
    subject_id,
    attributes,
    status,
    read_at,
    email_count,
    last_emailed_at,
    created_at,
    updated_at
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"organization_id":      group.OrganizationID,
			"recipient_profile_id": group.RecipientProfileID,
			"notification_type":    group.Type,
			"pending":              NotificationStatusPending,
			"debounce_before":      debounceBefore,
			"now":                  now,
			"task_assigned":        NotificationTypeTaskAssigned,
			"task_mentioned":       NotificationTypeTaskMentioned,
			"none":                 NotificationEmailDeliveryNone,
			"each":                 NotificationEmailDeliveryEach,
			"reminders":            NotificationEmailDeliveryReminders,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot update due each notification emails: %w", err)
	}

	claimed, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Notification])
	if err != nil {
		return fmt.Errorf("cannot collect due each notification emails: %w", err)
	}

	*ns = claimed

	return nil
}

func (ns *Notifications) UpdateDueReminders(
	ctx context.Context,
	conn pg.Tx,
	group NotificationEmailGroup,
	now time.Time,
	debounceBefore time.Time,
	reminderInterval time.Duration,
) error {
	q := `
UPDATE
    notifications
SET
    email_count = email_count + 1,
    last_emailed_at = @now,
    updated_at = @now
WHERE
    organization_id = @organization_id
    AND recipient_profile_id = @recipient_profile_id
    AND type = @notification_type
    AND status = @pending
    AND email_count < @max_notifications
    AND (
        (
            email_count = 0
            AND created_at < @debounce_before
        )
        OR (
            email_count > 0
            AND EXTRACT(ISODOW FROM @now AT TIME ZONE 'UTC') BETWEEN 1 AND 5
            AND @now > (
                last_emailed_at
                    + make_interval(secs => @reminder_interval_seconds * email_count)
                    + CASE EXTRACT(ISODOW FROM (last_emailed_at + make_interval(secs => @reminder_interval_seconds * email_count)) AT TIME ZONE 'UTC')
                        WHEN 6 THEN INTERVAL '2 days'
                        WHEN 7 THEN INTERVAL '1 day'
                        ELSE INTERVAL '0'
                    END
            )
        )
    )
    AND COALESCE(
        (
            SELECT
                notification_preferences.email_delivery
            FROM
                notification_preferences
            WHERE
                notification_preferences.recipient_profile_id = notifications.recipient_profile_id
                AND notification_preferences.notification_type = notifications.type
        ),
        CASE notifications.type
            WHEN @task_assigned::notification_type THEN @none::notification_email_delivery
            WHEN @task_mentioned::notification_type THEN @none::notification_email_delivery
            ELSE @reminders::notification_email_delivery
        END
    ) = @reminders::notification_email_delivery
RETURNING
    id,
    organization_id,
    recipient_profile_id,
    type,
    subject_id,
    attributes,
    status,
    read_at,
    email_count,
    last_emailed_at,
    created_at,
    updated_at
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"organization_id":           group.OrganizationID,
			"recipient_profile_id":      group.RecipientProfileID,
			"notification_type":         group.Type,
			"pending":                   NotificationStatusPending,
			"max_notifications":         NotificationEmailMaxCount,
			"debounce_before":           debounceBefore,
			"now":                       now,
			"reminder_interval_seconds": reminderInterval.Seconds(),
			"task_assigned":             NotificationTypeTaskAssigned,
			"task_mentioned":            NotificationTypeTaskMentioned,
			"none":                      NotificationEmailDeliveryNone,
			"reminders":                 NotificationEmailDeliveryReminders,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot update due reminders notification emails: %w", err)
	}

	claimed, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Notification])
	if err != nil {
		return fmt.Errorf("cannot collect due reminders notification emails: %w", err)
	}

	*ns = claimed

	return nil
}

// UpdateOpenReminders advances the other pending reminders in the group so
// one email covers the batch.
func (ns *Notifications) UpdateOpenReminders(
	ctx context.Context,
	conn pg.Tx,
	group NotificationEmailGroup,
	now time.Time,
) error {
	q := `
UPDATE
    notifications
SET
    email_count = email_count + 1,
    last_emailed_at = @now,
    updated_at = @now
WHERE
    organization_id = @organization_id
    AND recipient_profile_id = @recipient_profile_id
    AND type = @notification_type
    AND status = @pending
    AND email_count < @max_notifications
    AND last_emailed_at IS DISTINCT FROM @now
    AND COALESCE(
        (
            SELECT
                notification_preferences.email_delivery
            FROM
                notification_preferences
            WHERE
                notification_preferences.recipient_profile_id = notifications.recipient_profile_id
                AND notification_preferences.notification_type = notifications.type
        ),
        CASE notifications.type
            WHEN @task_assigned::notification_type THEN @none::notification_email_delivery
            WHEN @task_mentioned::notification_type THEN @none::notification_email_delivery
            ELSE @reminders::notification_email_delivery
        END
    ) = @reminders::notification_email_delivery
`

	_, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"organization_id":      group.OrganizationID,
			"recipient_profile_id": group.RecipientProfileID,
			"notification_type":    group.Type,
			"pending":              NotificationStatusPending,
			"max_notifications":    NotificationEmailMaxCount,
			"now":                  now,
			"task_assigned":        NotificationTypeTaskAssigned,
			"task_mentioned":       NotificationTypeTaskMentioned,
			"none":                 NotificationEmailDeliveryNone,
			"reminders":            NotificationEmailDeliveryReminders,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot update open reminder emails: %w", err)
	}

	return nil
}
