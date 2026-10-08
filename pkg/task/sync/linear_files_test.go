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

package tasksync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

type fakeAttachments struct {
	file  Attachment
	calls int
}

func (f *fakeAttachments) ReadAttachment(
	context.Context,
	coredata.Scoper,
	gid.GID,
	gid.GID,
) (Attachment, error) {
	f.calls++

	return f.file, nil
}

func TestEmbedLinearFiles_UploadsAndRestoresFileID(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType)
	content, err := MarkdownToContent("![diagram](" + prosemirror.AttachmentPath(fileID.String()) + ")")
	require.NoError(t, err)

	markdown, err := ContentToMarkdown(content)
	require.NoError(t, err)
	require.Contains(t, markdown, prosemirror.AttachmentPath(fileID.String()))

	var srv *httptest.Server

	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"fileUpload": map[string]any{
					"success": true,
					"uploadFile": map[string]any{
						"uploadUrl": srv.URL + "/upload",
						"assetUrl":  "https://uploads.linear.app/file/diagram",
						"headers":   []map[string]string{},
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	client := linear.NewClient(srv.Client(), srv.URL)
	client.SetUploadHTTPForTest(srv.Client())

	reader := &fakeAttachments{file: Attachment{Name: "diagram.png", MimeType: "image/png", Body: []byte("png")}}
	svc := &Service{
		attachments:  reader,
		linearAssets: linearAssetCache{urls: map[string]string{}},
	}

	remote, files, changed, err := svc.embedLinearFiles(
		t.Context(),
		client,
		nil,
		gid.Nil,
		gid.Nil,
		markdown,
		content,
		nil,
	)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, 1, reader.calls)
	assert.Contains(t, remote, "https://uploads.linear.app/file/diagram")
	assert.NotContains(t, remote, prosemirror.AttachmentPath(fileID.String()))
	assert.Equal(t, "https://uploads.linear.app/file/diagram", files[fileID.String()])

	restored, err := MarkdownToContent(restoreLinearFiles(remote, files))
	require.NoError(t, err)
	assert.Contains(t, restored, fileID.String())

	_, _, changed, err = svc.embedLinearFiles(
		t.Context(),
		client,
		nil,
		gid.Nil,
		gid.Nil,
		markdown,
		content,
		files,
	)
	require.NoError(t, err)
	assert.False(t, changed)
	assert.Equal(t, 1, reader.calls)
}

func TestMergeAttachments_KeepsAppActor(t *testing.T) {
	t.Parallel()

	metadata, err := mergeAttachments(
		[]byte(`{"app_actor_id":"app-1","attachment_id":"att-1"}`),
		map[string]string{"file-1": "https://uploads.linear.app/file/a"},
	)
	require.NoError(t, err)
	assert.True(t, isAppActor(metadata, "app-1"))
	assert.Equal(
		t,
		map[string]string{"file-1": "https://uploads.linear.app/file/a"},
		attachmentsFromMetadata(metadata),
	)
}

func TestLinearFilesMissing(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType)
	content, err := MarkdownToContent("![diagram](" + prosemirror.AttachmentPath(fileID.String()) + ")")
	require.NoError(t, err)

	missing, err := linearFilesMissing(nil, content)
	require.NoError(t, err)
	assert.True(t, missing)

	missing, err = linearFilesMissing(map[string]string{fileID.String(): "https://uploads.linear.app/file/a"}, content)
	require.NoError(t, err)
	assert.False(t, missing)
}
