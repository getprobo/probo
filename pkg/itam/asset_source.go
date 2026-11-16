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

package itam

import (
	"context"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
)

// NoteConnectorAccount ensures a due asset source when the provider has
// a discoverer. Account deletion needs no matching call: the foreign key
// clears connector_account_id.
func NoteConnectorAccount(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	provider coredata.ConnectorProvider,
	account *coredata.ConnectorAccount,
) error {
	if account == nil || account.ExternalAccountID == "" || !hasDiscoverer(provider) {
		return nil
	}

	source := &coredata.AssetSource{
		OrganizationID:     account.OrganizationID,
		Provider:           provider,
		ExternalAccountID:  account.ExternalAccountID,
		ConnectorAccountID: &account.ID,
	}

	if err := source.Ensure(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot note asset source: %w", err)
	}

	return nil
}
