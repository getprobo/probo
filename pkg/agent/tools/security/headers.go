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

package security

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/internal/netcheck"
)

type (
	headersParams struct {
		URL string `json:"url" jsonschema:"The URL to check security headers for (e.g. https://example.com)"`
	}

	headerCheck struct {
		Present bool   `json:"present"`
		Value   string `json:"value,omitempty"`
	}

	headersResult struct {
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
)

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

func CheckSecurityHeadersTool() agent.Tool {
	client := httpclient.DefaultPooledClient(httpclient.WithSSRFProtection())
	client.Timeout = 10 * time.Second
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	followClient := httpclient.DefaultPooledClient(httpclient.WithSSRFProtection())
	followClient.Timeout = 10 * time.Second

	return agent.FunctionTool(
		"check_security_headers",
		"Check security-related HTTP headers for a URL (HSTS, CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, Permissions-Policy, Cross-Origin-*-Policy). Also checks if HTTP redirects to HTTPS.",
		func(ctx context.Context, p headersParams) (agent.ToolResult, error) {
			if err := netcheck.ValidatePublicURL(p.URL); err != nil {
				return agent.ResultJSON(
					headersResult{
						ErrorDetail: fmt.Sprintf("URL not allowed: %s", err),
					},
				), nil
			}

			// First check the HTTP version to detect HTTP→HTTPS redirect.
			redirectsToHTTPS := false

			parsedURL, err := url.Parse(p.URL)
			if err != nil {
				return agent.ResultJSON(
					headersResult{
						ErrorDetail: fmt.Sprintf("cannot parse URL: %s", err),
					},
				), nil
			}

			httpParsed := *parsedURL
			httpParsed.Scheme = "http"
			httpURL := httpParsed.String()

			httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, httpURL, nil)
			if err == nil {
				httpResp, err := client.Do(httpReq)
				if err == nil {
					_ = httpResp.Body.Close()
					if httpResp.StatusCode >= 300 && httpResp.StatusCode < 400 {
						loc := httpResp.Header.Get("Location")
						if strings.HasPrefix(loc, "https://") {
							redirectsToHTTPS = true
						}
					}
				}
			}

			// Now check the HTTPS version for the actual security headers.
			httpsParsed := *parsedURL
			httpsParsed.Scheme = "https"
			httpsURL := httpsParsed.String()

			httpsReq, err := http.NewRequestWithContext(ctx, http.MethodGet, httpsURL, nil)
			if err != nil {
				return agent.ResultJSON(
					headersResult{
						ErrorDetail: fmt.Sprintf("cannot create request for %s: %s", httpsURL, err),
					},
				), nil
			}

			resp, err := followClient.Do(httpsReq)
			if err != nil {
				return agent.ResultJSON(
					headersResult{
						ErrorDetail: fmt.Sprintf("cannot fetch %s: %s", httpsURL, err),
					},
				), nil
			}

			defer func() { _ = resp.Body.Close() }()

			result := headersFromResponse(resp)
			result.RedirectsToHTTPS = redirectsToHTTPS

			return agent.ResultJSON(result), nil
		},
	)
}
