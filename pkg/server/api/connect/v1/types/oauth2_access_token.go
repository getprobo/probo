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

package types

import (
	"errors"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	CreateOAuth2AccessTokenInput struct {
		Name      string    `json:"name"`
		ExpiresAt time.Time `json:"expiresAt"`
		Scopes    []string  `json:"scopes"`
	}

	OAuth2AccessTokenConnection struct {
		TotalCount int
		Edges      []*OAuth2AccessTokenEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func (in *CreateOAuth2AccessTokenInput) ParsedScopes() (coredata.OAuth2Scopes, error) {
	if len(in.Scopes) == 0 {
		return nil, errors.New("scopes are required")
	}

	scopes := make(coredata.OAuth2Scopes, len(in.Scopes))
	for i, scopeString := range in.Scopes {
		scopes[i] = coredata.OAuth2Scope(scopeString)
	}

	return scopes, nil
}

func NewOAuth2AccessTokenConnection(
	p *page.Page[*coredata.OAuth2AccessToken, coredata.OAuth2AccessTokenOrderField],
	resolver any,
	parentID gid.GID,
) *OAuth2AccessTokenConnection {
	edges := make([]*OAuth2AccessTokenEdge, len(p.Data))
	for i, token := range p.Data {
		edges[i] = NewOAuth2AccessTokenEdge(token, p.Cursor.OrderBy.Field)
	}

	return &OAuth2AccessTokenConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
	}
}

func NewOAuth2AccessTokenEdge(
	token *coredata.OAuth2AccessToken,
	orderField coredata.OAuth2AccessTokenOrderField,
) *OAuth2AccessTokenEdge {
	return &OAuth2AccessTokenEdge{
		Node:   NewOAuth2AccessToken(token),
		Cursor: token.CursorKey(orderField),
	}
}

func NewOAuth2AccessToken(token *coredata.OAuth2AccessToken) *OAuth2AccessToken {
	scopes := make([]string, len(token.Scopes))
	for i, scope := range token.Scopes {
		scopes[i] = string(scope)
	}

	return &OAuth2AccessToken{
		ID:        token.ID,
		Name:      token.Name,
		Scopes:    scopes,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
	}
}
