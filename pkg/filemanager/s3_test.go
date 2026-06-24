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

package filemanager_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
)

func newTestS3Service(t *testing.T, handler http.HandlerFunc) *filemanager.Service {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	s3Client := awss3.NewFromConfig(
		aws.Config{
			Region:      "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider("access-key", "secret-key", ""),
		},
		func(o *awss3.Options) {
			o.BaseEndpoint = aws.String(srv.URL)
			o.UsePathStyle = true
		},
	)

	return filemanager.NewService(nil, nil, s3Client)
}

func TestOpenFile_StreamsBody(t *testing.T) {
	t.Parallel()

	const (
		etag    = `"abc123"`
		content = "hello world"
	)

	svc := newTestS3Service(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("ETag", etag)
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = io.WriteString(w, content)
		},
	)

	file := &coredata.File{
		BucketName: "uploads",
		FileKey:    "tenant/file",
		MimeType:   "text/plain",
		FileSize:   int64(len(content)),
	}

	obj, err := svc.OpenFile(context.Background(), file, filemanager.FileConditions{})
	require.NoError(t, err)
	require.NotNil(t, obj)
	require.False(t, obj.NotModified)

	defer func() { _ = obj.Body.Close() }()

	assert.Equal(t, etag, obj.ETag)
	assert.Equal(t, "text/plain", obj.ContentType)
	assert.Equal(t, int64(len(content)), obj.ContentLength)

	body, err := io.ReadAll(obj.Body)
	require.NoError(t, err)
	assert.Equal(t, content, string(body))
}

func TestOpenFile_RangeRequestReturnsPartialContent(t *testing.T) {
	t.Parallel()

	const (
		etag    = `"abc123"`
		content = "hello world"
	)

	svc := newTestS3Service(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "bytes=0-4", r.Header.Get("Range"))

			w.Header().Set("ETag", etag)
			w.Header().Set("Content-Range", "bytes 0-4/11")
			w.Header().Set("Content-Length", "5")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = io.WriteString(w, content[:5])
		},
	)

	file := &coredata.File{
		BucketName: "uploads",
		FileKey:    "tenant/file",
		MimeType:   "text/plain",
		FileSize:   int64(len(content)),
	}

	obj, err := svc.OpenFile(
		context.Background(),
		file,
		filemanager.FileConditions{Range: "bytes=0-4"},
	)
	require.NoError(t, err)
	require.NotNil(t, obj)

	defer func() { _ = obj.Body.Close() }()

	assert.True(t, obj.PartialContent)
	assert.False(t, obj.NotModified)
	assert.Equal(t, "bytes 0-4/11", obj.ContentRange)
	assert.Equal(t, int64(5), obj.ContentLength)

	body, err := io.ReadAll(obj.Body)
	require.NoError(t, err)
	assert.Equal(t, content[:5], string(body))
}

func TestOpenFile_IfRangeMatchHonorsRange(t *testing.T) {
	t.Parallel()

	const (
		etag    = `"abc123"`
		content = "hello world"
	)

	svc := newTestS3Service(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodHead {
				w.Header().Set("ETag", etag)
				w.WriteHeader(http.StatusOK)

				return
			}

			assert.Equal(t, "bytes=0-4", r.Header.Get("Range"))

			w.Header().Set("ETag", etag)
			w.Header().Set("Content-Range", "bytes 0-4/11")
			w.Header().Set("Content-Length", "5")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = io.WriteString(w, content[:5])
		},
	)

	file := &coredata.File{
		BucketName: "uploads",
		FileKey:    "tenant/file",
		MimeType:   "text/plain",
		FileSize:   int64(len(content)),
	}

	obj, err := svc.OpenFile(
		context.Background(),
		file,
		filemanager.FileConditions{Range: "bytes=0-4", IfRange: etag},
	)
	require.NoError(t, err)
	require.NotNil(t, obj)

	defer func() { _ = obj.Body.Close() }()

	assert.True(t, obj.PartialContent)
	assert.Equal(t, "bytes 0-4/11", obj.ContentRange)

	body, err := io.ReadAll(obj.Body)
	require.NoError(t, err)
	assert.Equal(t, content[:5], string(body))
}

func TestOpenFile_IfRangeMismatchServesFullContent(t *testing.T) {
	t.Parallel()

	const (
		staleETag   = `"old"`
		currentETag = `"new"`
		content     = "hello world"
	)

	svc := newTestS3Service(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodHead {
				w.Header().Set("ETag", currentETag)
				w.WriteHeader(http.StatusOK)

				return
			}

			// The stale If-Range guard must have dropped the Range so S3
			// returns the full object rather than a 206 of the fresh bytes.
			assert.Empty(t, r.Header.Get("Range"))

			w.Header().Set("ETag", currentETag)
			_, _ = io.WriteString(w, content)
		},
	)

	file := &coredata.File{
		BucketName: "uploads",
		FileKey:    "tenant/file",
		MimeType:   "text/plain",
		FileSize:   int64(len(content)),
	}

	obj, err := svc.OpenFile(
		context.Background(),
		file,
		filemanager.FileConditions{Range: "bytes=0-4", IfRange: staleETag},
	)
	require.NoError(t, err)
	require.NotNil(t, obj)

	defer func() { _ = obj.Body.Close() }()

	assert.False(t, obj.PartialContent)
	assert.Empty(t, obj.ContentRange)

	body, err := io.ReadAll(obj.Body)
	require.NoError(t, err)
	assert.Equal(t, content, string(body))
}

func TestOpenFile_RangeNotSatisfiable(t *testing.T) {
	t.Parallel()

	const content = "hello world"

	svc := newTestS3Service(
		t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Range", "bytes */11")
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		},
	)

	file := &coredata.File{
		BucketName: "uploads",
		FileKey:    "tenant/file",
		MimeType:   "text/plain",
		FileSize:   int64(len(content)),
	}

	obj, err := svc.OpenFile(
		context.Background(),
		file,
		filemanager.FileConditions{Range: "bytes=999-1000"},
	)
	require.NoError(t, err)
	require.NotNil(t, obj)
	assert.True(t, obj.RangeNotSatisfiable)
	assert.False(t, obj.PartialContent)
	assert.Nil(t, obj.Body)
	assert.Equal(t, int64(len(content)), obj.ContentLength)
}

func TestOpenFile_NotModifiedByETag(t *testing.T) {
	t.Parallel()

	const etag = `"abc123"`

	svc := newTestS3Service(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("If-None-Match") == etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			w.Header().Set("ETag", etag)
			_, _ = io.WriteString(w, "content")
		},
	)

	file := &coredata.File{
		BucketName: "uploads",
		FileKey:    "tenant/file",
		MimeType:   "text/plain",
	}

	obj, err := svc.OpenFile(context.Background(), file, filemanager.FileConditions{IfNoneMatch: etag})
	require.NoError(t, err)
	require.NotNil(t, obj)
	assert.True(t, obj.NotModified)
	assert.Nil(t, obj.Body)
}

func TestOpenFile_NotModifiedByModifiedSince(t *testing.T) {
	t.Parallel()

	svc := newTestS3Service(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("If-Modified-Since") != "" {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			_, _ = io.WriteString(w, "content")
		},
	)

	file := &coredata.File{
		BucketName: "uploads",
		FileKey:    "tenant/file",
		MimeType:   "text/plain",
	}

	obj, err := svc.OpenFile(
		context.Background(),
		file,
		filemanager.FileConditions{IfModifiedSince: time.Now()},
	)
	require.NoError(t, err)
	require.NotNil(t, obj)
	assert.True(t, obj.NotModified)
	assert.Nil(t, obj.Body)
}
