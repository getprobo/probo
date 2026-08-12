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

const measureEvidenceTestPDF = "%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF"

const measureEvidenceNodeSelection = `
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
	measure { id }
	task { id }
	canDelete: permission(action: "core:evidence:delete")
`

type (
	measureEvidenceWireNode struct {
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
		Measure struct {
			ID string `json:"id"`
		} `json:"measure"`
		Task *struct {
			ID string `json:"id"`
		} `json:"task"`
		CanDelete bool `json:"canDelete"`
	}

	measureEvidenceUploadResult struct {
		UploadMeasureEvidence struct {
			EvidenceEdge struct {
				Cursor string                  `json:"cursor"`
				Node   measureEvidenceWireNode `json:"node"`
			} `json:"evidenceEdge"`
		} `json:"uploadMeasureEvidence"`
	}

	measureEvidencesConnection struct {
		TotalCount int `json:"totalCount"`
		PageInfo   testutil.PageInfo
		Edges      []struct {
			Cursor string                  `json:"cursor"`
			Node   measureEvidenceWireNode `json:"node"`
		} `json:"edges"`
	}
)

func measureEvidenceUploadPDF(fileName string) testutil.UploadFile {
	return testutil.UploadFile{
		Filename:    fileName,
		ContentType: "application/pdf",
		Content:     []byte(measureEvidenceTestPDF),
	}
}

func uploadMeasureEvidence(
	t *testing.T,
	client *testutil.Client,
	measureID string,
	fileName string,
) measureEvidenceWireNode {
	t.Helper()

	const mutation = `
		mutation UploadMeasureEvidence($input: UploadMeasureEvidenceInput!) {
			uploadMeasureEvidence(input: $input) {
				evidenceEdge {
					cursor
					node {
						NODE
					}
				}
			}
		}
	`

	var result measureEvidenceUploadResult

	query := replaceMeasureEvidenceNodeSelection(mutation)

	err := client.ExecuteWithFile(
		query,
		map[string]any{
			"input": map[string]any{
				"measureId": measureID,
				"file":      nil,
			},
		},
		"input.file",
		measureEvidenceUploadPDF(fileName),
		&result,
	)
	require.NoError(t, err)

	return result.UploadMeasureEvidence.EvidenceEdge.Node
}

func uploadMeasureEvidenceExpectError(
	t *testing.T,
	client *testutil.Client,
	measureID string,
	fileName string,
) error {
	t.Helper()

	const mutation = `
		mutation UploadMeasureEvidence($input: UploadMeasureEvidenceInput!) {
			uploadMeasureEvidence(input: $input) {
				evidenceEdge { node { id } }
			}
		}
	`

	return client.ExecuteWithFile(
		mutation,
		map[string]any{
			"input": map[string]any{
				"measureId": measureID,
				"file":      nil,
			},
		},
		"input.file",
		measureEvidenceUploadPDF(fileName),
		nil,
	)
}

func deleteMeasureEvidence(t *testing.T, client *testutil.Client, evidenceID string) string {
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

func deleteMeasureEvidenceExpectError(t *testing.T, client *testutil.Client, evidenceID string) error {
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

func queryMeasureEvidences(
	t *testing.T,
	client *testutil.Client,
	measureID string,
) measureEvidencesConnection {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Measure {
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
			ID        string                     `json:"id"`
			Evidences measureEvidencesConnection `json:"evidences"`
		} `json:"node"`
	}

	err := client.Execute(
		replaceMeasureEvidenceNodeSelection(query),
		map[string]any{"id": measureID},
		&result,
	)
	require.NoError(t, err)

	return result.Node.Evidences
}

func queryEvidenceNode(
	t *testing.T,
	client *testutil.Client,
	evidenceID string,
) *measureEvidenceWireNode {
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
		Node *measureEvidenceWireNode `json:"node"`
	}

	err := client.Execute(
		replaceMeasureEvidenceNodeSelection(query),
		map[string]any{"id": evidenceID},
		&result,
	)
	require.NoError(t, err)

	return result.Node
}

func replaceMeasureEvidenceNodeSelection(query string) string {
	return strings.ReplaceAll(query, "NODE", measureEvidenceNodeSelection)
}
