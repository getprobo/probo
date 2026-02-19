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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestReport_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Report").Create()

	tests := []struct {
		name        string
		input       map[string]any
		assertField string
		assertValue string
	}{
		{
			name: "with full details",
			input: map[string]any{
				"name":  "SOC 2 Type II Report 2025",
				"state": "NOT_STARTED",
			},
			assertField: "name",
			assertValue: "SOC 2 Type II Report 2025",
		},
		{
			name: "with NOT_STARTED state",
			input: map[string]any{
				"name":  "Report NOT_STARTED",
				"state": "NOT_STARTED",
			},
			assertField: "state",
			assertValue: "NOT_STARTED",
		},
		{
			name: "with IN_PROGRESS state",
			input: map[string]any{
				"name":  "Report IN_PROGRESS",
				"state": "IN_PROGRESS",
			},
			assertField: "state",
			assertValue: "IN_PROGRESS",
		},
		{
			name: "with COMPLETED state",
			input: map[string]any{
				"name":  "Report COMPLETED",
				"state": "COMPLETED",
			},
			assertField: "state",
			assertValue: "COMPLETED",
		},
		{
			name: "with REJECTED state",
			input: map[string]any{
				"name":  "Report REJECTED",
				"state": "REJECTED",
			},
			assertField: "state",
			assertValue: "REJECTED",
		},
		{
			name: "with OUTDATED state",
			input: map[string]any{
				"name":  "Report OUTDATED",
				"state": "OUTDATED",
			},
			assertField: "state",
			assertValue: "OUTDATED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateReport($input: CreateReportInput!) {
					createReport(input: $input) {
						reportEdge {
							node {
								id
								name
								state
							}
						}
					}
				}
			`

			input := map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"frameworkId":    frameworkID,
			}
			for k, v := range tt.input {
				input[k] = v
			}

			var result struct {
				CreateReport struct {
					ReportEdge struct {
						Node struct {
							ID    string `json:"id"`
							Name  string `json:"name"`
							State string `json:"state"`
						} `json:"node"`
					} `json:"reportEdge"`
				} `json:"createReport"`
			}

			err := owner.Execute(query, map[string]any{"input": input}, &result)
			require.NoError(t, err)

			node := result.CreateReport.ReportEdge.Node
			assert.NotEmpty(t, node.ID)

			switch tt.assertField {
			case "name":
				assert.Equal(t, tt.assertValue, node.Name)
			case "state":
				assert.Equal(t, tt.assertValue, node.State)
			}
		})
	}
}

func TestReport_Create_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Report Validation").Create()

	tests := []struct {
		name              string
		input             map[string]any
		skipOrganization  bool
		skipFramework     bool
		wantErrorContains string
	}{
		{
			name: "missing organizationId",
			input: map[string]any{
				"name": "Test Report",
			},
			skipOrganization:  true,
			wantErrorContains: "organizationId",
		},
		{
			name: "missing frameworkId",
			input: map[string]any{
				"name": "Test Report",
			},
			skipFramework:     true,
			wantErrorContains: "frameworkId",
		},
		{
			name: "name with HTML tags",
			input: map[string]any{
				"name": "<script>alert('xss')</script>",
			},
			wantErrorContains: "HTML",
		},
		{
			name: "name with newline",
			input: map[string]any{
				"name": "Test\nReport",
			},
			wantErrorContains: "newline",
		},
		{
			name: "name with carriage return",
			input: map[string]any{
				"name": "Test\rReport",
			},
			wantErrorContains: "carriage return",
		},
		{
			name: "name with null byte",
			input: map[string]any{
				"name": "Test\x00Report",
			},
			wantErrorContains: "control character",
		},
		{
			name: "name with tab character",
			input: map[string]any{
				"name": "Test\tReport",
			},
			wantErrorContains: "control character",
		},
		{
			name: "name with zero-width space",
			input: map[string]any{
				"name": "Test\u200BReport",
			},
			wantErrorContains: "zero-width",
		},
		{
			name: "name with zero-width joiner",
			input: map[string]any{
				"name": "Test\u200DReport",
			},
			wantErrorContains: "zero-width",
		},
		{
			name: "name with right-to-left override",
			input: map[string]any{
				"name": "Test\u202EReport",
			},
			wantErrorContains: "bidirectional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateReport($input: CreateReportInput!) {
					createReport(input: $input) {
						reportEdge {
							node {
								id
							}
						}
					}
				}
			`

			input := make(map[string]any)
			if !tt.skipOrganization {
				input["organizationId"] = owner.GetOrganizationID().String()
			}
			if !tt.skipFramework {
				input["frameworkId"] = frameworkID
			}
			for k, v := range tt.input {
				input[k] = v
			}

			_, err := owner.Do(query, map[string]any{"input": input})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestReport_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Report Update").Create()

	tests := []struct {
		name        string
		setup       func() string
		input       func(id string) map[string]any
		assertField string
		assertValue string
	}{
		{
			name: "update name",
			setup: func() string {
				return factory.NewReport(owner, frameworkID).
					WithName("Report to Update").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{
					"id":   id,
					"name": "Updated Report Name",
				}
			},
			assertField: "name",
			assertValue: "Updated Report Name",
		},
		{
			name: "update to IN_PROGRESS state",
			setup: func() string {
				return factory.NewReport(owner, frameworkID).
					WithName("State Test").
					WithState("NOT_STARTED").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "state": "IN_PROGRESS"}
			},
			assertField: "state",
			assertValue: "IN_PROGRESS",
		},
		{
			name: "update to COMPLETED state",
			setup: func() string {
				return factory.NewReport(owner, frameworkID).
					WithName("State Test").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "state": "COMPLETED"}
			},
			assertField: "state",
			assertValue: "COMPLETED",
		},
		{
			name: "update to REJECTED state",
			setup: func() string {
				return factory.NewReport(owner, frameworkID).
					WithName("State Test").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "state": "REJECTED"}
			},
			assertField: "state",
			assertValue: "REJECTED",
		},
		{
			name: "update to OUTDATED state",
			setup: func() string {
				return factory.NewReport(owner, frameworkID).
					WithName("State Test").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "state": "OUTDATED"}
			},
			assertField: "state",
			assertValue: "OUTDATED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportID := tt.setup()

			query := `
				mutation UpdateReport($input: UpdateReportInput!) {
					updateReport(input: $input) {
						report {
							id
							name
							state
						}
					}
				}
			`

			var result struct {
				UpdateReport struct {
					Report struct {
						ID    string `json:"id"`
						Name  string `json:"name"`
						State string `json:"state"`
					} `json:"report"`
				} `json:"updateReport"`
			}

			err := owner.Execute(query, map[string]any{"input": tt.input(reportID)}, &result)
			require.NoError(t, err)

			report := result.UpdateReport.Report
			switch tt.assertField {
			case "name":
				assert.Equal(t, tt.assertValue, report.Name)
			case "state":
				assert.Equal(t, tt.assertValue, report.State)
			}
		})
	}
}

