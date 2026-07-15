// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/uri"
)

const (
	openIDConfigurationPath         = "/.well-known/openid-configuration"
	maxDiscoveryDocumentBytes int64 = 65536
)

func FetchServerMetadata(
	ctx context.Context,
	client *http.Client,
	issuerBaseURL string,
) (*ServerMetadata, error) {
	discoveryURL, err := discoveryDocumentURL(issuerBaseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot build discovery document URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create discovery request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch discovery document: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery endpoint returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxDiscoveryDocumentBytes))
	if err != nil {
		return nil, fmt.Errorf("cannot read discovery document: %w", err)
	}

	var metadata ServerMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("cannot decode discovery document: %w", err)
	}

	if metadata.AuthorizationEndpoint == "" {
		return nil, fmt.Errorf("discovery document does not advertise an authorization endpoint")
	}

	return &metadata, nil
}

func AuthorizationURLWithQuery(
	authorizationEndpoint uri.URI,
	query url.Values,
) (string, error) {
	u, err := url.Parse(authorizationEndpoint.String())
	if err != nil {
		return "", fmt.Errorf("cannot parse authorization endpoint: %w", err)
	}

	u.RawQuery = query.Encode()

	return u.String(), nil
}

func discoveryDocumentURL(issuerBaseURL string) (string, error) {
	return url.JoinPath(strings.TrimSuffix(issuerBaseURL, "/"), openIDConfigurationPath)
}
