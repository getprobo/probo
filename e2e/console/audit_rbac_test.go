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

func TestAudit_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	const createAuditMutation = `
		mutation CreateAudit($input: CreateAuditInput!) {
			createAudit(input: $input) {
				auditEdge { node { id } }
			}
		}
	`

	const updateAuditMutation = `
		mutation UpdateAudit($input: UpdateAuditInput!) {
			updateAudit(input: $input) {
				audit { id }
			}
		}
	`

	const deleteAuditMutation = `
		mutation DeleteAudit($input: DeleteAuditInput!) {
			deleteAudit(input: $input) {
				deletedAuditId
			}
		}
	`

	const readAuditQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Audit { id name }
			}
		}
	`

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
					allowMsg: "owner should be able to create audit",
				},
				{
					name:     "admin can create",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to create audit",
				},
				{
					name:         "viewer cannot create",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to create audit",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						frameworkID := factory.NewFramework(ownerClient).
							WithName(factory.SafeName("RBAC framework create " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							createAuditMutation,
							map[string]any{
								"input": map[string]any{
									"organizationId": client.GetOrganizationID().String(),
									"frameworkId":    frameworkID,
									"name":           factory.SafeName("RBAC create " + string(tt.role)),
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
					allowMsg:   "owner should be able to update audit",
				},
				{
					name:       "admin can update",
					role:       testutil.RoleAdmin,
					updateName: "Updated by Admin",
					allowMsg:   "admin should be able to update audit",
				},
				{
					name:         "viewer cannot update",
					role:         testutil.RoleViewer,
					forbidden:    true,
					updateName:   "Updated by Viewer",
					forbiddenMsg: "viewer should not be able to update audit",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						frameworkID := factory.NewFramework(ownerClient).
							WithName(factory.SafeName("RBAC framework update " + string(tt.role))).
							Create()
						auditID := factory.NewAudit(ownerClient, frameworkID).
							WithName(factory.SafeName("RBAC update " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							updateAuditMutation,
							map[string]any{
								"input": map[string]any{
									"id":   auditID,
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
					allowMsg: "owner should be able to delete audit",
				},
				{
					name:     "admin can delete",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to delete audit",
				},
				{
					name:         "viewer cannot delete",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to delete audit",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						frameworkID := factory.NewFramework(ownerClient).
							WithName(factory.SafeName("RBAC framework delete " + string(tt.role))).
							Create()
						auditID := factory.NewAudit(ownerClient, frameworkID).
							WithName(factory.SafeName("RBAC delete " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							deleteAuditMutation,
							map[string]any{
								"input": map[string]any{
									"auditId": auditID,
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
					allowMsg:    "owner should be able to read audit",
					nodePresent: "owner should receive audit data",
				},
				{
					name:        "admin can read",
					role:        testutil.RoleAdmin,
					allowMsg:    "admin should be able to read audit",
					nodePresent: "admin should receive audit data",
				},
				{
					name:        "viewer can read",
					role:        testutil.RoleViewer,
					allowMsg:    "viewer should be able to read audit",
					nodePresent: "viewer should receive audit data",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						frameworkID := factory.NewFramework(ownerClient).
							WithName(factory.SafeName("RBAC framework read " + string(tt.role))).
							Create()
						auditID := factory.NewAudit(ownerClient, frameworkID).
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
							readAuditQuery,
							map[string]any{
								"id": auditID,
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
func TestAudit_UploadReport_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	const uploadAuditReportMutation = `
		mutation UploadAuditReport($input: UploadAuditReportInput!) {
			uploadAuditReport(input: $input) {
				audit {
					id
				}
			}
		}
	`

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	uploadFile := testutil.UploadFile{
		Filename:    "report.pdf",
		ContentType: "application/pdf",
		Content:     pdfContent,
	}

	tests := []struct {
		name         string
		role         testutil.TestRole
		forbidden    bool
		allowMsg     string
		forbiddenMsg string
	}{
		{
			name:     "owner can upload",
			role:     testutil.RoleOwner,
			allowMsg: "owner should be able to upload report",
		},
		{
			name:     "admin can upload",
			role:     testutil.RoleAdmin,
			allowMsg: "admin should be able to upload report",
		},
		{
			name:         "viewer cannot upload",
			role:         testutil.RoleViewer,
			forbidden:    true,
			forbiddenMsg: "viewer should not be able to upload report",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				owner := org.Client(t, testutil.RoleOwner)

				frameworkID := factory.NewFramework(owner).
					WithName(factory.SafeName("RBAC framework upload " + string(tt.role))).
					Create()

				auditID := factory.NewAudit(owner, frameworkID).
					WithName(factory.SafeName("RBAC upload " + string(tt.role))).
					Create()

				client := org.Client(t, tt.role)

				err := client.ExecuteWithFile(
					uploadAuditReportMutation,
					map[string]any{
						"input": map[string]any{
							"auditId": auditID,
							"file":    nil,
						},
					},
					"input.file",
					uploadFile,
					nil,
				)
				if tt.forbidden {
					testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
				} else {
					require.NoError(t, err, tt.allowMsg)
				}
			},
		)
	}
}
func TestAudit_DeleteReport_RBAC(t *testing.T) {
	t.Parallel()

	uploadQuery := `
		mutation UploadAuditReport($input: UploadAuditReportInput!) {
			uploadAuditReport(input: $input) {
				audit { id }
			}
		}
	`

	deleteQuery := `
		mutation DeleteAuditReport($input: DeleteAuditReportInput!) {
			deleteAuditReport(input: $input) {
				audit { id }
			}
		}
	`

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	t.Run("viewer cannot delete report", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
		auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Delete Report Test").Create()

		// Owner uploads the report
		err := owner.ExecuteWithFile(uploadQuery, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
				"file":    nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, nil)
		require.NoError(t, err)

		// Viewer tries to delete
		_, err = viewer.Do(deleteQuery, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to delete report")
	})
}
