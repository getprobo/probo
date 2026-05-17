// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

var emptyJSONObject = json.RawMessage(`{}`)

type (
	Device struct {
		ID                     gid.GID              `db:"id"`
		TenantID               gid.TenantID         `db:"tenant_id"`
		OrganizationID         gid.GID              `db:"organization_id"`
		HardwareUUID           string               `db:"hardware_uuid"`
		SerialNumber           *string              `db:"serial_number"`
		Hostname               string               `db:"hostname"`
		Platform               DevicePlatform       `db:"platform"`
		OSVersion              string               `db:"os_version"`
		AgentVersion           string               `db:"agent_version"`
		APIKeyHash             []byte               `db:"api_key_hash"`
		AssignedUserIdentityID *gid.GID             `db:"assigned_user_identity_id"`
		Labels                 json.RawMessage      `db:"labels"`
		EnrolledAt             time.Time            `db:"enrolled_at"`
		LastSeenAt             time.Time            `db:"last_seen_at"`
		RevokedAt              *time.Time           `db:"revoked_at"`
		CreatedAt              time.Time            `db:"created_at"`
		UpdatedAt              time.Time            `db:"updated_at"`
	}

	Devices []*Device
)

func (d *Device) CursorKey(orderBy DeviceOrderField) page.CursorKey {
	switch orderBy {
	case DeviceOrderFieldCreatedAt:
		return page.NewCursorKey(d.ID, d.CreatedAt)
	case DeviceOrderFieldUpdatedAt:
		return page.NewCursorKey(d.ID, d.UpdatedAt)
	case DeviceOrderFieldHostname:
		return page.NewCursorKey(d.ID, d.Hostname)
	case DeviceOrderFieldLastSeenAt:
		return page.NewCursorKey(d.ID, d.LastSeenAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (d *Device) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM devices WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, d.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query device authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (d *Device) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	deviceID gid.GID,
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    hardware_uuid,
    serial_number,
    hostname,
    platform,
    os_version,
    agent_version,
    api_key_hash,
    assigned_user_identity_id,
    labels,
    enrolled_at,
    last_seen_at,
    revoked_at,
    created_at,
    updated_at
FROM
    devices
WHERE
    %s
    AND id = @device_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"device_id": deviceID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query device: %w", err)
	}

	device, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Device])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect device: %w", err)
	}

	*d = device
	return nil
}

// LoadByAPIKeyHash loads a non-revoked device by its API key hash without
// requiring a tenant scope (the device key itself is the credential).
func (d *Device) LoadByAPIKeyHash(
	ctx context.Context,
	conn pg.Querier,
	apiKeyHash []byte,
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    hardware_uuid,
    serial_number,
    hostname,
    platform,
    os_version,
    agent_version,
    api_key_hash,
    assigned_user_identity_id,
    labels,
    enrolled_at,
    last_seen_at,
    revoked_at,
    created_at,
    updated_at
FROM
    devices
WHERE
    api_key_hash = @api_key_hash
    AND revoked_at IS NULL
LIMIT 1;
`
	args := pgx.StrictNamedArgs{"api_key_hash": apiKeyHash}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query device by api key hash: %w", err)
	}

	device, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Device])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect device: %w", err)
	}

	*d = device
	return nil
}

func (d *Device) LoadByHardwareUUID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	hardwareUUID string,
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    hardware_uuid,
    serial_number,
    hostname,
    platform,
    os_version,
    agent_version,
    api_key_hash,
    assigned_user_identity_id,
    labels,
    enrolled_at,
    last_seen_at,
    revoked_at,
    created_at,
    updated_at
FROM
    devices
WHERE
    %s
    AND organization_id = @organization_id
    AND hardware_uuid = @hardware_uuid
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"hardware_uuid":   hardwareUUID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query device by hardware uuid: %w", err)
	}

	device, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Device])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect device: %w", err)
	}

	*d = device
	return nil
}

func (d Device) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	labels := d.Labels
	if len(labels) == 0 {
		labels = emptyJSONObject
	}

	q := `
INSERT INTO devices (
    id,
    tenant_id,
    organization_id,
    hardware_uuid,
    serial_number,
    hostname,
    platform,
    os_version,
    agent_version,
    api_key_hash,
    assigned_user_identity_id,
    labels,
    enrolled_at,
    last_seen_at,
    revoked_at,
    created_at,
    updated_at
) VALUES (
    @device_id,
    @tenant_id,
    @organization_id,
    @hardware_uuid,
    @serial_number,
    @hostname,
    @platform,
    @os_version,
    @agent_version,
    @api_key_hash,
    @assigned_user_identity_id,
    @labels,
    @enrolled_at,
    @last_seen_at,
    @revoked_at,
    @created_at,
    @updated_at
)
`
	args := pgx.StrictNamedArgs{
		"device_id":                 d.ID,
		"tenant_id":                 scope.GetTenantID(),
		"organization_id":           d.OrganizationID,
		"hardware_uuid":             d.HardwareUUID,
		"serial_number":             d.SerialNumber,
		"hostname":                  d.Hostname,
		"platform":                  d.Platform,
		"os_version":                d.OSVersion,
		"agent_version":             d.AgentVersion,
		"api_key_hash":              d.APIKeyHash,
		"assigned_user_identity_id": d.AssignedUserIdentityID,
		"labels":                    labels,
		"enrolled_at":               d.EnrolledAt,
		"last_seen_at":              d.LastSeenAt,
		"revoked_at":                d.RevokedAt,
		"created_at":                d.CreatedAt,
		"updated_at":                d.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert device: %w", err)
	}
	return nil
}

// Reenroll rotates the API key and refreshes the device metadata for an
// existing (organization_id, hardware_uuid) row.
func (d *Device) Reenroll(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	now := time.Now()

	q := `
UPDATE devices
SET
    serial_number = @serial_number,
    hostname = @hostname,
    platform = @platform,
    os_version = @os_version,
    agent_version = @agent_version,
    api_key_hash = @api_key_hash,
    revoked_at = NULL,
    last_seen_at = @now,
    updated_at = @now
WHERE %s
    AND id = @device_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"device_id":     d.ID,
		"serial_number": d.SerialNumber,
		"hostname":      d.Hostname,
		"platform":      d.Platform,
		"os_version":    d.OSVersion,
		"agent_version": d.AgentVersion,
		"api_key_hash":  d.APIKeyHash,
		"now":           now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot re-enroll device: %w", err)
	}

	d.LastSeenAt = now
	d.UpdatedAt = now
	d.RevokedAt = nil
	return nil
}

