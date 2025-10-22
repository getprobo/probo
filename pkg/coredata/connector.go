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
	"encoding/json"
	"fmt"
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/connector"
	"github.com/getprobo/probo/pkg/crypto/cipher"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	Connector struct {
		ID                  gid.GID              `db:"id"`
		OrganizationID      gid.GID              `db:"organization_id"`
		Provider            ConnectorProvider    `db:"provider"`
		Protocol            ConnectorProtocol    `db:"protocol"`
		Settings            map[string]any       `db:"settings"`
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

func (c *Connectors) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ConnectorOrderField],
	encryptionKey cipher.EncryptionKey,
	filter *ConnectorProviderFilter,
) error {
	if err := c.loadByOrganizationIDWithPagination(ctx, conn, scope, organizationID, cursor, filter); err != nil {
		return fmt.Errorf("cannot load connectors by organization ID: %w", err)
	}

	if err := c.decryptConnections(encryptionKey); err != nil {
		return fmt.Errorf("cannot decrypt connections: %w", err)
	}

	return nil
}

func (c *Connectors) LoadAllByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	encryptionKey cipher.EncryptionKey,
) error {
	if err := c.loadAllByOrganizationID(ctx, conn, scope, organizationID); err != nil {
		return fmt.Errorf("cannot load all connectors by organization ID: %w", err)
	}

	if err := c.decryptConnections(encryptionKey); err != nil {
		return fmt.Errorf("cannot decrypt connections: %w", err)
	}

	return nil
}

func (c *Connectors) LoadAllByOrganizationIDProtocolAndProvider(
	ctx context.Context,
	conn pg.Conn,
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

func (c *Connectors) LoadByOrganizationIDWithoutDecryptedConnection(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ConnectorOrderField],
	filter *ConnectorProviderFilter,
) error {
	return c.loadByOrganizationIDWithPagination(ctx, conn, scope, organizationID, cursor, filter)
}

func (c *Connectors) LoadAllByOrganizationIDWithoutDecryptedConnection(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) error {
	return c.loadAllByOrganizationID(ctx, conn, scope, organizationID)
}

func (c *Connector) Insert(
	ctx context.Context,
	conn pg.Conn,
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

	c.extractSlackSettings()

	connection, err := json.Marshal(c.Connection)
	if err != nil {
		return fmt.Errorf("cannot marshal connection: %w", err)
	}

	encryptedConnection, err := cipher.Encrypt(connection, encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt connection: %w", err)
	}

	args := pgx.StrictNamedArgs{
		"id":                   c.ID,
		"tenant_id":            scope.GetTenantID(),
		"organization_id":      c.OrganizationID,
		"provider":             c.Provider,
		"protocol":             c.Protocol,
		"settings":             c.Settings,
		"encrypted_connection": encryptedConnection,
		"created_at":           c.CreatedAt,
		"updated_at":           c.UpdatedAt,
	}

	_, err = conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert connector: %w", err)
	}

	c.EncryptedConnection = encryptedConnection
	c.populateSlackSettings()

	return nil
}

func (c *Connector) populateSlackSettings() {
	if c.Provider != ConnectorProviderSlack {
		return
	}

	slackConn, ok := c.Connection.(*connector.SlackConnection)
	if !ok {
		return
	}

	if channel, ok := c.Settings["channel"].(string); ok {
		slackConn.Settings.Channel = channel
	}
	if channelID, ok := c.Settings["channel_id"].(string); ok {
		slackConn.Settings.ChannelID = channelID
	}
}

func (c *Connector) extractSlackSettings() {
	if c.Provider != ConnectorProviderSlack {
		return
	}

	slackConn, ok := c.Connection.(*connector.SlackConnection)
	if !ok {
		return
	}

	c.Settings = make(map[string]any)
	if slackConn.Settings.Channel != "" {
		c.Settings["channel"] = slackConn.Settings.Channel
	}
	if slackConn.Settings.ChannelID != "" {
		c.Settings["channel_id"] = slackConn.Settings.ChannelID
	}
}

func (c *Connectors) loadByOrganizationIDWithPagination(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ConnectorOrderField],
	filter *ConnectorProviderFilter,
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
	conn pg.Conn,
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
	conn pg.Conn,
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

		cnnctr.populateSlackSettings()
	}

	return nil
}
