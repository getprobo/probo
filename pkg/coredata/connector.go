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
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

// jsonRawMessageOrNull is a json.RawMessage that scans NULL as an empty
// slice and serialises an empty/nil value as SQL NULL.  This avoids the
// need for *json.RawMessage and keeps the zero-value useful.
type jsonRawMessageOrNull json.RawMessage

func (j *jsonRawMessageOrNull) Scan(src any) error {
	if src == nil {
		*j = nil
		return nil
	}

	switch v := src.(type) {
	case []byte:
		cp := make(jsonRawMessageOrNull, len(v))
		copy(cp, v)
		*j = cp

		return nil
	case string:
		*j = jsonRawMessageOrNull(v)
		return nil
	default:
		return fmt.Errorf("unsupported type for jsonRawMessageOrNull: %T", src)
	}
}

type (
	Connector struct {
		ID                  gid.GID              `db:"id"`
		OrganizationID      gid.GID              `db:"organization_id"`
		Provider            ConnectorProvider    `db:"provider"`
		Protocol            ConnectorProtocol    `db:"protocol"`
		RawSettings         jsonRawMessageOrNull `db:"settings"`
		Connection          connector.Connection `db:"-"`
		EncryptedConnection []byte               `db:"encrypted_connection"`
		CreatedAt           time.Time            `db:"created_at"`
		UpdatedAt           time.Time            `db:"updated_at"`
	}

	Connectors []*Connector
)

func (c *Connector) CursorKey(orderBy ConnectorOrderField) page.CursorKey {
	switch orderBy {
	case ConnectorOrderFieldCreatedAt:
		return page.CursorKey{ID: c.ID, Value: c.CreatedAt}
	case ConnectorOrderFieldProvider:
		return page.CursorKey{ID: c.ID, Value: c.Provider}
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (c *Connector) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM connectors WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
}

// LoadSlackMessagingConnector resolves the Slack connector the legacy
// messaging fallback sends with (probot delivers via its own
// installation tokens, not this table). The pick is deterministic
// under several Slack rows: channel-configured settings win — only the
// legacy messaging connect flow ever captured one — then oldest
// created_at, then id. Returns ErrResourceNotFound if no OAuth2 Slack
// row exists.
func (c *Connector) LoadSlackMessagingConnector(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
	organizationID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    provider,
    protocol,
    settings,
    encrypted_connection,
    created_at,
    updated_at
FROM
    connectors
WHERE
    %s
    AND organization_id = @organization_id
    AND provider = @provider
    AND protocol = @protocol
ORDER BY
    (COALESCE(settings->>'channel_id', '') <> '') DESC,
    created_at ASC,
    id ASC
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"provider":        ConnectorProviderSlack,
		"protocol":        ConnectorProtocolOAuth2,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query connectors: %w", err)
	}

	loadedConnector, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Connector])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect connector row: %w", err)
	}

	*c = loadedConnector

	if err := c.decryptConnection(encryptionKey); err != nil {
		return fmt.Errorf("cannot decrypt connection: %w", err)
	}

	return nil
}

func (c *Connectors) LoadByOrganizationIDWithoutDecryptedConnection(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ConnectorOrderField],
	filter *ConnectorFilter,
) error {
	return c.loadByOrganizationIDWithPagination(ctx, conn, scope, organizationID, cursor, filter)
}

func (c *Connectors) LoadAllByOrganizationIDWithoutDecryptedConnection(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	return c.loadAllByOrganizationID(ctx, conn, scope, organizationID)
}

func (c *Connector) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorID gid.GID,
	encryptionKey cipher.EncryptionKey,
) error {
	if err := c.LoadMetadataByID(ctx, conn, scope, connectorID); err != nil {
		return err
	}

	// Decrypt the connection
	if len(c.EncryptedConnection) > 0 {
		decryptedConnection, err := cipher.Decrypt(c.EncryptedConnection, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt connection: %w", err)
		}

		c.Connection, err = connector.UnmarshalConnection(c.Protocol.String(), c.Provider.String(), decryptedConnection)
		if err != nil {
			return fmt.Errorf("cannot unmarshal connection: %w", err)
		}

		if c.Provider == ConnectorProviderSlack {
			if slackConn, ok := c.Connection.(*connector.SlackConnection); ok {
				settings, _ := ConnectorSettings[SlackConnectorSettings](c)
				slackConn.Settings.Channel = settings.Channel
				slackConn.Settings.ChannelID = settings.ChannelID
			}
		}
	}

	return nil
}

