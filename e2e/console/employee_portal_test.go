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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestEmployeePortal_Create(t *testing.T) {
	t.Parallel()

	t.Run("with required fields", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		const query = `
			mutation CreateEmployeePortal($input: CreateEmployeePortalInput!) {
				createEmployeePortal(input: $input) {
					employeePortalEdge {
						node {
							id
							name
							active
							capabilities {
								deviceAgent
							}
							createdAt
							updatedAt
						}
					}
				}
			}
		`

		name := factory.SafeName("Employee Portal")

		var result struct {
			CreateEmployeePortal struct {
				EmployeePortalEdge struct {
					Node struct {
						ID           string `json:"id"`
						Name         string `json:"name"`
						Active       bool   `json:"active"`
						Capabilities struct {
							DeviceAgent bool `json:"deviceAgent"`
						} `json:"capabilities"`
						CreatedAt string `json:"createdAt"`
						UpdatedAt string `json:"updatedAt"`
					} `json:"node"`
				} `json:"employeePortalEdge"`
			} `json:"createEmployeePortal"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           name,
			},
		}, &result)

		require.NoError(t, err)

		node := result.CreateEmployeePortal.EmployeePortalEdge.Node
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, name, node.Name)
		assert.True(t, node.Active)
		assert.True(t, node.Capabilities.DeviceAgent)
		assert.NotEmpty(t, node.CreatedAt)
		assert.NotEmpty(t, node.UpdatedAt)
	})
}

func TestEmployeePortal_Update(t *testing.T) {
	t.Parallel()

	t.Run("partial update name and active", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		portalID := factory.CreateEmployeePortal(owner)
		newName := factory.SafeName("Updated Portal")

		const query = `
			mutation UpdateEmployeePortal($input: UpdateEmployeePortalInput!) {
				updateEmployeePortal(input: $input) {
					employeePortal {
						id
						name
						active
						capabilities {
							deviceAgent
						}
					}
				}
			}
		`

		var result struct {
			UpdateEmployeePortal struct {
				EmployeePortal struct {
					ID           string `json:"id"`
					Name         string `json:"name"`
					Active       bool   `json:"active"`
					Capabilities struct {
						DeviceAgent bool `json:"deviceAgent"`
					} `json:"capabilities"`
				} `json:"employeePortal"`
			} `json:"updateEmployeePortal"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"employeePortalId": portalID,
				"name":             newName,
				"active":           false,
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, portalID, result.UpdateEmployeePortal.EmployeePortal.ID)
		assert.Equal(t, newName, result.UpdateEmployeePortal.EmployeePortal.Name)
		assert.False(t, result.UpdateEmployeePortal.EmployeePortal.Active)
		assert.True(t, result.UpdateEmployeePortal.EmployeePortal.Capabilities.DeviceAgent)
	})

	t.Run("disable device agent", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		portalID := factory.CreateEmployeePortal(owner)

		const query = `
			mutation UpdateEmployeePortal($input: UpdateEmployeePortalInput!) {
				updateEmployeePortal(input: $input) {
					employeePortal {
						id
						capabilities {
							deviceAgent
						}
					}
				}
			}
		`

		var result struct {
			UpdateEmployeePortal struct {
				EmployeePortal struct {
					ID           string `json:"id"`
					Capabilities struct {
						DeviceAgent bool `json:"deviceAgent"`
					} `json:"capabilities"`
				} `json:"employeePortal"`
			} `json:"updateEmployeePortal"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"employeePortalId": portalID,
				"capabilities": map[string]any{
					"deviceAgent": false,
				},
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, portalID, result.UpdateEmployeePortal.EmployeePortal.ID)
		assert.False(t, result.UpdateEmployeePortal.EmployeePortal.Capabilities.DeviceAgent)
	})
}

func TestEmployeePortal_UpdateBrand(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	portalID := factory.CreateEmployeePortal(owner)

	const uploadMutation = `
		mutation UpdateEmployeePortalBrand($input: UpdateEmployeePortalBrandInput!) {
			updateEmployeePortalBrand(input: $input) {
				employeePortal {
					id
					logo {
						id
						fileName
						downloadUrl
					}
				}
			}
		}
	`

	pngContent := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}

	var uploadResult struct {
		UpdateEmployeePortalBrand struct {
			EmployeePortal struct {
				ID   string `json:"id"`
				Logo *struct {
					ID          string `json:"id"`
					FileName    string `json:"fileName"`
					DownloadURL string `json:"downloadUrl"`
				} `json:"logo"`
			} `json:"employeePortal"`
		} `json:"updateEmployeePortalBrand"`
	}

	err := owner.ExecuteWithFile(uploadMutation, map[string]any{
		"input": map[string]any{
			"employeePortalId": portalID,
			"logoFile":         nil,
		},
	}, "input.logoFile", testutil.UploadFile{
		Filename:    "employee-portal-logo.png",
		ContentType: "image/png",
		Content:     pngContent,
	}, &uploadResult)
	require.NoError(t, err)
	require.NotNil(t, uploadResult.UpdateEmployeePortalBrand.EmployeePortal.Logo)
	assert.Equal(t, "employee-portal-logo.png", uploadResult.UpdateEmployeePortalBrand.EmployeePortal.Logo.FileName)
	assert.True(
		t,
		strings.Contains(uploadResult.UpdateEmployeePortalBrand.EmployeePortal.Logo.DownloadURL, "/api/files/v1/public/"),
		"downloadUrl must route through the public files API, got %q",
		uploadResult.UpdateEmployeePortalBrand.EmployeePortal.Logo.DownloadURL,
	)
}

func TestEmployeePortal_Delete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		portalID := factory.CreateEmployeePortal(owner)

		const query = `
			mutation DeleteEmployeePortal($input: DeleteEmployeePortalInput!) {
				deleteEmployeePortal(input: $input) {
					deletedEmployeePortalId
				}
			}
		`

		var result struct {
			DeleteEmployeePortal struct {
				DeletedEmployeePortalID string `json:"deletedEmployeePortalId"`
			} `json:"deleteEmployeePortal"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"employeePortalId": portalID,
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, portalID, result.DeleteEmployeePortal.DeletedEmployeePortalID)
	})
}