func TestReport_Update_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Report Update Validation").Create()
	baseReportID := factory.NewReport(owner, frameworkID).WithName("Validation Test Report").Create()

	tests := []struct {
		name              string
		setup             func() string
		input             func(id string) map[string]any
		wantErrorContains string
	}{
		{
			name:  "invalid ID format",
			setup: func() string { return "invalid-id-format" },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test"}
			},
			wantErrorContains: "base64",
		},
		{
			name:  "name with HTML tags",
			setup: func() string { return baseReportID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "<script>alert('xss')</script>"}
			},
			wantErrorContains: "HTML",
		},
		{
			name:  "name with newline",
			setup: func() string { return baseReportID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\nReport"}
			},
			wantErrorContains: "newline",
		},
		{
			name:  "name with carriage return",
			setup: func() string { return baseReportID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\rReport"}
			},
			wantErrorContains: "carriage return",
		},
		{
			name:  "name with null byte",
			setup: func() string { return baseReportID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\x00Report"}
			},
			wantErrorContains: "control character",
		},
		{
			name:  "name with tab character",
			setup: func() string { return baseReportID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\tReport"}
			},
			wantErrorContains: "control character",
		},
		{
			name:  "name with zero-width space",
			setup: func() string { return baseReportID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\u200BReport"}
			},
			wantErrorContains: "zero-width",
		},
		{
			name:  "name with zero-width joiner",
			setup: func() string { return baseReportID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\u200DReport"}
			},
			wantErrorContains: "zero-width",
		},
		{
			name:  "name with right-to-left override",
			setup: func() string { return baseReportID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\u202EReport"}
			},
			wantErrorContains: "bidirectional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportID := tt.setup()

			query := `
				mutation UpdateReport($input: UpdateReportInput!) {
					updateReport(input: $input) {
						report {
							id
						}
					}
				}
			`

			_, err := owner.Do(query, map[string]any{"input": tt.input(reportID)})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestReport_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Report Delete").Create()

	t.Run("delete existing report", func(t *testing.T) {
		reportID := factory.NewReport(owner, frameworkID).WithName("Report to Delete").Create()

		query := `
			mutation DeleteReport($input: DeleteReportInput!) {
				deleteReport(input: $input) {
					deletedReportId
				}
			}
		`

		var result struct {
			DeleteReport struct {
				DeletedReportID string `json:"deletedReportId"`
			} `json:"deleteReport"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"reportId": reportID},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, reportID, result.DeleteReport.DeletedReportID)
	})
}

func TestReport_Delete_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		reportID          string
		wantErrorContains string
	}{
		{
			name:              "invalid ID format",
			reportID:          "invalid-id-format",
			wantErrorContains: "base64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation DeleteReport($input: DeleteReportInput!) {
					deleteReport(input: $input) {
						deletedReportId
					}
				}
			`

			_, err := owner.Do(query, map[string]any{
				"input": map[string]any{"reportId": tt.reportID},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestReport_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Report List").Create()

	reportNames := []string{"Report A", "Report B", "Report C"}
	for _, name := range reportNames {
		factory.NewReport(owner, frameworkID).WithName(name).Create()
	}

	query := `
		query GetReports($id: ID!) {
			node(id: $id) {
				... on Organization {
					reports(first: 10) {
						edges {
							node {
								id
								name
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
			Reports struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"reports"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Reports.TotalCount, 3)
}

func TestReport_Query(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("query with non-existent ID returns error", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Report {
						id
						name
					}
				}
			}
		`

		err := owner.ExecuteShouldFail(query, map[string]any{
			"id": "V0wtM0tMNmJBQ1lBQUFBQUFackhLSTJfbXJJRUFZVXo", // Valid format but doesn't exist
		})
		require.Error(t, err, "Non-existent ID should return error")
	})
}

func TestReport_Timestamps(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Report Timestamps").Create()

	t.Run("createdAt and updatedAt are set on create", func(t *testing.T) {
		beforeCreate := time.Now().Add(-time.Second)

		query := `
			mutation CreateReport($input: CreateReportInput!) {
				createReport(input: $input) {
					reportEdge {
						node {
							id
							createdAt
							updatedAt
						}
					}
				}
			}
		`

		var result struct {
			CreateReport struct {
				ReportEdge struct {
					Node struct {
						ID        string    `json:"id"`
						CreatedAt time.Time `json:"createdAt"`
						UpdatedAt time.Time `json:"updatedAt"`
					} `json:"node"`
				} `json:"reportEdge"`
			} `json:"createReport"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"frameworkId":    frameworkID,
				"name":           "Timestamp Test Report",
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateReport.ReportEdge.Node
		testutil.AssertTimestampsOnCreate(t, node.CreatedAt, node.UpdatedAt, beforeCreate)
	})

	t.Run("updatedAt changes on update", func(t *testing.T) {
		reportID := factory.NewReport(owner, frameworkID).WithName("Timestamp Update Test").Create()

		getQuery := `
			query($id: ID!) {
				node(id: $id) {
					... on Report {
						createdAt
						updatedAt
					}
				}
			}
		`

		var getResult struct {
			Node struct {
				CreatedAt time.Time `json:"createdAt"`
				UpdatedAt time.Time `json:"updatedAt"`
			} `json:"node"`
		}

		err := owner.Execute(getQuery, map[string]any{"id": reportID}, &getResult)
		require.NoError(t, err)

		initialCreatedAt := getResult.Node.CreatedAt
		initialUpdatedAt := getResult.Node.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		updateQuery := `
			mutation UpdateReport($input: UpdateReportInput!) {
				updateReport(input: $input) {
					report {
						createdAt
						updatedAt
					}
				}
			}
		`

		var updateResult struct {
			UpdateReport struct {
				Report struct {
					CreatedAt time.Time `json:"createdAt"`
					UpdatedAt time.Time `json:"updatedAt"`
				} `json:"report"`
			} `json:"updateReport"`
		}

		err = owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"id":   reportID,
				"name": "Updated Timestamp Test",
			},
		}, &updateResult)
		require.NoError(t, err)

		report := updateResult.UpdateReport.Report
		testutil.AssertTimestampsOnUpdate(t, report.CreatedAt, report.UpdatedAt, initialCreatedAt, initialUpdatedAt)
	})
}

func TestReport_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Report SubResolvers").Create()
	reportID := factory.NewReport(owner, frameworkID).WithName("SubResolver Test Report").Create()

	t.Run("framework sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Report {
						id
						framework {
							id
							name
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID        string `json:"id"`
				Framework struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"framework"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": reportID}, &result)
		require.NoError(t, err)
		assert.Equal(t, frameworkID, result.Node.Framework.ID)
	})

	t.Run("organization sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Report {
						id
						organization {
							id
							name
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID           string `json:"id"`
				Organization struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"organization"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": reportID}, &result)
		require.NoError(t, err)
		assert.Equal(t, owner.GetOrganizationID().String(), result.Node.Organization.ID)
		assert.NotEmpty(t, result.Node.Organization.Name)
	})
}

func TestReport_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("create", func(t *testing.T) {
		t.Run("owner can create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()

			_, err := owner.Do(`
				mutation CreateReport($input: CreateReportInput!) {
					createReport(input: $input) {
						reportEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": owner.GetOrganizationID().String(),
					"frameworkId":    frameworkID,
					"name":           "RBAC Test Report",
				},
			})
			require.NoError(t, err, "owner should be able to create report")
		})

		t.Run("admin can create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()

			_, err := admin.Do(`
				mutation CreateReport($input: CreateReportInput!) {
					createReport(input: $input) {
						reportEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": admin.GetOrganizationID().String(),
					"frameworkId":    frameworkID,
					"name":           "RBAC Test Report",
				},
			})
			require.NoError(t, err, "admin should be able to create report")
		})

		t.Run("viewer cannot create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()

			_, err := viewer.Do(`
				mutation CreateReport($input: CreateReportInput!) {
					createReport(input: $input) {
						reportEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": viewer.GetOrganizationID().String(),
					"frameworkId":    frameworkID,
					"name":           "RBAC Test Report",
				},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to create report")
		})
	})

	t.Run("update", func(t *testing.T) {
		t.Run("owner can update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			reportID := factory.NewReport(owner, frameworkID).WithName("RBAC Update Test").Create()

			_, err := owner.Do(`
				mutation UpdateReport($input: UpdateReportInput!) {
					updateReport(input: $input) {
						report { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   reportID,
					"name": "Updated by Owner",
				},
			})
			require.NoError(t, err, "owner should be able to update report")
		})

		t.Run("admin can update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			reportID := factory.NewReport(owner, frameworkID).WithName("RBAC Update Test").Create()

			_, err := admin.Do(`
				mutation UpdateReport($input: UpdateReportInput!) {
					updateReport(input: $input) {
						report { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   reportID,
					"name": "Updated by Admin",
				},
			})
			require.NoError(t, err, "admin should be able to update report")
		})

		t.Run("viewer cannot update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			reportID := factory.NewReport(owner, frameworkID).WithName("RBAC Update Test").Create()

			_, err := viewer.Do(`
				mutation UpdateReport($input: UpdateReportInput!) {
					updateReport(input: $input) {
						report { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   reportID,
					"name": "Updated by Viewer",
				},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to update report")
		})
	})

	t.Run("delete", func(t *testing.T) {
		t.Run("owner can delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			reportID := factory.NewReport(owner, frameworkID).WithName("RBAC Delete Test").Create()

			_, err := owner.Do(`
				mutation DeleteReport($input: DeleteReportInput!) {
					deleteReport(input: $input) {
						deletedReportId
					}
				}
			`, map[string]any{
				"input": map[string]any{"reportId": reportID},
			})
			require.NoError(t, err, "owner should be able to delete report")
		})

		t.Run("admin can delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			reportID := factory.NewReport(owner, frameworkID).WithName("RBAC Delete Test").Create()

			_, err := admin.Do(`
				mutation DeleteReport($input: DeleteReportInput!) {
					deleteReport(input: $input) {
						deletedReportId
					}
				}
			`, map[string]any{
				"input": map[string]any{"reportId": reportID},
			})
			require.NoError(t, err, "admin should be able to delete report")
		})

		t.Run("viewer cannot delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			reportID := factory.NewReport(owner, frameworkID).WithName("RBAC Delete Test").Create()

			_, err := viewer.Do(`
				mutation DeleteReport($input: DeleteReportInput!) {
					deleteReport(input: $input) {
						deletedReportId
					}
				}
			`, map[string]any{
				"input": map[string]any{"reportId": reportID},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to delete report")
		})
	})

	t.Run("read", func(t *testing.T) {
		t.Run("owner can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			reportID := factory.NewReport(owner, frameworkID).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := owner.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Report { id name }
					}
				}
			`, map[string]any{"id": reportID}, &result)
			require.NoError(t, err, "owner should be able to read report")
			require.NotNil(t, result.Node, "owner should receive report data")
		})

		t.Run("admin can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			reportID := factory.NewReport(owner, frameworkID).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := admin.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Report { id name }
					}
				}
			`, map[string]any{"id": reportID}, &result)
			require.NoError(t, err, "admin should be able to read report")
			require.NotNil(t, result.Node, "admin should receive report data")
		})

		t.Run("viewer can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			reportID := factory.NewReport(owner, frameworkID).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := viewer.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Report { id name }
					}
				}
			`, map[string]any{"id": reportID}, &result)
			require.NoError(t, err, "viewer should be able to read report")
			require.NotNil(t, result.Node, "viewer should receive report data")
		})
	})
}

