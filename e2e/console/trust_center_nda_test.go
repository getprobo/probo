// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
