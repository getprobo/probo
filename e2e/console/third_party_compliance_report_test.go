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

func TestThirdPartyComplianceReport_Upload(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	thirdPartyID := factory.NewThirdParty(owner).WithName("Compliance Report Upload ThirdParty").Create()

	const query = `
		mutation UploadThirdPartyComplianceReport($input: UploadThirdPartyComplianceReportInput!) {
			uploadThirdPartyComplianceReport(input: $input) {
				thirdPartyComplianceReportEdge {
					node {
						id
						reportName
						reportDate
						validUntil
					}
				}
			}
		}
	`

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	var result struct {
		UploadThirdPartyComplianceReport struct {
			ThirdPartyComplianceReportEdge struct {
				Node struct {
					ID         string  `json:"id"`
					ReportName string  `json:"reportName"`
					ReportDate string  `json:"reportDate"`
					ValidUntil *string `json:"validUntil"`
				} `json:"node"`
			} `json:"thirdPartyComplianceReportEdge"`
		} `json:"uploadThirdPartyComplianceReport"`
	}

	err := owner.ExecuteWithFile(
		query,
		map[string]any{
			"input": map[string]any{
				"thirdPartyId": thirdPartyID,
				"reportName":   "SOC 2 Type II",
				"reportDate":   "2024-01-01T00:00:00Z",
				"validUntil":   "2025-01-01T00:00:00Z",
				"file":         nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "soc2-report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		},
		&result,
	)
	require.NoError(t, err)

	node := result.UploadThirdPartyComplianceReport.ThirdPartyComplianceReportEdge.Node
	assert.NotEmpty(t, node.ID)
	assert.Equal(t, "SOC 2 Type II", node.ReportName)
	assert.NotEmpty(t, node.ReportDate)
	assert.NotNil(t, node.ValidUntil)
}
func TestThirdPartyComplianceReport_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	thirdPartyID := factory.NewThirdParty(owner).WithName("Compliance Report List ThirdParty").Create()

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	uploadQuery := `
		mutation UploadThirdPartyComplianceReport($input: UploadThirdPartyComplianceReportInput!) {
			uploadThirdPartyComplianceReport(input: $input) {
				thirdPartyComplianceReportEdge {
					node { id }
				}
			}
		}
	`

	var uploadResult struct {
		UploadThirdPartyComplianceReport struct {
			ThirdPartyComplianceReportEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyComplianceReportEdge"`
		} `json:"uploadThirdPartyComplianceReport"`
	}

	err := owner.ExecuteWithFile(
		uploadQuery,
		map[string]any{
			"input": map[string]any{
				"thirdPartyId": thirdPartyID,
				"reportName":   "ISO 27001",
				"reportDate":   "2024-06-01T00:00:00Z",
				"file":         nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "iso27001.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		},
		&uploadResult,
	)
	require.NoError(t, err)

	reportID := uploadResult.UploadThirdPartyComplianceReport.ThirdPartyComplianceReportEdge.Node.ID
	require.NotEmpty(t, reportID)

	const listQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on ThirdParty {
					id
					complianceReports(first: 10) {
						edges {
							node {
								id
								reportName
								reportDate
							}
						}
					}
				}
			}
		}
	`

	var listResult struct {
		Node struct {
			ID                string `json:"id"`
			ComplianceReports struct {
				Edges []struct {
					Node struct {
						ID         string `json:"id"`
						ReportName string `json:"reportName"`
						ReportDate string `json:"reportDate"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"complianceReports"`
		} `json:"node"`
	}

	err = owner.Execute(listQuery, map[string]any{"id": thirdPartyID}, &listResult)
	require.NoError(t, err)

	require.Len(t, listResult.Node.ComplianceReports.Edges, 1)
	assert.Equal(t, reportID, listResult.Node.ComplianceReports.Edges[0].Node.ID)
	assert.Equal(t, "ISO 27001", listResult.Node.ComplianceReports.Edges[0].Node.ReportName)
}

func TestThirdPartyComplianceReport_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	thirdPartyID := factory.NewThirdParty(owner).WithName("Compliance Report Delete ThirdParty").Create()

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	uploadQuery := `
		mutation UploadThirdPartyComplianceReport($input: UploadThirdPartyComplianceReportInput!) {
			uploadThirdPartyComplianceReport(input: $input) {
				thirdPartyComplianceReportEdge {
					node { id }
				}
			}
		}
	`

	var uploadResult struct {
		UploadThirdPartyComplianceReport struct {
			ThirdPartyComplianceReportEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyComplianceReportEdge"`
		} `json:"uploadThirdPartyComplianceReport"`
	}

	err := owner.ExecuteWithFile(uploadQuery, map[string]any{
		"input": map[string]any{
			"thirdPartyId": thirdPartyID,
			"reportName":   "PCI DSS",
			"reportDate":   "2024-03-01T00:00:00Z",
			"file":         nil,
		},
	}, "input.file", testutil.UploadFile{
		Filename:    "pci-dss.pdf",
		ContentType: "application/pdf",
		Content:     pdfContent,
	}, &uploadResult)
	require.NoError(t, err)

	reportID := uploadResult.UploadThirdPartyComplianceReport.ThirdPartyComplianceReportEdge.Node.ID
	require.NotEmpty(t, reportID)

	const deleteQuery = `
		mutation DeleteThirdPartyComplianceReport($input: DeleteThirdPartyComplianceReportInput!) {
			deleteThirdPartyComplianceReport(input: $input) {
				deletedThirdPartyComplianceReportId
			}
		}
	`

	var deleteResult struct {
		DeleteThirdPartyComplianceReport struct {
			DeletedThirdPartyComplianceReportID string `json:"deletedThirdPartyComplianceReportId"`
		} `json:"deleteThirdPartyComplianceReport"`
	}

	err = owner.Execute(
		deleteQuery,
		map[string]any{
			"input": map[string]any{
				"reportId": reportID,
			},
		},
		&deleteResult,
	)
	require.NoError(t, err)
	assert.Equal(t, reportID, deleteResult.DeleteThirdPartyComplianceReport.DeletedThirdPartyComplianceReportID)
}

