// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"net/url"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/netx"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/uri"
)

type (
	OAuth2Client struct {
		ID                      gid.GID                             `db:"id"`
		ExternalClientID        string                              `db:"external_client_id"`
		OrganizationID          *gid.GID                            `db:"organization_id"`
		ClientSecretHash        []byte                              `db:"client_secret_hash"`
		ClientName              string                              `db:"client_name"`
		Visibility              OAuth2ClientVisibility              `db:"visibility"`
		RedirectURIs            []uri.URI                           `db:"redirect_uris"`
		Scopes                  OAuth2Scopes                        `db:"scopes"`
		GrantTypes              OAuth2GrantTypes                    `db:"grant_types"`
		ResponseTypes           OAuth2ResponseTypes                 `db:"response_types"`
		TokenEndpointAuthMethod OAuth2ClientTokenEndpointAuthMethod `db:"token_endpoint_auth_method"`
		LogoURI                 *uri.URI                            `db:"logo_uri"`
		ClientURI               *uri.URI                            `db:"client_uri"`
		Contacts                []string                            `db:"contacts"`
		CreatedAt               time.Time                           `db:"created_at"`
		UpdatedAt               time.Time                           `db:"updated_at"`
	}

	OAuth2Clients []*OAuth2Client
)

func (c *OAuth2Client) IsRedirectURIAllowed(rawURI string) bool {
	if slices.Contains(c.RedirectURIs, uri.URI(rawURI)) {
		return true
	}

	// RFC 8252 §7.3: native apps using a loopback redirect URI choose an
	// ephemeral port at request time, so the port must be ignored when
	// matching against the registered loopback redirect URIs.
	requested, err := url.Parse(rawURI)
	if err != nil || !netx.IsLoopback(requested.Hostname()) {
		return false
	}

	for _, registered := range c.RedirectURIs {
		candidate, err := url.Parse(registered.String())
		if err != nil {
			continue
		}

		if candidate.Scheme != requested.Scheme {
			continue
		}

		if !netx.IsLoopback(candidate.Hostname()) {
			continue
		}

		if candidate.Hostname() != requested.Hostname() {
			continue
		}

		if candidate.Path == requested.Path && candidate.RawQuery == requested.RawQuery {
			return true
		}
	}

	return false
}

func (c *OAuth2Client) HasGrantType(grantType OAuth2GrantType) bool {
	return slices.Contains(c.GrantTypes, grantType)
}

func (c *OAuth2Client) AreScopesAllowed(scopes OAuth2Scopes) bool {
	return c.Scopes.ContainsAll(scopes.Values())
}

func (c *OAuth2Client) CursorKey(orderBy OAuth2ClientOrderField) page.CursorKey {
	switch orderBy {
	case OAuth2ClientOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (c *OAuth2Client) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
    id,
    organization_id
FROM
    iam_oauth2_clients
WHERE
    id = ANY(@resource_ids::text[])
`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query oauth2 client authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID, len(resourceIDs))

	for rows.Next() {
		var (
			id             gid.GID
			organizationID *gid.GID
		)

		err = rows.Scan(&id, &organizationID)
		if err != nil {
			return nil, fmt.Errorf("cannot scan oauth2 client authorization attributes: %w", err)
		}

		attrs := make(map[string]string)
		if organizationID != nil {
			attrs["organization_id"] = organizationID.String()
		}

		attrsByID[id] = attrs
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate oauth2 client authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (c *OAuth2Client) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	clientID gid.GID,
) error {
	q := `
SELECT
	id,
	COALESCE(external_client_id, '') AS external_client_id,
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

func (c *OAuth2Client) LoadByExternalClientID(
	ctx context.Context,
	conn pg.Querier,
	externalClientID string,
) error {
	q := `
SELECT
	id,
	COALESCE(external_client_id, '') AS external_client_id,
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
	external_client_id = @external_client_id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"external_client_id": externalClientID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_oauth2_clients by external_client_id: %w", err)
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
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[OAuth2ClientOrderField],
) error {
	q := `
SELECT
	id,
	COALESCE(external_client_id, '') AS external_client_id,
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
	conn pg.Querier,
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
	conn pg.Tx,
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

func (c *OAuth2Client) UpsertCIMD(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
INSERT INTO iam_oauth2_clients (
	id,
	tenant_id,
	organization_id,
	external_client_id,
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
	NULL,
	NULL,
	@external_client_id,
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
ON CONFLICT (external_client_id) WHERE external_client_id IS NOT NULL DO UPDATE SET
	client_name = EXCLUDED.client_name,
	redirect_uris = EXCLUDED.redirect_uris,
	grant_types = EXCLUDED.grant_types,
	response_types = EXCLUDED.response_types,
	token_endpoint_auth_method = EXCLUDED.token_endpoint_auth_method,
	logo_uri = EXCLUDED.logo_uri,
	client_uri = EXCLUDED.client_uri,
	updated_at = EXCLUDED.updated_at
RETURNING
	id,
	external_client_id,
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
`

	args := pgx.StrictNamedArgs{
		"id":                         c.ID,
		"external_client_id":         c.ExternalClientID,
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

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert cimd oauth2_client: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OAuth2Client])
	if err != nil {
		return fmt.Errorf("cannot collect upsert result: %w", err)
	}

	*c = row

	return nil
}

func (c *OAuth2Client) Update(
	ctx context.Context,
	conn pg.Tx,
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
	conn pg.Tx,
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

func NewCIMDClient(
	externalClientID string,
	name string,
	redirectURIs []string,
	scopes OAuth2Scopes,
	logoURI, clientURI *string,
	now time.Time,
) (*OAuth2Client, error) {
	uris := make([]uri.URI, 0, len(redirectURIs))
	for _, raw := range redirectURIs {
		uris = append(uris, uri.URI(raw))
	}

	client := &OAuth2Client{
		ID:               gid.New(gid.NewTenantID(), OAuth2ClientEntityType),
		ExternalClientID: externalClientID,
		ClientName:       name,
		Visibility:       OAuth2ClientVisibilityPublic,
		RedirectURIs:     uris,
		Scopes:           scopes,
		GrantTypes: OAuth2GrantTypes{
			OAuth2GrantTypeAuthorizationCode,
			OAuth2GrantTypeRefreshToken,
		},
		ResponseTypes: OAuth2ResponseTypes{
			OAuth2ResponseTypeCode,
		},
		TokenEndpointAuthMethod: OAuth2ClientTokenEndpointAuthMethodNone,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	if logoURI != nil && *logoURI != "" {
		u := uri.URI(*logoURI)
		client.LogoURI = &u
	}

	if clientURI != nil && *clientURI != "" {
		u := uri.URI(*clientURI)
		client.ClientURI = &u
	}

	return client, nil
}
