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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
)

const (
	maxLinearUploadBytes = 50 << 20
	linearUploadTimeout  = 2 * time.Minute
)

// UploadFile stores body in Linear and returns the asset URL to embed in
// markdown. The signed upload URL is not logged and is not requested with the
// OAuth client, because that client would attach a bearer token the signature
// does not include.
func (c *Client) UploadFile(ctx context.Context, filename, contentType string, body []byte) (string, error) {
	if len(body) == 0 {
		return "", fmt.Errorf("cannot upload an empty file to Linear")
	}

	if len(body) > maxLinearUploadBytes {
		return "", fmt.Errorf("cannot upload file to Linear: file exceeds %d bytes", maxLinearUploadBytes)
	}

	filename = linearUploadName(filename)
	contentType = strings.TrimSpace(contentType)

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	const query = `
mutation TaskSyncLinearFileUpload($contentType: String!, $filename: String!, $size: Int!) {
  fileUpload(contentType: $contentType, filename: $filename, size: $size) {
    success
    uploadFile {
      uploadUrl
      assetUrl
      headers {
        key
        value
      }
    }
  }
}
`

	var resp struct {
		Data struct {
			FileUpload struct {
				Success    bool `json:"success"`
				UploadFile *struct {
					UploadURL string `json:"uploadUrl"`
					AssetURL  string `json:"assetUrl"`
					Headers   []struct {
						Key   string `json:"key"`
						Value string `json:"value"`
					} `json:"headers"`
				} `json:"uploadFile"`
			} `json:"fileUpload"`
		} `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	err := c.do(
		ctx,
		query,
		map[string]any{
			"contentType": contentType,
			"filename":    filename,
			"size":        len(body),
		},
		&resp,
	)
	if err != nil {
		return "", err
	}

	upload := resp.Data.FileUpload.UploadFile
	if !resp.Data.FileUpload.Success || upload == nil || upload.UploadURL == "" || upload.AssetURL == "" {
		return "", fmt.Errorf("cannot upload file to Linear: upload URL is missing")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, upload.UploadURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("cannot upload file to Linear: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Cache-Control", "public, max-age=31536000")

	for _, header := range upload.Headers {
		setUploadHeader(req.Header, header.Key, header.Value)
	}

	httpResp, err := c.uploadClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot upload file to Linear: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	_, _ = io.Copy(io.Discard, httpResp.Body)

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot upload file to Linear: unexpected status %d", httpResp.StatusCode)
	}

	return upload.AssetURL, nil
}

// SetUploadHTTPForTest replaces the client that PUTs file bytes.
func (c *Client) SetUploadHTTPForTest(httpClient *http.Client) {
	c.uploadHTTP = httpClient
}

func (c *Client) uploadClient() *http.Client {
	if c.uploadHTTP != nil {
		return c.uploadHTTP
	}

	client := httpclient.DefaultPooledClient(httpclient.WithSSRFProtection())
	client.Timeout = linearUploadTimeout

	return client
}

func linearUploadName(name string) string {
	name = path.Base(strings.TrimSpace(name))
	if name == "." || name == "/" || name == "" {
		return "file"
	}

	return name
}

func setUploadHeader(header http.Header, key, value string) {
	if strings.TrimSpace(key) == "" || strings.ContainsAny(key, "\r\n") || strings.ContainsAny(value, "\r\n") {
		return
	}

	header.Set(key, value)
}
