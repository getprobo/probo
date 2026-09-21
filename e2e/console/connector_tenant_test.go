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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// TestConnector_TenantIsolation pins the gate the settings integrations pages
// sit behind. Both of them resolve a connector through node(id:), and the
// details page is reached by a GID in the URL, so this is what stands between
// a guessed identifier and another customer's vendor credential — or its
// deletion, now that settings is where disconnecting happens.
func TestConnector_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1 := testutil.NewClient(t, testutil.RoleOwner)
	org2 := testutil.NewClient(t, testutil.RoleOwner)

	connectorID := factory.NewConnector(org1, org1.GetOrganizationID().String()).Create()

	t.Run("cannot read a connector from another organization", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err := org2.Execute(
			`query($id: ID!) { node(id: $id) { ... on Connector { id } } }`,
			map[string]any{"id": connectorID},
			&result,
		)
		if err == nil {
			assert.Nil(t, result.Node)
		}
	})

	t.Run("cannot disconnect another organization's connector", func(t *testing.T) {
		t.Parallel()

		var result struct{}

		err := org2.Execute(
			`mutation($input: DeleteConnectorInput!) {
				deleteConnector(input: $input) { deletedConnectorId }
			}`,
			map[string]any{"input": map[string]any{"connectorId": connectorID}},
			&result,
		)
		assert.Error(t, err)
	})
}
