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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/api"
)

func TestDoUpload_FilePartContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		wantType string
	}{
		{
			name:     "pdf",
			filename: "report.pdf",
			wantType: "application/pdf",
		},
		{
			name:     "markdown",
			filename: "notes.md",
			wantType: "text/markdown",
		},
		{
			name:     "csv",
			filename: "rows.csv",
			wantType: "text/csv",
		},
		{
			name:     "docx",
			filename: "policy.docx",
			wantType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		},
		{
			name:     "quoted filename",
			filename: `report "final".pdf`,
			wantType: "application/pdf",
		},
		{
			name:     "unknown extension",
			filename: "blob.bin",
			wantType: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			type uploadPart struct {
				contentType string
				err         error
			}

			parts := make(chan uploadPart, 1)

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseMultipartForm(1 << 20); err != nil {
					parts <- uploadPart{err: err}
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				_, header, err := r.FormFile("0")
				if err != nil {
					parts <- uploadPart{err: err}
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				parts <- uploadPart{contentType: header.Header.Get("Content-Type")}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
			}))
			t.Cleanup(srv.Close)

			client := api.NewClient(srv.URL, "token", "", time.Second)

			_, err := client.DoUpload(
				"mutation { ok }",
				map[string]any{
					"input": map[string]any{
						"file": nil,
					},
				},
				"variables.input.file",
				tt.filename,
				strings.NewReader("file-bytes"),
			)
			require.NoError(t, err)

			part := <-parts
			require.NoError(t, part.err)
			assert.Equal(t, tt.wantType, part.contentType)
		})
	}
}
