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

func TestMeasure_TenantIsolation(t *testing.T) {
	t.Parallel()

	// Create two separate organizations with their own owners
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a measure in org1
	measureID := factory.NewMeasure(org1Owner).WithName("Org1 Measure").Create()

	t.Run("cannot read measure from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Measure {
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

		err := org2Owner.Execute(query, map[string]any{"id": measureID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "measure")
	})

	t.Run("cannot update measure from another organization", func(t *testing.T) {
		query := `
			mutation UpdateMeasure($input: UpdateMeasureInput!) {
				updateMeasure(input: $input) {
					measure { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   measureID,
				"name": "Hijacked Measure",
			},
		})
		require.Error(t, err, "Should not be able to update measure from another org")
	})

	t.Run("cannot delete measure from another organization", func(t *testing.T) {
		query := `
			mutation DeleteMeasure($input: DeleteMeasureInput!) {
				deleteMeasure(input: $input) {
					deletedMeasureId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"measureId": measureID,
			},
		})
		require.Error(t, err, "Should not be able to delete measure from another org")
	})

	t.Run("cannot list measures from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						measures(first: 100) {
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
				Measures struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"measures"`
			} `json:"node"`
		}

		// Query org1's measures as org2
		err := org2Owner.Execute(query, map[string]any{
			"id": org1Owner.GetOrganizationID().String(),
		}, &result)

		// Should either error or return empty list (can't access other org's data)
		if err == nil {
			// If no error, the measure from org1 should not be in the list
			for _, edge := range result.Node.Measures.Edges {
				assert.NotEqual(t, measureID, edge.Node.ID, "Should not see measure from another org")
			}
		}
	})
}
