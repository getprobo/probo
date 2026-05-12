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
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
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
func (c *Connector) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM connectors WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, c.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query connector authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (c *Connectors) LoadAllByOrganizationIDProtocolAndProvider(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	protocol ConnectorProtocol,
	provider ConnectorProvider,
	encryptionKey cipher.EncryptionKey,
) error {
	if err := c.loadAllByOrganizationIDProtocolAndProvider(ctx, conn, scope, organizationID, protocol, provider); err != nil {
		return fmt.Errorf("cannot load all connectors by organization ID, protocol and provider: %w", err)
	}

	if err := c.decryptConnections(encryptionKey); err != nil {
		return fmt.Errorf("cannot decrypt connections: %w", err)
	}

	return nil
}

// LoadOneByOrganizationIDAndProvider loads the effective OAuth2
// connector for an (organization, provider) pair, picking the row with
// the widest stored scope set. Ties are broken by most recent
// updated_at. Returns ErrResourceNotFound if no OAuth2 row exists.
func (c *Connector) LoadOneByOrganizationIDAndProvider(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
	organizationID gid.GID,
	provider ConnectorProvider,
) error {
	var connectors Connectors
	if err := connectors.LoadAllByOrganizationIDProtocolAndProvider(
		ctx,
		conn,
		scope,
		organizationID,
		ConnectorProtocolOAuth2,
		provider,
		encryptionKey,
	); err != nil {
		return fmt.Errorf("cannot load connectors: %w", err)
	}

	if len(connectors) == 0 {
		return ErrResourceNotFound
	}

	// Widest-scope-wins, tiebreak by most recent updated_at.
	sort.Slice(connectors, func(i, j int) bool {
		ci, cj := connectorScopeCount(connectors[i]), connectorScopeCount(connectors[j])
		if ci != cj {
			return ci > cj
		}
		return connectors[i].UpdatedAt.After(connectors[j].UpdatedAt)
	})

	*c = *connectors[0]
	return nil
}

// connectorScopeCount returns the number of scopes granted on a
// decrypted connector's connection. Returns 0 if the connection is nil.
// Used by the widest-scope selector.
func connectorScopeCount(c *Connector) int {
	if c == nil || c.Connection == nil {
		return 0
	}
	return len(c.Connection.Scopes())
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
				settings, _ := c.SlackSettings()
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

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete connector: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
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
			_ = c.SetSettings(&SlackConnectorSettings{
				Channel:   slackConn.Settings.Channel,
				ChannelID: slackConn.Settings.ChannelID,
			})
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

func (c *Connectors) loadAllByOrganizationIDProtocolAndProvider(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	protocol ConnectorProtocol,
	provider ConnectorProvider,
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
    AND protocol = @protocol
    AND provider = @provider
ORDER BY
	created_at ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"protocol":        protocol,
		"provider":        provider,
	}
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
			_ = c.SetSettings(&SlackConnectorSettings{
				Channel:   slackConn.Settings.Channel,
				ChannelID: slackConn.Settings.ChannelID,
			})
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

func (c *Connectors) decryptConnections(encryptionKey cipher.EncryptionKey) error {
	for _, cnnctr := range *c {
		if len(cnnctr.EncryptedConnection) == 0 {
			continue
		}

		decryptedConnection, err := cipher.Decrypt(cnnctr.EncryptedConnection, encryptionKey)
		if err != nil {
			return fmt.Errorf("cannot decrypt connection for %s: %w", cnnctr.Provider, err)
		}

		cnnctr.Connection, err = connector.UnmarshalConnection(cnnctr.Protocol.String(), cnnctr.Provider.String(), decryptedConnection)
		if err != nil {
			return fmt.Errorf("cannot unmarshal connection for %s: %w", cnnctr.Provider, err)
		}

		if cnnctr.Provider == ConnectorProviderSlack {
			if slackConn, ok := cnnctr.Connection.(*connector.SlackConnection); ok {
				settings, _ := cnnctr.SlackSettings()
				slackConn.Settings.Channel = settings.Channel
				slackConn.Settings.ChannelID = settings.ChannelID
			}
		}
	}

	return nil
}
