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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadFile_PutsBytesAndReturnsAssetURL(t *testing.T) {
	t.Parallel()

	var (
		srv      *httptest.Server
		putBody  []byte
		putType  string
		putCache string
		putAuth  string
		putExtra string
	)

	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			putBody = body
			putType = r.Header.Get("Content-Type")
			putCache = r.Header.Get("Cache-Control")
			putAuth = r.Header.Get("Authorization")
			putExtra = r.Header.Get("X-Upload-Token")

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

	client := NewClient(srv.Client(), srv.URL)
	client.uploadHTTP = srv.Client()

	assetURL, err := client.UploadFile(t.Context(), "diagram.png", "image/png", []byte("png-bytes"))
	require.NoError(t, err)
	assert.Equal(t, "https://uploads.linear.app/file/abc", assetURL)
	assert.Equal(t, []byte("png-bytes"), putBody)
	assert.Equal(t, "image/png", putType)
	assert.Equal(t, "public, max-age=31536000", putCache)
	assert.Empty(t, putAuth)
	assert.Equal(t, "signed", putExtra)
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
