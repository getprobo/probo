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

func TestCreateAPIKeyConnector_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	t.Run("viewer cannot create connector", func(t *testing.T) {
		t.Parallel()

		_, err := viewer.Do(`
			mutation($input: CreateAPIKeyConnectorInput!) {
				createAPIKeyConnector(input: $input) {
					connector { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": viewer.GetOrganizationID().String(),
				"provider":       "BREX",
				"apiKey":         "bxt_test-key",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create connector")
	})
}

// TestConnectorPermission pins the field the integrations list needs to decide
// whether to offer disconnect. A viewer may list connectors but not delete
// one, and without asking per connector the console's only way to find that
// out is to attempt the mutation and show the user an error it could have
// predicted.
//
// The question is asked through Organization.connectors rather than node(id:)
// on purpose: node resolution is gated on core:connector:get, which a viewer
// does not hold, so the list is the only place a viewer sees a connector at
// all — and it is where the delete control lives.
func TestConnectorPermission(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	factory.NewConnector(owner, owner.GetOrganizationID().String()).Create()

	const query = `
		query($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					connectors {
						id
						canDelete: permission(action: "core:connector:delete")
					}
				}
			}
		}
	`

	canDelete := func(c *testutil.Client) bool {
		c.T.Helper()

		var result struct {
			Node struct {
				Connectors []struct {
					ID        string `json:"id"`
					CanDelete bool   `json:"canDelete"`
				} `json:"connectors"`
			} `json:"node"`
		}

		require.NoError(t, c.Execute(query, map[string]any{
			"organizationId": c.GetOrganizationID().String(),
		}, &result))
		require.Len(t, result.Node.Connectors, 1)

		return result.Node.Connectors[0].CanDelete
	}

	assert.True(t, canDelete(owner), "an owner disconnects vendors")
	assert.False(t, canDelete(viewer), "a viewer sees the integration but is not offered disconnect")
}
