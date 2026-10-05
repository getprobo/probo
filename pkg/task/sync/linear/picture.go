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
	"net/url"
	"strings"
	"time"
)

const pictureUploadTimeout = 60 * time.Second

type (
	UploadedPicture struct {
		AssetURL     string
		AttachmentID string
	}

	uploadHeader struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
)

// pictureUploadClient puts bytes on Linear's signed upload URL. It is separate
// from the GraphQL client so the OAuth transport does not attach a bearer
// token that would break the signed request.
var pictureUploadClient = &http.Client{Timeout: pictureUploadTimeout}

// IsEntityMissing reports whether a Linear GraphQL error means the entity is
// already gone.
func IsEntityMissing(err error) bool {
	return linearEntityMissing(err)
}

// UploadPicture stores an image in Linear and attaches it to the issue.
// The attachment survives later description updates from task sync.
func (c *Client) UploadPicture(
	ctx context.Context,
	issueID string,
	filename string,
	contentType string,
	body []byte,
) (*UploadedPicture, error) {
	if issueID == "" {
		return nil, fmt.Errorf("cannot upload picture to Linear: issue id is required")
	}

	if filename == "" {
		return nil, fmt.Errorf("cannot upload picture to Linear: filename is required")
	}

	if contentType == "" {
		return nil, fmt.Errorf("cannot upload picture to Linear: content type is required")
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("cannot upload picture to Linear: file is empty")
	}

	uploadURL, assetURL, headers, err := c.requestPictureUpload(ctx, filename, contentType, len(body))
	if err != nil {
		return nil, err
	}

	if err := c.uploadURLAllowed(uploadURL); err != nil {
		return nil, err
	}

	if err := putLinearUpload(ctx, uploadURL, contentType, headers, body); err != nil {
		return nil, err
	}

	attachmentID, err := c.LinkAttachment(ctx, issueID, assetURL, filename)
	if err != nil {
		return nil, fmt.Errorf("cannot attach picture to Linear issue: %w", err)
	}

	return &UploadedPicture{
		AssetURL:     assetURL,
		AttachmentID: attachmentID,
	}, nil
}

func (c *Client) requestPictureUpload(
	ctx context.Context,
	filename string,
	contentType string,
	size int,
) (string, string, []uploadHeader, error) {
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
					UploadURL string         `json:"uploadUrl"`
					AssetURL  string         `json:"assetUrl"`
					Headers   []uploadHeader `json:"headers"`
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
			"size":        size,
		},
		&resp,
	)
	if err != nil {
		return "", "", nil, err
	}

	uploaded := resp.Data.FileUpload.UploadFile
	if !resp.Data.FileUpload.Success || uploaded == nil || uploaded.UploadURL == "" || uploaded.AssetURL == "" {
		return "", "", nil, fmt.Errorf("cannot upload picture to Linear: file upload unsuccessful")
	}

	return uploaded.UploadURL, uploaded.AssetURL, uploaded.Headers, nil
}

func (c *Client) uploadURLAllowed(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("cannot upload picture to Linear: invalid upload URL")
	}

	switch parsed.Scheme {
	case "https":
		return nil
	case "http":
		if strings.HasPrefix(c.endpoint, "http://") {
			return nil
		}
	}

	return fmt.Errorf("cannot upload picture to Linear: unsupported upload URL")
}

func putLinearUpload(
	ctx context.Context,
	uploadURL string,
	contentType string,
	headers []uploadHeader,
	body []byte,
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot create Linear upload request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Cache-Control", "public, max-age=31536000")

	for _, header := range headers {
		if header.Key == "" {
			continue
		}

		req.Header.Set(header.Key, header.Value)
	}

	resp, err := pictureUploadClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot upload picture to Linear: %w", err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("cannot upload picture to Linear: unexpected status %d", resp.StatusCode)
	}

	return nil
}
