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

func TestDoUpload_SetsContentTypeFromFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		filename        string
		wantContentType string
	}{
		{
			name:            "pdf",
			filename:        "report.pdf",
			wantContentType: "application/pdf",
		},
		{
			name:            "markdown",
			filename:        "notes.md",
			wantContentType: "text/markdown",
		},
		{
			name:            "avi",
			filename:        "clip.avi",
			wantContentType: "video/x-msvideo",
		},
		{
			name:            "unknown extension",
			filename:        "payload.exe",
			wantContentType: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				contentType string
				mapPaths    []string
				formErr     error
			)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if err != nil {
					formErr = err
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				if mediaType != "multipart/form-data" {
					http.Error(w, "expected multipart", http.StatusBadRequest)

					return
				}

				reader := multipart.NewReader(r.Body, params["boundary"])

				for {
					part, err := reader.NextPart()
					if err == io.EOF {
						break
					}

					if err != nil {
						formErr = err
						http.Error(w, err.Error(), http.StatusBadRequest)

						return
					}

					body, err := io.ReadAll(part)
					if err != nil {
						formErr = err
						http.Error(w, err.Error(), http.StatusBadRequest)

						return
					}

					switch part.FormName() {
					case "map":
						var fileMap map[string][]string
						if err := json.Unmarshal(body, &fileMap); err != nil {
							formErr = err
							http.Error(w, err.Error(), http.StatusBadRequest)

							return
						}

						mapPaths = fileMap["0"]
					case "0":
						contentType = part.Header.Get("Content-Type")
					}
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
			}))
			t.Cleanup(server.Close)

			client := api.NewClient(server.URL, "token", "/graphql", 5*time.Second)
			_, err := client.DoUpload(
				`mutation { ok }`,
				map[string]any{"input": map[string]any{"file": nil}},
				"variables.input.file",
				tt.filename,
				strings.NewReader("file-body"),
			)
			require.NoError(t, err)
			require.NoError(t, formErr)
			assert.Equal(t, tt.wantContentType, contentType)
			assert.Equal(t, []string{"variables.input.file"}, mapPaths)
		})
	}
}
