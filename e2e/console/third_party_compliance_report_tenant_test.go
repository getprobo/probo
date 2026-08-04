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
