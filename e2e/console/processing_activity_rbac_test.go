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

func TestProcessingActivity_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	const createProcessingActivityMutation = `
		mutation CreateProcessingActivity($input: CreateProcessingActivityInput!) {
			createProcessingActivity(input: $input) {
				processingActivityEdge { node { id } }
			}
		}
	`

	const updateProcessingActivityMutation = `
		mutation UpdateProcessingActivity($input: UpdateProcessingActivityInput!) {
			updateProcessingActivity(input: $input) {
				processingActivity { id }
			}
		}
	`

	const deleteProcessingActivityMutation = `
		mutation DeleteProcessingActivity($input: DeleteProcessingActivityInput!) {
			deleteProcessingActivity(input: $input) {
				deletedProcessingActivityId
			}
		}
	`

	const readProcessingActivityQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on ProcessingActivity { id name }
			}
		}
	`

	processingActivityCreateInput := func(
		client *testutil.Client,
		name string,
	) map[string]any {
		return map[string]any{
			"organizationId":                       client.GetOrganizationID().String(),
			"name":                                 name,
			"lawfulBasis":                          "CONSENT",
			"specialOrCriminalData":                "NO",
			"internationalTransfers":               false,
			"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
			"transferImpactAssessmentNeeded":       "NOT_NEEDED",
			"role":                                 "CONTROLLER",
		}
	}

	t.Run(
		"create",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name         string
				role         testutil.TestRole
				forbidden    bool
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:     "owner can create",
					role:     testutil.RoleOwner,
					allowMsg: "owner should be able to create processing activity",
				},
				{
					name:     "admin can create",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to create processing activity",
				},
				{
					name:         "viewer cannot create",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to create processing activity",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							createProcessingActivityMutation,
							map[string]any{
								"input": processingActivityCreateInput(
									client,
									factory.SafeName("RBAC create "+string(tt.role)),
								),
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
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
				name         string
				role         testutil.TestRole
				forbidden    bool
				updateName   string
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:       "owner can update",
					role:       testutil.RoleOwner,
					updateName: "Updated by Owner",
					allowMsg:   "owner should be able to update processing activity",
				},
				{
					name:       "admin can update",
					role:       testutil.RoleAdmin,
					updateName: "Updated by Admin",
					allowMsg:   "admin should be able to update processing activity",
				},
				{
					name:         "viewer cannot update",
					role:         testutil.RoleViewer,
					forbidden:    true,
					updateName:   "Updated by Viewer",
					forbiddenMsg: "viewer should not be able to update processing activity",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						paID := factory.NewProcessingActivity(ownerClient).
							WithName(factory.SafeName("RBAC update " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							updateProcessingActivityMutation,
							map[string]any{
								"input": map[string]any{
									"id":   paID,
									"name": tt.updateName,
								},
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
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
				name         string
				role         testutil.TestRole
				forbidden    bool
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:     "owner can delete",
					role:     testutil.RoleOwner,
					allowMsg: "owner should be able to delete processing activity",
				},
				{
					name:     "admin can delete",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to delete processing activity",
				},
				{
					name:         "viewer cannot delete",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to delete processing activity",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						paID := factory.NewProcessingActivity(ownerClient).
							WithName(factory.SafeName("RBAC delete " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							deleteProcessingActivityMutation,
							map[string]any{
								"input": map[string]any{
									"processingActivityId": paID,
								},
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
						}
					},
				)
			}
		},
	)

	t.Run(
		"read",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name        string
				role        testutil.TestRole
				allowMsg    string
				nodePresent string
			}{
				{
					name:        "owner can read",
					role:        testutil.RoleOwner,
					allowMsg:    "owner should be able to read processing activity",
					nodePresent: "owner should receive processing activity data",
				},
				{
					name:        "admin can read",
					role:        testutil.RoleAdmin,
					allowMsg:    "admin should be able to read processing activity",
					nodePresent: "admin should receive processing activity data",
				},
				{
					name:        "viewer can read",
					role:        testutil.RoleViewer,
					allowMsg:    "viewer should be able to read processing activity",
					nodePresent: "viewer should receive processing activity data",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						paID := factory.NewProcessingActivity(ownerClient).
							WithName(factory.SafeName("RBAC read " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						var result struct {
							Node *struct {
								ID   string `json:"id"`
								Name string `json:"name"`
							} `json:"node"`
						}

						err := client.Execute(
							readProcessingActivityQuery,
							map[string]any{
								"id": paID,
							},
							&result,
						)
						require.NoError(t, err, tt.allowMsg)
						require.NotNil(t, result.Node, tt.nodePresent)
					},
				)
			}
		},
	)
}
func TestProcessingActivity_DPIA_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	const createDPIAMutation = `
		mutation($input: CreateDataProtectionImpactAssessmentInput!) {
			createDataProtectionImpactAssessment(input: $input) {
				dataProtectionImpactAssessment { id }
			}
		}
	`

	const updateDPIAMutation = `
		mutation($input: UpdateDataProtectionImpactAssessmentInput!) {
			updateDataProtectionImpactAssessment(input: $input) {
				dataProtectionImpactAssessment { id }
			}
		}
	`

	const deleteDPIAMutation = `
		mutation($input: DeleteDataProtectionImpactAssessmentInput!) {
			deleteDataProtectionImpactAssessment(input: $input) {
				deletedDataProtectionImpactAssessmentId
			}
		}
	`

	type dpiaCreateResult struct {
		CreateDataProtectionImpactAssessment struct {
			DataProtectionImpactAssessment struct {
				ID string `json:"id"`
			} `json:"dataProtectionImpactAssessment"`
		} `json:"createDataProtectionImpactAssessment"`
	}

	t.Run(
		"owner can manage DPIA",
		func(t *testing.T) {
			t.Parallel()

			ownerClient := org.Client(t, testutil.RoleOwner)

			paID := factory.NewProcessingActivity(ownerClient).
				WithName(factory.SafeName("DPIA RBAC Owner Test")).
				Create()

			var createResult dpiaCreateResult

			err := ownerClient.Execute(
				createDPIAMutation,
				map[string]any{
					"input": map[string]any{
						"processingActivityId": paID,
						"description":          "Owner DPIA",
					},
				},
				&createResult,
			)
			require.NoError(t, err, "owner should be able to create DPIA")

			dpiaID := createResult.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessment.ID

			_, err = ownerClient.Do(
				updateDPIAMutation,
				map[string]any{
					"input": map[string]any{
						"id":          dpiaID,
						"description": "Updated by owner",
					},
				},
			)
			require.NoError(t, err, "owner should be able to update DPIA")

			_, err = ownerClient.Do(
				deleteDPIAMutation,
				map[string]any{
					"input": map[string]any{
						"dataProtectionImpactAssessmentId": dpiaID,
					},
				},
			)
			require.NoError(t, err, "owner should be able to delete DPIA")
		},
	)

	t.Run(
		"admin can manage DPIA",
		func(t *testing.T) {
			t.Parallel()

			ownerClient := org.Client(t, testutil.RoleOwner)
			adminClient := org.Client(t, testutil.RoleAdmin)

			paID := factory.NewProcessingActivity(ownerClient).
				WithName(factory.SafeName("DPIA RBAC Admin Test")).
				Create()

			var createResult dpiaCreateResult

			err := adminClient.Execute(
				createDPIAMutation,
				map[string]any{
					"input": map[string]any{
						"processingActivityId": paID,
						"description":          "Admin DPIA",
					},
				},
				&createResult,
			)
			require.NoError(t, err, "admin should be able to create DPIA")

			dpiaID := createResult.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessment.ID

			_, err = adminClient.Do(
				updateDPIAMutation,
				map[string]any{
					"input": map[string]any{
						"id":          dpiaID,
						"description": "Updated by admin",
					},
				},
			)
			require.NoError(t, err, "admin should be able to update DPIA")

			_, err = adminClient.Do(
				deleteDPIAMutation,
				map[string]any{
					"input": map[string]any{
						"dataProtectionImpactAssessmentId": dpiaID,
					},
				},
			)
			require.NoError(t, err, "admin should be able to delete DPIA")
		},
	)

	t.Run(
		"viewer cannot manage DPIA",
		func(t *testing.T) {
			t.Parallel()

			ownerClient := org.Client(t, testutil.RoleOwner)
			viewer := org.Client(t, testutil.RoleViewer)

			paID := factory.NewProcessingActivity(ownerClient).
				WithName(factory.SafeName("DPIA RBAC Viewer Test")).
				Create()

			_, err := viewer.Do(
				createDPIAMutation,
				map[string]any{
					"input": map[string]any{
						"processingActivityId": paID,
						"description":          "Viewer DPIA",
					},
				},
			)
			testutil.RequireForbiddenError(t, err, "viewer should not be able to create DPIA")

			var createResult dpiaCreateResult

			err = ownerClient.Execute(
				createDPIAMutation,
				map[string]any{
					"input": map[string]any{
						"processingActivityId": paID,
						"description":          "Owner created DPIA",
					},
				},
				&createResult,
			)
			require.NoError(t, err)

			dpiaID := createResult.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessment.ID

			_, err = viewer.Do(
				updateDPIAMutation,
				map[string]any{
					"input": map[string]any{
						"id":          dpiaID,
						"description": "Updated by viewer",
					},
				},
			)
			testutil.RequireForbiddenError(t, err, "viewer should not be able to update DPIA")

			_, err = viewer.Do(
				deleteDPIAMutation,
				map[string]any{
					"input": map[string]any{
						"dataProtectionImpactAssessmentId": dpiaID,
					},
				},
			)
			testutil.RequireForbiddenError(t, err, "viewer should not be able to delete DPIA")
		},
	)
}
func TestProcessingActivity_TIA_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	const createTIAMutation = `
		mutation($input: CreateTransferImpactAssessmentInput!) {
			createTransferImpactAssessment(input: $input) {
				transferImpactAssessment { id }
			}
		}
	`

	const updateTIAMutation = `
		mutation($input: UpdateTransferImpactAssessmentInput!) {
			updateTransferImpactAssessment(input: $input) {
				transferImpactAssessment { id }
			}
		}
	`

	const deleteTIAMutation = `
		mutation($input: DeleteTransferImpactAssessmentInput!) {
			deleteTransferImpactAssessment(input: $input) {
				deletedTransferImpactAssessmentId
			}
		}
	`

	type tiaCreateResult struct {
		CreateTransferImpactAssessment struct {
			TransferImpactAssessment struct {
				ID string `json:"id"`
			} `json:"transferImpactAssessment"`
		} `json:"createTransferImpactAssessment"`
	}

	t.Run(
		"owner can manage TIA",
		func(t *testing.T) {
			t.Parallel()

			ownerClient := org.Client(t, testutil.RoleOwner)

			paID := factory.NewProcessingActivity(ownerClient).
				WithName(factory.SafeName("TIA RBAC Owner Test")).
				Create()

			var createResult tiaCreateResult

			err := ownerClient.Execute(
				createTIAMutation,
				map[string]any{
					"input": map[string]any{
						"processingActivityId": paID,
						"dataSubjects":         "Owner TIA subjects",
					},
				},
				&createResult,
			)
			require.NoError(t, err, "owner should be able to create TIA")

			tiaID := createResult.CreateTransferImpactAssessment.TransferImpactAssessment.ID

			_, err = ownerClient.Do(
				updateTIAMutation,
				map[string]any{
					"input": map[string]any{
						"id":           tiaID,
						"dataSubjects": "Updated by owner",
					},
				},
			)
			require.NoError(t, err, "owner should be able to update TIA")

			_, err = ownerClient.Do(
				deleteTIAMutation,
				map[string]any{
					"input": map[string]any{
						"transferImpactAssessmentId": tiaID,
					},
				},
			)
			require.NoError(t, err, "owner should be able to delete TIA")
		},
	)

	t.Run(
		"admin can manage TIA",
		func(t *testing.T) {
			t.Parallel()

			ownerClient := org.Client(t, testutil.RoleOwner)
			adminClient := org.Client(t, testutil.RoleAdmin)

			paID := factory.NewProcessingActivity(ownerClient).
				WithName(factory.SafeName("TIA RBAC Admin Test")).
				Create()

			var createResult tiaCreateResult

			err := adminClient.Execute(
				createTIAMutation,
				map[string]any{
					"input": map[string]any{
						"processingActivityId": paID,
						"dataSubjects":         "Admin TIA subjects",
					},
				},
				&createResult,
			)
			require.NoError(t, err, "admin should be able to create TIA")

			tiaID := createResult.CreateTransferImpactAssessment.TransferImpactAssessment.ID

			_, err = adminClient.Do(
				updateTIAMutation,
				map[string]any{
					"input": map[string]any{
						"id":           tiaID,
						"dataSubjects": "Updated by admin",
					},
				},
			)
			require.NoError(t, err, "admin should be able to update TIA")

			_, err = adminClient.Do(
				deleteTIAMutation,
				map[string]any{
					"input": map[string]any{
						"transferImpactAssessmentId": tiaID,
					},
				},
			)
			require.NoError(t, err, "admin should be able to delete TIA")
		},
	)

	t.Run(
		"viewer cannot manage TIA",
		func(t *testing.T) {
			t.Parallel()

			ownerClient := org.Client(t, testutil.RoleOwner)
			viewer := org.Client(t, testutil.RoleViewer)

			paID := factory.NewProcessingActivity(ownerClient).
				WithName(factory.SafeName("TIA RBAC Viewer Test")).
				Create()

			_, err := viewer.Do(
				createTIAMutation,
				map[string]any{
					"input": map[string]any{
						"processingActivityId": paID,
						"dataSubjects":         "Viewer TIA subjects",
					},
				},
			)
			testutil.RequireForbiddenError(t, err, "viewer should not be able to create TIA")

			var createResult tiaCreateResult

			err = ownerClient.Execute(
				createTIAMutation,
				map[string]any{
					"input": map[string]any{
						"processingActivityId": paID,
						"dataSubjects":         "Owner created TIA",
					},
				},
				&createResult,
			)
			require.NoError(t, err)

			tiaID := createResult.CreateTransferImpactAssessment.TransferImpactAssessment.ID

			_, err = viewer.Do(
				updateTIAMutation,
				map[string]any{
					"input": map[string]any{
						"id":           tiaID,
						"dataSubjects": "Updated by viewer",
					},
				},
			)
			testutil.RequireForbiddenError(t, err, "viewer should not be able to update TIA")

			_, err = viewer.Do(
				deleteTIAMutation,
				map[string]any{
					"input": map[string]any{
						"transferImpactAssessmentId": tiaID,
					},
				},
			)
			testutil.RequireForbiddenError(t, err, "viewer should not be able to delete TIA")
		},
	)
}