// LoadMetadataByID loads connector metadata without decrypting the connection.
// Use this when you only need provider, organization, or other metadata.
func (c *Connector) LoadMetadataByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    provider,
    protocol,
    settings,
    encrypted_connection,
    created_at,
    updated_at
FROM
    connectors
WHERE
    %s
    AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": connectorID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query connectors: %w", err)
	}

	loadedConnector, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Connector])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect connector row: %w", err)
	}

	*c = loadedConnector

	return nil
}

// LoadByOrganizationIDProviderAndSettingForUpdate loads the connector in
// organizationID for provider whose settings carry settingValue at settingKey,
// locking the row for the caller's transaction. It is the idempotency lookup for
// an app-install callback: the vendor tenant id is the ceremony's durable
// output, so a repeat install of the same tenant must find its existing row
// rather than add a second.
//
// It deliberately tolerates duplicates. The (organization_id, provider,
// protocol) unique index was dropped in 20260819T142937Z and the retired API-key
// create path took a customer-typed website id, so an organization may ALREADY
// hold two rows for one tenant. The query therefore orders and takes one row
// instead of asserting there is exactly one: pgx.CollectExactlyOneRow -- the
// house pattern for a singular loader -- returns ErrTooManyRows on exactly that
// pre-existing data, which would turn a tolerated duplicate into a 500 on every
// re-install of that tenant.
//
// The (created_at, id) tiebreak is what makes the choice deterministic: every
// caller converges on the same, oldest row, so two concurrent completions cannot
// each adopt a different duplicate and diverge.
//
// Callers must already hold the advisory lock from LockConnectorInstallResource:
// FOR UPDATE on zero rows locks nothing, so this loader alone does not serialize
// two concurrent first installs of the same tenant.
//
// It does not decrypt the connection. Returns ErrResourceNotFound when there is
// no such connector.
func (c *Connector) LoadByOrganizationIDProviderAndSettingForUpdate(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	organizationID gid.GID,
	provider ConnectorProvider,
	settingKey string,
	settingValue string,
) error {
	q := `
SELECT
    id,
    organization_id,
    provider,
    protocol,
    settings,
    encrypted_connection,
    created_at,
    updated_at
FROM
    connectors
WHERE
    %s
    AND organization_id = @organization_id
    AND provider = @provider
    AND settings ->> @setting_key = @setting_value
ORDER BY
    created_at, id
LIMIT 1
FOR UPDATE
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"provider":        provider,
		"setting_key":     settingKey,
		"setting_value":   settingValue,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query connectors: %w", err)
	}

	loadedConnector, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Connector])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect connector row: %w", err)
	}

	*c = loadedConnector

	return nil
}

// LockConnectorInstallResource serializes concurrent install completions that
// target the same (organization, provider, vendor tenant id) for the rest of the
// transaction. It must be taken BEFORE the find-or-create load: two states are
// two distinct claim rows, and FOR UPDATE on a not-yet-existing row locks
// nothing, so without it two tabs both miss and both insert.
//
// resourceID must be the canonical spelling of the vendor tenant id: two
// spellings take two different locks and defeat the mechanism.
func LockConnectorInstallResource(
	ctx context.Context,
	conn pg.Tx,
	organizationID gid.GID,
	provider ConnectorProvider,
	resourceID string,
) error {
	// hashtext is int4, so unrelated keys can collide; a collision costs two
	// installs a serialization, never correctness. The namespace prefix keeps
	// this key space clear of BusinessFunction.Insert's, which hashes a bare
	// organization id.
	q := `
SELECT pg_advisory_xact_lock(
    hashtext(
        'connector-install:'
        || @organization_id::text
        || ':' || @provider::text
        || ':' || @resource_id::text
    )
)
`

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"provider":        provider,
		"resource_id":     resourceID,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot acquire connector install advisory lock: %w", err)
	}

	return nil
}

func (c *Connector) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM connectors
WHERE %s AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23503" || pgErr.Code == "23001" {
				return ErrResourceInUse
			}
		}

		return fmt.Errorf("cannot delete connector: %w", err)
	}

	return nil
}

func (c *Connector) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
) error {
	q := `
INSERT INTO connectors (
	id,
	tenant_id,
	organization_id,
	provider,
	protocol,
	settings,
	encrypted_connection,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@provider,
	@protocol,
	@settings,
	@encrypted_connection,
	@created_at,
	@updated_at
)
`

	if c.Connection == nil {
		return fmt.Errorf("connection is nil")
	}

	if c.Provider == ConnectorProviderSlack {
		if slackConn, ok := c.Connection.(*connector.SlackConnection); ok {
			_ = c.SetSettings(
				&SlackConnectorSettings{
					Channel:   slackConn.Settings.Channel,
					ChannelID: slackConn.Settings.ChannelID,
				},
			)
		}
	}

	connection, err := json.Marshal(c.Connection)
	if err != nil {
		return fmt.Errorf("cannot marshal connection: %w", err)
	}

	encryptedConnection, err := cipher.Encrypt(connection, encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt connection: %w", err)
	}

	var settingsArg any
	if len(c.RawSettings) > 0 {
		settingsArg = []byte(c.RawSettings)
	}

	args := pgx.StrictNamedArgs{
		"id":                   c.ID,
		"tenant_id":            scope.GetTenantID(),
		"organization_id":      c.OrganizationID,
		"provider":             c.Provider,
		"protocol":             c.Protocol,
		"settings":             settingsArg,
		"encrypted_connection": encryptedConnection,
		"created_at":           c.CreatedAt,
		"updated_at":           c.UpdatedAt,
	}

	_, err = conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert connector: %w", err)
	}

	c.EncryptedConnection = encryptedConnection

	return nil
}

func (c *Connectors) loadByOrganizationIDWithPagination(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ConnectorOrderField],
	filter *ConnectorFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
    provider,
    protocol,
    settings,
    encrypted_connection,
	created_at,
	updated_at
FROM
    connectors
WHERE
	%s
    AND organization_id = @organization_id
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query connectors: %w", err)
	}

	connectors, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Connector])
	if err != nil {
		return fmt.Errorf("cannot collect connectors: %w", err)
	}

	*c = connectors

	return nil
}

func (c *Connectors) loadAllByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    provider,
    protocol,
    settings,
    encrypted_connection,
	created_at,
	updated_at
FROM
    connectors
WHERE
	%s
    AND organization_id = @organization_id
ORDER BY
	created_at ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query connectors: %w", err)
	}

	connectors, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Connector])
	if err != nil {
		return fmt.Errorf("cannot collect connectors: %w", err)
	}

	*c = connectors

	return nil
}

func (c *Connector) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
) error {
	q := `
UPDATE connectors
SET
    settings = @settings,
    encrypted_connection = @encrypted_connection,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	if c.Connection == nil {
		return fmt.Errorf("connection is nil")
	}

	if c.Provider == ConnectorProviderSlack {
		if slackConn, ok := c.Connection.(*connector.SlackConnection); ok {
			_ = c.SetSettings(
				&SlackConnectorSettings{
					Channel:   slackConn.Settings.Channel,
					ChannelID: slackConn.Settings.ChannelID,
				},
			)
		}
	}

	connection, err := json.Marshal(c.Connection)
	if err != nil {
		return fmt.Errorf("cannot marshal connection: %w", err)
	}

	encryptedConnection, err := cipher.Encrypt(connection, encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt connection: %w", err)
	}

	var settingsArg any
	if len(c.RawSettings) > 0 {
		settingsArg = []byte(c.RawSettings)
	}

	args := pgx.StrictNamedArgs{
		"id":                   c.ID,
		"settings":             settingsArg,
		"encrypted_connection": encryptedConnection,
		"updated_at":           c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update connector: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	c.EncryptedConnection = encryptedConnection

	return nil
}

// decryptConnection decrypts and unmarshals the connector's encrypted
// connection blob, hydrating Slack channel settings from the settings
// column. A connector without a blob is left with a nil Connection.
func (c *Connector) decryptConnection(encryptionKey cipher.EncryptionKey) error {
	if len(c.EncryptedConnection) == 0 {
		return nil
	}

	decryptedConnection, err := cipher.Decrypt(c.EncryptedConnection, encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt connection for %s: %w", c.Provider, err)
	}

	c.Connection, err = connector.UnmarshalConnection(c.Protocol.String(), c.Provider.String(), decryptedConnection)
	if err != nil {
		return fmt.Errorf("cannot unmarshal connection for %s: %w", c.Provider, err)
	}

	if c.Provider == ConnectorProviderSlack {
		if slackConn, ok := c.Connection.(*connector.SlackConnection); ok {
			settings, _ := ConnectorSettings[SlackConnectorSettings](c)
			slackConn.Settings.Channel = settings.Channel
			slackConn.Settings.ChannelID = settings.ChannelID
		}
	}

	return nil
}
