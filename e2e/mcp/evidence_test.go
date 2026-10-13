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

package mcp_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_InternalControlEvidence_FileURL(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	internalControlID := factory.NewInternalControl(owner).
		WithName(factory.SafeName("MCP evidence download")).
		WithCategory("EVIDENCE").
		Create()
	fileContent := []byte("%PDF-1.4\nMCP evidence download\n%%EOF")

	var uploadResult struct {
		UploadInternalControlEvidence struct {
			EvidenceEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"evidenceEdge"`
		} `json:"uploadInternalControlEvidence"`
	}

	err := owner.ExecuteWithFile(
		`
			mutation UploadInternalControlEvidence($input: UploadInternalControlEvidenceInput!) {
				uploadInternalControlEvidence(input: $input) {
					evidenceEdge {
						node {
							id
						}
					}
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"internalControlId": internalControlID,
				"file":              nil,
			},
		},
		"input.file",
		testutil.UploadFile{
			Filename:    "mcp-evidence.pdf",
			ContentType: "application/pdf",
			Content:     fileContent,
		},
		&uploadResult,
	)
	require.NoError(t, err)

	evidenceID := uploadResult.UploadInternalControlEvidence.EvidenceEdge.Node.ID
	require.NotEmpty(t, evidenceID)

	var listResult struct {
		Evidences []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			URL  string `json:"url"`
			File *struct {
				FileName    string `json:"file_name"`
				DownloadURL string `json:"download_url"`
			} `json:"file"`
		} `json:"evidences"`
	}
	mc.CallToolInto(
		"listInternalControlEvidences",
		map[string]any{
			"internal_control_id": internalControlID,
		},
		&listResult,
	)
	require.Len(t, listResult.Evidences, 1)
	assert.Equal(t, evidenceID, listResult.Evidences[0].ID)
	assert.Equal(t, "FILE", listResult.Evidences[0].Type)
	require.NotNil(t, listResult.Evidences[0].File)
	assert.Equal(t, "mcp-evidence.pdf", listResult.Evidences[0].File.FileName)
	assert.NotEmpty(t, listResult.Evidences[0].URL)
	assert.Equal(t, listResult.Evidences[0].File.DownloadURL, listResult.Evidences[0].URL)

	response, err := owner.HTTPClient().Get(listResult.Evidences[0].URL)
	require.NoError(t, err)

	defer func() { _ = response.Body.Close() }()

	downloaded, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, fileContent, downloaded)
}
