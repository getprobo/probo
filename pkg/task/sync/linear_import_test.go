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
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

type memoryAttachmentStore struct {
	files []Attachment
}

func (m *memoryAttachmentStore) StoreAttachment(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	file Attachment,
) (StoredAttachment, error) {
	m.files = append(m.files, file)

	return StoredAttachment{
		ID:       gid.New(gid.NewTenantID(), coredata.FileEntityType),
		Name:     file.Name,
		MimeType: file.MimeType,
		Size:     int64(len(file.Body)),
	}, nil
}

func TestImportLinearMarkdown_StoresImageAndFile(t *testing.T) {
	t.Parallel()

	const (
		imageURL = "https://uploads.linear.app/org/file/diagram.png"
		fileURL  = "https://uploads.linear.app/org/file/notes.pdf"
		linkURL  = "https://uploads.linear.app/org/file/187622421.png"
	)

	store := &memoryAttachmentStore{}
	client := linear.NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			var (
				body string
				kind string
			)

			switch req.URL.Path {
			case "/org/file/diagram.png", "/org/file/187622421.png":
				body = "png-bytes"
				kind = "image/png"
			case "/org/file/notes.pdf":
				body = "pdf-bytes"
				kind = "application/pdf"
			default:
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}

			header := make(http.Header)
			header.Set("Content-Type", kind)

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    req,
			}, nil
		}),
	}, "https://api.linear.app")

	markdown := strings.Join([]string{
		"![diagram](" + imageURL + ")",
		"",
		"[notes.pdf](" + fileURL + ")",
		"",
		"[187622421.png](" + linkURL + ")",
		"",
		"See [notes.pdf](" + fileURL + ") too",
	}, "\n")

	svc := &Service{stores: store}
	rewritten, files, metas, err := svc.importLinearMarkdown(
		t.Context(),
		client,
		nil,
		gid.Nil,
		gid.Nil,
		markdown,
		nil,
	)
	require.NoError(t, err)
	assert.NotContains(t, rewritten, "uploads.linear.app")
	assert.Contains(t, rewritten, "![187622421.png](")
	assert.Contains(t, rewritten, "[notes.pdf](")
	assert.NotContains(t, rewritten, "![notes.pdf](")
	assert.Len(t, store.files, 3)
	assert.Len(t, files, 3)

	content, err := MarkdownToContent(rewritten)
	require.NoError(t, err)

	content, err = annotateStoredFiles(content, metas)
	require.NoError(t, err)

	assert.Contains(t, content, `"type":"image"`)
	assert.Contains(t, content, `"type":"file"`)
	assert.Contains(t, content, `"mimeType":"application/pdf"`)
	assert.Contains(t, content, `"size":9`)
	assert.NotContains(t, content, "uploads.linear.app")

	for id := range files {
		if metas[id].MimeType == "image/png" {
			assert.Contains(t, content, id)
		}
	}
}

func TestImportLinearMarkdown_ReusesKnownAsset(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType)
	assetURL := "https://uploads.linear.app/org/file/diagram.png"
	store := &memoryAttachmentStore{}
	svc := &Service{stores: store}

	rewritten, files, metas, err := svc.importLinearMarkdown(
		t.Context(),
		nil,
		nil,
		gid.Nil,
		gid.Nil,
		"![diagram]("+assetURL+")",
		map[string]string{fileID.String(): assetURL},
	)
	require.NoError(t, err)
	assert.Nil(t, metas)
	assert.Empty(t, store.files)
	assert.Equal(t, assetURL, files[fileID.String()])
	assert.Contains(t, rewritten, prosemirror.AttachmentPath(fileID.String()))
	assert.NotContains(t, rewritten, assetURL)
}

func TestImportLinearMarkdown_ReusesFileAfterTransientFailure(t *testing.T) {
	t.Parallel()

	const (
		imageURL = "https://uploads.linear.app/org/file/diagram.png"
		fileURL  = "https://uploads.linear.app/org/file/notes.pdf"
	)

	var (
		downloads   int
		pdfAttempts int
	)

	store := &memoryAttachmentStore{}
	client := linear.NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			downloads++

			header := make(http.Header)

			if req.URL.Path == "/org/file/notes.pdf" {
				pdfAttempts++
				if pdfAttempts == 1 {
					return &http.Response{
						StatusCode: http.StatusBadGateway,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     header,
						Request:    req,
					}, nil
				}

				header.Set("Content-Type", "application/pdf")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(strings.NewReader("pdf-bytes")),
					Request:    req,
				}, nil
			}

			header.Set("Content-Type", "image/png")

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader("png-bytes")),
				Request:    req,
			}, nil
		}),
	}, "https://api.linear.app")

	markdown := "![diagram](" + imageURL + ")\n\n[notes.pdf](" + fileURL + ")"
	svc := &Service{stores: store}
	_, _, _, err := svc.importLinearMarkdown(
		t.Context(),
		client,
		nil,
		gid.Nil,
		gid.Nil,
		markdown,
		nil,
	)
	require.Error(t, err)
	assert.Len(t, store.files, 1)

	rewritten, files, _, err := svc.importLinearMarkdown(
		t.Context(),
		client,
		nil,
		gid.Nil,
		gid.Nil,
		markdown,
		nil,
	)
	require.NoError(t, err)
	assert.Len(t, store.files, 2)
	assert.Len(t, files, 2)
	assert.NotContains(t, rewritten, "uploads.linear.app")
	assert.Equal(t, 3, downloads)
}

