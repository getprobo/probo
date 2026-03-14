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

type corsParams struct {
	URL    string `json:"url" jsonschema:"description=The URL to check CORS headers for"`
	Origin string `json:"origin" jsonschema:"description=The Origin header value to send in the preflight request (e.g. https://evil.com)"`
}

type corsResult struct {
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

func CheckCORSTool() (agent.Tool, error) {
	return agent.FunctionTool[corsParams](
		"check_cors",
		"Send a CORS preflight (OPTIONS) request to a URL with a given Origin and analyze the Access-Control-* response headers, flagging wildcard origins and origin reflection.",
		func(ctx context.Context, p corsParams) (agent.ToolResult, error) {
			client := &http.Client{Timeout: 10 * time.Second}

			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodOptions,
				p.URL,
				nil,
			)
			if err != nil {
				data, _ := json.Marshal(corsResult{
					ErrorDetail: fmt.Sprintf("cannot build request: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			req.Header.Set("Origin", p.Origin)
			req.Header.Set("Access-Control-Request-Method", "GET")

			resp, err := client.Do(req)
			if err != nil {
				data, _ := json.Marshal(corsResult{
					ErrorDetail: fmt.Sprintf("cannot fetch %s: %s", p.URL, err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}
			defer resp.Body.Close()

			allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")

			result := corsResult{
				AllowOrigin:      allowOrigin,
				AllowMethods:     splitTrimmed(resp.Header.Get("Access-Control-Allow-Methods"), ","),
				AllowHeaders:     splitTrimmed(resp.Header.Get("Access-Control-Allow-Headers"), ","),
				AllowCredentials: strings.EqualFold(resp.Header.Get("Access-Control-Allow-Credentials"), "true"),
				ExposeHeaders:    splitTrimmed(resp.Header.Get("Access-Control-Expose-Headers"), ","),
				MaxAge:           resp.Header.Get("Access-Control-Max-Age"),
				WildcardOrigin:   allowOrigin == "*",
				ReflectsOrigin:   allowOrigin == p.Origin,
			}

			data, _ := json.Marshal(result)

			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
