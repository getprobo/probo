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

	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/internal/netcheck"
)

type (
	cspParams struct {
		URL string `json:"url" jsonschema:"The URL to analyze the Content-Security-Policy header for"`
	}

	cspDirective struct {
		Name   string   `json:"name"`
		Values []string `json:"values"`
	}

	cspResult struct {
		Present         bool           `json:"present"`
		ReportOnly      bool           `json:"report_only"`
		RawHeader       string         `json:"raw_header,omitempty"`
		Directives      []cspDirective `json:"directives,omitempty"`
		HasUnsafeEval   bool           `json:"has_unsafe_eval"`
		HasUnsafeInline bool           `json:"has_unsafe_inline"`
		HasWildcard     bool           `json:"has_wildcard"`
		ErrorDetail     string         `json:"error_detail,omitempty"`
	}
)

func parseCSPDirectives(raw string) []cspDirective {
	var directives []cspDirective

	for part := range strings.SplitSeq(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		tokens := strings.Fields(part)
		if len(tokens) == 0 {
			continue
		}

		directives = append(
			directives,
			cspDirective{
				Name:   tokens[0],
				Values: tokens[1:],
			},
		)
	}

	return directives
}

func AnalyzeCSPTool() agent.Tool {
	client := httpclient.DefaultPooledClient(httpclient.WithSSRFProtection())
	client.Timeout = 10 * time.Second

	return agent.FunctionTool(
		"analyze_csp",
		"Analyze the Content-Security-Policy header for a URL, parsing directives and flagging unsafe patterns like unsafe-eval, unsafe-inline, and wildcard sources.",
		func(ctx context.Context, p cspParams) (agent.ToolResult, error) {
			if err := netcheck.ValidatePublicURL(p.URL); err != nil {
				return agent.ResultJSON(
					cspResult{
						ErrorDetail: fmt.Sprintf("URL not allowed: %s", err),
					},
				), nil
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URL, nil)
			if err != nil {
				return agent.ResultJSON(
					cspResult{
						ErrorDetail: fmt.Sprintf("cannot create request for %s: %s", p.URL, err),
					},
				), nil
			}

			resp, err := client.Do(req)
			if err != nil {
				return agent.ResultJSON(
					cspResult{
						ErrorDetail: fmt.Sprintf("cannot fetch %s: %s", p.URL, err),
					},
				), nil
			}

			defer func() { _ = resp.Body.Close() }()

			raw := resp.Header.Get("Content-Security-Policy")
			reportOnly := false

			if raw == "" {
				raw = resp.Header.Get("Content-Security-Policy-Report-Only")
				if raw != "" {
					reportOnly = true
				}
			}

			if raw == "" {
				return agent.ResultJSON(cspResult{Present: false}), nil
			}

			directives := parseCSPDirectives(raw)

			var hasUnsafeEval, hasUnsafeInline, hasWildcard bool

			for _, d := range directives {
				for _, v := range d.Values {
					switch v {
					case "'unsafe-eval'":
						hasUnsafeEval = true
					case "'unsafe-inline'":
						hasUnsafeInline = true
					case "*":
						hasWildcard = true
					}
				}
			}

			result := cspResult{
				Present:         true,
				ReportOnly:      reportOnly,
				RawHeader:       raw,
				Directives:      directives,
				HasUnsafeEval:   hasUnsafeEval,
				HasUnsafeInline: hasUnsafeInline,
				HasWildcard:     hasWildcard,
			}

			return agent.ResultJSON(result), nil
		},
	)
}