func TestEmployeePortal_List(t *testing.T) {
	t.Parallel()

	t.Run("lists portals via organization", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		factory.CreateEmployeePortal(owner)
		factory.CreateEmployeePortal(owner)

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						employeePortals(first: 10) {
							totalCount
							edges {
								node {
									id
									name
								}
							}
							pageInfo {
								hasNextPage
								hasPreviousPage
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				EmployeePortals struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						HasNextPage     bool `json:"hasNextPage"`
						HasPreviousPage bool `json:"hasPreviousPage"`
					} `json:"pageInfo"`
				} `json:"employeePortals"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &result)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Node.EmployeePortals.TotalCount, 2)
		assert.GreaterOrEqual(t, len(result.Node.EmployeePortals.Edges), 2)
	})
}

func TestEmployeePortal_Node(t *testing.T) {
	t.Parallel()

	t.Run("fetch by ID", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		name := factory.SafeName("Employee Portal")
		portalID := factory.CreateEmployeePortal(owner, factory.Attrs{"name": name})

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on EmployeePortal {
						id
						name
						active
						capabilities {
							deviceAgent
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID           string `json:"id"`
				Name         string `json:"name"`
				Active       bool   `json:"active"`
				Capabilities struct {
					DeviceAgent bool `json:"deviceAgent"`
				} `json:"capabilities"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": portalID}, &result)
		require.NoError(t, err)
		assert.Equal(t, portalID, result.Node.ID)
		assert.Equal(t, name, result.Node.Name)
		assert.True(t, result.Node.Active)
		assert.True(t, result.Node.Capabilities.DeviceAgent)
	})
}

