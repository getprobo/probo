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
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_UploadPicture(t *testing.T) {
	t.Parallel()

	var (
		handlerErr error
		uploaded   []byte
		putType    string
		putHeader  string
	)

	uploadServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				handlerErr = err
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			uploaded = body
			putType = r.Header.Get("Content-Type")
			putHeader = r.Header.Get("X-Linear-Upload")
			w.WriteHeader(http.StatusOK)
		}),
	)
	t.Cleanup(uploadServer.Close)

	graphqlServer := newLinearGraphQLServer(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			req, ok := serveDecodedGraphQL(w, r, &handlerErr)
			if !ok {
				return
			}

			switch {
			case strings.Contains(req.Query, "fileUpload"):
				writeJSON(
					w,
					map[string]any{
						"data": map[string]any{
							"fileUpload": map[string]any{
								"success": true,
								"uploadFile": map[string]any{
									"uploadUrl": uploadServer.URL + "/put",
									"assetUrl":  "https://uploads.linear.app/pic.png",
									"headers": []any{
										map[string]any{
											"key":   "X-Linear-Upload",
											"value": "signed",
										},
									},
								},
							},
						},
					},
				)
			case strings.Contains(req.Query, "attachmentCreate"):
				writeJSON(
					w,
					map[string]any{
						"data": map[string]any{
							"attachmentCreate": map[string]any{
								"success": true,
								"attachment": map[string]any{
									"id": "attachment-1",
								},
							},
						},
					},
				)
			default:
				handlerErr = fmt.Errorf("unexpected query")
				http.Error(w, "unexpected query", http.StatusBadRequest)
			}
		},
	)

	picture, err := NewClient(graphqlServer.Client(), graphqlServer.URL).UploadPicture(
		context.Background(),
		"issue-1",
		"pic.png",
		"image/png",
		[]byte("png-bytes"),
	)

	require.NoError(t, handlerErr)
	require.NoError(t, err)
	require.NotNil(t, picture)
	assert.Equal(t, "https://uploads.linear.app/pic.png", picture.AssetURL)
	assert.Equal(t, "attachment-1", picture.AttachmentID)
	assert.Equal(t, []byte("png-bytes"), uploaded)
	assert.Equal(t, "image/png", putType)
	assert.Equal(t, "signed", putHeader)
}

func TestClient_uploadURLAllowed(t *testing.T) {
	t.Parallel()

	secure := NewClient(http.DefaultClient, "https://api.linear.app/graphql")
	require.NoError(t, secure.uploadURLAllowed("https://uploads.linear.app/a"))
	require.Error(t, secure.uploadURLAllowed("http://uploads.linear.app/a"))
	require.Error(t, secure.uploadURLAllowed("not a url"))

	local := NewClient(http.DefaultClient, "http://127.0.0.1:9/graphql")
	require.NoError(t, local.uploadURLAllowed("http://127.0.0.1:1/upload"))
}
