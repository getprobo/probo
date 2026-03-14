// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package security

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/agent"
)

type headersParams struct {
	URL string `json:"url" jsonschema:"description=The URL to check security headers for (e.g. https://example.com)"`
}

type headerCheck struct {
	Present bool   `json:"present"`
	Value   string `json:"value,omitempty"`
}

type headersResult struct {
	HSTS                      headerCheck `json:"strict_transport_security"`
	CSP                       headerCheck `json:"content_security_policy"`
	XFrameOptions             headerCheck `json:"x_frame_options"`
	XContentTypeOptions       headerCheck `json:"x_content_type_options"`
	ReferrerPolicy            headerCheck `json:"referrer_policy"`
	PermissionsPolicy         headerCheck `json:"permissions_policy"`
	CrossOriginOpenerPolicy   headerCheck `json:"cross_origin_opener_policy"`
	CrossOriginEmbedderPolicy headerCheck `json:"cross_origin_embedder_policy"`
	CrossOriginResourcePolicy headerCheck `json:"cross_origin_resource_policy"`
	RedirectsToHTTPS          bool        `json:"redirects_to_https"`
	ErrorDetail               string      `json:"error_detail,omitempty"`
}

func checkHeader(h http.Header, name string) headerCheck {
	v := h.Get(name)
	return headerCheck{
		Present: v != "",
		Value:   v,
	}
}

func headersFromResponse(resp *http.Response) headersResult {
	return headersResult{
		HSTS:                      checkHeader(resp.Header, "Strict-Transport-Security"),
		CSP:                       checkHeader(resp.Header, "Content-Security-Policy"),
		XFrameOptions:             checkHeader(resp.Header, "X-Frame-Options"),
		XContentTypeOptions:       checkHeader(resp.Header, "X-Content-Type-Options"),
		ReferrerPolicy:            checkHeader(resp.Header, "Referrer-Policy"),
		PermissionsPolicy:         checkHeader(resp.Header, "Permissions-Policy"),
		CrossOriginOpenerPolicy:   checkHeader(resp.Header, "Cross-Origin-Opener-Policy"),
		CrossOriginEmbedderPolicy: checkHeader(resp.Header, "Cross-Origin-Embedder-Policy"),
		CrossOriginResourcePolicy: checkHeader(resp.Header, "Cross-Origin-Resource-Policy"),
	}
}

func CheckSecurityHeadersTool() (agent.Tool, error) {
	return agent.FunctionTool[headersParams](
		"check_security_headers",
		"Check security-related HTTP headers for a URL (HSTS, CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, Permissions-Policy, Cross-Origin-*-Policy). Also checks if HTTP redirects to HTTPS.",
		func(ctx context.Context, p headersParams) (agent.ToolResult, error) {
			client := &http.Client{
				Timeout: 10 * time.Second,
				CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			// First check the HTTP version to detect HTTP→HTTPS redirect.
			redirectsToHTTPS := false
			httpURL := p.URL
			if strings.HasPrefix(httpURL, "https://") {
				httpURL = "http://" + strings.TrimPrefix(httpURL, "https://")
			}

			httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, httpURL, nil)
			if err == nil {
				httpResp, err := client.Do(httpReq)
				if err == nil {
					httpResp.Body.Close()
					if httpResp.StatusCode >= 300 && httpResp.StatusCode < 400 {
						loc := httpResp.Header.Get("Location")
						if strings.HasPrefix(loc, "https://") {
							redirectsToHTTPS = true
						}
					}
				}
			}

			// Now check the HTTPS version for the actual security headers.
			httpsURL := p.URL
			if strings.HasPrefix(httpsURL, "http://") {
				httpsURL = "https://" + strings.TrimPrefix(httpsURL, "http://")
			}

			followClient := &http.Client{Timeout: 10 * time.Second}
			httpsReq, err := http.NewRequestWithContext(ctx, http.MethodGet, httpsURL, nil)
			if err != nil {
				data, _ := json.Marshal(headersResult{
					ErrorDetail: fmt.Sprintf("cannot create request for %s: %s", httpsURL, err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}
			resp, err := followClient.Do(httpsReq)
			if err != nil {
				data, _ := json.Marshal(headersResult{
					ErrorDetail: fmt.Sprintf("cannot fetch %s: %s", httpsURL, err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}
			defer resp.Body.Close()

			result := headersFromResponse(resp)
			result.RedirectsToHTTPS = redirectsToHTTPS

			data, _ := json.Marshal(result)

			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