func TestEmployeePortal_CreatedWithOrganization(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					employeePortals(first: 10) {
						totalCount
						edges {
							node {
								id
								name
								organization { id }
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			EmployeePortals struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID           string `json:"id"`
						Name         string `json:"name"`
						Organization struct {
							ID string `json:"id"`
						} `json:"organization"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"employeePortals"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	require.Equal(t, 1, result.Node.EmployeePortals.TotalCount)
	require.Len(t, result.Node.EmployeePortals.Edges, 1)
	assert.Equal(t, owner.GetOrganizationID().String(), result.Node.EmployeePortals.Edges[0].Node.Organization.ID)
}

func TestEmployeePortal_EmployeeCanGetDefaultPortal(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	portalID := factory.DefaultEmployeePortalID(owner)

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on EmployeePortal {
					id
					organization { id }
				}
			}
		}
	`

	var result struct {
		Node struct {
			ID           string `json:"id"`
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
		} `json:"node"`
	}

	err := employee.Execute(query, map[string]any{"id": portalID}, &result)
	require.NoError(t, err)
	assert.Equal(t, portalID, result.Node.ID)
	assert.Equal(t, owner.GetOrganizationID().String(), result.Node.Organization.ID)
}

func TestEmployeePortal_ConnectAssumeAndList(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	portalID := factory.DefaultEmployeePortalID(owner)
	root := testutil.NewClientWithNewSession(t, owner)

	const listQuery = `
		query {
			viewer {
				profiles(first: 100, filter: { states: [ACTIVE] }) {
					edges {
						node {
							organization {
								id
								employeePortals(
									first: 1
									orderBy: { field: CREATED_AT, direction: ASC }
								) {
									edges {
										node { id }
									}
								}
							}
						}
					}
				}
			}
		}
	`

	var listResult struct {
		Viewer struct {
			Profiles struct {
				Edges []struct {
					Node struct {
						Organization struct {
							ID              string `json:"id"`
							EmployeePortals struct {
								Edges []struct {
									Node struct {
										ID string `json:"id"`
									} `json:"node"`
								} `json:"edges"`
							} `json:"employeePortals"`
						} `json:"organization"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"profiles"`
		} `json:"viewer"`
	}

	err := root.ExecuteConnect(listQuery, nil, &listResult)
	require.NoError(t, err)

	var listedPortalID string
	for _, edge := range listResult.Viewer.Profiles.Edges {
		if edge.Node.Organization.ID == owner.GetOrganizationID().String() {
			require.NotEmpty(t, edge.Node.Organization.EmployeePortals.Edges)
			listedPortalID = edge.Node.Organization.EmployeePortals.Edges[0].Node.ID
			break
		}
	}
	require.NotEmpty(t, listedPortalID)
	assert.Equal(t, portalID, listedPortalID)

	const assumeMutation = `
		mutation($input: AssumeEmployeePortalSessionInput!) {
			assumeEmployeePortalSession(input: $input) {
				organizationId
				result {
					__typename
					... on OrganizationSessionCreated {
						session { id }
					}
				}
			}
		}
	`

	var assumeResult struct {
		AssumeEmployeePortalSession struct {
			OrganizationID string `json:"organizationId"`
			Result         struct {
				Typename string `json:"__typename"`
			} `json:"result"`
		} `json:"assumeEmployeePortalSession"`
	}

	err = root.ExecuteConnect(assumeMutation, map[string]any{
		"input": map[string]any{
			"employeePortalId": portalID,
			"continue":         testutil.GetBaseURL(),
		},
	}, &assumeResult)
	require.NoError(t, err)
	assert.Equal(t, owner.GetOrganizationID().String(), assumeResult.AssumeEmployeePortalSession.OrganizationID)
	assert.Equal(t, "OrganizationSessionCreated", assumeResult.AssumeEmployeePortalSession.Result.Typename)

	const nodeQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on EmployeePortal {
					id
					organization { id }
				}
			}
		}
	`

	var nodeResult struct {
		Node struct {
			ID           string `json:"id"`
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
		} `json:"node"`
	}

	err = root.Execute(nodeQuery, map[string]any{"id": portalID}, &nodeResult)
	require.NoError(t, err)
	assert.Equal(t, portalID, nodeResult.Node.ID)
	assert.Equal(t, owner.GetOrganizationID().String(), nodeResult.Node.Organization.ID)
}
