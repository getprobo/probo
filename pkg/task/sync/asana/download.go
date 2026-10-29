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

package asana

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const maxDownloadRedirects = 3

func (c *Client) Download(ctx context.Context, rawURL string) ([]byte, string, error) {
	if c == nil || c.downloadHTTP == nil {
		return nil, "", fmt.Errorf("cannot download asana file: client is not configured")
	}

	if !c.downloadAllowed(rawURL) {
		return nil, "", fmt.Errorf("cannot download asana file: URL is not an Asana upload")
	}

	current := rawURL

	for range maxDownloadRedirects + 1 {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, current, nil)
		if err != nil {
			return nil, "", fmt.Errorf("cannot build asana download: %w", err)
		}

		resp, err := c.downloadHTTP.Do(req)
		if err != nil {
			return nil, "", fmt.Errorf("cannot download asana file: %w", err)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			body, err := io.ReadAll(io.LimitReader(resp.Body, maxFileSize))
			_ = resp.Body.Close()

			if err != nil {
				return nil, "", fmt.Errorf("cannot read asana file: %w", err)
			}

			return body, resp.Header.Get("Content-Type"), nil
		}

		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			next := resp.Header.Get("Location")
			_ = resp.Body.Close()

			resolved, err := resolveDownloadURL(current, next)
			if err != nil || !c.downloadAllowed(resolved) {
				return nil, "", fmt.Errorf("cannot download asana file: redirect is not an Asana upload")
			}

			current = resolved

			continue
		}

		status := resp.StatusCode
		_ = resp.Body.Close()

		return nil, "", fmt.Errorf("cannot download asana file: status %d", status)
	}

	return nil, "", fmt.Errorf("cannot download asana file: too many redirects")
}

func resolveDownloadURL(current, next string) (string, error) {
	base, err := url.Parse(current)
	if err != nil {
		return "", fmt.Errorf("cannot parse asana download url: %w", err)
	}

	ref, err := url.Parse(next)
	if err != nil {
		return "", fmt.Errorf("cannot parse asana download redirect: %w", err)
	}

	return base.ResolveReference(ref).String(), nil
}

func (c *Client) downloadAllowed(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return false
	}

	if c.allowedDownloadHost != "" {
		return parsed.Host == c.allowedDownloadHost && (parsed.Scheme == "http" || parsed.Scheme == "https")
	}

	if parsed.Scheme != "https" {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "app.asana.com" || host == "asana.com" || strings.HasSuffix(host, ".asana.com") {
		return true
	}

	return asanaUploadBucket(host)
}

func asanaUploadBucket(host string) bool {
	if !strings.HasSuffix(host, ".amazonaws.com") {
		return false
	}

	name := strings.TrimSuffix(host, ".amazonaws.com")

	bucket := name
	if before, _, ok := strings.Cut(name, ".s3"); ok {
		bucket = before
	}

	return strings.HasPrefix(bucket, "asana")
}
