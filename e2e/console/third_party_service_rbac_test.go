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

func TestThirdPartyService_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	const createThirdPartyServiceMutation = `
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

	const updateThirdPartyServiceMutation = `
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

	const deleteThirdPartyServiceMutation = `
		mutation DeleteThirdPartyService($input: DeleteThirdPartyServiceInput!) {
			deleteThirdPartyService(input: $input) {
				deletedThirdPartyServiceId
			}
		}
	`

	const listThirdPartyServicesQuery = `
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

	type (
		createThirdPartyServiceResponse struct {
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

		updateThirdPartyServiceResponse struct {
			UpdateThirdPartyService struct {
				ThirdPartyService struct {
					ID          string  `json:"id"`
					Name        string  `json:"name"`
					Description *string `json:"description"`
				} `json:"thirdPartyService"`
			} `json:"updateThirdPartyService"`
		}

		deleteThirdPartyServiceResponse struct {
			DeleteThirdPartyService struct {
				DeletedThirdPartyServiceID string `json:"deletedThirdPartyServiceId"`
			} `json:"deleteThirdPartyService"`
		}

		listThirdPartyServicesResponse struct {
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
	)

	createService := func(
		t *testing.T,
		owner *testutil.Client,
		thirdPartyID string,
		name string,
		description string,
	) string {
		t.Helper()

		input := map[string]any{
			"thirdPartyId": thirdPartyID,
			"name":         name,
		}
		if description != "" {
			input["description"] = description
		}

		var m createThirdPartyServiceResponse

		err := owner.Execute(
			createThirdPartyServiceMutation,
			map[string]any{
				"input": input,
			},
			&m,
		)
		require.NoError(t, err)
		require.NotEmpty(
			t,
			m.CreateThirdPartyService.ThirdPartyServiceEdge.Node.ID,
		)

		return m.CreateThirdPartyService.ThirdPartyServiceEdge.Node.ID
	}

	t.Run(
		"create",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name        string
				role        testutil.TestRole
				serviceName string
				description *string
				forbidden   bool
				assertOK    func(t *testing.T, m *createThirdPartyServiceResponse)
			}{
				{
					name:        "owner can create thirdParty service",
					role:        testutil.RoleOwner,
					serviceName: "Amazon S3",
					description: new("Simple Storage Service"),
					assertOK: func(t *testing.T, m *createThirdPartyServiceResponse) {
						assert.NotEmpty(t, m.CreateThirdPartyService.ThirdPartyServiceEdge.Node.ID)
						assert.Equal(t, "Amazon S3", m.CreateThirdPartyService.ThirdPartyServiceEdge.Node.Name)
						assert.Equal(
							t,
							"Simple Storage Service",
							*m.CreateThirdPartyService.ThirdPartyServiceEdge.Node.Description,
						)
					},
				},
				{
					name:        "admin can create thirdParty service",
					role:        testutil.RoleAdmin,
					serviceName: "Amazon EC2",
					description: new("Elastic Compute Cloud"),
				},
				{
					name:        "viewer cannot create thirdParty service",
					role:        testutil.RoleViewer,
					serviceName: "Should Fail",
					forbidden:   true,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						owner := org.Client(t, testutil.RoleOwner)

						thirdPartyID := factory.NewThirdParty(owner).
							WithName(factory.SafeName("RBAC create third party " + string(tt.role))).
							WithCategory("CLOUD_PROVIDER").
							Create()

						client := org.Client(t, tt.role)

						input := map[string]any{
							"thirdPartyId": thirdPartyID,
							"name":         tt.serviceName,
						}
						if tt.description != nil {
							input["description"] = *tt.description
						}

						var m createThirdPartyServiceResponse

						err := client.Execute(
							createThirdPartyServiceMutation,
							map[string]any{
								"input": input,
							},
							&m,
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(
								t,
								err,
								"Viewer should not be able to create thirdParty service",
							)

							return
						}

						require.NoError(t, err)

						if tt.assertOK != nil {
							tt.assertOK(t, &m)
						}
					},
				)
			}
		},
	)

	t.Run(
		"update",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name        string
				role        testutil.TestRole
				updateName  string
				description *string
				forbidden   bool
				assertOK    func(t *testing.T, serviceID string, m *updateThirdPartyServiceResponse)
			}{
				{
					name:        "owner can update thirdParty service",
					role:        testutil.RoleOwner,
					updateName:  "Updated Cloud Storage",
					description: new("Updated description"),
					assertOK: func(t *testing.T, serviceID string, m *updateThirdPartyServiceResponse) {
						assert.Equal(t, serviceID, m.UpdateThirdPartyService.ThirdPartyService.ID)
						assert.Equal(t, "Updated Cloud Storage", m.UpdateThirdPartyService.ThirdPartyService.Name)
						assert.Equal(
							t,
							"Updated description",
							*m.UpdateThirdPartyService.ThirdPartyService.Description,
						)
					},
				},
				{
					name:       "admin can update thirdParty service",
					role:       testutil.RoleAdmin,
					updateName: "Admin Updated Storage",
				},
				{
					name:       "viewer cannot update thirdParty service",
					role:       testutil.RoleViewer,
					updateName: "Should Fail",
					forbidden:  true,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						owner := org.Client(t, testutil.RoleOwner)

						thirdPartyID := factory.NewThirdParty(owner).
							WithName(factory.SafeName("RBAC update third party " + string(tt.role))).
							WithCategory("CLOUD_PROVIDER").
							Create()

						serviceID := createService(
							t,
							owner,
							thirdPartyID,
							"Cloud Storage",
							"Initial description",
						)

						client := org.Client(t, tt.role)

						input := map[string]any{
							"id":   serviceID,
							"name": tt.updateName,
						}
						if tt.description != nil {
							input["description"] = *tt.description
						}

						var m updateThirdPartyServiceResponse

						err := client.Execute(
							updateThirdPartyServiceMutation,
							map[string]any{
								"input": input,
							},
							&m,
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(
								t,
								err,
								"Viewer should not be able to update thirdParty service",
							)

							return
						}

						require.NoError(t, err)

						if tt.assertOK != nil {
							tt.assertOK(t, serviceID, &m)
						}
					},
				)
			}
		},
	)

	t.Run(
		"delete",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name      string
				role      testutil.TestRole
				forbidden bool
			}{
				{
					name:      "viewer cannot delete thirdParty service",
					role:      testutil.RoleViewer,
					forbidden: true,
				},
				{
					name: "admin can delete thirdParty service",
					role: testutil.RoleAdmin,
				},
				{
					name: "owner can delete thirdParty service",
					role: testutil.RoleOwner,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						owner := org.Client(t, testutil.RoleOwner)

						thirdPartyID := factory.NewThirdParty(owner).
							WithName(factory.SafeName("RBAC delete third party " + string(tt.role))).
							WithCategory("CLOUD_PROVIDER").
							Create()

						serviceID := createService(
							t,
							owner,
							thirdPartyID,
							factory.SafeName("Service to delete"),
							"",
						)

						client := org.Client(t, tt.role)

						var m deleteThirdPartyServiceResponse

						err := client.Execute(
							deleteThirdPartyServiceMutation,
							map[string]any{
								"input": map[string]any{
									"thirdPartyServiceId": serviceID,
								},
							},
							&m,
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(
								t,
								err,
								"Viewer should not be able to delete thirdParty service",
							)

							return
						}

						require.NoError(t, err)
						assert.Equal(t, serviceID, m.DeleteThirdPartyService.DeletedThirdPartyServiceID)
					},
				)
			}
		},
	)

	t.Run(
		"list",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name           string
				role           testutil.TestRole
				serviceCount   int
				minListedCount int
			}{
				{
					name:           "owner can list thirdParty services",
					role:           testutil.RoleOwner,
					serviceCount:   3,
					minListedCount: 3,
				},
				{
					name:         "viewer can list thirdParty services",
					role:         testutil.RoleViewer,
					serviceCount: 1,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						owner := org.Client(t, testutil.RoleOwner)

						thirdPartyID := factory.NewThirdParty(owner).
							WithName(factory.SafeName("RBAC list third party " + string(tt.role))).
							WithCategory("CLOUD_PROVIDER").
							Create()

						listServiceNames := []string{
							"Service A",
							"Service B",
							"Service C",
						}
						if tt.serviceCount == 1 {
							listServiceNames = []string{
								factory.SafeName("Service list " + string(tt.role)),
							}
						}

						for _, serviceName := range listServiceNames {
							createService(
								t,
								owner,
								thirdPartyID,
								serviceName,
								"",
							)
						}

						client := org.Client(t, tt.role)

						var q listThirdPartyServicesResponse

						err := client.Execute(
							listThirdPartyServicesQuery,
							map[string]any{
								"id": thirdPartyID,
							},
							&q,
						)
						require.NoError(t, err)

						if tt.minListedCount > 0 {
							assert.GreaterOrEqual(t, len(q.Node.Services.Edges), tt.minListedCount)
						}
					},
				)
			}
		},
	)
}
