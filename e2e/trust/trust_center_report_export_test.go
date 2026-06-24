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

package trust_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const exportReportPDFMutation = `
	mutation ExportReportPDF($input: ExportReportPDFInput!) {
		exportReportPDF(input: $input) {
			data
		}
	}
`

const nodeQuery = `
	query Node($id: ID!) {
		node(id: $id) {
			__typename
		}
	}
`

// TestTrustCenter_ExportReportPDF_TenantIsolation verifies that a public
// audit-report PDF can only be exported through its own organization's trust
// center. A visitor on another organization's trust center must not be able to
// download it by supplying the foreign report GID (cross-tenant IDOR).
func TestTrustCenter_ExportReportPDF_TenantIsolation(t *testing.T) {
	t.Parallel()

	victimOwner := testutil.NewClient(t, testutil.RoleOwner)
	attackerOwner := testutil.NewClient(t, testutil.RoleOwner)

	victimTrustCenterID, victimReportID := setupPublicAuditReport(t, victimOwner)
	attackerTrustCenterID, _ := setupPublicAuditReport(t, attackerOwner)

	t.Run("owning trust center can export its report", func(t *testing.T) {
		var result struct {
			ExportReportPDF struct {
				Data string `json:"data"`
			} `json:"exportReportPDF"`
		}

		err := victimOwner.ExecuteTrust(victimTrustCenterID, exportReportPDFMutation, map[string]any{
			"input": map[string]any{"reportId": victimReportID},
		}, &result)
		require.NoError(t, err, "the owning trust center must serve its own public report")
		assert.True(
			t,
			strings.HasPrefix(result.ExportReportPDF.Data, "data:application/pdf;base64,"),
			"expected a base64 PDF data URL, got %q",
			result.ExportReportPDF.Data,
		)
	})

	t.Run("foreign trust center cannot export another org's report", func(t *testing.T) {
		err := attackerOwner.ExecuteTrust(attackerTrustCenterID, exportReportPDFMutation, map[string]any{
			"input": map[string]any{"reportId": victimReportID},
		}, nil)
		require.Error(t, err, "a foreign trust center must not export another org's report")
		assert.Contains(
			t,
			err.Error(),
			"not found",
			"cross-tenant report GID must be rejected as not found",
		)
	})
}

// TestTrustCenter_Node_TenantIsolation exercises the generic node(id:) resolver:
// a visitor on one organization's trust center must not resolve a node that
// belongs to another organization, even with a valid foreign GID.
func TestTrustCenter_Node_TenantIsolation(t *testing.T) {
	t.Parallel()

	victimOwner := testutil.NewClient(t, testutil.RoleOwner)
	attackerOwner := testutil.NewClient(t, testutil.RoleOwner)

	victimTrustCenterID, _ := setupPublicAuditReport(t, victimOwner)
	attackerTrustCenterID, _ := setupPublicAuditReport(t, attackerOwner)

	t.Run("owning trust center resolves its own node", func(t *testing.T) {
		var result struct {
			Node struct {
				Typename string `json:"__typename"`
			} `json:"node"`
		}

		err := victimOwner.ExecuteTrust(victimTrustCenterID, nodeQuery, map[string]any{
			"id": victimTrustCenterID,
		}, &result)
		require.NoError(t, err, "the owning trust center must resolve its own node")
		assert.NotEmpty(t, result.Node.Typename, "expected the node to resolve to a concrete type")
	})

	t.Run("foreign trust center cannot resolve another org's node", func(t *testing.T) {
		err := attackerOwner.ExecuteTrust(attackerTrustCenterID, nodeQuery, map[string]any{
			"id": victimTrustCenterID,
		}, nil)
		require.Error(t, err, "a foreign trust center must not resolve another org's node")
		assert.Contains(
			t,
			err.Error(),
			"not found",
			"cross-tenant GID must be rejected as not found",
		)
	})
}

// setupPublicAuditReport creates an audit with an uploaded report file, marks it
// as publicly visible on the trust center, activates the trust center, and
// returns the trust center ID and the report file ID.
func setupPublicAuditReport(t *testing.T, owner *testutil.Client) (trustCenterID string, reportID string) {
	t.Helper()

	frameworkID := factory.NewFramework(owner).WithName(factory.SafeName("Framework")).Create()
	auditID := factory.NewAudit(owner, frameworkID).WithName(factory.SafeName("Audit")).Create()

	const uploadMutation = `
		mutation UploadAuditReport($input: UploadAuditReportInput!) {
			uploadAuditReport(input: $input) {
				audit {
					reportFile { id }
				}
			}
		}
	`

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	var uploadResult struct {
		UploadAuditReport struct {
			Audit struct {
				ReportFile struct {
					ID string `json:"id"`
				} `json:"reportFile"`
			} `json:"audit"`
		} `json:"uploadAuditReport"`
	}

	err := owner.ExecuteWithFile(uploadMutation, map[string]any{
		"input": map[string]any{
			"auditId": auditID,
			"file":    nil,
		},
	}, "input.file", testutil.UploadFile{
		Filename:    "audit-report.pdf",
		ContentType: "application/pdf",
		Content:     pdfContent,
	}, &uploadResult)
	require.NoError(t, err)

	reportID = uploadResult.UploadAuditReport.Audit.ReportFile.ID
	require.NotEmpty(t, reportID)

	const setVisibilityMutation = `
		mutation UpdateAudit($input: UpdateAuditInput!) {
			updateAudit(input: $input) {
				audit { id }
			}
		}
	`

	err = owner.Execute(setVisibilityMutation, map[string]any{
		"input": map[string]any{
			"id":                    auditID,
			"trustCenterVisibility": "PUBLIC",
		},
	}, nil)
	require.NoError(t, err)

	trustCenterID = lookupTrustCenterID(t, owner)
	activateTrustCenter(t, owner, trustCenterID)

	return trustCenterID, reportID
}
