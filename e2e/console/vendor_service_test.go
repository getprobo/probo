// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

func TestVendorService_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a vendor first
	createVendorMutation := `
		mutation CreateVendor($input: CreateVendorInput!) {
			createVendor(input: $input) {
				vendorEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createVendorResult struct {
		CreateVendor struct {
			VendorEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"vendorEdge"`
		} `json:"createVendor"`
	}

	err := owner.Execute(createVendorMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "AWS",
			"category":       "CLOUD_PROVIDER",
		},
	}, &createVendorResult)
	require.NoError(t, err)

	vendorID := createVendorResult.CreateVendor.VendorEdge.Node.ID

	tests := []struct {
		name      string
		role      testutil.TestRole
		variables func() map[string]any
		check     func(t *testing.T, err error, m *struct {
			CreateVendorService struct {
				VendorServiceEdge struct {
					Node struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"node"`
				} `json:"vendorServiceEdge"`
			} `json:"createVendorService"`
		})
	}{
		{
			name: "Owner can create vendor service",
			role: testutil.RoleOwner,
			variables: func() map[string]any {
				return map[string]any{
					"input": map[string]any{
						"vendorId":    vendorID,
						"name":        "Amazon S3",
						"description": "Simple Storage Service",
					},
				}
			},
			check: func(t *testing.T, err error, m *struct {
				CreateVendorService struct {
					VendorServiceEdge struct {
						Node struct {
							ID          string  `json:"id"`
							Name        string  `json:"name"`
							Description *string `json:"description"`
						} `json:"node"`
					} `json:"vendorServiceEdge"`
				} `json:"createVendorService"`
			}) {
				require.NoError(t, err)
				assert.NotEmpty(t, m.CreateVendorService.VendorServiceEdge.Node.ID)
				assert.Equal(t, "Amazon S3", m.CreateVendorService.VendorServiceEdge.Node.Name)
				assert.Equal(t, "Simple Storage Service", *m.CreateVendorService.VendorServiceEdge.Node.Description)
			},
		},
		{
			name: "Admin can create vendor service",
			role: testutil.RoleAdmin,
			variables: func() map[string]any {
				return map[string]any{
					"input": map[string]any{
						"vendorId":    vendorID,
						"name":        "Amazon EC2",
						"description": "Elastic Compute Cloud",
					},
				}
			},
			check: func(t *testing.T, err error, m *struct {
				CreateVendorService struct {
					VendorServiceEdge struct {
						Node struct {
							ID          string  `json:"id"`
							Name        string  `json:"name"`
							Description *string `json:"description"`
						} `json:"node"`
					} `json:"vendorServiceEdge"`
				} `json:"createVendorService"`
			}) {
				require.NoError(t, err)
			},
		},
		{
			name: "Viewer cannot create vendor service",
			role: testutil.RoleViewer,
			variables: func() map[string]any {
				return map[string]any{
					"input": map[string]any{
						"vendorId": vendorID,
						"name":     "Should Fail",
					},
				}
			},
			check: func(t *testing.T, err error, m *struct {
				CreateVendorService struct {
					VendorServiceEdge struct {
						Node struct {
							ID          string  `json:"id"`
							Name        string  `json:"name"`
							Description *string `json:"description"`
						} `json:"node"`
					} `json:"vendorServiceEdge"`
				} `json:"createVendorService"`
			}) {
				require.Error(t, err, "Viewer should not be able to create vendor service")
			},
		},
	}

	createVendorServiceMutation := `
		mutation CreateVendorService($input: CreateVendorServiceInput!) {
			createVendorService(input: $input) {
				vendorServiceEdge {
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
				CreateVendorService struct {
					VendorServiceEdge struct {
						Node struct {
							ID          string  `json:"id"`
							Name        string  `json:"name"`
							Description *string `json:"description"`
						} `json:"node"`
					} `json:"vendorServiceEdge"`
				} `json:"createVendorService"`
			}

			err := client.Execute(createVendorServiceMutation, tt.variables(), &m)
			tt.check(t, err, &m)
		})
	}
}

func TestVendorService_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a vendor first
	createVendorMutation := `
		mutation CreateVendor($input: CreateVendorInput!) {
			createVendor(input: $input) {
				vendorEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createVendorResult struct {
		CreateVendor struct {
			VendorEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"vendorEdge"`
		} `json:"createVendor"`
	}

	err := owner.Execute(createVendorMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Google Cloud",
			"category":       "CLOUD_PROVIDER",
		},
	}, &createVendorResult)
	require.NoError(t, err)

	vendorID := createVendorResult.CreateVendor.VendorEdge.Node.ID

	// Create a vendor service
	createServiceMutation := `
		mutation CreateVendorService($input: CreateVendorServiceInput!) {
			createVendorService(input: $input) {
				vendorServiceEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createServiceResult struct {
		CreateVendorService struct {
			VendorServiceEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"vendorServiceEdge"`
		} `json:"createVendorService"`
	}

	err = owner.Execute(createServiceMutation, map[string]any{
		"input": map[string]any{
			"vendorId":    vendorID,
			"name":        "Cloud Storage",
			"description": "Initial description",
		},
	}, &createServiceResult)
	require.NoError(t, err)

	serviceID := createServiceResult.CreateVendorService.VendorServiceEdge.Node.ID

	tests := []struct {
		name      string
		role      testutil.TestRole
		variables func() map[string]any
		check     func(t *testing.T, err error, m *struct {
			UpdateVendorService struct {
				VendorService struct {
					ID          string  `json:"id"`
					Name        string  `json:"name"`
					Description *string `json:"description"`
				} `json:"vendorService"`
			} `json:"updateVendorService"`
		})
	}{
		{
			name: "Owner can update vendor service",
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
				UpdateVendorService struct {
					VendorService struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"vendorService"`
				} `json:"updateVendorService"`
			}) {
				require.NoError(t, err)
				assert.Equal(t, serviceID, m.UpdateVendorService.VendorService.ID)
				assert.Equal(t, "Updated Cloud Storage", m.UpdateVendorService.VendorService.Name)
				assert.Equal(t, "Updated description", *m.UpdateVendorService.VendorService.Description)
			},
		},
		{
			name: "Admin can update vendor service",
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
				UpdateVendorService struct {
					VendorService struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"vendorService"`
				} `json:"updateVendorService"`
			}) {
				require.NoError(t, err)
			},
		},
		{
			name: "Viewer cannot update vendor service",
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
				UpdateVendorService struct {
					VendorService struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"vendorService"`
				} `json:"updateVendorService"`
			}) {
				require.Error(t, err, "Viewer should not be able to update vendor service")
			},
		},
	}

	updateVendorServiceMutation := `
		mutation UpdateVendorService($input: UpdateVendorServiceInput!) {
			updateVendorService(input: $input) {
				vendorService {
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
				UpdateVendorService struct {
					VendorService struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"vendorService"`
				} `json:"updateVendorService"`
			}

			err := client.Execute(updateVendorServiceMutation, tt.variables(), &m)
			tt.check(t, err, &m)
		})
	}
}

func TestVendorService_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a vendor first
	createVendorMutation := `
		mutation CreateVendor($input: CreateVendorInput!) {
			createVendor(input: $input) {
				vendorEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createVendorResult struct {
		CreateVendor struct {
			VendorEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"vendorEdge"`
		} `json:"createVendor"`
	}

	err := owner.Execute(createVendorMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Azure",
			"category":       "CLOUD_PROVIDER",
		},
	}, &createVendorResult)
	require.NoError(t, err)

	vendorID := createVendorResult.CreateVendor.VendorEdge.Node.ID

	createService := func() string {
		createServiceMutation := `
			mutation CreateVendorService($input: CreateVendorServiceInput!) {
				createVendorService(input: $input) {
					vendorServiceEdge {
						node {
							id
						}
					}
				}
			}
		`

		var m struct {
			CreateVendorService struct {
				VendorServiceEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"vendorServiceEdge"`
			} `json:"createVendorService"`
		}

		err := owner.Execute(createServiceMutation, map[string]any{
			"input": map[string]any{
				"vendorId": vendorID,
				"name":     "Service to delete",
			},
		}, &m)
		require.NoError(t, err)

		return m.CreateVendorService.VendorServiceEdge.Node.ID
	}

	tests := []struct {
		name      string
		role      testutil.TestRole
		variables func(serviceID string) map[string]any
		check     func(t *testing.T, err error, serviceID string, m *struct {
			DeleteVendorService struct {
				DeletedVendorServiceID string `json:"deletedVendorServiceId"`
			} `json:"deleteVendorService"`
		})
	}{
		{
			name: "Viewer cannot delete vendor service",
			role: testutil.RoleViewer,
			variables: func(serviceID string) map[string]any {
				return map[string]any{
					"input": map[string]any{
						"vendorServiceId": serviceID,
					},
				}
			},
			check: func(t *testing.T, err error, serviceID string, m *struct {
				DeleteVendorService struct {
					DeletedVendorServiceID string `json:"deletedVendorServiceId"`
				} `json:"deleteVendorService"`
			}) {
				require.Error(t, err, "Viewer should not be able to delete vendor service")
			},
		},
		{
			name: "Admin can delete vendor service",
			role: testutil.RoleAdmin,
			variables: func(serviceID string) map[string]any {
				return map[string]any{
					"input": map[string]any{
						"vendorServiceId": serviceID,
					},
				}
			},
			check: func(t *testing.T, err error, serviceID string, m *struct {
				DeleteVendorService struct {
					DeletedVendorServiceID string `json:"deletedVendorServiceId"`
				} `json:"deleteVendorService"`
			}) {
				require.NoError(t, err)
				assert.Equal(t, serviceID, m.DeleteVendorService.DeletedVendorServiceID)
			},
		},
		{
			name: "Owner can delete vendor service",
			role: testutil.RoleOwner,
			variables: func(serviceID string) map[string]any {
				return map[string]any{
					"input": map[string]any{
						"vendorServiceId": serviceID,
					},
				}
			},
			check: func(t *testing.T, err error, serviceID string, m *struct {
				DeleteVendorService struct {
					DeletedVendorServiceID string `json:"deletedVendorServiceId"`
				} `json:"deleteVendorService"`
			}) {
				require.NoError(t, err)
				assert.Equal(t, serviceID, m.DeleteVendorService.DeletedVendorServiceID)
			},
		},
	}

	deleteVendorServiceMutation := `
		mutation DeleteVendorService($input: DeleteVendorServiceInput!) {
			deleteVendorService(input: $input) {
				deletedVendorServiceId
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
				DeleteVendorService struct {
					DeletedVendorServiceID string `json:"deletedVendorServiceId"`
				} `json:"deleteVendorService"`
			}

			err := client.Execute(deleteVendorServiceMutation, tt.variables(serviceID), &m)
			tt.check(t, err, serviceID, &m)
		})
	}
}

func TestVendorService_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a vendor first
	createVendorMutation := `
		mutation CreateVendor($input: CreateVendorInput!) {
			createVendor(input: $input) {
				vendorEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createVendorResult struct {
		CreateVendor struct {
			VendorEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"vendorEdge"`
		} `json:"createVendor"`
	}

	err := owner.Execute(createVendorMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Vendor for Services",
			"category":       "CLOUD_PROVIDER",
		},
	}, &createVendorResult)
	require.NoError(t, err)

	vendorID := createVendorResult.CreateVendor.VendorEdge.Node.ID

	// Create multiple services
	createServiceMutation := `
		mutation CreateVendorService($input: CreateVendorServiceInput!) {
			createVendorService(input: $input) {
				vendorServiceEdge {
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
			CreateVendorService struct {
				VendorServiceEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"vendorServiceEdge"`
			} `json:"createVendorService"`
		}

		err := owner.Execute(createServiceMutation, map[string]any{
			"input": map[string]any{
				"vendorId": vendorID,
				"name":     name,
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
			name: "Owner can list vendor services",
			role: testutil.RoleOwner,
			variables: func() map[string]any {
				return map[string]any{
					"id": vendorID,
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
			name: "Viewer can list vendor services",
			role: testutil.RoleViewer,
			variables: func() map[string]any {
				return map[string]any{
					"id": vendorID,
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

	listVendorServicesQuery := `
		query ListVendorServices($id: ID!) {
			node(id: $id) {
				... on Vendor {
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

			err := client.Execute(listVendorServicesQuery, tt.variables(), &q)
			tt.check(t, err, &q)
		})
	}
}
