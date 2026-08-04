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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
)

// withRequestChecksumCalculation sets the S3 client's RequestChecksumCalculation
// for a test built with newTestS3Service.
func withRequestChecksumCalculation(v aws.RequestChecksumCalculation) func(*aws.Config) {
	return func(cfg *aws.Config) {
		cfg.RequestChecksumCalculation = v
	}
}

func newTestS3Service(
	t *testing.T,
	handler http.HandlerFunc,
	opts ...func(*aws.Config),
) *filemanager.Service {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg := aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider("access-key", "secret-key", ""),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	s3Client := awss3.NewFromConfig(
		cfg,
		func(o *awss3.Options) {
			o.BaseEndpoint = aws.String(srv.URL)
			o.UsePathStyle = true
		},
	)

	return filemanager.NewService(nil, nil, s3Client, log.NewLogger(log.WithOutput(io.Discard)))
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

// unseekableReader hides io.Seeker so the SDK takes the streaming path, which is
// what a real multipart form upload looks like.
type unseekableReader struct{ r io.Reader }

func (u unseekableReader) Read(p []byte) (int, error) { return u.r.Read(p) }

// requestCarriesChecksum reports whether an upload request asks the server to
// verify an AWS checksum, in any of the forms the SDK can use: a declared
// algorithm, a precomputed header, or a chunked body with a trailing checksum.
func requestCarriesChecksum(h http.Header) bool {
	if h.Get("X-Amz-Sdk-Checksum-Algorithm") != "" || h.Get("X-Amz-Trailer") != "" {
		return true
	}

	if strings.Contains(h.Get("Content-Encoding"), "aws-chunked") {
		return true
	}

	for name := range h {
		if strings.HasPrefix(strings.ToLower(name), "x-amz-checksum-") {
			return true
		}
	}

	return false
}

// The transfer manager keeps its own RequestChecksumCalculation, which takes
// precedence over the S3 client's. Before the accompanying fix it ignored the
// client entirely and always applied CRC32, so S3-compatible endpoints that do
// not implement AWS's checksum scheme (Google Cloud Storage's S3-interop
// endpoint among them) rejected every upload with an opaque
// `SignatureDoesNotMatch`. This asserts the client's setting is honoured in
// both directions, so the regression cannot come back silently.
func TestPutFile_HonoursRequestChecksumCalculation(t *testing.T) {
	t.Parallel()

	const content = "evidence payload"

	for _, tc := range []struct {
		name         string
		checksum     aws.RequestChecksumCalculation
		wantChecksum bool
	}{
		{
			name:         "when required omits checksum",
			checksum:     aws.RequestChecksumCalculationWhenRequired,
			wantChecksum: false,
		},
		{
			name:         "when supported adds checksum",
			checksum:     aws.RequestChecksumCalculationWhenSupported,
			wantChecksum: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				mu        sync.Mutex
				putHeader http.Header
			)

			svc := newTestS3Service(
				t,
				func(w http.ResponseWriter, r *http.Request) {
					if r.Method == http.MethodPut {
						mu.Lock()
						putHeader = r.Header.Clone()
						mu.Unlock()

						_, _ = io.Copy(io.Discard, r.Body)

						w.Header().Set("ETag", `"abc123"`)
						w.WriteHeader(http.StatusOK)

						return
					}

					// HeadObject, which PutFile issues to learn the stored size.
					w.Header().Set("Content-Length", "16")
					w.WriteHeader(http.StatusOK)
				},
				withRequestChecksumCalculation(tc.checksum),
			)

			file := &coredata.File{
				BucketName: "uploads",
				FileKey:    "tenant/evidence",
				MimeType:   "text/plain",
				FileSize:   int64(len(content)),
			}

			_, err := svc.PutFile(
				context.Background(),
				file,
				unseekableReader{r: strings.NewReader(content)},
				map[string]string{},
			)
			require.NoError(t, err)

			mu.Lock()
			defer mu.Unlock()

			require.NotNil(t, putHeader, "expected an upload request to reach the server")
			assert.Equal(t, tc.wantChecksum, requestCarriesChecksum(putHeader))
		})
	}
}
