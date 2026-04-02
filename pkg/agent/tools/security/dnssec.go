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
	"strings"

	"codeberg.org/miekg/dns"
	"go.probo.inc/probo/pkg/agent"
)

type dnssecParams struct {
	Domain string `json:"domain" jsonschema:"The domain to check DNSSEC for (e.g. example.com)"`
}

type dnssecResult struct {
	Enabled     bool   `json:"enabled"`
	HasDNSKEY   bool   `json:"has_dnskey"`
	KeyCount    int    `json:"key_count,omitempty"`
	Details     string `json:"details,omitempty"`
	ErrorDetail string `json:"error_detail,omitempty"`
}

func CheckDNSSECTool() (agent.Tool, error) {
	return agent.FunctionTool[dnssecParams](
		"check_dnssec",
		"Check if DNSSEC is enabled for a domain by looking up DNSKEY records.",
		func(ctx context.Context, p dnssecParams) (agent.ToolResult, error) {
			fqdn := p.Domain
			if !strings.HasSuffix(fqdn, ".") {
				fqdn = fqdn + "."
			}

			msg := &dns.Msg{
				MsgHeader: dns.MsgHeader{
					ID:               dns.ID(),
					RecursionDesired: true,
					UDPSize:          4096,
					Security:         true, // DNSSEC OK (DO) bit
				},
			}
			msg.Question = []dns.RR{
				&dns.DNSKEY{
					Hdr: dns.Header{
						Name:  fqdn,
						Class: dns.ClassINET,
					},
				},
			}

			client := dns.NewClient()
			resp, _, err := client.Exchange(ctx, msg, "udp", defaultResolverAddr)
			if err == nil && resp.Truncated {
				resp, _, err = client.Exchange(ctx, msg, "tcp", defaultResolverAddr)
			}
			if err != nil {
				data, _ := json.Marshal(dnssecResult{
					Enabled:     false,
					ErrorDetail: fmt.Sprintf("cannot query DNSKEY records: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			if resp.Rcode != dns.RcodeSuccess {
				data, _ := json.Marshal(dnssecResult{
					Enabled:     false,
					ErrorDetail: fmt.Sprintf("DNS query failed: %s", dns.RcodeToString[resp.Rcode]),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			var keyCount int
			var keyDetails []string
			for _, answer := range resp.Answer {
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

			data, _ := json.Marshal(result)

			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