type blockingAttachmentStore struct {
	mu      sync.Mutex
	files   []Attachment
	started chan struct{}
	release chan struct{}
}

func (b *blockingAttachmentStore) StoreAttachment(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	file Attachment,
) (StoredAttachment, error) {
	b.mu.Lock()
	b.files = append(b.files, file)
	count := len(b.files)
	b.mu.Unlock()

	if count == 1 {
		close(b.started)
		<-b.release
	}

	return StoredAttachment{
		ID:       gid.New(gid.NewTenantID(), coredata.FileEntityType),
		Name:     file.Name,
		MimeType: file.MimeType,
		Size:     int64(len(file.Body)),
	}, nil
}

func TestImportLinearMarkdown_SingleFlightStoresOnce(t *testing.T) {
	t.Parallel()

	assetURL := "https://uploads.linear.app/org/file/diagram.png"
	store := &blockingAttachmentStore{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	client := linear.NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			header := make(http.Header)
			header.Set("Content-Type", "image/png")

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader("png-bytes")),
				Request:    req,
			}, nil
		}),
	}, "https://api.linear.app")

	svc := &Service{stores: store}
	importFile := func() (string, error) {
		_, files, _, err := svc.importLinearMarkdown(
			t.Context(),
			client,
			nil,
			gid.Nil,
			gid.Nil,
			"![diagram]("+assetURL+")",
			nil,
		)
		if err != nil {
			return "", err
		}

		for id := range files {
			return id, nil
		}

		return "", nil
	}

	var (
		wg        sync.WaitGroup
		firstID   string
		secondID  string
		firstErr  error
		secondErr error
	)

	wg.Go(func() {
		firstID, firstErr = importFile()
	})

	<-store.started

	wg.Go(func() {
		secondID, secondErr = importFile()
	})

	close(store.release)
	wg.Wait()

	require.NoError(t, firstErr)
	require.NoError(t, secondErr)
	assert.Equal(t, firstID, secondID)
	assert.NotEmpty(t, firstID)
	assert.Len(t, store.files, 1)
}

func TestImportLinearMarkdown_BestEffortKeepsTaskText(t *testing.T) {
	t.Parallel()

	assetURL := "https://uploads.linear.app/org/file/notes.pdf"
	client := linear.NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	}, "https://api.linear.app")

	svc := &Service{stores: &memoryAttachmentStore{}}
	rewritten, _, _, err := svc.importLinearMarkdownBestEffort(
		t.Context(),
		client,
		nil,
		gid.Nil,
		gid.Nil,
		"[notes.pdf]("+assetURL+")",
		nil,
	)
	require.NoError(t, err)
	assert.Contains(t, rewritten, assetURL)
}

func TestImportLinearFiles_ReturnsCreatedFilesOnFailure(t *testing.T) {
	t.Parallel()

	const (
		imageURL = "https://uploads.linear.app/org/file/diagram.png"
		failURL  = "https://uploads.linear.app/org/file/broken.pdf"
	)

	store := &memoryAttachmentStore{}
	client := linear.NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/org/file/diagram.png":
				header := make(http.Header)
				header.Set("Content-Type", "image/png")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(strings.NewReader("png-bytes")),
					Request:    req,
				}, nil
			case "/org/file/broken.pdf":
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			default:
				t.Errorf("unexpected Linear asset request %s", req.URL.Path)

				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
		}),
	}, "https://api.linear.app")

	markdown := "![diagram](" + imageURL + ")\n\n[broken.pdf](" + failURL + ")"

	svc := &Service{stores: store}
	_, files, metas, created, err := svc.importLinearFiles(
		t.Context(),
		client,
		nil,
		gid.Nil,
		gid.Nil,
		markdown,
		nil,
		false,
	)

	require.Error(t, err)
	assert.Nil(t, files)
	assert.Nil(t, metas)
	require.Len(t, store.files, 1, "the image must have been stored before the failure")
	require.Len(t, created, 1, "the already-stored file must be reported so the caller can discard it")
	assert.Equal(t, imageURL, created[0].assetURL)
	assert.Equal(t, gid.Nil, created[0].connectorID)
	assert.NotEqual(t, "", created[0].id.String())
}