func TestReport_MaxLength_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Max Length").Create()
	longName := strings.Repeat("a", 1001)

	t.Run("create", func(t *testing.T) {
		query := `
			mutation CreateReport($input: CreateReportInput!) {
				createReport(input: $input) {
					reportEdge {
						node { id }
					}
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"frameworkId":    frameworkID,
				"name":           longName,
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("update", func(t *testing.T) {
		reportID := factory.NewReport(owner, frameworkID).WithName("Max Length Test").Create()

		query := `
			mutation UpdateReport($input: UpdateReportInput!) {
				updateReport(input: $input) {
					report { id }
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   reportID,
				"name": longName,
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})
}

func TestReport_Pagination(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Pagination").Create()

	for i := 0; i < 5; i++ {
		factory.NewReport(owner, frameworkID).
			WithName(fmt.Sprintf("Pagination Report %d", i)).
			Create()
	}

	t.Run("first/after pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						reports(first: 2) {
							edges {
								node { id name }
								cursor
							}
							pageInfo {
								hasNextPage
								hasPreviousPage
								startCursor
								endCursor
							}
							totalCount
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				Reports struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
						Cursor string `json:"cursor"`
					} `json:"edges"`
					PageInfo   testutil.PageInfo `json:"pageInfo"`
					TotalCount int               `json:"totalCount"`
				} `json:"reports"`
			} `json:"node"`
		}

		err := owner.Execute(
			query,
			map[string]any{
				"id": owner.GetOrganizationID().String(),
			},
			&result,
		)
		require.NoError(t, err)

		testutil.AssertFirstPage(t, len(result.Node.Reports.Edges), result.Node.Reports.PageInfo, 2, true)
		assert.GreaterOrEqual(t, result.Node.Reports.TotalCount, 5)

		testutil.AssertHasMorePages(t, result.Node.Reports.PageInfo)
		queryAfter := `
			query($id: ID!, $after: CursorKey) {
				node(id: $id) {
					... on Organization {
						reports(first: 2, after: $after) {
							edges {
								node { id name }
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

		var resultAfter struct {
			Node struct {
				Reports struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"reports"`
			} `json:"node"`
		}

		err = owner.Execute(queryAfter, map[string]any{
			"id":    owner.GetOrganizationID().String(),
			"after": *result.Node.Reports.PageInfo.EndCursor,
		}, &resultAfter)
		require.NoError(t, err)

		testutil.AssertMiddlePage(t, len(resultAfter.Node.Reports.Edges), resultAfter.Node.Reports.PageInfo, 2)
	})

	t.Run("last/before pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						reports(last: 2) {
							edges {
								node { id name }
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
				Reports struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"reports"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		testutil.AssertLastPage(t, len(result.Node.Reports.Edges), result.Node.Reports.PageInfo, 2, true)
	})
}

