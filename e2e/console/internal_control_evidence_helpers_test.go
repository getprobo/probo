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
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const internalControlEvidenceTestPDF = "%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF"

const internalControlEvidenceNodeSelection = `
	id
	size
	state
	type
	description
	url
	createdAt
	updatedAt
	file {
		id
		fileName
		mimeType
		size
		downloadUrl
	}
	internalControl { id }
	task { id }
	canDelete: permission(action: "core:evidence:delete")
`

type (
	internalControlEvidenceWireNode struct {
		ID          string    `json:"id"`
		Size        int       `json:"size"`
		State       string    `json:"state"`
		Type        string    `json:"type"`
		Description *string   `json:"description"`
		URL         *string   `json:"url"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
		File        *struct {
			ID          string `json:"id"`
			FileName    string `json:"fileName"`
			MimeType    string `json:"mimeType"`
			Size        int64  `json:"size"`
			DownloadURL string `json:"downloadUrl"`
		} `json:"file"`
		InternalControl struct {
			ID string `json:"id"`
		} `json:"internalControl"`
		Task *struct {
			ID string `json:"id"`
		} `json:"task"`
		CanDelete bool `json:"canDelete"`
	}

	internalControlEvidenceUploadResult struct {
		UploadInternalControlEvidence struct {
			EvidenceEdge struct {
				Cursor string                          `json:"cursor"`
				Node   internalControlEvidenceWireNode `json:"node"`
			} `json:"evidenceEdge"`
		} `json:"uploadInternalControlEvidence"`
	}

	internalControlEvidencesConnection struct {
		TotalCount int `json:"totalCount"`
		PageInfo   testutil.PageInfo
		Edges      []struct {
			Cursor string                          `json:"cursor"`
			Node   internalControlEvidenceWireNode `json:"node"`
		} `json:"edges"`
	}
)

func internalControlEvidenceUploadPDF(fileName string) testutil.UploadFile {
	return testutil.UploadFile{
		Filename:    fileName,
		ContentType: "application/pdf",
		Content:     []byte(internalControlEvidenceTestPDF),
	}
}

func uploadInternalControlEvidence(
	t *testing.T,
	client *testutil.Client,
	internalControlID string,
	fileName string,
) internalControlEvidenceWireNode {
	t.Helper()

	const mutation = `
		mutation UploadInternalControlEvidence($input: UploadInternalControlEvidenceInput!) {
			uploadInternalControlEvidence(input: $input) {
				evidenceEdge {
					cursor
					node {
						NODE
					}
				}
			}
		}
	`

	var result internalControlEvidenceUploadResult

	query := replaceInternalControlEvidenceNodeSelection(mutation)

	err := client.ExecuteWithFile(
		query,
		map[string]any{
			"input": map[string]any{
				"internalControlId": internalControlID,
				"file":              nil,
			},
		},
		"input.file",
		internalControlEvidenceUploadPDF(fileName),
		&result,
	)
	require.NoError(t, err)

	return result.UploadInternalControlEvidence.EvidenceEdge.Node
}

func uploadInternalControlEvidenceExpectError(
	t *testing.T,
	client *testutil.Client,
	internalControlID string,
	fileName string,
) error {
	t.Helper()

	const mutation = `
		mutation UploadInternalControlEvidence($input: UploadInternalControlEvidenceInput!) {
			uploadInternalControlEvidence(input: $input) {
				evidenceEdge { node { id } }
			}
		}
	`

	return client.ExecuteWithFile(
		mutation,
		map[string]any{
			"input": map[string]any{
				"internalControlId": internalControlID,
				"file":              nil,
			},
		},
		"input.file",
		internalControlEvidenceUploadPDF(fileName),
		nil,
	)
}

func deleteInternalControlEvidence(t *testing.T, client *testutil.Client, evidenceID string) string {
	t.Helper()

	const mutation = `
		mutation DeleteEvidence($input: DeleteEvidenceInput!) {
			deleteEvidence(input: $input) {
				deletedEvidenceId
			}
		}
	`

	var result struct {
		DeleteEvidence struct {
			DeletedEvidenceID string `json:"deletedEvidenceId"`
		} `json:"deleteEvidence"`
	}

	err := client.Execute(
		mutation,
		map[string]any{
			"input": map[string]any{
				"evidenceId": evidenceID,
			},
		},
		&result,
	)
	require.NoError(t, err)

	return result.DeleteEvidence.DeletedEvidenceID
}

func deleteInternalControlEvidenceExpectError(t *testing.T, client *testutil.Client, evidenceID string) error {
	t.Helper()

	const mutation = `
		mutation DeleteEvidence($input: DeleteEvidenceInput!) {
			deleteEvidence(input: $input) {
				deletedEvidenceId
			}
		}
	`

	return client.Execute(
		mutation,
		map[string]any{
			"input": map[string]any{
				"evidenceId": evidenceID,
			},
		},
		nil,
	)
}

func queryInternalControlEvidences(
	t *testing.T,
	client *testutil.Client,
	internalControlID string,
) internalControlEvidencesConnection {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on InternalControl {
					id
					evidences(first: 10) {
						totalCount
						pageInfo {
							hasNextPage
							hasPreviousPage
							startCursor
							endCursor
						}
						edges {
							cursor
							node {
								NODE
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			ID        string                             `json:"id"`
			Evidences internalControlEvidencesConnection `json:"evidences"`
		} `json:"node"`
	}

	err := client.Execute(
		replaceInternalControlEvidenceNodeSelection(query),
		map[string]any{"id": internalControlID},
		&result,
	)
	require.NoError(t, err)

	return result.Node.Evidences
}

func queryEvidenceNode(
	t *testing.T,
	client *testutil.Client,
	evidenceID string,
) *internalControlEvidenceWireNode {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Evidence {
					NODE
				}
			}
		}
	`

	var result struct {
		Node *internalControlEvidenceWireNode `json:"node"`
	}

	err := client.Execute(
		replaceInternalControlEvidenceNodeSelection(query),
		map[string]any{"id": evidenceID},
		&result,
	)
	require.NoError(t, err)

	return result.Node
}

func replaceInternalControlEvidenceNodeSelection(query string) string {
	return strings.ReplaceAll(query, "NODE", internalControlEvidenceNodeSelection)
}
