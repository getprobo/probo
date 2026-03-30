// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	OAuth2Client struct {
		ID                      gid.GID                             `db:"id"`
		OrganizationID          gid.GID                             `db:"organization_id"`
		ClientSecretHash        []byte                              `db:"client_secret_hash"`
		ClientName              string                              `db:"client_name"`
		Visibility              OAuth2ClientVisibility              `db:"visibility"`
		RedirectURIs            []string                            `db:"redirect_uris"`
		Scopes                  OAuth2Scopes                        `db:"scopes"`
		GrantTypes              OAuth2GrantTypes                    `db:"grant_types"`
		ResponseTypes           OAuth2ResponseTypes                 `db:"response_types"`
		TokenEndpointAuthMethod OAuth2ClientTokenEndpointAuthMethod `db:"token_endpoint_auth_method"`
		LogoURI                 *string                             `db:"logo_uri"`
		ClientURI               *string                             `db:"client_uri"`
		Contacts                []string                            `db:"contacts"`
		CreatedAt               time.Time                           `db:"created_at"`
		UpdatedAt               time.Time                           `db:"updated_at"`
	}

	OAuth2Clients []*OAuth2Client
)

func (c *OAuth2Client) CursorKey(orderBy OAuth2ClientOrderField) page.CursorKey {
	switch orderBy {
	case OAuth2ClientOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (c *OAuth2Client) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `
SELECT
	organization_id
FROM
	iam_oauth2_clients
WHERE
	id = $1
LIMIT 1;
`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, c.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query oauth2 client authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (c *OAuth2Client) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	clientID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	client_secret_hash,
	client_name,
	visibility,
	redirect_uris,
	scopes,
	grant_types,
	response_types,
	token_endpoint_auth_method,
	logo_uri,
	client_uri,
	contacts,
	created_at,
	updated_at
FROM
	iam_oauth2_clients
WHERE
	%s
	AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": clientID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_oauth2_clients: %w", err)
	}

	client, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OAuth2Client])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect oauth2_client: %w", err)
	}

	*c = client
	return nil
}

func (c *OAuth2Client) LoadByIDNoScope(
	ctx context.Context,
	conn pg.Conn,
	clientID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	client_secret_hash,
	client_name,
	visibility,
	redirect_uris,
	scopes,
	grant_types,
	response_types,
	token_endpoint_auth_method,
	logo_uri,
	client_uri,
	contacts,
	created_at,
	updated_at
FROM
	iam_oauth2_clients
WHERE
	id = @id
LIMIT 1;
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"id": clientID})
	if err != nil {
		return fmt.Errorf("cannot query iam_oauth2_clients: %w", err)
	}

	client, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OAuth2Client])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect oauth2_client: %w", err)
	}

	*c = client
	return nil
}

func (c *OAuth2Clients) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[OAuth2ClientOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	client_secret_hash,
	client_name,
	visibility,
	redirect_uris,
	scopes,
	grant_types,
	response_types,
	token_endpoint_auth_method,
	logo_uri,
	client_uri,
	contacts,
	created_at,
	updated_at
FROM
	iam_oauth2_clients
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(
		q,
		scope.SQLFragment(),
		cursor.SQLFragment(),
	)

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_oauth2_clients: %w", err)
	}

	clients, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[OAuth2Client])
	if err != nil {
		return fmt.Errorf("cannot collect oauth2_clients: %w", err)
	}

	*c = clients
	return nil
}

func (c *OAuth2Clients) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	iam_oauth2_clients
WHERE
	%s
	AND organization_id = @organization_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count oauth2_clients: %w", err)
	}

	return count, nil
}

func (c *OAuth2Client) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO iam_oauth2_clients (
	id,
	tenant_id,
	organization_id,
	client_secret_hash,
	client_name,
	visibility,
	redirect_uris,
	scopes,
	grant_types,
	response_types,
	token_endpoint_auth_method,
	logo_uri,
	client_uri,
	contacts,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@client_secret_hash,
	@client_name,
	@visibility,
	@redirect_uris,
	@scopes,
	@grant_types,
	@response_types,
	@token_endpoint_auth_method,
	@logo_uri,
	@client_uri,
	@contacts,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                         c.ID,
		"tenant_id":                  scope.GetTenantID(),
		"organization_id":            c.OrganizationID,
		"client_secret_hash":         c.ClientSecretHash,
		"client_name":                c.ClientName,
		"visibility":                 c.Visibility,
		"redirect_uris":              c.RedirectURIs,
		"scopes":                     c.Scopes,
		"grant_types":                c.GrantTypes,
		"response_types":             c.ResponseTypes,
		"token_endpoint_auth_method": c.TokenEndpointAuthMethod,
		"logo_uri":                   c.LogoURI,
		"client_uri":                 c.ClientURI,
		"contacts":                   c.Contacts,
		"created_at":                 c.CreatedAt,
		"updated_at":                 c.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert oauth2_client: %w", err)
	}

	return nil
}

func (c *OAuth2Client) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE iam_oauth2_clients
SET
	client_name = @client_name,
	visibility = @visibility,
	redirect_uris = @redirect_uris,
	scopes = @scopes,
	grant_types = @grant_types,
	response_types = @response_types,
	token_endpoint_auth_method = @token_endpoint_auth_method,
	logo_uri = @logo_uri,
	client_uri = @client_uri,
	contacts = @contacts,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                         c.ID,
		"client_name":                c.ClientName,
		"visibility":                 c.Visibility,
		"redirect_uris":              c.RedirectURIs,
		"scopes":                     c.Scopes,
		"grant_types":                c.GrantTypes,
		"response_types":             c.ResponseTypes,
		"token_endpoint_auth_method": c.TokenEndpointAuthMethod,
		"logo_uri":                   c.LogoURI,
		"client_uri":                 c.ClientURI,
		"contacts":                   c.Contacts,
		"updated_at":                 c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update oauth2_client: %w", err)
	}

	return nil
}

func (c *OAuth2Client) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM iam_oauth2_clients
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete oauth2_client: %w", err)
	}

	return nil
}