func TestReport_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(org1Owner).WithName("Org1 Framework").Create()
	reportID := factory.NewReport(org1Owner, frameworkID).WithName("Org1 Report").Create()

	t.Run("cannot read report from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Report {
						id
						name
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": reportID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "report")
	})

	t.Run("cannot update report from another organization", func(t *testing.T) {
		query := `
			mutation UpdateReport($input: UpdateReportInput!) {
				updateReport(input: $input) {
					report { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   reportID,
				"name": "Hijacked Report",
			},
		})
		require.Error(t, err, "Should not be able to update report from another org")
	})

	t.Run("cannot delete report from another organization", func(t *testing.T) {
		query := `
			mutation DeleteReport($input: DeleteReportInput!) {
				deleteReport(input: $input) {
					deletedReportId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
			},
		})
		require.Error(t, err, "Should not be able to delete report from another org")
	})

	t.Run("cannot list reports from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						reports(first: 100) {
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

		var result struct {
			Node struct {
				Reports struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"reports"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{
			"id": org1Owner.GetOrganizationID().String(),
		}, &result)

		if err == nil {
			for _, edge := range result.Node.Reports.Edges {
				assert.NotEqual(t, reportID, edge.Node.ID, "Should not see report from another org")
			}
		}
	})
}

func TestReport_Ordering(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Ordering").Create()

	factory.NewReport(owner, frameworkID).WithName("AAA Order Test").Create()
	factory.NewReport(owner, frameworkID).WithName("ZZZ Order Test").Create()

	t.Run("order by created_at descending", func(t *testing.T) {
		query := `
			query($id: ID!, $orderBy: ReportOrder) {
				node(id: $id) {
					... on Organization {
						reports(first: 100, orderBy: $orderBy) {
							edges {
								node {
									id
									createdAt
								}
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				Reports struct {
					Edges []struct {
						Node struct {
							ID        string    `json:"id"`
							CreatedAt time.Time `json:"createdAt"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"reports"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
			"orderBy": map[string]any{
				"field":     "CREATED_AT",
				"direction": "DESC",
			},
		}, &result)
		require.NoError(t, err)

		times := make([]time.Time, len(result.Node.Reports.Edges))
		for i, edge := range result.Node.Reports.Edges {
			times[i] = edge.Node.CreatedAt
		}
		testutil.AssertTimesOrderedDescending(t, times, "createdAt")
	})
}

