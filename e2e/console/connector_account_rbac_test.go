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

// TestConnectorAccount_RBAC pins that discovery is not a read.
//
// It makes a live outbound call into the customer's cloud, so a role that may
// list Probo's own connector rows does not thereby get to talk to AWS. A
// viewer holds core:connector:list and nothing else here.
func TestConnectorAccount_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	orgID := owner.GetOrganizationID().String()

	connectorID := factory.NewConnector(owner, orgID).Create()
	accountIDs := factory.EnableConnectorAccounts(owner, orgID, connectorID, "111111111111")
	require.Len(t, accountIDs, 1)

	t.Run("viewer cannot discover accounts", func(t *testing.T) {
		t.Parallel()

		var result struct{}

		err := viewer.Execute(discoveredAccountsQuery, map[string]any{"id": connectorID}, &result)
		assert.Error(t, err)
	})

	t.Run("viewer cannot enable accounts", func(t *testing.T) {
		t.Parallel()

		const enableMutation = `
			mutation($input: EnableConnectorAccountsInput!) {
				enableConnectorAccounts(input: $input) {
					connectorAccountEdges { node { id } }
				}
			}
		`

		var result struct{}

		err := viewer.Execute(enableMutation, map[string]any{
			"input": map[string]any{
				"organizationId": orgID,
				"connectorId":    connectorID,
				"accounts":       []map[string]any{{"externalAccountId": "999999999999"}},
			},
		}, &result)
		assert.Error(t, err)
	})

	t.Run("viewer cannot disable an account", func(t *testing.T) {
		t.Parallel()

		var result struct{}

		err := viewer.Execute(disableConnectorAccountMutation, map[string]any{
			"input": map[string]any{"connectorAccountId": accountIDs[0]},
		}, &result)
		assert.Error(t, err)
	})
}
