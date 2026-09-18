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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// TestConnectorAccount_TenantIsolation pins that an account GID from another
// organization is not reachable, and that enabling cannot be pointed at a
// connector the caller's organization does not own.
func TestConnectorAccount_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1 := testutil.NewClient(t, testutil.RoleOwner)
	org2 := testutil.NewClient(t, testutil.RoleOwner)
	org1ID := org1.GetOrganizationID().String()
	org2ID := org2.GetOrganizationID().String()

	connectorID := factory.NewConnector(org1, org1ID).Create()
	accountIDs := factory.EnableConnectorAccounts(org1, org1ID, connectorID, "111111111111")
	require.Len(t, accountIDs, 1)

	t.Run("cannot read an account from another organization", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err := org2.Execute(
			`query($id: ID!) { node(id: $id) { ... on ConnectorAccount { id } } }`,
			map[string]any{"id": accountIDs[0]},
			&result,
		)
		if err == nil {
			assert.Nil(t, result.Node)
		}
	})

	t.Run("cannot enable accounts on another organization's connector", func(t *testing.T) {
		t.Parallel()

		const enableMutation = `
			mutation($input: EnableConnectorAccountsInput!) {
				enableConnectorAccounts(input: $input) {
					connectorAccountEdges { node { id } }
				}
			}
		`

		var result struct{}

		err := org2.Execute(enableMutation, map[string]any{
			"input": map[string]any{
				"organizationId": org2ID,
				"connectorId":    connectorID,
				"accounts":       []map[string]any{{"externalAccountId": "999999999999"}},
			},
		}, &result)
		assert.Error(t, err)
	})
}
