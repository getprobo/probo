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