// UpdateHeartbeat updates the volatile last-seen / version columns. Used by
// the heartbeat handler on every check-in.
func (d *Device) UpdateHeartbeat(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	now := time.Now()

	q := `
UPDATE devices
SET
    hostname = @hostname,
    os_version = @os_version,
    agent_version = @agent_version,
    last_seen_at = @last_seen_at,
    updated_at = @updated_at
WHERE %s
    AND id = @device_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"device_id":     d.ID,
		"hostname":      d.Hostname,
		"os_version":    d.OSVersion,
		"agent_version": d.AgentVersion,
		"last_seen_at":  now,
		"updated_at":    now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot update device heartbeat: %w", err)
	}

	d.LastSeenAt = now
	d.UpdatedAt = now
	return nil
}

// Revoke marks the device as revoked. Subsequent agent calls authenticated
// with the device's API key will be rejected by the middleware.
func (d *Device) Revoke(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	now := time.Now()

	q := `
UPDATE devices
SET
    revoked_at = COALESCE(revoked_at, @now),
    updated_at = @now
WHERE %s
    AND id = @device_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"device_id": d.ID,
		"now":       now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot revoke device: %w", err)
	}

	if d.RevokedAt == nil {
		d.RevokedAt = &now
	}
	d.UpdatedAt = now
	return nil
}

func (d *Device) AssignUser(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	identityID *gid.GID,
) error {
	now := time.Now()

	q := `
UPDATE devices
SET
    assigned_user_identity_id = @identity_id,
    updated_at = @now
WHERE %s
    AND id = @device_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"device_id":   d.ID,
		"identity_id": identityID,
		"now":         now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot assign device user: %w", err)
	}

	d.AssignedUserIdentityID = identityID
	d.UpdatedAt = now
	return nil
}

func (ds *Devices) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[DeviceOrderField],
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    hardware_uuid,
    serial_number,
    hostname,
    platform,
    os_version,
    agent_version,
    api_key_hash,
    assigned_user_identity_id,
    labels,
    enrolled_at,
    last_seen_at,
    revoked_at,
    created_at,
    updated_at
FROM
    devices
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query devices: %w", err)
	}

	devices, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Device])
	if err != nil {
		return fmt.Errorf("cannot collect devices: %w", err)
	}

	*ds = devices
	return nil
}

func (ds *Devices) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id) FROM devices
WHERE %s AND organization_id = @organization_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count devices: %w", err)
	}
	return count, nil
}
