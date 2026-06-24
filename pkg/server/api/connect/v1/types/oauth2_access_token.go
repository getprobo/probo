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
