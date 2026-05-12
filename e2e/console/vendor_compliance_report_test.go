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

func TestVendorComplianceReport_Upload(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	vendorID := factory.NewVendor(owner).WithName("Compliance Report Upload Vendor").Create()

	const query = `
		mutation UploadVendorComplianceReport($input: UploadVendorComplianceReportInput!) {
			uploadVendorComplianceReport(input: $input) {
				vendorComplianceReportEdge {
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
		UploadVendorComplianceReport struct {
			VendorComplianceReportEdge struct {
				Node struct {
					ID         string  `json:"id"`
					ReportName string  `json:"reportName"`
					ReportDate string  `json:"reportDate"`
					ValidUntil *string `json:"validUntil"`
				} `json:"node"`
			} `json:"vendorComplianceReportEdge"`
		} `json:"uploadVendorComplianceReport"`
	}

	err := owner.ExecuteWithFile(
		query,
		map[string]any{
			"input": map[string]any{
				"vendorId":   vendorID,
				"reportName": "SOC 2 Type II",
				"reportDate": "2024-01-01T00:00:00Z",
				"validUntil": "2025-01-01T00:00:00Z",
				"file":       nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "soc2-report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		},
		&result,
	)
	require.NoError(t, err)

	node := result.UploadVendorComplianceReport.VendorComplianceReportEdge.Node
	assert.NotEmpty(t, node.ID)
	assert.Equal(t, "SOC 2 Type II", node.ReportName)
	assert.NotEmpty(t, node.ReportDate)
	assert.NotNil(t, node.ValidUntil)
}
func TestVendorComplianceReport_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	vendorID := factory.NewVendor(owner).WithName("Compliance Report List Vendor").Create()

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	uploadQuery := `
		mutation UploadVendorComplianceReport($input: UploadVendorComplianceReportInput!) {
			uploadVendorComplianceReport(input: $input) {
				vendorComplianceReportEdge {
					node { id }
				}
			}
		}
	`

	var uploadResult struct {
		UploadVendorComplianceReport struct {
			VendorComplianceReportEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"vendorComplianceReportEdge"`
		} `json:"uploadVendorComplianceReport"`
	}

	err := owner.ExecuteWithFile(
		uploadQuery,
		map[string]any{
			"input": map[string]any{
				"vendorId":   vendorID,
				"reportName": "ISO 27001",
				"reportDate": "2024-06-01T00:00:00Z",
				"file":       nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "iso27001.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		},
		&uploadResult,
	)
	require.NoError(t, err)

	reportID := uploadResult.UploadVendorComplianceReport.VendorComplianceReportEdge.Node.ID
	require.NotEmpty(t, reportID)

	const listQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Vendor {
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

	err = owner.Execute(listQuery, map[string]any{"id": vendorID}, &listResult)
	require.NoError(t, err)

	require.Len(t, listResult.Node.ComplianceReports.Edges, 1)
	assert.Equal(t, reportID, listResult.Node.ComplianceReports.Edges[0].Node.ID)
	assert.Equal(t, "ISO 27001", listResult.Node.ComplianceReports.Edges[0].Node.ReportName)
}

func TestVendorComplianceReport_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	vendorID := factory.NewVendor(owner).WithName("Compliance Report Delete Vendor").Create()

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	uploadQuery := `
		mutation UploadVendorComplianceReport($input: UploadVendorComplianceReportInput!) {
			uploadVendorComplianceReport(input: $input) {
				vendorComplianceReportEdge {
					node { id }
				}
			}
		}
	`

	var uploadResult struct {
		UploadVendorComplianceReport struct {
			VendorComplianceReportEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"vendorComplianceReportEdge"`
		} `json:"uploadVendorComplianceReport"`
	}

	err := owner.ExecuteWithFile(uploadQuery, map[string]any{
		"input": map[string]any{
			"vendorId":   vendorID,
			"reportName": "PCI DSS",
			"reportDate": "2024-03-01T00:00:00Z",
			"file":       nil,
		},
	}, "input.file", testutil.UploadFile{
		Filename:    "pci-dss.pdf",
		ContentType: "application/pdf",
		Content:     pdfContent,
	}, &uploadResult)
	require.NoError(t, err)

	reportID := uploadResult.UploadVendorComplianceReport.VendorComplianceReportEdge.Node.ID
	require.NotEmpty(t, reportID)

	const deleteQuery = `
		mutation DeleteVendorComplianceReport($input: DeleteVendorComplianceReportInput!) {
			deleteVendorComplianceReport(input: $input) {
				deletedVendorComplianceReportId
			}
		}
	`

	var deleteResult struct {
		DeleteVendorComplianceReport struct {
			DeletedVendorComplianceReportID string `json:"deletedVendorComplianceReportId"`
		} `json:"deleteVendorComplianceReport"`
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
	assert.Equal(t, reportID, deleteResult.DeleteVendorComplianceReport.DeletedVendorComplianceReportID)
}