func TestReport_UploadReport(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Upload").Create()

	t.Run("upload valid PDF report", func(t *testing.T) {
		reportID := factory.NewReport(owner, frameworkID).WithName("Upload Test").Create()

		query := `
			mutation UploadReportFile($input: UploadReportFileInput!) {
				uploadReportFile(input: $input) {
					report {
						id
						file {
							id
							fileName
							size
						}
					}
				}
			}
		`

		// Create a minimal valid PDF content
		pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

		var result struct {
			UploadReportFile struct {
				Report struct {
					ID   string `json:"id"`
					File *struct {
						ID       string `json:"id"`
						FileName string `json:"fileName"`
						Size     int    `json:"size"`
					} `json:"file"`
				} `json:"report"`
			} `json:"uploadReportFile"`
		}

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
				"file":     nil, // Will be replaced by the file
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report-file.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, reportID, result.UploadReportFile.Report.ID)
		require.NotNil(t, result.UploadReportFile.Report.File)
		assert.Equal(t, "report-file.pdf", result.UploadReportFile.Report.File.FileName)
		assert.Equal(t, len(pdfContent), result.UploadReportFile.Report.File.Size)
	})

	t.Run("upload replaces existing report", func(t *testing.T) {
		reportID := factory.NewReport(owner, frameworkID).WithName("Replace Report Test").Create()

		query := `
			mutation UploadReportFile($input: UploadReportFileInput!) {
				uploadReportFile(input: $input) {
					report {
						id
						file {
							id
							fileName
						}
					}
				}
			}
		`

		pdfContent1 := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

		var result1 struct {
			UploadReportFile struct {
				Report struct {
					ID   string `json:"id"`
					File *struct {
						ID       string `json:"id"`
						FileName string `json:"fileName"`
					} `json:"file"`
				} `json:"report"`
			} `json:"uploadReportFile"`
		}

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
				"file":     nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "first-report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent1,
		}, &result1)
		require.NoError(t, err)
		firstFileID := result1.UploadReportFile.Report.File.ID

		pdfContent2 := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Version /1.4 >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

		var result2 struct {
			UploadReportFile struct {
				Report struct {
					ID   string `json:"id"`
					File *struct {
						ID       string `json:"id"`
						FileName string `json:"fileName"`
					} `json:"file"`
				} `json:"report"`
			} `json:"uploadReportFile"`
		}

		err = owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
				"file":     nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "second-report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent2,
		}, &result2)
		require.NoError(t, err)

		assert.Equal(t, "second-report.pdf", result2.UploadReportFile.Report.File.FileName)
		assert.NotEqual(t, firstFileID, result2.UploadReportFile.Report.File.ID, "File ID should change when replaced")
	})
}

