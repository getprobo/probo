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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

type fakeAttachments struct {
	file  Attachment
	calls int
}

type failingAttachments struct{}

type selectiveAttachments struct {
	files map[string]Attachment
}

func (s *selectiveAttachments) ReadAttachment(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	fileID gid.GID,
) (Attachment, error) {
	file, ok := s.files[fileID.String()]
	if !ok {
		return Attachment{}, errors.New("missing object")
	}

	return file, nil
}

func (failingAttachments) ReadAttachment(
	context.Context,
	coredata.Scoper,
	gid.GID,
	gid.GID,
) (Attachment, error) {
	return Attachment{}, errors.New("missing object")
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

// setLinearUploadHTTP points a Linear client's file upload at the test server.
// The client is otherwise created with SSRF protection, which refuses the
// loopback address httptest uses. The field stays unexported so production
// code cannot replace that client.
func setLinearUploadHTTP(client *linear.Client, httpClient *http.Client) {
	field := reflect.ValueOf(client).Elem().FieldByName("uploadHTTP")
	target := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr()))
	target.Elem().Set(reflect.ValueOf(httpClient))
}

func TestEmbedLinearFiles_UploadsAndRestoresFileID(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType)
	content, err := MarkdownToContent("![diagram](" + prosemirror.AttachmentPath(fileID.String()) + ")")
	require.NoError(t, err)

	markdown, err := ContentToMarkdown(content)
	require.NoError(t, err)
	require.Contains(t, markdown, prosemirror.AttachmentPath(fileID.String()))

	var (
		srv        *httptest.Server
		mu         sync.Mutex
		putBody    []byte
		uploadName string
		uploadType string
	)

	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			mu.Lock()
			putBody = body
			mu.Unlock()

			w.WriteHeader(http.StatusOK)

			return
		}

		var payload struct {
			Variables struct {
				Filename    string `json:"filename"`
				ContentType string `json:"contentType"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		mu.Lock()
		uploadName = payload.Variables.Filename
		uploadType = payload.Variables.ContentType
		mu.Unlock()

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
	setLinearUploadHTTP(client, srv.Client())

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

	mu.Lock()
	body, name, mime := putBody, uploadName, uploadType
	mu.Unlock()

	assert.Equal(t, []byte("png"), body)
	assert.Equal(t, "diagram.png", name)
	assert.Equal(t, "image/png", mime)
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

func TestEmbedLinearFiles_KnownFilesSkipStorage(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType)
	content, err := MarkdownToContent("![diagram](" + prosemirror.AttachmentPath(fileID.String()) + ")")
	require.NoError(t, err)

	markdown, err := ContentToMarkdown(content)
	require.NoError(t, err)

	const assetURL = "https://uploads.linear.app/file/diagram"

	t.Run("known url", func(t *testing.T) {
		t.Parallel()

		svc := &Service{}
		remote, files, changed, err := svc.embedLinearFiles(
			t.Context(),
			nil,
			nil,
			gid.Nil,
			gid.Nil,
			markdown,
			content,
			map[string]string{fileID.String(): assetURL},
		)
		require.NoError(t, err)
		assert.False(t, changed)
		assert.Contains(t, remote, assetURL)
		assert.Equal(t, assetURL, files[fileID.String()])
	})

	t.Run("cached url", func(t *testing.T) {
		t.Parallel()

		svc := &Service{}
		svc.storeLinearAsset(gid.Nil, "", fileID.String(), assetURL)

		remote, _, changed, err := svc.embedLinearFiles(
			t.Context(),
			nil,
			nil,
			gid.Nil,
			gid.Nil,
			markdown,
			content,
			nil,
		)
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Contains(t, remote, assetURL)
	})

	t.Run("previous connection", func(t *testing.T) {
		t.Parallel()

		svc := &Service{}
		svc.storeLinearAsset(gid.Nil, "old-connection", fileID.String(), assetURL)

		client := linear.NewClient(http.DefaultClient, "https://linear.test")
		client.SetConnectionGeneration("new-connection")

		_, _, _, err := svc.embedLinearFiles(
			t.Context(),
			client,
			nil,
			gid.Nil,
			gid.Nil,
			markdown,
			content,
			nil,
		)
		require.Error(t, err)
		assert.ErrorContains(t, err, "storage is not configured")
		assert.Empty(t, svc.cachedLinearAsset(gid.Nil, "old-connection", fileID.String()))
	})

	t.Run("upload without storage", func(t *testing.T) {
		t.Parallel()

		svc := &Service{}
		_, _, _, err := svc.embedLinearFiles(
			t.Context(),
			nil,
			nil,
			gid.Nil,
			gid.Nil,
			markdown,
			content,
			nil,
		)
		require.Error(t, err)
		assert.ErrorContains(t, err, "storage is not configured")
		assert.False(t, isPartialLinearEmbed(err))
	})
}

func TestLinearAssetGeneration_ChangesWithAccessToken(t *testing.T) {
	t.Parallel()

	assert.Empty(t, linearAssetGeneration(""))
	assert.NotEqual(t, linearAssetGeneration("old-token"), linearAssetGeneration("new-token"))
	assert.Equal(t, linearAssetGeneration("same-token"), linearAssetGeneration("same-token"))
}

func TestEmbedLinearFiles_KeepsTextWhenOneAttachmentFails(t *testing.T) {
	t.Parallel()

	goodID := gid.New(gid.NewTenantID(), coredata.FileEntityType)
	badID := gid.New(gid.NewTenantID(), coredata.FileEntityType)
	goodPath := prosemirror.AttachmentPath(goodID.String())
	badPath := prosemirror.AttachmentPath(badID.String())
	content, err := MarkdownToContent("![diagram](" + goodPath + ")\n\n![other](" + badPath + ")")
	require.NoError(t, err)

	markdown, err := ContentToMarkdown(content)
	require.NoError(t, err)

	const assetURL = "https://uploads.linear.app/file/diagram"

	svc := &Service{attachments: failingAttachments{}}

	remote, files, changed, err := svc.embedLinearFiles(
		t.Context(),
		nil,
		nil,
		gid.Nil,
		gid.Nil,
		markdown,
		content,
		map[string]string{goodID.String(): assetURL},
	)
	require.Error(t, err)
	assert.True(t, isPartialLinearEmbed(err))
	assert.ErrorContains(t, err, badID.String())
	assert.False(t, changed)
	assert.Contains(t, remote, assetURL)
	assert.Contains(t, remote, badPath)
	assert.NotContains(t, remote, goodPath)
	assert.Equal(t, assetURL, files[goodID.String()])
	assert.Empty(t, files[badID.String()])
}

func TestEmbedLinearFiles_UploadsOneFileWhenAnotherFails(t *testing.T) {
	t.Parallel()

	goodID := gid.New(gid.NewTenantID(), coredata.FileEntityType)
	badID := gid.New(gid.NewTenantID(), coredata.FileEntityType)
	goodPath := prosemirror.AttachmentPath(goodID.String())
	badPath := prosemirror.AttachmentPath(badID.String())
	content, err := MarkdownToContent("![diagram](" + goodPath + ")\n\n![other](" + badPath + ")")
	require.NoError(t, err)

	markdown, err := ContentToMarkdown(content)
	require.NoError(t, err)

	const assetURL = "https://uploads.linear.app/file/diagram"

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
						"assetUrl":  assetURL,
						"headers":   []map[string]string{},
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	client := linear.NewClient(srv.Client(), srv.URL)
	setLinearUploadHTTP(client, srv.Client())

	svc := &Service{
		attachments: &selectiveAttachments{files: map[string]Attachment{
			goodID.String(): {Name: "diagram.png", MimeType: "image/png", Body: []byte("png")},
		}},
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
	require.Error(t, err)
	assert.True(t, isPartialLinearEmbed(err))
	assert.ErrorContains(t, err, badID.String())
	assert.True(t, changed)
	assert.Equal(t, assetURL, files[goodID.String()])
	assert.Empty(t, files[badID.String()])
	assert.Contains(t, remote, assetURL)
	assert.Contains(t, remote, badPath)
	assert.NotContains(t, remote, goodPath)
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

func TestSaveLinearAttachments_ReplacedLink(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	taskID := gid.New(tenantID, coredata.TaskEntityType)
	connectorID := gid.New(tenantID, coredata.ConnectorEntityType)
	now := time.Now().UTC().Truncate(time.Microsecond)
	files := map[string]string{"file-1": "https://uploads.linear.app/old"}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $4)`,
			organizationID,
			tenantID,
			"linear-attachments-"+organizationID.String(),
			now,
		); err != nil {
			return err
		}

		if _, err := tx.Exec(
			ctx,
			`INSERT INTO tasks (
			     id, tenant_id, organization_id, name, content, reference_id,
			     state, priority, rank, created_at, updated_at
			 )
			 VALUES (
			     $1, $2, $3, 'links', '{"type":"doc","content":[]}'::jsonb, $4,
			     'TODO', 'MEDIUM', 1, $5, $5
			 )`,
			taskID,
			tenantID,
			organizationID,
			taskID.String(),
			now,
		); err != nil {
			return err
		}

		if _, err := tx.Exec(
			ctx,
			`INSERT INTO connectors (
			     id, tenant_id, organization_id, provider, protocol, name,
			     settings, encrypted_connection, created_at, updated_at
			 )
			 VALUES (
			     $1, $2, $3, 'LINEAR_SYNC', 'OAUTH2', 'Linear sync',
			     '{}', '\x', $4, $4
			 )`,
			connectorID,
			tenantID,
			organizationID,
			now,
		); err != nil {
			return err
		}

		_, err := tx.Exec(
			ctx,
			`INSERT INTO task_external_links (
			     task_id, tenant_id, organization_id, connector_id, provider,
			     external_id, external_identifier, external_url, destination, origin,
			     metadata, created_at, updated_at
			 )
			 VALUES (
			     $1, $2, $3, $4, 'LINEAR_SYNC',
			     'new-issue', 'ENG-2', 'https://linear.app/issue/ENG-2', '{}', 'PROBO',
			     '{}', $5, $5
			 )`,
			taskID,
			tenantID,
			organizationID,
			connectorID,
			now,
		)

		return err
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			if _, err := tx.Exec(ctx, `DELETE FROM task_external_links WHERE task_id = $1`, taskID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, taskID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM connectors WHERE id = $1`, connectorID); err != nil {
				return err
			}

			_, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID)

			return err
		})
	})

	svc := &Service{pg: client}
	previousConnectorID := gid.New(tenantID, coredata.ConnectorEntityType)

	require.NoError(t, svc.saveLinearAttachments(ctx, scope, taskID, previousConnectorID, "old-issue", files))

	stored, err := svc.linearAttachments(ctx, scope, taskID)
	require.NoError(t, err)
	assert.Empty(t, stored)

	svc.storeLinearAsset(connectorID, "", "file-1", files["file-1"])

	require.NoError(t, svc.saveLinearAttachments(ctx, scope, taskID, connectorID, "new-issue", files))

	stored, err = svc.linearAttachments(ctx, scope, taskID)
	require.NoError(t, err)
	assert.Equal(t, files, stored)
	assert.Empty(t, svc.cachedLinearAsset(connectorID, "", "file-1"))
}

func TestApplyInboundComment_CreateRestoresKnownFile(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	taskID := gid.New(tenantID, coredata.TaskEntityType)
	connectorID := gid.New(tenantID, coredata.ConnectorEntityType)
	fileID := gid.New(tenantID, coredata.FileEntityType)
	now := time.Now().UTC().Truncate(time.Microsecond)
	assetURL := "https://uploads.linear.app/file/diagram"
	metadata, err := json.Marshal(map[string]any{
		"attachments": map[string]string{fileID.String(): assetURL},
	})
	require.NoError(t, err)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $4)`,
			organizationID,
			tenantID,
			"inbound-comment-"+organizationID.String(),
			now,
		); err != nil {
			return err
		}

		if _, err := tx.Exec(
			ctx,
			`INSERT INTO tasks (
			     id, tenant_id, organization_id, name, content, reference_id,
			     state, priority, rank, created_at, updated_at
			 )
			 VALUES (
			     $1, $2, $3, 'links', '{"type":"doc","content":[]}'::jsonb, $4,
			     'TODO', 'MEDIUM', 1, $5, $5
			 )`,
			taskID,
			tenantID,
			organizationID,
			taskID.String(),
			now,
		); err != nil {
			return err
		}

		if _, err := tx.Exec(
			ctx,
			`INSERT INTO connectors (
			     id, tenant_id, organization_id, provider, protocol, name,
			     settings, encrypted_connection, created_at, updated_at
			 )
			 VALUES (
			     $1, $2, $3, 'LINEAR_SYNC', 'OAUTH2', 'Linear sync',
			     '{}', '\x', $4, $4
			 )`,
			connectorID,
			tenantID,
			organizationID,
			now,
		); err != nil {
			return err
		}

		_, err := tx.Exec(
			ctx,
			`INSERT INTO task_external_links (
			     task_id, tenant_id, organization_id, connector_id, provider,
			     external_id, external_identifier, external_url, destination, origin,
			     metadata, created_at, updated_at
			 )
			 VALUES (
			     $1, $2, $3, $4, 'LINEAR_SYNC',
			     'issue-1', 'ENG-1', 'https://linear.app/issue/ENG-1', '{}', 'PROBO',
			     $5, $6, $6
			 )`,
			taskID,
			tenantID,
			organizationID,
			connectorID,
			metadata,
			now,
		)

		return err
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			if _, err := tx.Exec(ctx, `DELETE FROM task_comment_external_links WHERE task_id = $1`, taskID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM task_comments WHERE task_id = $1`, taskID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM task_external_links WHERE task_id = $1`, taskID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, taskID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM connectors WHERE id = $1`, connectorID); err != nil {
				return err
			}

			_, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID)

			return err
		})
	})

	link := &coredata.TaskExternalLink{
		OrganizationID: organizationID,
		TaskID:         taskID,
		ConnectorID:    connectorID,
		Metadata:       metadata,
	}
	data := &linear.CommentWebhookData{
		ID:   "comment-1",
		Body: "See ![diagram](" + assetURL + ")",
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return (&Service{}).applyInboundCommentTx(ctx, tx, scope, link, &linear.WebhookEnvelope{}, data, nil)
	}))

	var content string

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return conn.QueryRow(
			ctx,
			`SELECT content FROM task_comments WHERE task_id = $1`,
			taskID,
		).Scan(&content)
	}))

	assert.Contains(t, content, fileID.String())
	assert.NotContains(t, content, assetURL)
}
