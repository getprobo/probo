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
				return agent.ResultJSON(
					dnssecResult{
						Enabled:     false,
						ErrorDetail: fmt.Sprintf("cannot query DNSKEY records: %s", err),
					},
				), nil
			}

			var (
				keyCount   int
				keyDetails []string
			)

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