func TestReport_UploadReport_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Upload Validation").Create()

	t.Run("reject non document file", func(t *testing.T) {
		reportID := factory.NewReport(owner, frameworkID).WithName("Invalid File Test").Create()

		query := `
			mutation UploadReportFile($input: UploadReportFileInput!) {
				uploadReportFile(input: $input) {
					report {
						id
					}
				}
			}
		`

		// Try to upload a text file
		textContent := []byte("This is not a document file")

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
				"file":     nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "not-a-document.txt",
			ContentType: "text/plain",
			Content:     textContent,
		}, nil)
		require.Error(t, err, "Should reject non-document file")
	})

	t.Run("reject file with wrong extension but document content-type", func(t *testing.T) {
		reportID := factory.NewReport(owner, frameworkID).WithName("Wrong Extension Test").Create()

		query := `
			mutation UploadReportFile($input: UploadReportFileInput!) {
				uploadReportFile(input: $input) {
					report {
						id
					}
				}
			}
		`

		// Try to upload with wrong extension
		textContent := []byte("Not a real document")

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
				"file":     nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "fake.exe",
			ContentType: "application/pdf",
			Content:     textContent,
		}, nil)
		require.Error(t, err, "Should reject file with wrong extension")
	})

	t.Run("reject empty file", func(t *testing.T) {
		reportID := factory.NewReport(owner, frameworkID).WithName("Empty File Test").Create()

		query := `
			mutation UploadReportFile($input: UploadReportFileInput!) {
				uploadReportFile(input: $input) {
					report {
						id
					}
				}
			}
		`

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
				"file":     nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "empty.pdf",
			ContentType: "application/pdf",
			Content:     []byte{},
		}, nil)
		require.Error(t, err, "Should reject empty file")
	})

	t.Run("reject invalid report ID", func(t *testing.T) {
		query := `
			mutation UploadReportFile($input: UploadReportFileInput!) {
				uploadReportFile(input: $input) {
					report {
						id
					}
				}
			}
		`

		pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"reportId": "invalid-id",
				"file":     nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, nil)
		require.Error(t, err, "Should reject invalid report ID")
	})
}

