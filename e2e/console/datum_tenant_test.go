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

func TestDatum_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	profileID := factory.CreateUser(org1Owner)
	datumID := factory.NewDatum(org1Owner, profileID).WithName("Org1 Datum").Create()

	t.Run("cannot read datum from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Datum {
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

		err := org2Owner.Execute(query, map[string]any{"id": datumID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "datum")
	})

	t.Run("cannot update datum from another organization", func(t *testing.T) {
		query := `
			mutation UpdateDatum($input: UpdateDatumInput!) {
				updateDatum(input: $input) {
					datum { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   datumID,
				"name": "Hijacked Datum",
			},
		})
		require.Error(t, err, "Should not be able to update datum from another org")
	})

	t.Run("cannot delete datum from another organization", func(t *testing.T) {
		query := `
			mutation DeleteDatum($input: DeleteDatumInput!) {
				deleteDatum(input: $input) {
					deletedDatumId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"datumId": datumID,
			},
		})
		require.Error(t, err, "Should not be able to delete datum from another org")
	})

	t.Run("cannot list data from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						data(first: 100) {
							edges {
								node {
									id
									name
								}
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				Data struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"data"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{
			"id": org1Owner.GetOrganizationID().String(),
		}, &result)
		if err == nil {
			for _, edge := range result.Node.Data.Edges {
				assert.NotEqual(t, datumID, edge.Node.ID, "Should not see datum from another org")
			}
		}
	})

	t.Run("cannot create datum referencing a thirdParty from another organization", func(t *testing.T) {
		org2ThirdPartyID := factory.NewThirdParty(org2Owner).WithName("Org2 ThirdParty").Create()

		_, err := org1Owner.Do(`
			mutation($input: CreateDatumInput!) {
				createDatum(input: $input) {
					datumEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId":     org1Owner.GetOrganizationID().String(),
				"name":               factory.SafeName("Datum"),
				"dataClassification": "CONFIDENTIAL",
				"ownerId":            profileID,
				"thirdPartyIds":      []string{org2ThirdPartyID},
			},
		})
		require.Error(t, err, "must not accept a thirdPartyId belonging to another organization")
	})

	t.Run("cannot update datum to reference a thirdParty from another organization", func(t *testing.T) {
		org2ThirdPartyID := factory.NewThirdParty(org2Owner).WithName("Org2 ThirdParty for Update").Create()
		otherDatumID := factory.NewDatum(org1Owner, profileID).WithName("Org1 Datum for ThirdPartyIDs").Create()

		_, err := org1Owner.Do(`
			mutation($input: UpdateDatumInput!) {
				updateDatum(input: $input) {
					datum { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":            otherDatumID,
				"thirdPartyIds": []string{org2ThirdPartyID},
			},
		})
		require.Error(t, err, "must not accept a thirdPartyId belonging to another organization")
	})
}
