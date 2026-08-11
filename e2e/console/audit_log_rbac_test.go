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

func TestAuditLog_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Generate an audit log entry.
	factory.NewThirdParty(owner).WithName(factory.SafeName("RBACThirdParty")).Create()

	const query = `
		query($orgId: ID!) {
			node(id: $orgId) {
				... on Organization {
					auditLogEntries(first: 10) {
						edges {
							node {
								id
								action
							}
						}
						totalCount
					}
				}
			}
		}
	`

	t.Run("viewer can list audit log entries", func(t *testing.T) {
		t.Parallel()
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		var result struct {
			Node struct {
				AuditLogEntries struct {
					Edges []struct {
						Node struct {
							ID     string `json:"id"`
							Action string `json:"action"`
						} `json:"node"`
					} `json:"edges"`
					TotalCount int `json:"totalCount"`
				} `json:"auditLogEntries"`
			} `json:"node"`
		}

		err := viewer.Execute(query, map[string]any{
			"orgId": viewer.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Node.AuditLogEntries.TotalCount, 1)
	})

	t.Run("admin can list audit log entries", func(t *testing.T) {
		t.Parallel()
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)

		var result struct {
			Node struct {
				AuditLogEntries struct {
					TotalCount int `json:"totalCount"`
				} `json:"auditLogEntries"`
			} `json:"node"`
		}

		err := admin.Execute(query, map[string]any{
			"orgId": admin.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Node.AuditLogEntries.TotalCount, 1)
	})
}
