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
	"encoding/base64"
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

func TestMCP_UploadInternalControlEvidence(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	internalControlID := factory.NewInternalControl(owner).
		WithName(factory.SafeName("MCP evidence upload")).
		WithCategory("EVIDENCE").
		Create()
	fileContent := []byte(`{"check":"s3_account_block_public_access","result":true}`)

	type uploadResult struct {
		Evidence struct {
			ID                string `json:"id"`
			InternalControlID string `json:"internal_control_id"`
			Type              string `json:"type"`
			State             string `json:"state"`
			URL               string `json:"url"`
			File              *struct {
				FileName string `json:"file_name"`
				MimeType string `json:"mime_type"`
			} `json:"file"`
		} `json:"evidence"`
	}

	var uploaded uploadResult
	mc.CallToolInto("uploadInternalControlEvidence", map[string]any{
		"internal_control_id": internalControlID,
		"file_name":           "aws-evidence.json",
		"content_base64":      base64.StdEncoding.EncodeToString(fileContent),
	}, &uploaded)
	require.NotEmpty(t, uploaded.Evidence.ID)
	assert.Equal(t, internalControlID, uploaded.Evidence.InternalControlID)
	assert.Equal(t, "FILE", uploaded.Evidence.Type)
	assert.Equal(t, "FULFILLED", uploaded.Evidence.State)
	require.NotNil(t, uploaded.Evidence.File)
	assert.Equal(t, "aws-evidence.json", uploaded.Evidence.File.FileName)
	assert.Equal(t, "application/json", uploaded.Evidence.File.MimeType)
	require.NotEmpty(t, uploaded.Evidence.URL)

	response, err := owner.HTTPClient().Get(uploaded.Evidence.URL)
	require.NoError(t, err)

	defer func() { _ = response.Body.Close() }()

	downloaded, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, fileContent, downloaded)

	var listResult struct {
		Evidences []struct {
			ID string `json:"id"`
		} `json:"evidences"`
	}
	mc.CallToolInto("listInternalControlEvidences", map[string]any{
		"internal_control_id": internalControlID,
	}, &listResult)
	require.Len(t, listResult.Evidences, 1)
	assert.Equal(t, uploaded.Evidence.ID, listResult.Evidences[0].ID)
}

func TestMCP_UploadInternalControlEvidence_RejectsInvalidInput(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	internalControlID := factory.NewInternalControl(owner).
		WithName(factory.SafeName("MCP evidence invalid")).
		WithCategory("EVIDENCE").
		Create()

	msg := mc.CallToolExpectToolError("uploadInternalControlEvidence", map[string]any{
		"internal_control_id": internalControlID,
		"file_name":           "evidence.json",
		"content_base64":      "not base64!",
	})
	assert.Contains(t, msg, "base64")

	msg = mc.CallToolExpectToolError("uploadInternalControlEvidence", map[string]any{
		"internal_control_id": internalControlID,
		"file_name":           "payload.exe",
		"content_base64":      base64.StdEncoding.EncodeToString([]byte("MZ")),
	})
	assert.Contains(t, msg, "mime_type")

	msg = mc.CallToolExpectToolError("uploadInternalControlEvidence", map[string]any{
		"internal_control_id": internalControlID,
		"file_name":           "empty.json",
		"content_base64":      "",
	})
	assert.NotContains(t, msg, "pq:")
	assert.NotContains(t, msg, "sql:")
}

func TestMCP_UploadInternalControlEvidence_PermissionDenied(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.NewInternalControl(owner).
		WithName(factory.SafeName("MCP evidence denied")).
		WithCategory("EVIDENCE").
		Create()

	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	viewerMC := testutil.NewMCPClient(t, viewer)

	msg := viewerMC.CallToolExpectToolError("uploadInternalControlEvidence", map[string]any{
		"internal_control_id": internalControlID,
		"file_name":           "evidence.json",
		"content_base64":      base64.StdEncoding.EncodeToString([]byte(`{}`)),
	})
	assert.Contains(t, msg, "permission denied")
}
