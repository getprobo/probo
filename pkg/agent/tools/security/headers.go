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
	"time"

	"go.probo.inc/probo/pkg/agent"
)

type headersParams struct {
	URL string `json:"url" jsonschema:"description=The URL to check security headers for"`
}

type headerCheck struct {
	Present bool   `json:"present"`
	Value   string `json:"value,omitempty"`
}

type headersResult struct {
	HSTS                headerCheck `json:"strict_transport_security"`
	CSP                 headerCheck `json:"content_security_policy"`
	XFrameOptions       headerCheck `json:"x_frame_options"`
	XContentTypeOptions headerCheck `json:"x_content_type_options"`
	ReferrerPolicy      headerCheck `json:"referrer_policy"`
	PermissionsPolicy   headerCheck `json:"permissions_policy"`
	ErrorDetail         string      `json:"error_detail,omitempty"`
}

func checkHeader(h http.Header, name string) headerCheck {
	v := h.Get(name)
	return headerCheck{
		Present: v != "",
		Value:   v,
	}
}

func CheckSecurityHeadersTool() (agent.Tool, error) {
	return agent.FunctionTool[headersParams](
		"check_security_headers",
		"Check security-related HTTP headers for a URL (HSTS, CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, Permissions-Policy).",
		func(ctx context.Context, p headersParams) (agent.ToolResult, error) {
			client := &http.Client{Timeout: 10 * time.Second}

			resp, err := client.Get(p.URL)
			if err != nil {
				data, _ := json.Marshal(headersResult{
					ErrorDetail: fmt.Sprintf("cannot fetch %s: %s", p.URL, err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}
			defer resp.Body.Close()

			result := headersResult{
				HSTS:                checkHeader(resp.Header, "Strict-Transport-Security"),
				CSP:                 checkHeader(resp.Header, "Content-Security-Policy"),
				XFrameOptions:       checkHeader(resp.Header, "X-Frame-Options"),
				XContentTypeOptions: checkHeader(resp.Header, "X-Content-Type-Options"),
				ReferrerPolicy:      checkHeader(resp.Header, "Referrer-Policy"),
				PermissionsPolicy:   checkHeader(resp.Header, "Permissions-Policy"),
			}

			data, _ := json.Marshal(result)

			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
