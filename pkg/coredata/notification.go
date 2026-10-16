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
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	Notification struct {
		ID                 gid.GID          `db:"id"`
		OrganizationID     gid.GID          `db:"organization_id"`
		RecipientProfileID gid.GID          `db:"recipient_profile_id"`
		Type               NotificationType `db:"type"`
		SubjectID          gid.GID          `db:"subject_id"`
		Attributes         json.RawMessage  `db:"attributes"`
		ReadAt             *time.Time       `db:"read_at"`
		CreatedAt          time.Time        `db:"created_at"`
		UpdatedAt          time.Time        `db:"updated_at"`
	}

	Notifications []*Notification

	notificationAttributes struct {
		Title string `json:"title"`
	}
)

func (n Notification) CursorKey(orderBy NotificationOrderField) page.CursorKey {
	switch orderBy {
	case NotificationOrderFieldCreatedAt:
		return page.NewCursorKey(n.ID, n.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (n Notification) Title() (string, error) {
	var attributes notificationAttributes
	if err := json.Unmarshal(n.Attributes, &attributes); err != nil {
		return "", fmt.Errorf("cannot decode notification attributes: %w", err)
	}

	return attributes.Title, nil
}

func (n *Notification) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
    id,
    organization_id
FROM
    notifications
WHERE
    id = ANY(@resource_ids)
`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query notification authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan notification authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate notification authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (n *Notification) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	notificationID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    recipient_profile_id,
    type,
    subject_id,
    attributes,
    read_at,
    created_at,
    updated_at
FROM
    notifications
WHERE
    %s
    AND id = @notification_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"notification_id": notificationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query notification: %w", err)
	}

	notification, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Notification])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect notification: %w", err)
	}

	*n = notification

	return nil
}

func (n *Notification) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	if n.Attributes == nil {
		n.Attributes = json.RawMessage(`{}`)
	}

	q := `
INSERT INTO
    notifications (
        tenant_id,
        id,
        organization_id,
        recipient_profile_id,
        type,
        subject_id,
        attributes,
        read_at,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @notification_id,
    @organization_id,
    @recipient_profile_id,
    @type,
    @subject_id,
    @attributes,
    @read_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":            scope.GetTenantID(),
		"notification_id":      n.ID,
		"organization_id":      n.OrganizationID,
		"recipient_profile_id": n.RecipientProfileID,
		"type":                 n.Type,
		"subject_id":           n.SubjectID,
		"attributes":           n.Attributes,
		"read_at":              n.ReadAt,
		"created_at":           n.CreatedAt,
		"updated_at":           n.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert notification: %w", err)
	}

	return nil
}

func (n *Notification) MarkRead(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	recipientProfileID gid.GID,
	readAt time.Time,
) error {
	q := `
UPDATE
    notifications
SET
    read_at = @read_at,
    updated_at = @updated_at
WHERE
    %s
    AND id = @notification_id
    AND recipient_profile_id = @recipient_profile_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"notification_id":      n.ID,
		"recipient_profile_id": recipientProfileID,
		"read_at":              readAt,
		"updated_at":           readAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot mark notification read: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	n.ReadAt = &readAt
	n.UpdatedAt = readAt

	return nil
}

func (n *Notifications) LoadByRecipientProfileID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	recipientProfileID gid.GID,
	cursor *page.Cursor[NotificationOrderField],
	filter *NotificationFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
    recipient_profile_id,
    type,
    subject_id,
    attributes,
    read_at,
    created_at,
    updated_at
FROM
    notifications
WHERE
    %s
    AND organization_id = @organization_id
    AND recipient_profile_id = @recipient_profile_id
    AND %s
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id":      organizationID,
		"recipient_profile_id": recipientProfileID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query notifications: %w", err)
	}

	notifications, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Notification])
	if err != nil {
		return fmt.Errorf("cannot collect notifications: %w", err)
	}

	*n = notifications

	return nil
}

func (n *Notifications) CountByRecipientProfileID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	recipientProfileID gid.GID,
	filter *NotificationFilter,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    notifications
WHERE
    %s
    AND organization_id = @organization_id
    AND recipient_profile_id = @recipient_profile_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id":      organizationID,
		"recipient_profile_id": recipientProfileID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count notifications: %w", err)
	}

	return count, nil
}
