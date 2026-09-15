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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	ServiceAccountOrderBy OrderBy[coredata.ServiceAccountOrderField]

	ServiceAccountConnection struct {
		TotalCount int
		Edges      []*ServiceAccountEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	ServiceAccountCredentialOrderBy OrderBy[coredata.ServiceAccountCredentialOrderField]

	ServiceAccountCredentialConnection struct {
		TotalCount int
		Edges      []*ServiceAccountCredentialEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewServiceAccountConnection(
	p *page.Page[*coredata.ServiceAccount, coredata.ServiceAccountOrderField],
	resolver any,
	parentID gid.GID,
) *ServiceAccountConnection {
	edges := make([]*ServiceAccountEdge, len(p.Data))
	for i, account := range p.Data {
		edges[i] = NewServiceAccountEdge(account, p.Cursor.OrderBy.Field)
	}

	return &ServiceAccountConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
	}
}

func NewServiceAccountEdge(
	account *coredata.ServiceAccount,
	orderField coredata.ServiceAccountOrderField,
) *ServiceAccountEdge {
	return &ServiceAccountEdge{
		Cursor: account.CursorKey(orderField),
		Node:   NewServiceAccount(account),
	}
}

func NewServiceAccount(account *coredata.ServiceAccount) *ServiceAccount {
	return &ServiceAccount{
		ID:             account.ID,
		OrganizationID: account.OrganizationID,
		Name:           account.Name,
		Description:    account.Description,
		Scopes:         account.Scopes,
		DisabledAt:     account.DisabledAt,
		CreatedAt:      account.CreatedAt,
		UpdatedAt:      account.UpdatedAt,
	}
}

func NewServiceAccountCredentialConnection(
	p *page.Page[*coredata.ServiceAccountCredential, coredata.ServiceAccountCredentialOrderField],
	resolver any,
	parentID gid.GID,
) *ServiceAccountCredentialConnection {
	edges := make([]*ServiceAccountCredentialEdge, len(p.Data))
	for i, credential := range p.Data {
		edges[i] = NewServiceAccountCredentialEdge(credential, p.Cursor.OrderBy.Field)
	}

	return &ServiceAccountCredentialConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
	}
}

func NewServiceAccountCredentialEdge(
	credential *coredata.ServiceAccountCredential,
	orderField coredata.ServiceAccountCredentialOrderField,
) *ServiceAccountCredentialEdge {
	return &ServiceAccountCredentialEdge{
		Cursor: credential.CursorKey(orderField),
		Node:   NewServiceAccountCredential(credential),
	}
}

func NewServiceAccountCredential(
	credential *coredata.ServiceAccountCredential,
) *ServiceAccountCredential {
	return &ServiceAccountCredential{
		ID:               credential.ID,
		ServiceAccountID: credential.ServiceAccountID,
		Name:             credential.Name,
		Scopes:           credential.Scopes,
		ExpiresAt:        credential.ExpiresAt,
		LastUsedAt:       credential.LastUsedAt,
		RevokedAt:        credential.RevokedAt,
		CreatedAt:        credential.CreatedAt,
	}
}
