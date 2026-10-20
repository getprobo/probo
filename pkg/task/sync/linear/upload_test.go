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

package linear

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinearUploadName_StripsDirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "unix path", in: "docs/policy.pdf", want: "policy.pdf"},
		{name: "windows path", in: `C:\Users\alice\Documents\policy.pdf`, want: "policy.pdf"},
		{name: "unc path", in: `\\server\share\report.pdf`, want: "report.pdf"},
		{name: "bare name", in: "policy.pdf", want: "policy.pdf"},
		{name: "blank", in: "  ", want: "file"},
		{name: "backslash only", in: `\`, want: "file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, linearUploadName(tt.in))
		})
	}
}

func TestUploadFile_PutsBytesAndReturnsAssetURL(t *testing.T) {
	t.Parallel()

	var (
		srv         *httptest.Server
		mu          sync.Mutex
		putBody     []byte
		putType     string
		putCache    string
		putAuth     string
		putExtra    string
		graphqlAuth string
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
			putType = r.Header.Get("Content-Type")
			putCache = r.Header.Get("Cache-Control")
			putAuth = r.Header.Get("Authorization")
			putExtra = r.Header.Get("X-Upload-Token")
			mu.Unlock()

			w.WriteHeader(http.StatusOK)

			return
		}

		mu.Lock()
		graphqlAuth = r.Header.Get("Authorization")
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"fileUpload": map[string]any{
					"success": true,
					"uploadFile": map[string]any{
						"uploadUrl": srv.URL + "/upload",
						"assetUrl":  "https://uploads.linear.app/file/abc",
						"headers": []map[string]string{
							{"key": "X-Upload-Token", "value": "signed"},
						},
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	shared := srv.Client()

	transport := shared.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	oauthClient := &http.Client{
		CheckRedirect: shared.CheckRedirect,
		Timeout:       shared.Timeout,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req = req.Clone(req.Context())
			req.Header.Set("Authorization", "Bearer oauth-token")

			return transport.RoundTrip(req)
		}),
	}
	uploadClient := &http.Client{
		CheckRedirect: shared.CheckRedirect,
		Timeout:       shared.Timeout,
		Transport:     transport,
	}

	client := NewClient(oauthClient, srv.URL)
	client.uploadHTTP = uploadClient

	assetURL, err := client.UploadFile(t.Context(), "diagram.png", "image/png", []byte("png-bytes"))
	require.NoError(t, err)

	mu.Lock()
	body, contentType, cache, auth, extra, queryAuth := putBody, putType, putCache, putAuth, putExtra, graphqlAuth
	mu.Unlock()

	assert.Equal(t, "https://uploads.linear.app/file/abc", assetURL)
	assert.Equal(t, []byte("png-bytes"), body)
	assert.Equal(t, "image/png", contentType)
	assert.Equal(t, "public, max-age=31536000", cache)
	assert.Equal(t, "Bearer oauth-token", queryAuth)
	assert.Empty(t, auth)
	assert.Equal(t, "signed", extra)
}

func TestUploadFile_RejectsEmptyBody(t *testing.T) {
	t.Parallel()

	var (
		mu    sync.Mutex
		calls int
	)

	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client(), srv.URL)
	_, err := client.UploadFile(t.Context(), "diagram.png", "image/png", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty file")

	mu.Lock()
	defer mu.Unlock()

	assert.Zero(t, calls)
}

func TestUploadFile_RejectsOversizedBody(t *testing.T) {
	t.Parallel()

	var (
		mu    sync.Mutex
		calls int
	)

	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client(), srv.URL)
	_, err := client.UploadFile(t.Context(), "diagram.png", "image/png", make([]byte, (50<<20)+1))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds 52428800 bytes")

	mu.Lock()
	defer mu.Unlock()

	assert.Zero(t, calls)
}

func TestUploadFile_RejectsIncompleteUpload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		success    bool
		uploadFile any
	}{
		{
			name:    "success false",
			success: false,
			uploadFile: map[string]any{
				"uploadUrl": "https://uploads.linear.app/upload",
				"assetUrl":  "https://uploads.linear.app/file/abc",
				"headers":   []map[string]string{},
			},
		},
		{
			name:       "missing upload file",
			success:    true,
			uploadFile: nil,
		},
		{
			name:    "missing asset url",
			success: true,
			uploadFile: map[string]any{
				"uploadUrl": "https://uploads.linear.app/upload",
				"assetUrl":  "",
				"headers":   []map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				mu   sync.Mutex
				puts int
			)

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPut {
					mu.Lock()
					puts++
					mu.Unlock()
					w.WriteHeader(http.StatusOK)

					return
				}

				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"fileUpload": map[string]any{
							"success":    tt.success,
							"uploadFile": tt.uploadFile,
						},
					},
				})
			}))
			t.Cleanup(srv.Close)

			client := NewClient(srv.Client(), srv.URL)
			client.uploadHTTP = srv.Client()

			_, err := client.UploadFile(t.Context(), "diagram.png", "image/png", []byte("png-bytes"))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "upload URL or asset URL is missing")

			mu.Lock()
			defer mu.Unlock()

			assert.Zero(t, puts)
		})
	}
}

func TestUploadFile_RejectsFailedPut(t *testing.T) {
	t.Parallel()

	var srv *httptest.Server

	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"fileUpload": map[string]any{
					"success": true,
					"uploadFile": map[string]any{
						"uploadUrl": srv.URL + "/upload",
						"assetUrl":  "https://uploads.linear.app/file/abc",
						"headers":   []map[string]string{},
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client(), srv.URL)
	client.uploadHTTP = srv.Client()

	_, err := client.UploadFile(t.Context(), "notes.pdf", "application/pdf", []byte("pdf"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 403")
}

func TestUploadFile_PutErrorOmitsSignedURL(t *testing.T) {
	t.Parallel()

	const signedURL = "https://uploads.linear.app/upload?signature=secret-token"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"fileUpload": map[string]any{
					"success": true,
					"uploadFile": map[string]any{
						"uploadUrl": signedURL,
						"assetUrl":  "https://uploads.linear.app/file/abc",
						"headers":   []map[string]string{},
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	refused := errors.New("connection refused")
	client := NewClient(srv.Client(), srv.URL)
	client.uploadHTTP = &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, refused
		}),
	}

	_, err := client.UploadFile(t.Context(), "diagram.png", "image/png", []byte("png-bytes"))
	require.Error(t, err)
	assert.ErrorIs(t, err, refused)
	assert.NotContains(t, err.Error(), "secret-token")
	assert.NotContains(t, err.Error(), "uploads.linear.app")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
