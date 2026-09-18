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

package coredata_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

func newAccountSource(
	scope coredata.Scoper,
	organizationID gid.GID,
	connectorID *gid.GID,
	connectorAccountID *gid.GID,
	name string,
) *coredata.AccessReviewSource {
	now := time.Now().UTC()

	return &coredata.AccessReviewSource{
		ID:                 gid.New(scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
		OrganizationID:     organizationID,
		ConnectorID:        connectorID,
		ConnectorAccountID: connectorAccountID,
		Name:               name,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func insertAccountSource(
	ctx context.Context,
	client *pg.Client,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
) (bool, error) {
	var inserted bool

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		inserted, err = source.Insert(ctx, tx, scope)

		return err
	})

	return inserted, err
}

// TestAccessReviewSource_ConnectorAccount covers the two things the index swap
// changed at the SQL level. Source lifecycle through the API is covered in
// e2e/console/connector_account_test.go.
func TestAccessReviewSource_ConnectorAccount(t *testing.T) {
	t.Parallel()

	// The 42P10 regression. The conflict target must name exactly the columns
	// of the new composite index or every insert fails at plan time — CSV
	// sources included, which have no connector at all. It is immediate and
	// total, and no other test reaches a CSV insert at this layer.
	t.Run("csv sources still insert after the index swap", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		ctx := context.Background()
		scope, organizationID := seedConnectorOrg(t, ctx, client)

		for range 2 {
			source := newAccountSource(scope, organizationID, nil, nil, "CSV")

			inserted, err := insertAccountSource(ctx, client, scope, source)
			require.NoError(t, err)
			assert.True(t, inserted)
		}
	})

	// The capability the old unconditional index refused.
	t.Run("two accounts of one connector back two sources", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		ctx := context.Background()
		scope, organizationID := seedConnectorOrg(t, ctx, client)

		var key cipher.EncryptionKey

		connectorID, err := insertConnector(ctx, client, scope, organizationID, coredata.ConnectorProviderAWS, key)
		require.NoError(t, err)

		for _, externalID := range []string{"111111111111", "222222222222"} {
			account := newConnectorAccount(scope, organizationID, connectorID, &externalID, externalID)
			require.NoError(t, upsertConnectorAccount(ctx, client, scope, account))

			source := newAccountSource(scope, organizationID, &connectorID, &account.ID, externalID)

			inserted, err := insertAccountSource(ctx, client, scope, source)
			require.NoError(t, err)
			assert.True(t, inserted, "account %s should get its own source", externalID)
		}
	})
}
