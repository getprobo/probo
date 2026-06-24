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

// trustCenterID looks up the caller's own organization's trust center id.
func trustCenterID(t *testing.T, c *testutil.Client) string {
	t.Helper()

	var result struct {
		Node struct {
			TrustCenter struct {
				ID string `json:"id"`
			} `json:"trustCenter"`
		} `json:"node"`
	}

	err := c.Execute(`
		query($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					trustCenter { id }
				}
			}
		}
	`, map[string]any{
		"organizationId": c.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	require.NotEmpty(t, result.Node.TrustCenter.ID)

	return result.Node.TrustCenter.ID
}

func TestComplianceFramework_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	trustCenterID := trustCenterID(t, owner)
	frameworkID := factory.CreateFramework(owner)

	var result struct {
		CreateComplianceFramework struct {
			ComplianceFrameworkEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"complianceFrameworkEdge"`
		} `json:"createComplianceFramework"`
	}

	err := owner.Execute(`
		mutation($input: CreateComplianceFrameworkInput!) {
			createComplianceFramework(input: $input) {
				complianceFrameworkEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"trustCenterId": trustCenterID,
			"frameworkId":   frameworkID,
		},
	}, &result)
	require.NoError(t, err)
	require.NotEmpty(t, result.CreateComplianceFramework.ComplianceFrameworkEdge.Node.ID)
}

// TestComplianceFramework_TenantIsolation covers GHSA-c74x-79w6-63jh's
// structural sibling: ComplianceFrameworkService.Create must not accept a
// frameworkId belonging to another organization -- the FK is tenant-agnostic
// (ON DELETE CASCADE) so a cross-tenant reference would let org A pin a link
// to org B's framework and would let org B silently cascade-delete org A's
// compliance page entry.
func TestComplianceFramework_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1TrustCenterID := trustCenterID(t, org1Owner)
	org2FrameworkID := factory.CreateFramework(org2Owner)

	_, err := org1Owner.Do(`
		mutation($input: CreateComplianceFrameworkInput!) {
			createComplianceFramework(input: $input) {
				complianceFrameworkEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"trustCenterId": org1TrustCenterID,
			"frameworkId":   org2FrameworkID,
		},
	})
	require.Error(t, err, "must not accept a frameworkId belonging to another organization")
}
