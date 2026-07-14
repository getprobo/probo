// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

var emptyJSONObject = json.RawMessage(`{}`)

type (
	Device struct {
		ID             gid.GID         `db:"id"`
		TenantID       gid.TenantID    `db:"tenant_id"`
		OrganizationID gid.GID         `db:"organization_id"`
		State          DeviceState     `db:"state"`
		HardwareUUID   *string         `db:"hardware_uuid"`
		SerialNumber   *string         `db:"serial_number"`
		Hostname       *string         `db:"hostname"`
		Platform       *DevicePlatform `db:"platform"`
		OSVersion      *string         `db:"os_version"`
		AgentVersion   *string         `db:"agent_version"`
		APIKeyHash     []byte          `db:"api_key_hash"`
		OwnerID        *gid.GID        `db:"owner_profile_id"`
		Labels         json.RawMessage `db:"labels"`
		EnrolledAt     *time.Time      `db:"enrolled_at"`
		LastSeenAt     *time.Time      `db:"last_seen_at"`
		RevokedAt      *time.Time      `db:"revoked_at"`
		CreatedAt      time.Time       `db:"created_at"`
		UpdatedAt      time.Time       `db:"updated_at"`
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
		hostname := ""
		if d.Hostname != nil {
			hostname = *d.Hostname
		}

		return page.NewCursorKey(d.ID, hostname)
	case DeviceOrderFieldLastSeenAt:
		lastSeen := time.Time{}
		if d.LastSeenAt != nil {
			lastSeen = *d.LastSeenAt
		}

		return page.NewCursorKey(d.ID, lastSeen)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (d *Device) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
    id,
    organization_id,
    owner_profile_id
FROM
    devices
WHERE
    id = ANY(@resource_ids::text[])
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"resource_ids": resourceIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot query device authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID, len(resourceIDs))
	ownerProfileByDeviceID := make(map[gid.GID]*gid.GID, len(resourceIDs))
	profileIDSet := make(map[gid.GID]struct{})

	for rows.Next() {
		var (
			id, organizationID gid.GID
			ownerProfileID     *gid.GID
		)
		if err := rows.Scan(&id, &organizationID, &ownerProfileID); err != nil {
			return nil, fmt.Errorf("cannot scan device authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
		ownerProfileByDeviceID[id] = ownerProfileID

		if ownerProfileID != nil {
			profileIDSet[*ownerProfileID] = struct{}{}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate device authorization attributes: %w", err)
	}

	if len(profileIDSet) > 0 {
		profileIDs := make([]gid.GID, 0, len(profileIDSet))
		for profileID := range profileIDSet {
			profileIDs = append(profileIDs, profileID)
		}

		var profile MembershipProfile

		profileAttrsByID, err := profile.AuthorizationAttributes(ctx, conn, profileIDs)
		if err != nil {
			return nil, fmt.Errorf("cannot load profile authorization attributes: %w", err)
		}

		identityByProfileID := make(map[gid.GID]string, len(profileAttrsByID))
		for profileID, profileAttrs := range profileAttrsByID {
			if identityID, ok := profileAttrs["identity_id"]; ok {
				identityByProfileID[profileID] = identityID
			}
		}

		for deviceID, ownerProfileID := range ownerProfileByDeviceID {
			if ownerProfileID == nil {
				continue
			}

			if ownerIdentityID, ok := identityByProfileID[*ownerProfileID]; ok {
				attrsByID[deviceID]["owner_id"] = ownerIdentityID
			}
		}
	}

	return attrsByID, nil
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
	state,
	hardware_uuid,
	serial_number,
	hostname,
	platform,
	os_version,
	agent_version,
	api_key_hash,
	owner_profile_id,
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

func (d *Device) LoadByIDForUpdate(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	deviceID gid.GID,
) error {
	q := `
SELECT
	id,
	tenant_id,
	organization_id,
	state,
	hardware_uuid,
	serial_number,
	hostname,
	platform,
	os_version,
	agent_version,
	api_key_hash,
	owner_profile_id,
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
LIMIT 1
FOR UPDATE;
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
	state,
	hardware_uuid,
	serial_number,
	hostname,
	platform,
	os_version,
	agent_version,
	api_key_hash,
	owner_profile_id,
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
	AND state != @revoked_state
LIMIT 1;
`

	args := pgx.StrictNamedArgs{
		"api_key_hash":  apiKeyHash,
		"revoked_state": DeviceStateRevoked,
	}

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
	state,
	hardware_uuid,
	serial_number,
	hostname,
	platform,
	os_version,
	agent_version,
	api_key_hash,
	owner_profile_id,
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
    state,
    hardware_uuid,
    serial_number,
    hostname,
    platform,
    os_version,
    agent_version,
    api_key_hash,
    owner_profile_id,
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
    @state,
    @hardware_uuid,
    @serial_number,
    @hostname,
    @platform,
    @os_version,
    @agent_version,
    @api_key_hash,
    @owner_profile_id,
    @labels,
    @enrolled_at,
    @last_seen_at,
    @revoked_at,
    @created_at,
    @updated_at
)
`
	args := pgx.StrictNamedArgs{
		"device_id":        d.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  d.OrganizationID,
		"state":            d.State,
		"hardware_uuid":    d.HardwareUUID,
		"serial_number":    d.SerialNumber,
		"hostname":         d.Hostname,
		"platform":         d.Platform,
		"os_version":       d.OSVersion,
		"agent_version":    d.AgentVersion,
		"api_key_hash":     d.APIKeyHash,
		"owner_profile_id": d.OwnerID,
		"labels":           labels,
		"enrolled_at":      d.EnrolledAt,
		"last_seen_at":     d.LastSeenAt,
		"revoked_at":       d.RevokedAt,
		"created_at":       d.CreatedAt,
		"updated_at":       d.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert device: %w", err)
	}

	return nil
}

func (d *Device) SetAPIKeyHash(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	apiKeyHash []byte,
) error {
	now := time.Now()

	q := fmt.Sprintf(`
UPDATE devices
SET
    api_key_hash = @api_key_hash,
    updated_at = @updated_at
WHERE %s
    AND id = @device_id
    AND state = @pending_state
    AND api_key_hash IS NULL
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"device_id":     d.ID,
		"api_key_hash":  apiKeyHash,
		"updated_at":    now,
		"pending_state": DeviceStatePending,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "devices_api_key_hash_idx" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot set device api key hash: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	d.APIKeyHash = apiKeyHash
	d.UpdatedAt = now

	return nil
}

func (d *Device) Activate(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	now := time.Now()

	q := fmt.Sprintf(`
UPDATE devices
SET
    hardware_uuid = @hardware_uuid,
    serial_number = @serial_number,
    hostname = @hostname,
    platform = @platform,
    os_version = @os_version,
    agent_version = @agent_version,
    state = @active_state,
    enrolled_at = @now,
    last_seen_at = @now,
    updated_at = @now
WHERE %s
    AND id = @device_id
    AND state = @pending_state
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"device_id":     d.ID,
		"hardware_uuid": d.HardwareUUID,
		"serial_number": d.SerialNumber,
		"hostname":      d.Hostname,
		"platform":      d.Platform,
		"os_version":    d.OSVersion,
		"agent_version": d.AgentVersion,
		"active_state":  DeviceStateActive,
		"pending_state": DeviceStatePending,
		"now":           now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "devices_org_hardware_uuid_idx" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot activate device: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	d.State = DeviceStateActive
	d.EnrolledAt = &now
	d.LastSeenAt = &now
	d.UpdatedAt = now

	return nil
}

func (d *Device) UpdateHeartbeat(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	now := time.Now()

	q := fmt.Sprintf(`
UPDATE devices
SET
    hostname = @hostname,
    os_version = @os_version,
    agent_version = @agent_version,
    last_seen_at = @last_seen_at,
    updated_at = @updated_at
WHERE %s
    AND id = @device_id
    AND state = @active_state
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"device_id":     d.ID,
		"hostname":      d.Hostname,
		"os_version":    d.OSVersion,
		"agent_version": d.AgentVersion,
		"last_seen_at":  now,
		"updated_at":    now,
		"active_state":  DeviceStateActive,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update device heartbeat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	d.LastSeenAt = &now
	d.UpdatedAt = now

	return nil
}

func (d *Device) Revoke(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	now := time.Now()

	q := fmt.Sprintf(`
UPDATE devices
SET
    state = @revoked_state,
    revoked_at = COALESCE(revoked_at, @now),
    updated_at = @now
WHERE %s
    AND id = @device_id
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"device_id":     d.ID,
		"revoked_state": DeviceStateRevoked,
		"now":           now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot revoke device: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	d.State = DeviceStateRevoked
	if d.RevokedAt == nil {
		d.RevokedAt = &now
	}

	d.UpdatedAt = now

	return nil
}

func (d *Device) AssignOwner(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	ownerProfileID *gid.GID,
) error {
	now := time.Now()

	q := fmt.Sprintf(`
UPDATE devices
SET
    owner_profile_id = @owner_profile_id,
    updated_at = @now
WHERE %s
    AND id = @device_id
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"device_id":        d.ID,
		"owner_profile_id": ownerProfileID,
		"now":              now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot assign device user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	d.OwnerID = ownerProfileID
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
	state,
	hardware_uuid,
	serial_number,
	hostname,
	platform,
	os_version,
	agent_version,
	api_key_hash,
	owner_profile_id,
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

func (ds *Devices) LoadByOrganizationIDAndOwnerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	ownerID gid.GID,
	cursor *page.Cursor[DeviceOrderField],
) error {
	q := `
SELECT
	id,
	tenant_id,
	organization_id,
	state,
	hardware_uuid,
	serial_number,
	hostname,
	platform,
	os_version,
	agent_version,
	api_key_hash,
	owner_profile_id,
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
	AND owner_profile_id = @owner_profile_id
	AND state = @active_state
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id":  organizationID,
		"owner_profile_id": ownerID,
		"active_state":     DeviceStateActive,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query devices by owner: %w", err)
	}

	devices, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Device])
	if err != nil {
		return fmt.Errorf("cannot collect devices by owner: %w", err)
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
	q := fmt.Sprintf(`
SELECT COUNT(id) FROM devices
WHERE %s AND organization_id = @organization_id
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count devices: %w", err)
	}

	return count, nil
}

func (ds *Devices) CountByOrganizationIDAndOwnerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	ownerID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	devices
WHERE
	%s
	AND organization_id = @organization_id
	AND owner_profile_id = @owner_profile_id
	AND state = @active_state
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id":  organizationID,
		"owner_profile_id": ownerID,
		"active_state":     DeviceStateActive,
	}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count devices by owner: %w", err)
	}

	return count, nil
}
