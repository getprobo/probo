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
	"go.probo.inc/probo/pkg/page"
)

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

func NewListServiceAccountsOutput(
	p *page.Page[*coredata.ServiceAccount, coredata.ServiceAccountOrderField],
) ListServiceAccountsOutput {
	accounts := make([]*ServiceAccount, 0, len(p.Data))
	for _, account := range p.Data {
		accounts = append(accounts, NewServiceAccount(account))
	}

	var nextCursor *page.CursorKey
	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListServiceAccountsOutput{
		NextCursor:      nextCursor,
		ServiceAccounts: accounts,
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

func NewListServiceAccountCredentialsOutput(
	p *page.Page[*coredata.ServiceAccountCredential, coredata.ServiceAccountCredentialOrderField],
) ListServiceAccountCredentialsOutput {
	credentials := make([]*ServiceAccountCredential, 0, len(p.Data))
	for _, credential := range p.Data {
		credentials = append(credentials, NewServiceAccountCredential(credential))
	}

	var nextCursor *page.CursorKey
	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListServiceAccountCredentialsOutput{
		NextCursor:                nextCursor,
		ServiceAccountCredentials: credentials,
	}
}
