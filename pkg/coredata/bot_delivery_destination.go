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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type BotDeliveryDestination struct {
	ID                    gid.GID    `db:"id"`
	OrganizationID        gid.GID    `db:"organization_id"`
	Provider              string     `db:"provider"`
	TargetNamespace       string     `db:"target_namespace"`
	TargetKey             string     `db:"target_key"`
	ExternalDestinationID string     `db:"external_destination_id"`
	ExternalName          string     `db:"external_name"`
	VerifiedAt            *time.Time `db:"verified_at"`
	CreatedAt             time.Time  `db:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at"`
}

func NewBotDeliveryDestination(
	scope Scoper,
	organizationID gid.GID,
	provider string,
	targetNamespace string,
	targetKey string,
) *BotDeliveryDestination {
	now := time.Now()

	return &BotDeliveryDestination{
		ID:              gid.New(scope.GetTenantID(), BotDeliveryDestinationEntityType),
		OrganizationID:  organizationID,
		Provider:        provider,
		TargetNamespace: targetNamespace,
		TargetKey:       targetKey,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (d *BotDeliveryDestination) LoadByTarget(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	provider string,
	targetNamespace string,
	targetKey string,
) error {
	q := `
SELECT
	id,
	organization_id,
	provider,
	target_namespace,
	target_key,
	external_destination_id,
	external_name,
	verified_at,
	created_at,
	updated_at
FROM bot_delivery_destinations
WHERE %s
	AND provider = @provider
	AND target_namespace = @target_namespace
	AND target_key = @target_key
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"provider":         provider,
		"target_namespace": targetNamespace,
		"target_key":       targetKey,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query bot delivery destination: %w", err)
	}

	destination, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[BotDeliveryDestination])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect bot delivery destination: %w", err)
	}

	*d = destination

	return nil
}

func (d *BotDeliveryDestination) Upsert(ctx context.Context, conn pg.Tx, scope Scoper) (bool, error) {
	q := `
INSERT INTO bot_delivery_destinations (
	id,
	tenant_id,
	organization_id,
	provider,
	target_namespace,
	target_key,
	external_destination_id,
	external_name,
	verified_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@provider,
	@target_namespace,
	@target_key,
	@external_destination_id,
	@external_name,
	@verified_at,
	@created_at,
	@updated_at
)
ON CONFLICT (provider, target_namespace, target_key) DO UPDATE SET
	organization_id = EXCLUDED.organization_id,
	external_destination_id = EXCLUDED.external_destination_id,
	external_name = EXCLUDED.external_name,
	verified_at = EXCLUDED.verified_at,
	updated_at = EXCLUDED.updated_at
RETURNING
	id,
	organization_id,
	provider,
	target_namespace,
	target_key,
	external_destination_id,
	external_name,
	verified_at,
	created_at,
	updated_at
`
	originalID := d.ID
	args := pgx.StrictNamedArgs{
		"id":                      d.ID,
		"tenant_id":               scope.GetTenantID(),
		"organization_id":         d.OrganizationID,
		"provider":                d.Provider,
		"target_namespace":        d.TargetNamespace,
		"target_key":              d.TargetKey,
		"external_destination_id": d.ExternalDestinationID,
		"external_name":           d.ExternalName,
		"verified_at":             d.VerifiedAt,
		"created_at":              d.CreatedAt,
		"updated_at":              d.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert bot delivery destination: %w", err)
	}
	defer rows.Close()

	destination, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[BotDeliveryDestination])
	if err != nil {
		return false, fmt.Errorf("cannot collect bot delivery destination upsert result: %w", err)
	}

	*d = destination

	return originalID == d.ID, nil
}

func (d *BotDeliveryDestination) MarkVerified(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE bot_delivery_destinations
SET
	verified_at = @verified_at,
	updated_at = @updated_at
WHERE %s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":          d.ID,
		"verified_at": d.VerifiedAt,
		"updated_at":  d.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot verify bot delivery destination: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func DeleteBotDeliveryDestinationsByProviderAndOrganizationID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	provider string,
	organizationID gid.GID,
) error {
	q := `
DELETE FROM bot_delivery_destinations
WHERE %s
	AND provider = @provider
	AND organization_id = @organization_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"provider":        provider,
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete bot delivery destinations: %w", err)
	}

	return nil
}

func (d *BotDeliveryDestination) Delete(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `DELETE FROM bot_delivery_destinations WHERE %s AND id = @id`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": d.ID}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete bot delivery destination: %w", err)
	}

	return nil
}
