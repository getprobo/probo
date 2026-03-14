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

type cspParams struct {
	URL string `json:"url" jsonschema:"description=The URL to analyze the Content-Security-Policy header for"`
}

type cspDirective struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type cspResult struct {
	Present        bool           `json:"present"`
	ReportOnly     bool           `json:"report_only"`
	RawHeader      string         `json:"raw_header,omitempty"`
	Directives     []cspDirective `json:"directives,omitempty"`
	HasUnsafeEval  bool           `json:"has_unsafe_eval"`
	HasUnsafeInline bool          `json:"has_unsafe_inline"`
	HasWildcard    bool           `json:"has_wildcard"`
	ErrorDetail    string         `json:"error_detail,omitempty"`
}

func parseCSPDirectives(raw string) []cspDirective {
	var directives []cspDirective

	for _, part := range strings.Split(raw, ";") {
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

func AnalyzeCSPTool() (agent.Tool, error) {
	return agent.FunctionTool[cspParams](
		"analyze_csp",
		"Analyze the Content-Security-Policy header for a URL, parsing directives and flagging unsafe patterns like unsafe-eval, unsafe-inline, and wildcard sources.",
		func(ctx context.Context, p cspParams) (agent.ToolResult, error) {
			client := &http.Client{Timeout: 10 * time.Second}

			resp, err := client.Get(p.URL)
			if err != nil {
				data, _ := json.Marshal(cspResult{
					ErrorDetail: fmt.Sprintf("cannot fetch %s: %s", p.URL, err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}
			defer resp.Body.Close()

			raw := resp.Header.Get("Content-Security-Policy")
			reportOnly := false

			if raw == "" {
				raw = resp.Header.Get("Content-Security-Policy-Report-Only")
				if raw != "" {
					reportOnly = true
				}
			}

			if raw == "" {
				data, _ := json.Marshal(cspResult{Present: false})
				return agent.ToolResult{Content: string(data)}, nil
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

			data, _ := json.Marshal(result)

			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
