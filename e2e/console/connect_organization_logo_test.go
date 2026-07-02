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

func TestConnectOrganization_LogoUpload(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()

	const uploadMutation = `
		mutation UpdateOrganization($input: UpdateOrganizationInput!) {
			updateOrganization(input: $input) {
				organization {
					id
					logo {
						id
						fileName
						downloadUrl
					}
				}
			}
		}
	`

	pngContent := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}

	var uploadResult struct {
		UpdateOrganization struct {
			Organization struct {
				ID   string `json:"id"`
				Logo *struct {
					ID          string `json:"id"`
					FileName    string `json:"fileName"`
					DownloadURL string `json:"downloadUrl"`
				} `json:"logo"`
			} `json:"organization"`
		} `json:"updateOrganization"`
	}

	err := owner.ExecuteConnectWithFile(uploadMutation, map[string]any{
		"input": map[string]any{
			"organizationId": organizationID,
			"logoFile":       nil,
		},
	}, "input.logoFile", testutil.UploadFile{
		Filename:    "org-logo.png",
		ContentType: "image/png",
		Content:     pngContent,
	}, &uploadResult)
	require.NoError(t, err)

	assert.Equal(t, organizationID, uploadResult.UpdateOrganization.Organization.ID)
	require.NotNil(t, uploadResult.UpdateOrganization.Organization.Logo)
	assert.Equal(t, "org-logo.png", uploadResult.UpdateOrganization.Organization.Logo.FileName)
	assert.NotEmpty(t, uploadResult.UpdateOrganization.Organization.Logo.DownloadURL)
	assert.True(
		t,
		strings.Contains(uploadResult.UpdateOrganization.Organization.Logo.DownloadURL, "/api/files/v1/public/"),
		"downloadUrl must route through the public files API, got %q",
		uploadResult.UpdateOrganization.Organization.Logo.DownloadURL,
	)

	const queryOrganization = `
		query GetOrganization($id: ID!) {
			node(id: $id) {
				... on Organization {
					logo {
						downloadUrl
						fileName
						mimeType
						size
					}
				}
			}
		}
	`

	var queryResult struct {
		Node struct {
			Logo *struct {
				DownloadURL string `json:"downloadUrl"`
				FileName    string `json:"fileName"`
				MimeType    string `json:"mimeType"`
				Size        int64  `json:"size"`
			} `json:"logo"`
		} `json:"node"`
	}

	err = owner.ExecuteConnect(queryOrganization, map[string]any{
		"id": organizationID,
	}, &queryResult)
	require.NoError(t, err)

	require.NotNil(t, queryResult.Node.Logo)
	assert.Equal(t, "org-logo.png", queryResult.Node.Logo.FileName)
	assert.Equal(t, "image/png", queryResult.Node.Logo.MimeType)
	assert.Equal(t, int64(len(pngContent)), queryResult.Node.Logo.Size)
	assert.True(
		t,
		strings.Contains(queryResult.Node.Logo.DownloadURL, "/api/files/v1/public/"),
		"downloadUrl must route through the public files API, got %q",
		queryResult.Node.Logo.DownloadURL,
	)
}