func TestImportLinearFiles_SharedHoldSurvivesOneRelease(t *testing.T) {
	t.Parallel()

	assetURL := "https://uploads.linear.app/org/file/diagram.png"
	svc := &Service{stores: &memoryAttachmentStore{}}
	client := pngLinearClient(t)

	importFile := func() []createdInboundFile {
		t.Helper()

		_, _, _, created, err := svc.importLinearFiles(
			t.Context(),
			client,
			nil,
			gid.Nil,
			gid.Nil,
			"![diagram]("+assetURL+")",
			nil,
			false,
		)
		require.NoError(t, err)

		return created
	}

	first := importFile()
	second := importFile()

	require.Len(t, first, 1)
	require.Len(t, second, 1)
	assert.Equal(t, first[0].id, second[0].id)
	assert.Len(t, svc.stores.(*memoryAttachmentStore).files, 1)

	assert.False(t, svc.releaseInboundFile(gid.Nil, assetURL, false))
	_, held := svc.inboundFiles.files[inboundFileKey(gid.Nil, assetURL)]
	assert.True(t, held)

	assert.True(t, svc.releaseInboundFile(gid.Nil, assetURL, false))
	_, held = svc.inboundFiles.files[inboundFileKey(gid.Nil, assetURL)]
	assert.False(t, held)
}

func TestImportLinearFiles_RetainedFileIsNotDropped(t *testing.T) {
	t.Parallel()

	assetURL := "https://uploads.linear.app/org/file/diagram.png"
	svc := &Service{stores: &memoryAttachmentStore{}}
	client := pngLinearClient(t)

	_, _, _, created, err := svc.importLinearFiles(
		t.Context(),
		client,
		nil,
		gid.Nil,
		gid.Nil,
		"![diagram]("+assetURL+")",
		nil,
		false,
	)
	require.NoError(t, err)
	require.Len(t, created, 1)

	svc.retainCreatedFiles(created)

	_, _, _, shared, err := svc.importLinearFiles(
		t.Context(),
		client,
		nil,
		gid.Nil,
		gid.Nil,
		"![diagram]("+assetURL+")",
		nil,
		false,
	)
	require.NoError(t, err)
	require.Len(t, shared, 1)

	assert.False(t, svc.releaseInboundFile(gid.Nil, assetURL, false))
	_, held := svc.inboundFiles.files[inboundFileKey(gid.Nil, assetURL)]
	assert.True(t, held)
	assert.Len(t, svc.stores.(*memoryAttachmentStore).files, 1)
}

func TestRememberInboundFile_QueuesDifferentRowForDelete(t *testing.T) {
	t.Parallel()

	assetURL := "https://uploads.linear.app/org/file/diagram.png"
	connectorID := gid.Nil
	key := inboundFileKey(connectorID, assetURL)
	newer := gid.New(gid.NewTenantID(), coredata.FileEntityType)
	older := gid.New(newer.TenantID(), coredata.FileEntityType)
	svc := &Service{}
	svc.inboundFiles.files = map[string]cachedInboundFile{
		key: {fileID: newer.String(), refs: 1},
	}

	olderFile := createdInboundFile{
		id:          older,
		connectorID: connectorID,
		assetURL:    assetURL,
	}
	assert.False(t, svc.rememberInboundFile(olderFile))
	assert.Equal(t, newer.String(), svc.inboundFiles.files[key].fileID)

	scope := coredata.NewScope(older.TenantID(), gid.Nil)
	svc.queueInboundDelete(olderFile, scope)

	pending := svc.takePendingInboundDeletes()
	require.Len(t, pending, 1)
	assert.Equal(t, older, pending[0].file.id)
	assert.Equal(t, older.TenantID(), pending[0].tenantID)
	assert.Empty(t, svc.inboundFiles.pending)
}

func TestInboundFileCleanupContext_OutlivesParentCancel(t *testing.T) {
	t.Parallel()

	parent, cancel := context.WithCancel(context.Background())
	cancel()

	ctx, stop := inboundFileCleanupContext(parent)
	defer stop()

	assert.NoError(t, ctx.Err())
	deadline, ok := ctx.Deadline()
	assert.True(t, ok)
	assert.WithinDuration(t, time.Now().Add(inboundFileCleanupTimeout), deadline, time.Second)
}

func TestImportLinearMarkdown_SkipsMissingAsset(t *testing.T) {
	t.Parallel()

	assetURL := "https://uploads.linear.app/org/file/gone.png"
	client := linear.NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewReader(nil)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	}, "https://api.linear.app")

	svc := &Service{stores: &memoryAttachmentStore{}}
	rewritten, _, _, err := svc.importLinearMarkdown(
		t.Context(),
		client,
		nil,
		gid.Nil,
		gid.Nil,
		"![gone]("+assetURL+")",
		nil,
	)
	require.NoError(t, err)
	assert.Contains(t, rewritten, assetURL)
}

func pngLinearClient(t *testing.T) *linear.Client {
	t.Helper()

	return linear.NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			header := make(http.Header)
			header.Set("Content-Type", "image/png")

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader("png-bytes")),
				Request:    req,
			}, nil
		}),
	}, "https://api.linear.app")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
