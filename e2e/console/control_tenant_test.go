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

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestControl_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.CreateFramework(org1Owner)
	controlID := factory.CreateControl(org1Owner, frameworkID)

	t.Run("cannot read control from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Control {
						id
						name
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": controlID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "control")
	})

	t.Run("cannot update control from another organization", func(t *testing.T) {
		_, err := org2Owner.Do(`
			mutation($input: UpdateControlInput!) {
				updateControl(input: $input) {
					control { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":   controlID,
				"name": "Hijacked Control",
			},
		})
		require.Error(t, err, "Should not be able to update control from another org")
	})

	t.Run("cannot delete control from another organization", func(t *testing.T) {
		_, err := org2Owner.Do(`
			mutation($input: DeleteControlInput!) {
				deleteControl(input: $input) {
					deletedControlId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"controlId": controlID,
			},
		})
		require.Error(t, err, "Should not be able to delete control from another org")
	})
}
