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

package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"unicode"

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/version"
)

// FileBody is the stored file returned by GET /api/files/v1/{id}.
// The caller closes Body.
type FileBody struct {
	Body        io.ReadCloser
	FileName    string
	ContentType string
}

// ParseFileID validates the file id argument of prb file download.
func ParseFileID(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if _, err := gid.ParseGID(ref); err != nil {
		return "", fmt.Errorf("cannot download file: %q is not a file id", ref)
	}

	return ref, nil
}

// DownloadFile fetches a private file from the signed-in host.
// A storage redirect on another host does not receive the bearer token.
func (c *Client) DownloadFile(ctx context.Context, fileID string) (*FileBody, error) {
	if c == nil {
		return nil, fmt.Errorf("cannot download file: client is not configured")
	}

	if _, err := gid.ParseGID(fileID); err != nil {
		return nil, fmt.Errorf("cannot download file: %w", err)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	c.refreshIfDue()

	resp, err := c.getFile(ctx, fileID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized && c.refresher != nil {
		_ = resp.Body.Close()

		if refreshErr := c.tryRefreshToken(); refreshErr == nil {
			resp, err = c.getFile(ctx, fileID)
			if err != nil {
				return nil, err
			}
		}
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()

		return nil, downloadStatusError(resp.StatusCode)
	}

	name := fileNameFromDisposition(resp.Header.Get("Content-Disposition"))
	if name == "" {
		name = fileID
	}

	return &FileBody{
		Body:        resp.Body,
		FileName:    name,
		ContentType: resp.Header.Get("Content-Type"),
	}, nil
}

func (c *Client) getFile(ctx context.Context, fileID string) (*http.Response, error) {
	endpoint, err := c.fileDownloadURL(fileID)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot download file: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", version.UserAgent("prb"))

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = c.httpClient.Timeout

	hc := &http.Client{
		Transport:     transport,
		CheckRedirect: redirectWithoutCrossHostAuth,
	}

	resp, err := hc.Do(req)
	if err != nil {
		if urlErr, ok := errors.AsType[*url.Error](err); ok {
			err = urlErr.Err
		}

		return nil, fmt.Errorf("cannot download file: %w", err)
	}

	return resp, nil
}

func (c *Client) fileDownloadURL(fileID string) (*url.URL, error) {
	host := c.host
	if !strings.Contains(host, "://") {
		host = "https://" + host
	}

	base, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("cannot download file: %w", err)
	}

	if base.Scheme != "http" && base.Scheme != "https" {
		return nil, fmt.Errorf("cannot download file: unsupported host scheme")
	}

	joined, err := url.JoinPath(base.Path, "api", "files", "v1", url.PathEscape(fileID))
	if err != nil {
		return nil, fmt.Errorf("cannot download file: %w", err)
	}

	endpoint := *base
	endpoint.Path = joined
	endpoint.RawPath = ""
	endpoint.RawQuery = ""
	endpoint.Fragment = ""

	return &endpoint, nil
}

func redirectWithoutCrossHostAuth(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return fmt.Errorf("stopped after 10 redirects")
	}

	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		return fmt.Errorf("cannot follow redirect to scheme %q", req.URL.Scheme)
	}

	if !strings.EqualFold(via[0].URL.Host, req.URL.Host) || !strings.EqualFold(via[0].URL.Scheme, req.URL.Scheme) {
		req.Header.Del("Authorization")
	}

	return nil
}

func downloadStatusError(status int) error {
	switch status {
	case http.StatusUnauthorized:
		return fmt.Errorf("cannot download file: authentication failed (HTTP 401), try 'prb auth login'")
	case http.StatusForbidden:
		return fmt.Errorf("cannot download file: access denied (HTTP 403)")
	case http.StatusNotFound:
		return fmt.Errorf("cannot download file: file not found")
	default:
		return fmt.Errorf("cannot download file: HTTP %d", status)
	}
}

func fileNameFromDisposition(header string) string {
	if strings.TrimSpace(header) == "" {
		return ""
	}

	_, params, err := mime.ParseMediaType(header)
	if err != nil {
		return ""
	}

	return safeFileName(params["filename"])
}

func safeFileName(name string) string {
	name = strings.ReplaceAll(strings.TrimSpace(name), "\\", "/")

	name = path.Base(name)
	if name == "." || name == ".." || name == "/" {
		return ""
	}

	return strings.Map(
		func(r rune) rune {
			if unicode.IsControl(r) {
				return '_'
			}

			return r
		},
		name,
	)
}
