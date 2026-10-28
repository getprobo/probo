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

package asana

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	client "go.probo.inc/probo/pkg/task/sync/asana/client"
)

func TestPullAsanaFilesLeavesTheInputMapUnchanged(t *testing.T) {
	t.Parallel()

	known := map[string]string{"file-1": "attachment-1"}
	content := prosemirror.FromPlainText("Ship the sync")

	next, updated, err := (&Service{}).pullAsanaFiles(
		context.Background(),
		nil,
		asanaAccount{},
		gid.Nil,
		content,
		"",
		known,
		nil,
		"task-1",
		"task",
	)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"file-1": "attachment-1"}, known)
	assert.Empty(t, updated)
	assert.NotEmpty(t, next)
}

func TestPullAsanaFilesRestoresKnownFilesMissingFromHTML(t *testing.T) {
	t.Parallel()

	keptID := gid.New(gid.TenantID{}, 1).String()
	removedID := gid.New(gid.TenantID{}, 2).String()
	previous, err := documentWithFiles(
		"Ship the sync",
		[]prosemirror.FileAttrs{
			{FileID: keptID, FileName: "kept.pdf", MimeType: "application/pdf", Size: 4},
			{FileID: removedID, FileName: "gone.pdf", MimeType: "application/pdf", Size: 8},
		},
	)
	require.NoError(t, err)

	known := map[string]string{
		keptID:    "attachment-kept",
		removedID: "attachment-gone",
	}

	next, updated, err := (&Service{}).pullAsanaFiles(
		context.Background(),
		nil,
		asanaAccount{},
		gid.Nil,
		prosemirror.FromPlainText("Updated from Asana"),
		previous,
		known,
		[]client.Attachment{
			{
				GID:        "attachment-kept",
				Name:       "kept.pdf",
				ParentGID:  "task-1",
				ParentType: "task",
			},
		},
		"task-1",
		"task",
	)
	require.NoError(t, err)
	assert.Equal(
		t,
		map[string]string{
			keptID:    "attachment-kept",
			removedID: "attachment-gone",
		},
		known,
	)
	assert.Equal(t, map[string]string{keptID: "attachment-kept"}, updated)
	assert.Contains(t, next, "Updated from Asana")

	node, err := prosemirror.Parse(next)
	require.NoError(t, err)
	require.Len(t, prosemirror.ReferencedFiles(node), 1)

	var fileNode prosemirror.Node

	for _, child := range node.Content {
		if child.Type == prosemirror.NodeFile {
			fileNode = child
		}
	}

	attrs, err := fileNode.FileAttrs()
	require.NoError(t, err)
	assert.Equal(t, keptID, attrs.FileID)
	assert.Equal(t, "kept.pdf", attrs.FileName)
}

func TestReadAsanaFilesReturnsMalformedMetadata(t *testing.T) {
	t.Parallel()

	meta, err := readAsanaFiles(json.RawMessage(`{"asana_files":`))
	require.Error(t, err)
	assert.Nil(t, meta.Files)
	assert.Nil(t, meta.CommentFiles)
}

func TestReadAsanaFilesReadsMaps(t *testing.T) {
	t.Parallel()

	meta, err := readAsanaFiles(json.RawMessage(`{"asana_files":{"file-1":"attachment-1"}}`))
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"file-1": "attachment-1"}, meta.Files)
	require.NotNil(t, meta.CommentFiles)
	assert.Empty(t, meta.CommentFiles)

	meta, err = readAsanaFiles(nil)
	require.NoError(t, err)
	assert.Nil(t, meta.Files)
	assert.Nil(t, meta.CommentFiles)
}

func documentWithFiles(text string, files []prosemirror.FileAttrs) (string, error) {
	content := []prosemirror.Node{{
		Type: prosemirror.NodeParagraph,
		Content: []prosemirror.Node{{
			Type: prosemirror.NodeText,
			Text: &text,
		}},
	}}

	for _, file := range files {
		attrs, err := json.Marshal(file)
		if err != nil {
			return "", err
		}

		content = append(content, prosemirror.Node{Type: prosemirror.NodeFile, Attrs: attrs})
	}

	encoded, err := json.Marshal(prosemirror.Node{Type: prosemirror.NodeDoc, Content: content})
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}
