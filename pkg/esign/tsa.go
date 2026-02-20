// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package esign

import (
	"bytes"
	"context"
	"crypto"
	"fmt"
	"io"
	"net/http"

	"github.com/digitorus/timestamp"
	"go.gearno.de/kit/httpclient"
)

// TSAClient sends RFC 3161 timestamp requests to a Trusted Timestamp Authority.
type TSAClient struct {
	URL        string
	HTTPClient *http.Client
}

// Timestamp sends an RFC 3161 TimeStampReq via HTTP POST to the TSA.
// The data parameter is the raw bytes to timestamp (typically the seal hex
// string as UTF-8 bytes). CreateRequest internally computes SHA-256(data)
// to build the MessageImprint. Returns the raw DER-encoded TimeStampResp bytes.
func (c *TSAClient) Timestamp(ctx context.Context, data []byte) ([]byte, error) {
	tsReq, err := timestamp.CreateRequest(
		bytes.NewReader(data),
		&timestamp.RequestOptions{
			Hash:         crypto.SHA256,
			Certificates: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("esign: cannot create timestamp request: %w", err)
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = httpclient.DefaultPooledClient()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(tsReq))
	if err != nil {
		return nil, fmt.Errorf("esign: cannot build TSA HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/timestamp-query")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("esign: TSA request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("esign: TSA returned HTTP %d", resp.StatusCode)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("esign: cannot read TSA response: %w", err)
	}

	// Validate the response: checks PKIStatus and parses the signed TSTInfo.
	if _, err := timestamp.ParseResponse(respBytes); err != nil {
		return nil, fmt.Errorf("esign: invalid TSA response: %w", err)
	}

	return respBytes, nil
}