func TestReport_UploadReport_RBAC(t *testing.T) {
	t.Parallel()

	query := `
		mutation UploadReportFile($input: UploadReportFileInput!) {
			uploadReportFile(input: $input) {
				report {
					id
				}
			}
		}
	`

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	t.Run("owner can upload", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
		reportID := factory.NewReport(owner, frameworkID).WithName("RBAC Upload Test").Create()

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
				"file":     nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, nil)
		require.NoError(t, err, "owner should be able to upload report")
	})

	t.Run("admin can upload", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
		frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
		reportID := factory.NewReport(owner, frameworkID).WithName("RBAC Upload Test").Create()

		err := admin.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
				"file":     nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, nil)
		require.NoError(t, err, "admin should be able to upload report")
	})

	t.Run("viewer cannot upload", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
		reportID := factory.NewReport(owner, frameworkID).WithName("RBAC Upload Test").Create()

		err := viewer.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
				"file":     nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, nil)
		testutil.RequireForbiddenError(t, err, "viewer should not be able to upload report")
	})
}

func TestReport_DeleteReport(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Delete Report").Create()

	t.Run("delete existing report", func(t *testing.T) {
		reportID := factory.NewReport(owner, frameworkID).WithName("Delete Report Test").Create()

		// First upload a report
		uploadQuery := `
			mutation UploadReportFile($input: UploadReportFileInput!) {
				uploadReportFile(input: $input) {
					report {
						id
						file {
							id
						}
					}
				}
			}
		`

		pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

		err := owner.ExecuteWithFile(uploadQuery, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
				"file":     nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, nil)
		require.NoError(t, err)

		deleteQuery := `
			mutation DeleteReportFile($input: DeleteReportFileInput!) {
				deleteReportFile(input: $input) {
					report {
						id
						file {
							id
						}
					}
				}
			}
		`

		var deleteResult struct {
			DeleteReportFile struct {
				Report struct {
					ID   string `json:"id"`
					File *struct {
						ID string `json:"id"`
					} `json:"file"`
				} `json:"report"`
			} `json:"deleteReportFile"`
		}

		err = owner.Execute(deleteQuery, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
			},
		}, &deleteResult)
		require.NoError(t, err)
		assert.Equal(t, reportID, deleteResult.DeleteReportFile.Report.ID)
		assert.Nil(t, deleteResult.DeleteReportFile.Report.File, "File should be nil after deletion")
	})
}

func TestReport_DeleteReport_RBAC(t *testing.T) {
	t.Parallel()

	uploadQuery := `
		mutation UploadReportFile($input: UploadReportFileInput!) {
			uploadReportFile(input: $input) {
				report { id }
			}
		}
	`

	deleteQuery := `
		mutation DeleteReportFile($input: DeleteReportFileInput!) {
			deleteReportFile(input: $input) {
				report { id }
			}
		}
	`

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	t.Run("viewer cannot delete report", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
		reportID := factory.NewReport(owner, frameworkID).WithName("RBAC Delete Report Test").Create()

		// Owner uploads the report
		err := owner.ExecuteWithFile(uploadQuery, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
				"file":     nil,
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
				"reportId": reportID,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to delete report")
	})
}
