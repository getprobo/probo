// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestAuditLog_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a thirdParty to generate an audit log entry.
	factory.NewThirdParty(owner).WithName(factory.SafeName("AuditThirdParty")).Create()

	const query = `
		query($orgId: ID!) {
			node(id: $orgId) {
				... on Organization {
					auditLogEntries(first: 10) {
						edges {
							node {
								id
								actorId
								actorType
								action
								resourceType
								resourceId
								createdAt
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			AuditLogEntries struct {
				Edges []struct {
					Node struct {
						ID           string `json:"id"`
						ActorID      string `json:"actorId"`
						ActorType    string `json:"actorType"`
						Action       string `json:"action"`
						ResourceType string `json:"resourceType"`
						ResourceID   string `json:"resourceId"`
						CreatedAt    string `json:"createdAt"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"auditLogEntries"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.AuditLogEntries.TotalCount, 1)

	// Find the thirdParty create entry.
	found := false
	for _, edge := range result.Node.AuditLogEntries.Edges {
		if edge.Node.Action == "core:third-party:create" {
			found = true
			assert.Equal(t, "USER", edge.Node.ActorType)
			assert.Equal(t, "ThirdParty", edge.Node.ResourceType)
			assert.NotEmpty(t, edge.Node.ActorID)
			assert.NotEmpty(t, edge.Node.ResourceID)
			assert.NotEmpty(t, edge.Node.CreatedAt)
			break
		}
	}
	assert.True(t, found, "expected to find core:thirdParty:create audit log entry")
}

func TestAuditLog_Filter(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create different resources to generate different audit log entries.
	factory.NewThirdParty(owner).WithName(factory.SafeName("FilterThirdParty")).Create()

	const query = `
		query($orgId: ID!, $filter: AuditLogEntryFilter) {
			node(id: $orgId) {
				... on Organization {
					auditLogEntries(first: 50, filter: $filter) {
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

	t.Run("filter by action", func(t *testing.T) {
		t.Parallel()

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

		err := owner.Execute(query, map[string]any{
			"orgId":  owner.GetOrganizationID().String(),
			"filter": map[string]any{"action": "core:third-party:create"},
		}, &result)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Node.AuditLogEntries.TotalCount, 1)
		for _, edge := range result.Node.AuditLogEntries.Edges {
			assert.Equal(t, "core:third-party:create", edge.Node.Action)
		}
	})

	t.Run("filter by resource type", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node struct {
				AuditLogEntries struct {
					Edges []struct {
						Node struct {
							ID           string `json:"id"`
							ResourceType string `json:"resourceType"`
						} `json:"node"`
					} `json:"edges"`
					TotalCount int `json:"totalCount"`
				} `json:"auditLogEntries"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"orgId":  owner.GetOrganizationID().String(),
			"filter": map[string]any{"resourceType": "ThirdParty"},
		}, &result)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Node.AuditLogEntries.TotalCount, 1)
		for _, edge := range result.Node.AuditLogEntries.Edges {
			assert.Equal(t, "ThirdParty", edge.Node.ResourceType)
		}
	})
}

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
