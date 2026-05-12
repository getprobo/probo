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
	"fmt"
	"strings"

	"codeberg.org/miekg/dns"
	"go.probo.inc/probo/pkg/agent"
)

type (
	dnssecParams struct {
		Domain string `json:"domain" jsonschema:"The domain to check DNSSEC for (e.g. example.com)"`
	}

	dnssecResult struct {
		Enabled     bool   `json:"enabled"`
		HasDNSKEY   bool   `json:"has_dnskey"`
		KeyCount    int    `json:"key_count,omitempty"`
		Details     string `json:"details,omitempty"`
		ErrorDetail string `json:"error_detail,omitempty"`
	}
)

func CheckDNSSECTool() agent.Tool {
	return agent.FunctionTool(
		"check_dnssec",
		"Check if DNSSEC is enabled for a domain by looking up DNSKEY records.",
		func(ctx context.Context, p dnssecParams) (agent.ToolResult, error) {
			fqdn := p.Domain
			if !strings.HasSuffix(fqdn, ".") {
				fqdn = fqdn + "."
			}

			client := dns.NewClient()
			answers, err := queryDNS(
				ctx,
				client,
				&dns.DNSKEY{
					Hdr: dns.Header{
						Name:  fqdn,
						Class: dns.ClassINET,
					},
				},
				withDNSSEC(),
			)
			if err != nil {
				return agent.ResultJSON(dnssecResult{
					Enabled:     false,
					ErrorDetail: fmt.Sprintf("cannot query DNSKEY records: %s", err),
				}), nil
			}

			var keyCount int
			var keyDetails []string
			for _, answer := range answers {
				if key, ok := answer.(*dns.DNSKEY); ok {
					keyCount++
					flags := "ZSK"
					// SEP (Secure Entry Point) flag is bit 15 (value 1)
					if key.Flags&0x0001 != 0 {
						flags = "KSK"
					}
					keyDetails = append(
						keyDetails,
						fmt.Sprintf("%s (algorithm=%d, flags=%d)", flags, key.Algorithm, key.Flags),
					)
				}
			}

			hasDNSKEY := keyCount > 0
			result := dnssecResult{
				Enabled:   hasDNSKEY,
				HasDNSKEY: hasDNSKEY,
				KeyCount:  keyCount,
				Details:   strings.Join(keyDetails, "; "),
			}

			if !hasDNSKEY {
				result.Details = "no DNSKEY records found"
			}

			return agent.ResultJSON(result), nil
		},
	)
}