func TestThirdPartyComplianceReport_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1ThirdPartyID := factory.NewThirdParty(org1Owner).WithName("Org1 ThirdParty for Report").Create()

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	var createResult struct {
		UploadThirdPartyComplianceReport struct {
			ThirdPartyComplianceReportEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyComplianceReportEdge"`
		} `json:"uploadThirdPartyComplianceReport"`
	}

	err := org1Owner.ExecuteWithFile(
		`
			mutation($input: UploadThirdPartyComplianceReportInput!) {
				uploadThirdPartyComplianceReport(input: $input) {
					thirdPartyComplianceReportEdge { node { id } }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"thirdPartyId": org1ThirdPartyID,
				"reportName":   "Org1 Report",
				"reportDate":   "2024-01-01T00:00:00Z",
				"file":         nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		},
		&createResult,
	)
	require.NoError(t, err)

	reportID := createResult.UploadThirdPartyComplianceReport.ThirdPartyComplianceReportEdge.Node.ID

	t.Run("cannot delete thirdPartyComplianceReport from another organization", func(t *testing.T) {
		_, err := org2Owner.Do(`
			mutation($input: DeleteThirdPartyComplianceReportInput!) {
				deleteThirdPartyComplianceReport(input: $input) {
					deletedThirdPartyComplianceReportId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"reportId": reportID,
			},
		})
		require.Error(t, err, "Should not be able to delete thirdPartyComplianceReport from another org")
	})

	t.Run("cannot upload thirdPartyComplianceReport on a thirdParty from another organization", func(t *testing.T) {
		var result struct{}

		err := org2Owner.ExecuteWithFile(
			`
				mutation($input: UploadThirdPartyComplianceReportInput!) {
					uploadThirdPartyComplianceReport(input: $input) {
						thirdPartyComplianceReportEdge { node { id } }
					}
				}
			`,
			map[string]any{
				"input": map[string]any{
					"thirdPartyId": org1ThirdPartyID,
					"reportName":   "Attacker Report",
					"reportDate":   "2024-01-01T00:00:00Z",
					"file":         nil,
				},
			}, "input.file", testutil.UploadFile{
				Filename:    "attacker.pdf",
				ContentType: "application/pdf",
				Content:     pdfContent,
			},
			&result,
		)
		require.Error(t, err, "must not accept a thirdPartyId belonging to another organization")
	})
}
