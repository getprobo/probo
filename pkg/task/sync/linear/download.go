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
	"net/url"
	"strings"
)

const linearAssetHost = "uploads.linear.app"

const maxLinearAssetRedirects = 3

// AssetUnavailableError is a Linear upload that no longer exists.
type AssetUnavailableError struct {
	Status int
}

func (e *AssetUnavailableError) Error() string {
	return fmt.Sprintf("cannot download Linear file: status %d", e.Status)
}

// AssetRejectedError is a Linear upload that cannot be stored.
type AssetRejectedError struct {
	Status int
}

func (e *AssetRejectedError) Error() string {
	return fmt.Sprintf("cannot download Linear file: status %d", e.Status)
}

// DownloadAsset reads a Linear upload. The OAuth token is sent only to
// uploads.linear.app. A redirect is followed without that token, because the
// next hop is a signed CDN URL that does not include it.
func (c *Client) DownloadAsset(ctx context.Context, rawURL string) ([]byte, string, error) {
	if c == nil {
		return nil, "", fmt.Errorf("cannot download Linear file: client is not configured")
	}

	if !c.linearAssetURLAllowed(rawURL) {
		return nil, "", fmt.Errorf("cannot download Linear file: URL is not a Linear upload")
	}

	current := rawURL
	auth := true

	// The first request is not a redirect. Three hops may follow it.
	for range maxLinearAssetRedirects + 1 {
		resp, err := c.downloadOnce(ctx, current, auth)
		if err != nil {
			return nil, "", err
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return readAssetBody(resp)
		}

		if isAssetRedirect(resp.StatusCode) {
			next, err := assetRedirectURL(resp, c.allowedAssetHost != "")
			if err != nil {
				discardAsset(resp)

				return nil, "", err
			}

			discardAsset(resp)

			current = next
			auth = false

			continue
		}

		status := resp.StatusCode
		discardAsset(resp)

		switch status {
		case http.StatusNotFound, http.StatusGone:
			return nil, "", &AssetUnavailableError{Status: status}
		case http.StatusBadRequest, http.StatusRequestEntityTooLarge, http.StatusUnsupportedMediaType:
			return nil, "", &AssetRejectedError{Status: status}
		default:
			return nil, "", fmt.Errorf("cannot download Linear file: unexpected status %d", status)
		}
	}

	return nil, "", fmt.Errorf("cannot download Linear file: too many redirects")
}

func (c *Client) linearAssetURLAllowed(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.User != nil {
		return false
	}

	host := c.assetHost()
	if !strings.EqualFold(parsed.Hostname(), host) {
		return false
	}

	if host == linearAssetHost && parsed.Scheme != "https" {
		return false
	}

	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return false
	}

	return parsed.Path != "" && parsed.Path != "/"
}

func (c *Client) assetHost() string {
	if c != nil && c.allowedAssetHost != "" {
		return c.allowedAssetHost
	}

	return linearAssetHost
}

func (c *Client) downloadOnce(ctx context.Context, rawURL string, auth bool) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot download Linear file: %w", withoutRequestURL(err))
	}

	var base *http.Client
	if auth {
		base = c.httpClient
	} else {
		base = c.uploadClient()
	}

	if base == nil {
		return nil, fmt.Errorf("cannot download Linear file: client is not configured")
	}

	client := *base
	client.CheckRedirect = refuseAssetRedirect

	if client.Timeout == 0 || client.Timeout > linearUploadTimeout {
		client.Timeout = linearUploadTimeout
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot download Linear file: %w", withoutRequestURL(err))
	}

	return resp, nil
}

func refuseAssetRedirect(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}

func isAssetRedirect(status int) bool {
	switch status {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
}

func assetRedirectURL(resp *http.Response, allowHTTP bool) (string, error) {
	location, err := resp.Location()
	if err != nil || location == nil {
		return "", fmt.Errorf("cannot download Linear file: redirect is missing a location")
	}

	if location.User != nil {
		return "", fmt.Errorf("cannot download Linear file: redirect URL is not allowed")
	}

	if location.Scheme != "https" && (!allowHTTP || location.Scheme != "http") {
		return "", fmt.Errorf("cannot download Linear file: redirect URL is not allowed")
	}

	return location.String(), nil
}

func readAssetBody(resp *http.Response) ([]byte, string, error) {
	defer discardAsset(resp)

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxLinearUploadBytes+1))
	if err != nil {
		return nil, "", fmt.Errorf("cannot download Linear file: %w", withoutRequestURL(err))
	}

	if len(body) == 0 {
		return nil, "", &AssetRejectedError{Status: resp.StatusCode}
	}

	if len(body) > maxLinearUploadBytes {
		return nil, "", &AssetRejectedError{Status: http.StatusRequestEntityTooLarge}
	}

	return body, resp.Header.Get("Content-Type"), nil
}

func discardAsset(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	_ = resp.Body.Close()
}
