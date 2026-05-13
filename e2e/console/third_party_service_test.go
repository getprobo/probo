// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestThirdPartyService_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a thirdParty first
	createThirdPartyMutation := `
		mutation CreateThirdParty($input: CreateThirdPartyInput!) {
			createThirdParty(input: $input) {
				thirdPartyEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createThirdPartyResult struct {
		CreateThirdParty struct {
			ThirdPartyEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyEdge"`
		} `json:"createThirdParty"`
	}

	err := owner.Execute(createThirdPartyMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "AWS",
			"category":       "CLOUD_PROVIDER",
		},
	}, &createThirdPartyResult)
	require.NoError(t, err)

	thirdPartyID := createThirdPartyResult.CreateThirdParty.ThirdPartyEdge.Node.ID

	tests := []struct {
		name      string
		role      testutil.TestRole
		variables func() map[string]any
		check     func(t *testing.T, err error, m *struct {
			CreateThirdPartyService struct {
				ThirdPartyServiceEdge struct {
					Node struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"node"`
				} `json:"thirdPartyServiceEdge"`
			} `json:"createThirdPartyService"`
		})
	}{
		{
			name: "Owner can create thirdParty service",
			role: testutil.RoleOwner,
			variables: func() map[string]any {
				return map[string]any{
					"input": map[string]any{
						"thirdPartyId": thirdPartyID,
						"name":         "Amazon S3",
						"description":  "Simple Storage Service",
					},
				}
			},
			check: func(t *testing.T, err error, m *struct {
				CreateThirdPartyService struct {
					ThirdPartyServiceEdge struct {
						Node struct {
							ID          string  `json:"id"`
							Name        string  `json:"name"`
							Description *string `json:"description"`
						} `json:"node"`
					} `json:"thirdPartyServiceEdge"`
				} `json:"createThirdPartyService"`
			}) {
				require.NoError(t, err)
				assert.NotEmpty(t, m.CreateThirdPartyService.ThirdPartyServiceEdge.Node.ID)
				assert.Equal(t, "Amazon S3", m.CreateThirdPartyService.ThirdPartyServiceEdge.Node.Name)
				assert.Equal(t, "Simple Storage Service", *m.CreateThirdPartyService.ThirdPartyServiceEdge.Node.Description)
			},
		},
		{
			name: "Admin can create thirdParty service",
			role: testutil.RoleAdmin,
			variables: func() map[string]any {
				return map[string]any{
					"input": map[string]any{
						"thirdPartyId": thirdPartyID,
						"name":         "Amazon EC2",
						"description":  "Elastic Compute Cloud",
					},
				}
			},
			check: func(t *testing.T, err error, m *struct {
				CreateThirdPartyService struct {
					ThirdPartyServiceEdge struct {
						Node struct {
							ID          string  `json:"id"`
							Name        string  `json:"name"`
							Description *string `json:"description"`
						} `json:"node"`
					} `json:"thirdPartyServiceEdge"`
				} `json:"createThirdPartyService"`
			}) {
				require.NoError(t, err)
			},
		},
		{
			name: "Viewer cannot create thirdParty service",
			role: testutil.RoleViewer,
			variables: func() map[string]any {
				return map[string]any{
					"input": map[string]any{
						"thirdPartyId": thirdPartyID,
						"name":         "Should Fail",
					},
				}
			},
			check: func(t *testing.T, err error, m *struct {
				CreateThirdPartyService struct {
					ThirdPartyServiceEdge struct {
						Node struct {
							ID          string  `json:"id"`
							Name        string  `json:"name"`
							Description *string `json:"description"`
						} `json:"node"`
					} `json:"thirdPartyServiceEdge"`
				} `json:"createThirdPartyService"`
			}) {
				require.Error(t, err, "Viewer should not be able to create thirdParty service")
			},
		},
	}

	createThirdPartyServiceMutation := `
		mutation CreateThirdPartyService($input: CreateThirdPartyServiceInput!) {
			createThirdPartyService(input: $input) {
				thirdPartyServiceEdge {
					node {
						id
						name
						description
					}
				}
			}
		}
	`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *testutil.Client
			if tt.role == testutil.RoleOwner {
				client = owner
			} else {
				client = testutil.NewClientInOrg(t, tt.role, owner)
			}

			var m struct {
				CreateThirdPartyService struct {
					ThirdPartyServiceEdge struct {
						Node struct {
							ID          string  `json:"id"`
							Name        string  `json:"name"`
							Description *string `json:"description"`
						} `json:"node"`
					} `json:"thirdPartyServiceEdge"`
				} `json:"createThirdPartyService"`
			}

			err := client.Execute(createThirdPartyServiceMutation, tt.variables(), &m)
			tt.check(t, err, &m)
		})
	}
}

func TestThirdPartyService_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a thirdParty first
	createThirdPartyMutation := `
		mutation CreateThirdParty($input: CreateThirdPartyInput!) {
			createThirdParty(input: $input) {
				thirdPartyEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createThirdPartyResult struct {
		CreateThirdParty struct {
			ThirdPartyEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyEdge"`
		} `json:"createThirdParty"`
	}

	err := owner.Execute(createThirdPartyMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Google Cloud",
			"category":       "CLOUD_PROVIDER",
		},
	}, &createThirdPartyResult)
	require.NoError(t, err)

	thirdPartyID := createThirdPartyResult.CreateThirdParty.ThirdPartyEdge.Node.ID

	// Create a thirdParty service
	createServiceMutation := `
		mutation CreateThirdPartyService($input: CreateThirdPartyServiceInput!) {
			createThirdPartyService(input: $input) {
				thirdPartyServiceEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createServiceResult struct {
		CreateThirdPartyService struct {
			ThirdPartyServiceEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyServiceEdge"`
		} `json:"createThirdPartyService"`
	}

	err = owner.Execute(createServiceMutation, map[string]any{
		"input": map[string]any{
			"thirdPartyId": thirdPartyID,
			"name":         "Cloud Storage",
			"description":  "Initial description",
		},
	}, &createServiceResult)
	require.NoError(t, err)

	serviceID := createServiceResult.CreateThirdPartyService.ThirdPartyServiceEdge.Node.ID

	tests := []struct {
		name      string
		role      testutil.TestRole
		variables func() map[string]any
		check     func(t *testing.T, err error, m *struct {
			UpdateThirdPartyService struct {
				ThirdPartyService struct {
					ID          string  `json:"id"`
					Name        string  `json:"name"`
					Description *string `json:"description"`
				} `json:"thirdPartyService"`
			} `json:"updateThirdPartyService"`
		})
	}{
		{
			name: "Owner can update thirdParty service",
			role: testutil.RoleOwner,
			variables: func() map[string]any {
				return map[string]any{
					"input": map[string]any{
						"id":          serviceID,
						"name":        "Updated Cloud Storage",
						"description": "Updated description",
					},
				}
			},
			check: func(t *testing.T, err error, m *struct {
				UpdateThirdPartyService struct {
					ThirdPartyService struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"thirdPartyService"`
				} `json:"updateThirdPartyService"`
			}) {
				require.NoError(t, err)
				assert.Equal(t, serviceID, m.UpdateThirdPartyService.ThirdPartyService.ID)
				assert.Equal(t, "Updated Cloud Storage", m.UpdateThirdPartyService.ThirdPartyService.Name)
				assert.Equal(t, "Updated description", *m.UpdateThirdPartyService.ThirdPartyService.Description)
			},
		},
		{
			name: "Admin can update thirdParty service",
			role: testutil.RoleAdmin,
			variables: func() map[string]any {
				return map[string]any{
					"input": map[string]any{
						"id":   serviceID,
						"name": "Admin Updated Storage",
					},
				}
			},
			check: func(t *testing.T, err error, m *struct {
				UpdateThirdPartyService struct {
					ThirdPartyService struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"thirdPartyService"`
				} `json:"updateThirdPartyService"`
			}) {
				require.NoError(t, err)
			},
		},
		{
			name: "Viewer cannot update thirdParty service",
			role: testutil.RoleViewer,
			variables: func() map[string]any {
				return map[string]any{
					"input": map[string]any{
						"id":   serviceID,
						"name": "Should Fail",
					},
				}
			},
			check: func(t *testing.T, err error, m *struct {
				UpdateThirdPartyService struct {
					ThirdPartyService struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"thirdPartyService"`
				} `json:"updateThirdPartyService"`
			}) {
				require.Error(t, err, "Viewer should not be able to update thirdParty service")
			},
		},
	}

	updateThirdPartyServiceMutation := `
		mutation UpdateThirdPartyService($input: UpdateThirdPartyServiceInput!) {
			updateThirdPartyService(input: $input) {
				thirdPartyService {
					id
					name
					description
				}
			}
		}
	`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *testutil.Client
			if tt.role == testutil.RoleOwner {
				client = owner
			} else {
				client = testutil.NewClientInOrg(t, tt.role, owner)
			}

			var m struct {
				UpdateThirdPartyService struct {
					ThirdPartyService struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"thirdPartyService"`
				} `json:"updateThirdPartyService"`
			}

			err := client.Execute(updateThirdPartyServiceMutation, tt.variables(), &m)
			tt.check(t, err, &m)
		})
	}
}

func TestThirdPartyService_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a thirdParty first
	createThirdPartyMutation := `
		mutation CreateThirdParty($input: CreateThirdPartyInput!) {
			createThirdParty(input: $input) {
				thirdPartyEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createThirdPartyResult struct {
		CreateThirdParty struct {
			ThirdPartyEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyEdge"`
		} `json:"createThirdParty"`
	}

	err := owner.Execute(createThirdPartyMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Azure",
			"category":       "CLOUD_PROVIDER",
		},
	}, &createThirdPartyResult)
	require.NoError(t, err)

	thirdPartyID := createThirdPartyResult.CreateThirdParty.ThirdPartyEdge.Node.ID

	createService := func() string {
		createServiceMutation := `
			mutation CreateThirdPartyService($input: CreateThirdPartyServiceInput!) {
				createThirdPartyService(input: $input) {
					thirdPartyServiceEdge {
						node {
							id
						}
					}
				}
			}
		`

		var m struct {
			CreateThirdPartyService struct {
				ThirdPartyServiceEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"thirdPartyServiceEdge"`
			} `json:"createThirdPartyService"`
		}

		err := owner.Execute(createServiceMutation, map[string]any{
			"input": map[string]any{
				"thirdPartyId": thirdPartyID,
				"name":         "Service to delete",
			},
		}, &m)
		require.NoError(t, err)

		return m.CreateThirdPartyService.ThirdPartyServiceEdge.Node.ID
	}

	tests := []struct {
		name      string
		role      testutil.TestRole
		variables func(serviceID string) map[string]any
		check     func(t *testing.T, err error, serviceID string, m *struct {
			DeleteThirdPartyService struct {
				DeletedThirdPartyServiceID string `json:"deletedThirdPartyServiceId"`
			} `json:"deleteThirdPartyService"`
		})
	}{
		{
			name: "Viewer cannot delete thirdParty service",
			role: testutil.RoleViewer,
			variables: func(serviceID string) map[string]any {
				return map[string]any{
					"input": map[string]any{
						"thirdPartyServiceId": serviceID,
					},
				}
			},
			check: func(t *testing.T, err error, serviceID string, m *struct {
				DeleteThirdPartyService struct {
					DeletedThirdPartyServiceID string `json:"deletedThirdPartyServiceId"`
				} `json:"deleteThirdPartyService"`
			}) {
				require.Error(t, err, "Viewer should not be able to delete thirdParty service")
			},
		},
		{
			name: "Admin can delete thirdParty service",
			role: testutil.RoleAdmin,
			variables: func(serviceID string) map[string]any {
				return map[string]any{
					"input": map[string]any{
						"thirdPartyServiceId": serviceID,
					},
				}
			},
			check: func(t *testing.T, err error, serviceID string, m *struct {
				DeleteThirdPartyService struct {
					DeletedThirdPartyServiceID string `json:"deletedThirdPartyServiceId"`
				} `json:"deleteThirdPartyService"`
			}) {
				require.NoError(t, err)
				assert.Equal(t, serviceID, m.DeleteThirdPartyService.DeletedThirdPartyServiceID)
			},
		},
		{
			name: "Owner can delete thirdParty service",
			role: testutil.RoleOwner,
			variables: func(serviceID string) map[string]any {
				return map[string]any{
					"input": map[string]any{
						"thirdPartyServiceId": serviceID,
					},
				}
			},
			check: func(t *testing.T, err error, serviceID string, m *struct {
				DeleteThirdPartyService struct {
					DeletedThirdPartyServiceID string `json:"deletedThirdPartyServiceId"`
				} `json:"deleteThirdPartyService"`
			}) {
				require.NoError(t, err)
				assert.Equal(t, serviceID, m.DeleteThirdPartyService.DeletedThirdPartyServiceID)
			},
		},
	}

	deleteThirdPartyServiceMutation := `
		mutation DeleteThirdPartyService($input: DeleteThirdPartyServiceInput!) {
			deleteThirdPartyService(input: $input) {
				deletedThirdPartyServiceId
			}
		}
	`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceID := createService()

			var client *testutil.Client
			if tt.role == testutil.RoleOwner {
				client = owner
			} else {
				client = testutil.NewClientInOrg(t, tt.role, owner)
			}

			var m struct {
				DeleteThirdPartyService struct {
					DeletedThirdPartyServiceID string `json:"deletedThirdPartyServiceId"`
				} `json:"deleteThirdPartyService"`
			}

			err := client.Execute(deleteThirdPartyServiceMutation, tt.variables(serviceID), &m)
			tt.check(t, err, serviceID, &m)
		})
	}
}

func TestThirdPartyService_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a thirdParty first
	createThirdPartyMutation := `
		mutation CreateThirdParty($input: CreateThirdPartyInput!) {
			createThirdParty(input: $input) {
				thirdPartyEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createThirdPartyResult struct {
		CreateThirdParty struct {
			ThirdPartyEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyEdge"`
		} `json:"createThirdParty"`
	}

	err := owner.Execute(createThirdPartyMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "ThirdParty for Services",
			"category":       "CLOUD_PROVIDER",
		},
	}, &createThirdPartyResult)
	require.NoError(t, err)

	thirdPartyID := createThirdPartyResult.CreateThirdParty.ThirdPartyEdge.Node.ID

	// Create multiple services
	createServiceMutation := `
		mutation CreateThirdPartyService($input: CreateThirdPartyServiceInput!) {
			createThirdPartyService(input: $input) {
				thirdPartyServiceEdge {
					node {
						id
					}
				}
			}
		}
	`

	services := []string{"Service A", "Service B", "Service C"}
	for _, name := range services {
		var m struct {
			CreateThirdPartyService struct {
				ThirdPartyServiceEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"thirdPartyServiceEdge"`
			} `json:"createThirdPartyService"`
		}

		err := owner.Execute(createServiceMutation, map[string]any{
			"input": map[string]any{
				"thirdPartyId": thirdPartyID,
				"name":         name,
			},
		}, &m)
		require.NoError(t, err)
	}

	tests := []struct {
		name      string
		role      testutil.TestRole
		variables func() map[string]any
		check     func(t *testing.T, err error, q *struct {
			Node struct {
				ID       string `json:"id"`
				Services struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"services"`
			} `json:"node"`
		})
	}{
		{
			name: "Owner can list thirdParty services",
			role: testutil.RoleOwner,
			variables: func() map[string]any {
				return map[string]any{
					"id": thirdPartyID,
				}
			},
			check: func(t *testing.T, err error, q *struct {
				Node struct {
					ID       string `json:"id"`
					Services struct {
						Edges []struct {
							Node struct {
								ID   string `json:"id"`
								Name string `json:"name"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"services"`
				} `json:"node"`
			}) {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(q.Node.Services.Edges), 3)
			},
		},
		{
			name: "Viewer can list thirdParty services",
			role: testutil.RoleViewer,
			variables: func() map[string]any {
				return map[string]any{
					"id": thirdPartyID,
				}
			},
			check: func(t *testing.T, err error, q *struct {
				Node struct {
					ID       string `json:"id"`
					Services struct {
						Edges []struct {
							Node struct {
								ID   string `json:"id"`
								Name string `json:"name"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"services"`
				} `json:"node"`
			}) {
				require.NoError(t, err)
			},
		},
	}

	listThirdPartyServicesQuery := `
		query ListThirdPartyServices($id: ID!) {
			node(id: $id) {
				... on ThirdParty {
					id
					services(first: 10) {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *testutil.Client
			if tt.role == testutil.RoleOwner {
				client = owner
			} else {
				client = testutil.NewClientInOrg(t, tt.role, owner)
			}

			var q struct {
				Node struct {
					ID       string `json:"id"`
					Services struct {
						Edges []struct {
							Node struct {
								ID   string `json:"id"`
								Name string `json:"name"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"services"`
				} `json:"node"`
			}

			err := client.Execute(listThirdPartyServicesQuery, tt.variables(), &q)
			tt.check(t, err, &q)
		})
	}
}
