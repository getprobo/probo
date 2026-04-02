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

type dmarcParams struct {
	Domain string `json:"domain" jsonschema:"The domain to check DMARC record for (e.g. example.com)"`
}

type dmarcResult struct {
	Found       bool   `json:"found"`
	RawRecord   string `json:"raw_record,omitempty"`
	Policy      string `json:"policy,omitempty"`
	Percentage  string `json:"pct,omitempty"`
	RUA         string `json:"rua,omitempty"`
	RUF         string `json:"ruf,omitempty"`
	ErrorDetail string `json:"error_detail,omitempty"`
}

func parseDMARCTag(record, tag string) string {
	for part := range strings.SplitSeq(record, ";") {
		part = strings.TrimSpace(part)
		if after, ok := strings.CutPrefix(part, tag+"="); ok {
			return after
		}
	}
	return ""
}

func CheckDMARCTool() (agent.Tool, error) {
	return agent.FunctionTool[dmarcParams](
		"check_dmarc",
		"Check the DMARC DNS record for a domain, returning the policy, percentage, and reporting addresses.",
		func(ctx context.Context, p dmarcParams) (agent.ToolResult, error) {
			fqdn := "_dmarc." + p.Domain
			if !strings.HasSuffix(fqdn, ".") {
				fqdn = fqdn + "."
			}

			msg := &dns.Msg{
				MsgHeader: dns.MsgHeader{
					ID:               dns.ID(),
					RecursionDesired: true,
				},
			}
			msg.Question = []dns.RR{
				&dns.TXT{
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
				data, _ := json.Marshal(dmarcResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("cannot lookup DMARC record: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			if resp.Rcode != dns.RcodeSuccess {
				data, _ := json.Marshal(dmarcResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("DNS query failed: %s", dns.RcodeToString[resp.Rcode]),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			for _, answer := range resp.Answer {
				txt, ok := answer.(*dns.TXT)
				if !ok {
					continue
				}

				record := strings.Join(txt.Txt, "")
				if !strings.HasPrefix(record, "v=DMARC1") {
					continue
				}

				result := dmarcResult{
					Found:      true,
					RawRecord:  record,
					Policy:     parseDMARCTag(record, "p"),
					Percentage: parseDMARCTag(record, "pct"),
					RUA:        parseDMARCTag(record, "rua"),
					RUF:        parseDMARCTag(record, "ruf"),
				}

				data, _ := json.Marshal(result)

				return agent.ToolResult{Content: string(data)}, nil
			}

			data, _ := json.Marshal(dmarcResult{Found: false})

			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
