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

package api_test

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/api"
)

func TestDoRaw_SendsOperationName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		want    string
		present bool
	}{
		{
			name:    "named query",
			query:   "query AssetList($id: ID!) { node(id: $id) { id } }",
			want:    "AssetList",
			present: true,
		},
		{
			name:    "name on the line before the arguments",
			query:   "query AccessReviewEntryList(\n  $id: ID!\n) { node(id: $id) { id } }",
			want:    "AccessReviewEntryList",
			present: true,
		},
		{
			name:    "named mutation",
			query:   "mutation AssetCreate($input: CreateAssetInput!) { createAsset(input: $input) { id } }",
			want:    "AssetCreate",
			present: true,
		},
		{
			name:  "anonymous query",
			query: "query { ok }",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				var got map[string]any

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)

						return
					}

					if err := json.Unmarshal(body, &got); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)

						return
					}

					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
				}))
				t.Cleanup(server.Close)

				client := api.NewClient(server.URL, "token", "/graphql", time.Second)
				_, err := client.DoRaw(tt.query, nil)
				require.NoError(t, err)
				assert.Equal(t, tt.query, got["query"])

				operationName, ok := got["operationName"]
				assert.Equal(t, tt.present, ok)
				if tt.present {
					assert.Equal(t, tt.want, operationName)
				}
			},
		)
	}
}

func TestDoUpload_SendsOperationName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		want    string
		present bool
	}{
		{
			name:    "named mutation",
			query:   "mutation UploadFile($input: UploadFileInput!) { uploadFile(input: $input) { id } }",
			want:    "UploadFile",
			present: true,
		},
		{
			name:  "anonymous mutation",
			query: "mutation { ok }",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				var got map[string]any

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)

						return
					}

					reader := multipart.NewReader(r.Body, params["boundary"])
					part, err := reader.NextPart()
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)

						return
					}

					body, err := io.ReadAll(part)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)

						return
					}

					if err := json.Unmarshal(body, &got); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)

						return
					}

					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
				}))
				t.Cleanup(server.Close)

				client := api.NewClient(server.URL, "token", "/graphql", time.Second)
				_, err := client.DoUpload(
					tt.query,
					map[string]any{"input": map[string]any{"file": nil}},
					"variables.input.file",
					"report.pdf",
					strings.NewReader("file-body"),
				)
				require.NoError(t, err)

				operationName, ok := got["operationName"]
				assert.Equal(t, tt.present, ok)
				if tt.present {
					assert.Equal(t, tt.want, operationName)
				}
			},
		)
	}
}
