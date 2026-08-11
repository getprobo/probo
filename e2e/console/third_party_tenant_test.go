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

func TestThirdParty_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	thirdPartyID := factory.NewThirdParty(org1Owner).WithName("Org1 ThirdParty").Create()

	t.Run("cannot read thirdParty from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
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

		err := org2Owner.Execute(query, map[string]any{"id": thirdPartyID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "thirdParty")
	})

	t.Run("cannot update thirdParty from another organization", func(t *testing.T) {
		query := `
			mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   thirdPartyID,
				"name": "Hijacked ThirdParty",
			},
		})
		require.Error(t, err, "Should not be able to update thirdParty from another org")
	})

	t.Run("cannot delete thirdParty from another organization", func(t *testing.T) {
		query := `
			mutation DeleteThirdParty($input: DeleteThirdPartyInput!) {
				deleteThirdParty(input: $input) {
					deletedThirdPartyId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"thirdPartyId": thirdPartyID,
			},
		})
		require.Error(t, err, "Should not be able to delete thirdParty from another org")
	})

	t.Run("cannot create thirdParty referencing an administrator from another organization", func(t *testing.T) {
		org2ProfileID := factory.CreateUser(org2Owner)

		_, err := org1Owner.Do(`
			mutation($input: CreateThirdPartyInput!) {
				createThirdParty(input: $input) {
					thirdPartyEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId":   org1Owner.GetOrganizationID().String(),
				"name":             factory.SafeName("ThirdParty"),
				"administratorIds": []string{org2ProfileID},
			},
		})
		require.Error(t, err, "must not accept an administratorId belonging to another organization")
	})

	t.Run("cannot update thirdParty to reference an administrator from another organization", func(t *testing.T) {
		org2ProfileID := factory.CreateUser(org2Owner)
		otherThirdPartyID := factory.NewThirdParty(org1Owner).WithName("Org1 ThirdParty for Administrators").Create()

		_, err := org1Owner.Do(`
			mutation($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":               otherThirdPartyID,
				"administratorIds": []string{org2ProfileID},
			},
		})
		require.Error(t, err, "must not accept an administratorId belonging to another organization")
	})

	t.Run("cannot create thirdParty referencing a parent thirdParty from another organization", func(t *testing.T) {
		org2ParentID := factory.NewThirdParty(org2Owner).WithName("Org2 Parent ThirdParty").Create()

		_, err := org1Owner.Do(`
			mutation($input: CreateThirdPartyInput!) {
				createThirdParty(input: $input) {
					thirdPartyEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId":     org1Owner.GetOrganizationID().String(),
				"name":               factory.SafeName("ThirdParty"),
				"parentThirdPartyId": org2ParentID,
			},
		})
		require.Error(t, err, "must not accept a parentThirdPartyId belonging to another organization")
	})
}
