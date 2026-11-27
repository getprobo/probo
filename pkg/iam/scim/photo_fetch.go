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

package scim

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/iam/avatar"
)

const photoFetchTimeout = 15 * time.Second

func (s *Service) readPhoto(ctx context.Context, value string) ([]byte, error) {
	if strings.HasPrefix(value, "data:") {
		return decodePhotoDataURI(value)
	}

	photoURL, err := url.Parse(value)
	if err != nil || photoURL.Host == "" || photoURL.Scheme != "https" {
		return nil, fmt.Errorf("cannot parse photo url: %w", errUnsupportedPhoto)
	}

	return s.fetchPhoto(ctx, photoURL)
}

func decodePhotoDataURI(value string) ([]byte, error) {
	meta, payload, ok := strings.Cut(value, ",")
	if !ok || !strings.Contains(meta, ";base64") {
		return nil, fmt.Errorf("cannot read photo data uri: %w", errUnsupportedPhoto)
	}

	body, err := io.ReadAll(io.LimitReader(
		base64.NewDecoder(base64.StdEncoding, strings.NewReader(payload)),
		avatar.MaxFileSize+1,
	))
	if err != nil || len(body) == 0 || len(body) > avatar.MaxFileSize {
		return nil, fmt.Errorf("cannot decode photo data uri: %w", errUnsupportedPhoto)
	}

	return body, nil
}

func (s *Service) fetchPhoto(ctx context.Context, photoURL *url.URL) ([]byte, error) {
	if s.httpClient == nil {
		return nil, fmt.Errorf("cannot fetch photo without http client: %w", errPhotoFetch)
	}

	ctx, cancel := context.WithTimeout(ctx, photoFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, photoURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create photo request: %w", errPhotoFetch)
	}

	req.Header.Set("Accept", "image/jpeg, image/png, */*")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot send photo request: %w", errPhotoFetch)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot get photo: status %d: %w", resp.StatusCode, errPhotoFetch)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, avatar.MaxFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("cannot read photo body: %w", errPhotoFetch)
	}

	if len(body) == 0 || len(body) > avatar.MaxFileSize {
		return nil, fmt.Errorf("cannot use photo body: %w", errUnsupportedPhoto)
	}

	return body, nil
}
