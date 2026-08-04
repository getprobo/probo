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

func TestAuditLog_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a thirdParty in org1 to generate audit log entries.
	factory.NewThirdParty(org1Owner).WithName(factory.SafeName("IsoThirdParty")).Create()

	const query = `
		query($orgId: ID!) {
			node(id: $orgId) {
				... on Organization {
					auditLogEntries(first: 50) {
						edges {
							node {
								id
								action
								resourceType
							}
						}
						totalCount
					}
				}
			}
		}
	`

	// org2 should not see org1's audit log entries about thirdParties.
	var result struct {
		Node struct {
			AuditLogEntries struct {
				Edges []struct {
					Node struct {
						ID           string `json:"id"`
						Action       string `json:"action"`
						ResourceType string `json:"resourceType"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"auditLogEntries"`
		} `json:"node"`
	}

	err := org2Owner.Execute(query, map[string]any{
		"orgId": org2Owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	for _, edge := range result.Node.AuditLogEntries.Edges {
		// org2 may have its own audit log entries (from user/org creation),
		// but should never see org1's thirdParty entries.
		if edge.Node.ResourceType == "ThirdParty" {
			t.Fatalf("org2 should not see org1's thirdParty audit log entries, but found: %s", edge.Node.Action)
		}
	}
}
