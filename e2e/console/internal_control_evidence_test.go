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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestInternalControlEvidence_UploadLifecycle(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.NewInternalControl(owner).
		WithName(factory.SafeName("internal control evidence lifecycle")).
		WithCategory("EVIDENCE").
		Create()

	pdfSize := int64(len(internalControlEvidenceTestPDF))
	beforeUpload := time.Now().UTC()

	uploaded := uploadInternalControlEvidence(t, owner, internalControlID, "lifecycle-evidence.pdf")
	require.NotEmpty(t, uploaded.ID)

	assert.Equal(t, "FULFILLED", uploaded.State)
	assert.Equal(t, "FILE", uploaded.Type)
	assert.Nil(t, uploaded.URL)
	assert.Nil(t, uploaded.Description)
	assert.Equal(t, internalControlID, uploaded.InternalControl.ID)
	assert.Nil(t, uploaded.Task)
	assert.True(t, uploaded.CanDelete)
	require.NotNil(t, uploaded.File)
	assert.Equal(t, "lifecycle-evidence.pdf", uploaded.File.FileName)
	assert.Equal(t, "application/pdf", uploaded.File.MimeType)
	assert.Equal(t, pdfSize, uploaded.File.Size)
	assert.NotEmpty(t, uploaded.File.ID)
	assert.NotEmpty(t, uploaded.File.DownloadURL)
	testutil.AssertTimestampsOnCreate(t, uploaded.CreatedAt, uploaded.UpdatedAt, beforeUpload)

	node := queryEvidenceNode(t, owner, uploaded.ID)
	require.NotNil(t, node)
	assert.Equal(t, uploaded.ID, node.ID)
	assert.Equal(t, internalControlID, node.InternalControl.ID)
	assert.Nil(t, node.Task)
	assert.True(t, node.CanDelete)

	list := queryInternalControlEvidences(t, owner, internalControlID)
	assert.Equal(t, 1, list.TotalCount)
	testutil.AssertFirstPage(t, len(list.Edges), list.PageInfo, 1, false)
	require.Len(t, list.Edges, 1)
	assert.Equal(t, uploaded.ID, list.Edges[0].Node.ID)
	assert.NotEmpty(t, list.Edges[0].Cursor)

	deletedID := deleteInternalControlEvidence(t, owner, uploaded.ID)
	assert.Equal(t, uploaded.ID, deletedID)

	var afterDeleteNode struct {
		Node *internalControlEvidenceWireNode `json:"node"`
	}

	err := owner.Execute(
		`
			query($id: ID!) {
				node(id: $id) {
					... on Evidence { id }
				}
			}
		`,
		map[string]any{"id": uploaded.ID},
		&afterDeleteNode,
	)
	testutil.AssertNodeNotAccessible(t, err, afterDeleteNode.Node == nil, "evidence")

	afterList := queryInternalControlEvidences(t, owner, internalControlID)
	assert.Equal(t, 0, afterList.TotalCount)
	assert.Empty(t, afterList.Edges)
}

func TestInternalControlEvidence_UploadRejectsEmptyFile(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	internalControlID := factory.NewInternalControl(owner).
		WithName(factory.SafeName("internal control evidence empty file")).
		Create()

	err := owner.ExecuteWithFile(
		`
			mutation UploadInternalControlEvidence($input: UploadInternalControlEvidenceInput!) {
				uploadInternalControlEvidence(input: $input) {
					evidenceEdge { node { id } }
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
			Filename:    "empty.pdf",
			ContentType: "application/pdf",
			Content:     []byte{},
		},
		nil,
	)
	testutil.RequireErrorCode(t, err, "INVALID")
}
