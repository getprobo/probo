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

func TestAudit_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(org1Owner).WithName("Org1 Framework").Create()
	auditID := factory.NewAudit(org1Owner, frameworkID).WithName("Org1 Audit").Create()

	t.Run("cannot read audit from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Audit {
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

		err := org2Owner.Execute(query, map[string]any{"id": auditID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "audit")
	})

	t.Run("cannot update audit from another organization", func(t *testing.T) {
		query := `
			mutation UpdateAudit($input: UpdateAuditInput!) {
				updateAudit(input: $input) {
					audit { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   auditID,
				"name": "Hijacked Audit",
			},
		})
		require.Error(t, err, "Should not be able to update audit from another org")
	})

	t.Run("cannot delete audit from another organization", func(t *testing.T) {
		query := `
			mutation DeleteAudit($input: DeleteAuditInput!) {
				deleteAudit(input: $input) {
					deletedAuditId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
			},
		})
		require.Error(t, err, "Should not be able to delete audit from another org")
	})

	t.Run("cannot list audits from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						audits(first: 100) {
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
				Audits struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"audits"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{
			"id": org1Owner.GetOrganizationID().String(),
		}, &result)
		if err == nil {
			for _, edge := range result.Node.Audits.Edges {
				assert.NotEqual(t, auditID, edge.Node.ID, "Should not see audit from another org")
			}
		}
	})
}
