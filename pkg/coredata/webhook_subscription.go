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
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	WebhookSubscription struct {
		ID                     gid.GID           `db:"id"`
		OrganizationID         gid.GID           `db:"organization_id"`
		EndpointURL            string            `db:"endpoint_url"`
		SelectedEvents         WebhookEventTypes `db:"selected_events"`
		EncryptedSigningSecret []byte            `db:"encrypted_signing_secret"`
		CreatedAt              time.Time         `db:"created_at"`
		UpdatedAt              time.Time         `db:"updated_at"`
	}

	WebhookSubscriptions []*WebhookSubscription
)

func (w *WebhookSubscription) GenerateSigningSecret(encryptionKey cipher.EncryptionKey) (string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("cannot generate signing secret: %w", err)
	}

	signingSecret := "whsec_" + hex.EncodeToString(secret)

	encrypted, err := cipher.Encrypt([]byte(signingSecret), encryptionKey)
	if err != nil {
		return "", fmt.Errorf("cannot encrypt signing secret: %w", err)
	}

	w.EncryptedSigningSecret = encrypted

	return signingSecret, nil
}

func (w *WebhookSubscription) DecryptSigningSecret(encryptionKey cipher.EncryptionKey) (string, error) {
	if len(w.EncryptedSigningSecret) == 0 {
		return "", fmt.Errorf("no encrypted signing secret")
	}

	plaintext, err := cipher.Decrypt(w.EncryptedSigningSecret, encryptionKey)
	if err != nil {
		return "", fmt.Errorf("cannot decrypt signing secret: %w", err)
	}

	return string(plaintext), nil
}

func (w WebhookSubscription) CursorKey(orderBy WebhookSubscriptionOrderField) page.CursorKey {
	switch orderBy {
	case WebhookSubscriptionOrderFieldCreatedAt:
		return page.NewCursorKey(w.ID, w.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (w *WebhookSubscription) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM webhook_subscriptions WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, w.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query webhook subscription authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (w *WebhookSubscription) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	webhookSubscriptionID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    endpoint_url,
    selected_events,
    encrypted_signing_secret,
    created_at,
    updated_at
FROM
    webhook_subscriptions
WHERE
    %s
    AND id = @webhook_subscription_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"webhook_subscription_id": webhookSubscriptionID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query webhook subscriptions: %w", err)
	}

	wc, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[WebhookSubscription])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect webhook subscription: %w", err)
	}

	*w = wc
	return nil
}

func (w *WebhookSubscriptions) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[WebhookSubscriptionOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    endpoint_url,
    selected_events,
    encrypted_signing_secret,
    created_at,
    updated_at
FROM
    webhook_subscriptions
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query webhook subscriptions: %w", err)
	}

	subscriptions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[WebhookSubscription])
	if err != nil {
		return fmt.Errorf("cannot collect webhook subscriptions: %w", err)
	}

	*w = subscriptions
	return nil
}

func (w *WebhookSubscriptions) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    webhook_subscriptions
WHERE
    %s
    AND organization_id = @organization_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count webhook subscriptions: %w", err)
	}

	return count, nil
}

func (w *WebhookSubscriptions) ExistsByOrganizationIDAndEventType(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	eventType WebhookEventType,
) (bool, error) {
	q := `
SELECT EXISTS (
    SELECT 1
    FROM webhook_subscriptions
    WHERE %s
        AND organization_id = @organization_id
        AND @event_type = ANY(selected_events)
)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"event_type":      eventType.String(),
	}
	maps.Copy(args, scope.SQLArguments())

	var exists bool
	if err := conn.QueryRow(ctx, q, args).Scan(&exists); err != nil {
		return false, fmt.Errorf("cannot check webhook subscription existence: %w", err)
	}

	return exists, nil
}

func (w *WebhookSubscription) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    webhook_subscriptions (
        tenant_id,
        id,
        organization_id,
        endpoint_url,
        selected_events,
        encrypted_signing_secret,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @webhook_subscription_id,
    @organization_id,
    @endpoint_url,
    @selected_events,
    @encrypted_signing_secret,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":               scope.GetTenantID(),
		"webhook_subscription_id": w.ID,
		"organization_id":         w.OrganizationID,
		"endpoint_url":            w.EndpointURL,
		"selected_events":         w.SelectedEvents,
		"encrypted_signing_secret": w.EncryptedSigningSecret,
		"created_at":              w.CreatedAt,
		"updated_at":              w.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		return fmt.Errorf("cannot insert webhook subscription: %w", err)
	}

	return nil
}

func (w *WebhookSubscription) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE webhook_subscriptions
SET
    endpoint_url = @endpoint_url,
    selected_events = @selected_events,
    updated_at = @updated_at
WHERE %s
    AND id = @webhook_subscription_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"webhook_subscription_id": w.ID,
		"endpoint_url":            w.EndpointURL,
		"selected_events":         w.SelectedEvents,
		"updated_at":              w.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update webhook subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (w *WebhookSubscriptions) LoadMatchingByOrganizationIDAndEventType(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	eventType WebhookEventType,
) error {
	q := `
SELECT
    id,
    organization_id,
    endpoint_url,
    selected_events,
    encrypted_signing_secret,
    created_at,
    updated_at
FROM
    webhook_subscriptions
WHERE
    %s
    AND organization_id = @organization_id
    AND @event_type = ANY(selected_events)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"event_type":      eventType.String(),
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query matching webhook subscriptions: %w", err)
	}

	subscriptions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[WebhookSubscription])
	if err != nil {
		return fmt.Errorf("cannot collect matching webhook subscriptions: %w", err)
	}

	*w = subscriptions
	return nil
}

func (w *WebhookSubscription) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM webhook_subscriptions
WHERE %s
    AND id = @webhook_subscription_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"webhook_subscription_id": w.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete webhook subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
