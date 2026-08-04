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

func TestAuditProgram_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	frameworkID := factory.NewFramework(owner).WithName("Framework for Audit Program").Create()

	query := `
		mutation CreateAuditProgram($input: CreateAuditProgramInput!) {
			createAuditProgram(input: $input) {
				auditProgramEdge {
					node {
						id
						name
						framework { id }
					}
				}
			}
		}
	`

	var result struct {
		CreateAuditProgram struct {
			AuditProgramEdge struct {
				Node struct {
					ID        string `json:"id"`
					Name      string `json:"name"`
					Framework struct {
						ID string `json:"id"`
					} `json:"framework"`
				} `json:"node"`
			} `json:"auditProgramEdge"`
		} `json:"createAuditProgram"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"frameworkId":    frameworkID,
			"name":           "ISO 27001 2024-2027",
		},
	}, &result)
	require.NoError(t, err)

	node := result.CreateAuditProgram.AuditProgramEdge.Node
	assert.NotEmpty(t, node.ID)
	assert.Equal(t, "ISO 27001 2024-2027", node.Name)
	assert.Equal(t, frameworkID, node.Framework.ID)
}

func TestAuditProgram_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	frameworkID := factory.NewFramework(owner).WithName("Framework for Audit Program Update").Create()
	programID := factory.NewAuditProgram(owner, frameworkID).WithName("Original Program").Create()

	query := `
		mutation UpdateAuditProgram($input: UpdateAuditProgramInput!) {
			updateAuditProgram(input: $input) {
				auditProgram {
					id
					name
				}
			}
		}
	`

	var result struct {
		UpdateAuditProgram struct {
			AuditProgram struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"auditProgram"`
		} `json:"updateAuditProgram"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":   programID,
			"name": "Updated Program",
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, programID, result.UpdateAuditProgram.AuditProgram.ID)
	assert.Equal(t, "Updated Program", result.UpdateAuditProgram.AuditProgram.Name)
}

func TestAuditProgram_Delete_DetachesAudits(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	frameworkID := factory.NewFramework(owner).WithName("Framework for Audit Program Delete").Create()
	programID := factory.NewAuditProgram(owner, frameworkID).WithName("Program to Delete").Create()
	auditID := factory.NewAudit(owner, frameworkID).
		WithName("Linked Audit").
		WithAuditProgramID(programID).
		Create()

	deleteQuery := `
		mutation DeleteAuditProgram($input: DeleteAuditProgramInput!) {
			deleteAuditProgram(input: $input) {
				deletedAuditProgramId
			}
		}
	`

	var deleteResult struct {
		DeleteAuditProgram struct {
			DeletedAuditProgramID string `json:"deletedAuditProgramId"`
		} `json:"deleteAuditProgram"`
	}

	err := owner.Execute(deleteQuery, map[string]any{
		"input": map[string]any{"auditProgramId": programID},
	}, &deleteResult)
	require.NoError(t, err)
	assert.Equal(t, programID, deleteResult.DeleteAuditProgram.DeletedAuditProgramID)

	getQuery := `
		query($id: ID!) {
			node(id: $id) {
				... on Audit {
					id
					auditProgram { id }
				}
			}
		}
	`

	var getResult struct {
		Node struct {
			ID           string `json:"id"`
			AuditProgram *struct {
				ID string `json:"id"`
			} `json:"auditProgram"`
		} `json:"node"`
	}

	err = owner.Execute(getQuery, map[string]any{"id": auditID}, &getResult)
	require.NoError(t, err)
	assert.Equal(t, auditID, getResult.Node.ID)
	assert.Nil(t, getResult.Node.AuditProgram)
}

func TestAuditProgram_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	frameworkID := factory.NewFramework(owner).WithName("Framework for Audit Program List").Create()
	programID := factory.NewAuditProgram(owner, frameworkID).WithName("Listed Program").Create()

	query := `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					auditPrograms(first: 50) {
						edges {
							node { id name }
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			AuditPrograms struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"auditPrograms"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": owner.GetOrganizationID().String()}, &result)
	require.NoError(t, err)

	found := false

	for _, edge := range result.Node.AuditPrograms.Edges {
		if edge.Node.ID == programID {
			found = true

			assert.Equal(t, "Listed Program", edge.Node.Name)

			break
		}
	}

	assert.True(t, found, "expected to find created audit program in list")
}

func TestAudit_Create_WithAuditProgram(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	frameworkID := factory.NewFramework(owner).WithName("Framework for Linked Audit").Create()
	programID := factory.NewAuditProgram(owner, frameworkID).WithName("Parent Program").Create()

	query := `
		mutation CreateAudit($input: CreateAuditInput!) {
			createAudit(input: $input) {
				auditEdge {
					node {
						id
						auditProgram { id name }
					}
				}
			}
		}
	`

	var result struct {
		CreateAudit struct {
			AuditEdge struct {
				Node struct {
					ID           string `json:"id"`
					AuditProgram *struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"auditProgram"`
				} `json:"node"`
			} `json:"auditEdge"`
		} `json:"createAudit"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"frameworkId":    frameworkID,
			"auditProgramId": programID,
			"name":           "Year 1 Internal",
		},
	}, &result)
	require.NoError(t, err)

	node := result.CreateAudit.AuditEdge.Node
	require.NotNil(t, node.AuditProgram)
	assert.Equal(t, programID, node.AuditProgram.ID)
	assert.Equal(t, "Parent Program", node.AuditProgram.Name)
}

func TestAuditProgram_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	frameworkID := factory.NewFramework(owner).WithName("Framework for Audit Program RBAC").Create()

	t.Run("viewer cannot create", func(t *testing.T) {
		query := `
			mutation CreateAuditProgram($input: CreateAuditProgramInput!) {
				createAuditProgram(input: $input) {
					auditProgramEdge { node { id } }
				}
			}
		`

		_, err := viewer.Do(query, map[string]any{
			"input": map[string]any{
				"organizationId": viewer.GetOrganizationID().String(),
				"frameworkId":    frameworkID,
				"name":           "RBAC Program",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create audit program")
	})

	t.Run("viewer can list", func(t *testing.T) {
		programID := factory.NewAuditProgram(owner, frameworkID).WithName("Readable Program").Create()

		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						auditPrograms(first: 50) {
							edges { node { id } }
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				AuditPrograms struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"auditPrograms"`
			} `json:"node"`
		}

		err := viewer.Execute(query, map[string]any{"id": owner.GetOrganizationID().String()}, &result)
		require.NoError(t, err)

		found := false

		for _, edge := range result.Node.AuditPrograms.Edges {
			if edge.Node.ID == programID {
				found = true

				break
			}
		}

		assert.True(t, found)
	})
}

func TestAuditProgram_TenantIsolation(t *testing.T) {
	t.Parallel()
	ownerA := testutil.NewClient(t, testutil.RoleOwner)
	ownerB := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(ownerA).WithName("Tenant A Framework").Create()
	programID := factory.NewAuditProgram(ownerA, frameworkID).WithName("Tenant A Program").Create()

	query := `
		query($id: ID!) {
			node(id: $id) {
				... on AuditProgram {
					id
				}
			}
		}
	`

	var result struct {
		Node *struct {
			ID string `json:"id"`
		} `json:"node"`
	}

	err := ownerB.Execute(query, map[string]any{"id": programID}, &result)
	testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "audit program")
}
