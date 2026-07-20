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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestCompliancePortal_UploadNDA(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()

	const compliancePortalQuery = `
		query($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					compliancePortal {
						id
					}
				}
			}
		}
	`

	var compliancePortalLookup struct {
		Node struct {
			CompliancePortal struct {
				ID string `json:"id"`
			} `json:"compliancePortal"`
		} `json:"node"`
	}

	err := owner.Execute(compliancePortalQuery, map[string]any{
		"organizationId": organizationID,
	}, &compliancePortalLookup)
	require.NoError(t, err)
	require.NotEmpty(t, compliancePortalLookup.Node.CompliancePortal.ID)

	compliancePortalID := compliancePortalLookup.Node.CompliancePortal.ID

	const uploadMutation = `
		mutation UploadCompliancePortalNDA($input: UploadCompliancePortalNDAInput!) {
			uploadCompliancePortalNDA(input: $input) {
				compliancePortal {
					id
					nda {
						id
						fileName
						downloadUrl
					}
				}
			}
		}
	`

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	var uploadResult struct {
		UploadCompliancePortalNDA struct {
			CompliancePortal struct {
				ID  string `json:"id"`
				Nda *struct {
					ID          string `json:"id"`
					FileName    string `json:"fileName"`
					DownloadURL string `json:"downloadUrl"`
				} `json:"nda"`
			} `json:"compliancePortal"`
		} `json:"uploadCompliancePortalNDA"`
	}

	err = owner.ExecuteWithFile(uploadMutation, map[string]any{
		"input": map[string]any{
			"compliancePortalId": compliancePortalID,
			"fileName":           "nda.pdf",
			"file":               nil,
		},
	}, "input.file", testutil.UploadFile{
		Filename:    "nda.pdf",
		ContentType: "application/pdf",
		Content:     pdfContent,
	}, &uploadResult)
	require.NoError(t, err)

	assert.Equal(t, compliancePortalID, uploadResult.UploadCompliancePortalNDA.CompliancePortal.ID)
	require.NotNil(t, uploadResult.UploadCompliancePortalNDA.CompliancePortal.Nda)
	assert.Equal(t, "nda.pdf", uploadResult.UploadCompliancePortalNDA.CompliancePortal.Nda.FileName)
	assert.NotEmpty(t, uploadResult.UploadCompliancePortalNDA.CompliancePortal.Nda.DownloadURL)
	assert.True(
		t,
		strings.Contains(uploadResult.UploadCompliancePortalNDA.CompliancePortal.Nda.DownloadURL, "/api/files/v1/"),
		"downloadUrl must route through the files API, got %q",
		uploadResult.UploadCompliancePortalNDA.CompliancePortal.Nda.DownloadURL,
	)
}
