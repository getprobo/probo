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
	"strings"
	"time"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/internal/netcheck"
)

type (
	corsParams struct {
		URL    string `json:"url" jsonschema:"The URL to check CORS headers for"`
		Origin string `json:"origin" jsonschema:"The Origin header value to send in the preflight request (e.g. https://evil.com)"`
	}

	corsResult struct {
		AllowOrigin      string   `json:"access_control_allow_origin,omitempty"`
		AllowMethods     []string `json:"access_control_allow_methods,omitempty"`
		AllowHeaders     []string `json:"access_control_allow_headers,omitempty"`
		AllowCredentials bool     `json:"access_control_allow_credentials"`
		ExposeHeaders    []string `json:"access_control_expose_headers,omitempty"`
		MaxAge           string   `json:"access_control_max_age,omitempty"`
		WildcardOrigin   bool     `json:"wildcard_origin"`
		ReflectsOrigin   bool     `json:"reflects_origin"`
		ErrorDetail      string   `json:"error_detail,omitempty"`
	}
)

func splitTrimmed(s, sep string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}

	return out
}

func CheckCORSTool() agent.Tool {
	return agent.FunctionTool(
		"check_cors",
		"Send a CORS preflight (OPTIONS) request to a URL with a given Origin and analyze the Access-Control-* response headers, flagging wildcard origins and origin reflection.",
		func(ctx context.Context, p corsParams) (agent.ToolResult, error) {
			if err := netcheck.ValidatePublicURL(p.URL); err != nil {
				return agent.ResultJSON(
					corsResult{
						ErrorDetail: fmt.Sprintf("URL not allowed: %s", err),
					},
				), nil
			}

			client := &http.Client{
				Timeout: 10 * time.Second,
				CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodOptions,
				p.URL,
				nil,
			)
			if err != nil {
				return agent.ResultJSON(
					corsResult{
						ErrorDetail: fmt.Sprintf("cannot build request: %s", err),
					},
				), nil
			}

			req.Header.Set("Origin", p.Origin)
			req.Header.Set("Access-Control-Request-Method", "GET")

			resp, err := client.Do(req)
			if err != nil {
				return agent.ResultJSON(
					corsResult{
						ErrorDetail: fmt.Sprintf("cannot fetch %s: %s", p.URL, err),
					},
				), nil
			}

			defer func() { _ = resp.Body.Close() }()

			allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")

			result := corsResult{
				AllowOrigin:      allowOrigin,
				AllowMethods:     splitTrimmed(resp.Header.Get("Access-Control-Allow-Methods"), ","),
				AllowHeaders:     splitTrimmed(resp.Header.Get("Access-Control-Allow-Headers"), ","),
				AllowCredentials: strings.EqualFold(resp.Header.Get("Access-Control-Allow-Credentials"), "true"),
				ExposeHeaders:    splitTrimmed(resp.Header.Get("Access-Control-Expose-Headers"), ","),
				MaxAge:           resp.Header.Get("Access-Control-Max-Age"),
				WildcardOrigin:   allowOrigin == "*",
				ReflectsOrigin:   p.Origin != "" && allowOrigin == p.Origin,
			}

			return agent.ResultJSON(result), nil
		},
	)
}
