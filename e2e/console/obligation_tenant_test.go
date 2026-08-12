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

func TestObligation_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	profileID := factory.CreateUser(org1Owner)

	var createResult struct {
		CreateObligation struct {
			ObligationEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"obligationEdge"`
		} `json:"createObligation"`
	}

	err := org1Owner.Execute(`
		mutation($input: CreateObligationInput!) {
			createObligation(input: $input) {
				obligationEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": org1Owner.GetOrganizationID().String(),
			"area":           "Risk Management",
			"requirement":    "Org1 Obligation",
			"ownerId":        profileID,
			"status":         "NON_COMPLIANT",
			"type":           "LEGAL",
		},
	}, &createResult)
	require.NoError(t, err)

	obligationID := createResult.CreateObligation.ObligationEdge.Node.ID

	t.Run("cannot read obligation from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Obligation {
						id
						requirement
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID          string `json:"id"`
				Requirement string `json:"requirement"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": obligationID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "obligation")
	})

	t.Run("cannot update obligation from another organization", func(t *testing.T) {
		_, err := org2Owner.Do(`
			mutation($input: UpdateObligationInput!) {
				updateObligation(input: $input) {
					obligation { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":   obligationID,
				"area": "Hijacked Obligation",
			},
		})
		require.Error(t, err, "Should not be able to update obligation from another org")
	})

	t.Run("cannot delete obligation from another organization", func(t *testing.T) {
		_, err := org2Owner.Do(`
			mutation($input: DeleteObligationInput!) {
				deleteObligation(input: $input) {
					deletedObligationId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"obligationId": obligationID,
			},
		})
		require.Error(t, err, "Should not be able to delete obligation from another org")
	})

	t.Run("cannot create obligation referencing an owner from another organization", func(t *testing.T) {
		org2ProfileID := factory.CreateUser(org2Owner)

		_, err := org1Owner.Do(`
			mutation($input: CreateObligationInput!) {
				createObligation(input: $input) {
					obligationEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": org1Owner.GetOrganizationID().String(),
				"area":           "Risk Management",
				"requirement":    factory.SafeName("Obligation"),
				"ownerId":        org2ProfileID,
				"status":         "NON_COMPLIANT",
				"type":           "LEGAL",
			},
		})
		require.Error(t, err, "must not accept an ownerId belonging to another organization")
	})

	t.Run("cannot update obligation to reference an owner from another organization", func(t *testing.T) {
		org2ProfileID := factory.CreateUser(org2Owner)

		_, err := org1Owner.Do(`
			mutation($input: UpdateObligationInput!) {
				updateObligation(input: $input) {
					obligation { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":      obligationID,
				"ownerId": org2ProfileID,
			},
		})
		require.Error(t, err, "must not accept an ownerId belonging to another organization")
	})
}
