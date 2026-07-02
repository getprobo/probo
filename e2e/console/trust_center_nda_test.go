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

func TestTrustCenter_UploadNDA(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()

	const trustCenterQuery = `
		query($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					trustCenter {
						id
					}
				}
			}
		}
	`

	var trustCenterLookup struct {
		Node struct {
			TrustCenter struct {
				ID string `json:"id"`
			} `json:"trustCenter"`
		} `json:"node"`
	}

	err := owner.Execute(trustCenterQuery, map[string]any{
		"organizationId": organizationID,
	}, &trustCenterLookup)
	require.NoError(t, err)
	require.NotEmpty(t, trustCenterLookup.Node.TrustCenter.ID)

	trustCenterID := trustCenterLookup.Node.TrustCenter.ID

	const uploadMutation = `
		mutation UploadTrustCenterNDA($input: UploadTrustCenterNDAInput!) {
			uploadTrustCenterNDA(input: $input) {
				trustCenter {
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
		UploadTrustCenterNDA struct {
			TrustCenter struct {
				ID  string `json:"id"`
				Nda *struct {
					ID          string `json:"id"`
					FileName    string `json:"fileName"`
					DownloadURL string `json:"downloadUrl"`
				} `json:"nda"`
			} `json:"trustCenter"`
		} `json:"uploadTrustCenterNDA"`
	}

	err = owner.ExecuteWithFile(uploadMutation, map[string]any{
		"input": map[string]any{
			"trustCenterId": trustCenterID,
			"fileName":      "nda.pdf",
			"file":          nil,
		},
	}, "input.file", testutil.UploadFile{
		Filename:    "nda.pdf",
		ContentType: "application/pdf",
		Content:     pdfContent,
	}, &uploadResult)
	require.NoError(t, err)

	assert.Equal(t, trustCenterID, uploadResult.UploadTrustCenterNDA.TrustCenter.ID)
	require.NotNil(t, uploadResult.UploadTrustCenterNDA.TrustCenter.Nda)
	assert.Equal(t, "nda.pdf", uploadResult.UploadTrustCenterNDA.TrustCenter.Nda.FileName)
	assert.NotEmpty(t, uploadResult.UploadTrustCenterNDA.TrustCenter.Nda.DownloadURL)
	assert.True(
		t,
		strings.Contains(uploadResult.UploadTrustCenterNDA.TrustCenter.Nda.DownloadURL, "/api/files/v1/"),
		"downloadUrl must route through the files API, got %q",
		uploadResult.UploadTrustCenterNDA.TrustCenter.Nda.DownloadURL,
	)
}
